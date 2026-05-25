// Package app parses flags, wires the loader/aligner/inspector/printer, and
// orchestrates a structalign run.
package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime/debug"
	"sort"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/loader"
	"github.com/peczenyj/structalign/internal/match"
	"github.com/peczenyj/structalign/internal/ui"
	"github.com/peczenyj/structalign/pkg/common"
)

// version is stamped at release time via -ldflags "-X ...app.version=...".
var version = "dev"

// resolveVersion returns the version to print for -version. A GoReleaser build
// stamps `version` via -ldflags; for a `go install <module>@vX.Y.Z` build that
// stamp is absent (still "dev"), so fall back to the module version embedded in
// the build info.
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}

// App holds the injectable dependencies and output streams.
type App struct {
	Loader    common.Loader
	Aligner   common.Aligner
	Inspector common.Inspector
	Stdout    io.Writer
	Stderr    io.Writer
}

// New returns an App wired with the production implementations.
func New(stdout, stderr io.Writer) *App {
	return &App{
		Aligner:   align.New(),
		Inspector: layout.New(),
		Stdout:    stdout,
		Stderr:    stderr,
	}
}

type options struct {
	diff            common.DiffStyle
	width           int
	color           string
	typeFilter      string
	inspect         bool
	verbose         bool
	tags            bool
	showVersion     bool
	exclude         string
	tests           bool
	generated       bool
	skipCachePadded bool
}

// Run parses args (excluding argv[0]) and executes. Returns the process exit
// code: 1 when diff-mode findings exist, 2 on usage/load error, else 0.
//
//nolint:gocyclo // orchestration naturally accumulates branches; splitting further would obscure the flow
func (a *App) Run(args []string) int {
	var opt options
	fs := flag.NewFlagSet("structalign", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	opt.diff = common.DiffUnified // zero value is DiffUnified; set for clarity
	fs.Var(&opt.diff, "diff", "diff style: unified | side | none")
	fs.IntVar(&opt.width, "width", 0, "column width per side for -diff=side (0 = auto from terminal)")
	fs.StringVar(&opt.color, "color", "auto", "colorize: auto | always | never")
	fs.StringVar(&opt.typeFilter, "type", "", "only consider named structs matching this comma-separated list of glob patterns; empty means all")
	fs.BoolVar(&opt.inspect, "inspect", false, "inspect layout instead of diffing")
	fs.BoolVar(&opt.verbose, "verbose", false, "in -inspect mode, show padding on its own _ line")
	fs.BoolVar(&opt.tags, "tags", false, "preserve struct field tags in output (default: strip them)")
	fs.BoolVar(&opt.showVersion, "version", false, "print version and exit")
	fs.StringVar(&opt.exclude, "exclude", "^unsafe$|^builtin$", "exclude packages whose import path matches this regexp")
	fs.BoolVar(&opt.tests, "tests", false, "also analyze _test.go files")
	fs.BoolVar(&opt.generated, "generated", false, "also analyze generated files (// Code generated ... DO NOT EDIT.)")
	fs.BoolVar(&opt.skipCachePadded, "skip-cache-padded", false, "skip structs containing a golang.org/x/sys/cpu.CacheLinePad field")
	fs.Usage = func() {
		fmt.Fprintf(a.Stderr, "structalign: print field-aligned struct reorderings (no file changes)\n\n")
		fmt.Fprintf(a.Stderr, "usage: structalign [flags] [packages]\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if opt.showVersion {
		fmt.Fprintln(a.Stdout, resolveVersion())
		return 0
	}
	if fs.NArg() == 0 {
		fs.Usage()
		return 2
	}

	excludeRE, err := regexp.Compile(opt.exclude)
	if err != nil {
		fmt.Fprintf(a.Stderr, "structalign: invalid -exclude regexp: %v\n", err)
		return 2
	}

	width := opt.width
	if width <= 0 {
		width = ui.ResolveWidth(stdoutFile(a.Stdout))
	}
	patterns := match.ParsePatterns(opt.typeFilter)
	printer := &ui.Printer{
		Out:   a.Stdout,
		Color: ui.WantColor(opt.color, stdoutFile(a.Stdout)),
		Width: width,
	}

	ld := a.Loader
	if ld == nil {
		ld = loader.New(opt.tests)
	}

	targets, err := ld.Load(fs.Args()...)
	if err != nil {
		fmt.Fprintf(a.Stderr, "structalign: %v\n", err)
		return 2
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].PkgPath < targets[j].PkgPath })

	total := 0
	for _, t := range targets {
		if excludeRE.MatchString(t.PkgPath) {
			continue
		}
		for _, e := range t.Errors {
			fmt.Fprintf(a.Stderr, "structalign: %s: %v\n", t.PkgPath, e)
		}
		if t.Types == nil || t.Sizes == nil {
			continue
		}
		o := common.Options{
			Patterns:         patterns,
			KeepTags:         opt.tags,
			IncludeGenerated: opt.generated,
			SkipCachePadded:  opt.skipCachePadded,
		}
		if opt.inspect {
			total += printer.RenderLayouts(a.Inspector.Layouts(t, o), opt.verbose, opt.tags)
		} else {
			findings, ferr := a.Aligner.Findings(t, o)
			if ferr != nil {
				fmt.Fprintf(a.Stderr, "structalign: %s: %v\n", t.PkgPath, ferr)
				continue
			}
			total += printer.RenderFindings(findings, opt.diff)
		}
	}

	if total == 0 {
		if opt.inspect {
			fmt.Fprintln(a.Stderr, "no matching structs found")
		} else {
			fmt.Fprintln(a.Stderr, "no struct reorderings found")
		}
	}
	if total > 0 && !opt.inspect {
		return 1
	}
	return 0
}

// stdoutFile returns the *os.File behind w for terminal queries, or os.Stdout as
// a harmless fallback when w is not a file (e.g. a test buffer); WantColor and
// ResolveWidth both degrade gracefully in that case.
func stdoutFile(w io.Writer) *os.File {
	if f, ok := w.(*os.File); ok {
		return f
	}
	return os.Stdout
}
