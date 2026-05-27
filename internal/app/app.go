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
	"strings"

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
	colorize        common.Colorize
	typeFilter      string
	inspect         bool
	verbose         bool
	tags            bool
	showVersion     bool
	exclude         string
	tests           bool
	generated       bool
	skipCachePadded bool
	summary         bool
	sort            bool
	threshold       int
	showNolint      bool
	nolintLinters   string
}

// savings is the absolute bytes a finding saves, or 0 when sizes are unknown or
// the proposed layout is not smaller. Shared by -summary (and, later, -sort and
// -threshold).
func savings(f common.Finding) int64 {
	if f.OldSize > 0 && f.NewSize > 0 && f.NewSize < f.OldSize {
		return f.OldSize - f.NewSize
	}
	return 0
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
	fs.Var(&opt.diff, "diff", fmt.Sprintf("diff style: %s (default %q)", opt.diff.Type(), opt.diff.String()))
	fs.IntVar(&opt.width, "width", 0, "column width per side for -diff=side (0 = auto from terminal)")
	opt.colorize = common.ColorizeAuto // zero value is ColorizeAuto; set for clarity
	fs.Var(&opt.colorize, "color", fmt.Sprintf("colorize: %s (default %q)", opt.colorize.Type(), opt.colorize.String()))
	fs.StringVar(&opt.typeFilter, "type", "", "only consider named structs matching this comma-separated list of glob patterns; empty means all")
	fs.BoolVar(&opt.inspect, "inspect", false, "inspect layout instead of diffing")
	fs.BoolVar(&opt.verbose, "verbose", false, "in -inspect mode, show padding on its own _ line")
	fs.BoolVar(&opt.tags, "tags", false, "preserve struct field tags in output (default: strip them)")
	fs.BoolVar(&opt.showVersion, "version", false, "print version and exit")
	fs.StringVar(&opt.exclude, "exclude", "^unsafe$|^builtin$", "exclude packages whose import path matches this regexp")
	fs.BoolVar(&opt.tests, "tests", false, "also analyze _test.go files")
	fs.BoolVar(&opt.generated, "generated", false, "also analyze generated files (// Code generated ... DO NOT EDIT.)")
	fs.BoolVar(&opt.skipCachePadded, "skip-cache-padded", false, "skip structs containing a golang.org/x/sys/cpu.CacheLinePad field")
	fs.BoolVar(&opt.summary, "summary", false, "in diff mode, print a one-line summary after the diffs")
	fs.BoolVar(&opt.sort, "sort", false, "present results largest-first (diff: by bytes saved; inspect: by struct size)")
	fs.IntVar(&opt.threshold, "threshold", 0, "in diff mode, only show structs that save at least this many bytes")
	fs.BoolVar(&opt.showNolint, "show-nolint", false, "show structs even when their type carries a recognized //nolint directive")
	fs.StringVar(&opt.nolintLinters, "nolint-linters", "fieldalignment", "comma-separated //nolint tokens that suppress a finding (bare //nolint always counts)")
	fs.Usage = func() {
		fmt.Fprintf(a.Stderr, "structalign: print field-aligned struct reorderings (no file changes)\n\n")
		fmt.Fprintf(a.Stderr, "usage: structalign [flags] [packages]\n\n")
		fs.PrintDefaults()
	}
	// Easter egg: fieldalignment has -fix; structalign deliberately only prints
	// suggestions and never edits files. Caught before parsing, so -fix is never
	// a registered flag and stays invisible in -help.
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "-fix" || arg == "--fix" {
			fmt.Fprintln(a.Stderr, "structalign: sorry, I don't do -fix — I only print the reordering, never touch your files.")
			fmt.Fprintln(a.Stderr, "For an in-place rewrite, use fieldalignment -fix.")
			return 2
		}
	}

	// Easter-egg theme flags: -cga/-green/-amber select a retro palette. Like
	// -fix, they are caught before parsing and stripped from args, so they stay
	// invisible in -help and never trip "flag provided but not defined".
	themeName := ""
	filtered := args[:0:0]
	afterDD := false
	for _, arg := range args {
		if afterDD || arg == "--" {
			afterDD = true
			filtered = append(filtered, arg)
			continue
		}

		// Strip egg flags: -cga/-green/-amber and their -flag=value forms.
		egg := ""
		for _, name := range []string{"cga", "green", "amber"} {
			if arg == "-"+name || arg == "--"+name || strings.HasPrefix(arg, "-"+name+"=") || strings.HasPrefix(arg, "--"+name+"=") {
				egg = name
				break
			}
		}

		if egg != "" {
			if !strings.Contains(arg, "=") {
				themeName = egg
			} else if _, val, _ := strings.Cut(arg, "="); val == "true" || val == "1" {
				themeName = egg
			}
			continue
		}
		filtered = append(filtered, arg)
	}
	args = filtered

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

	// Resolve the theme: egg flag wins over STRUCTALIGN_THEME, else default.
	if themeName == "" {
		themeName = os.Getenv("STRUCTALIGN_THEME")
	}
	theme := ui.DefaultTheme()
	if themeName != "" && themeName != "default" {
		if th, ok := ui.ThemeByName(themeName); ok {
			theme = th
		} else {
			fmt.Fprintf(a.Stderr, "structalign: unknown theme %q, using default\n", themeName)
		}
	}

	printer := &ui.Printer{
		Out:   a.Stdout,
		Color: ui.WantColor(opt.colorize, stdoutFile(a.Stdout)),
		Width: width,
		Theme: theme,
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

	var allFindings []common.Finding
	var allLayouts []common.Layout
	for _, t := range targets {
		if opt.exclude != "" && excludeRE.MatchString(t.PkgPath) {
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
			RespectNolint:    !opt.showNolint,
			NolintLinters:    match.SplitCSV(opt.nolintLinters),
		}
		if opt.inspect {
			allLayouts = append(allLayouts, a.Inspector.Layouts(t, o)...)
		} else {
			findings, ferr := a.Aligner.Findings(t, o)
			if ferr != nil {
				fmt.Fprintf(a.Stderr, "structalign: %s: %v\n", t.PkgPath, ferr)
				continue
			}
			allFindings = append(allFindings, findings...)
		}
	}

	if !opt.inspect && opt.threshold > 0 {
		min := int64(opt.threshold)
		kept := allFindings[:0]
		for _, f := range allFindings {
			if savings(f) >= min {
				kept = append(kept, f)
			}
		}
		allFindings = kept
	}

	if opt.sort {
		if opt.inspect {
			sort.SliceStable(allLayouts, func(i, j int) bool {
				return allLayouts[i].Total > allLayouts[j].Total
			})
		} else {
			sort.SliceStable(allFindings, func(i, j int) bool {
				return savings(allFindings[i]) > savings(allFindings[j])
			})
		}
	}

	var total int
	if opt.inspect {
		total = printer.RenderLayouts(allLayouts, opt.verbose, opt.tags)
	} else {
		total = printer.RenderFindings(allFindings, opt.diff)
	}

	if opt.summary && !opt.inspect {
		var bytesSaved int64
		for _, f := range allFindings {
			bytesSaved += savings(f)
		}
		printer.RenderSummary(total, bytesSaved)
	} else if total == 0 {
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

// stdoutFile returns the *os.File behind w for terminal queries, or nil when
// w is not a file (e.g. a test buffer); WantColor and ResolveWidth both
// handle nil by falling back to safe defaults (no color, 80-column width).
func stdoutFile(w io.Writer) *os.File {
	if f, ok := w.(*os.File); ok {
		return f
	}
	return nil
}
