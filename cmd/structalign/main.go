// Command structalign reuses the upstream golang.org/x/tools fieldalignment
// analyzer to show how a struct's fields could be reordered to use less
// memory -- but instead of rewriting your files (like `fieldalignment -fix`)
// it prints the suggested reordered struct and a diff.
//
// It does NOT reimplement the alignment algorithm: it runs the unmodified
// fieldalignment.Analyzer and intercepts the SuggestedFix that the analyzer
// already produces (a single TextEdit replacing the whole struct node with
// the optimally-ordered, gofmt'd struct). We diff the original source slice
// against that NewText.
//
// Usage:
//
//	structalign [flags] [packages]
//
// packages are Go package patterns (./..., import paths, directories, or single
// .go files) -- whatever the go tool accepts -- resolved via go/packages.
//
// Flags:
//
//	-diff=unified     diff style: unified | side | none  (default unified)
//	-color            colorize output (default: auto = on if stdout is a TTY)
//	-width=N          column width per side for -diff=side (default: auto)
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	diff "github.com/aymanbagabas/go-udiff"
	"golang.org/x/term"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

// version is the tool's version. It is overridden at release time via
// -ldflags "-X main.version=...", and defaults to "dev" for source builds.
var version = "dev"

// config holds all command-line options. Flags bind directly to these fields
// via flag.<Type>Var, so the rest of main works with values, not pointers.
type config struct {
	diff        string // -diff:    unified | side | none
	width       int    // -width:   per-side column width for side mode (0 = auto)
	color       string // -color:   auto | always | never
	typeFilter  string // -type:    comma-separated glob patterns, empty = all
	inspect     bool   // -inspect: print layout instead of diffing
	verbose     bool   // -verbose: in inspect mode, padding on its own `_` line
	tags        bool   // -tags:    preserve struct field tags in output
	showVersion bool   // -version: print version and exit
}

func main() {
	var cfg config
	flag.StringVar(&cfg.diff, "diff", "unified", "diff style: unified | side | none")
	flag.IntVar(&cfg.width, "width", 0, "column width per side for -diff=side (0 = auto from terminal)")
	flag.StringVar(&cfg.color, "color", "auto", "colorize: auto | always | never")
	flag.StringVar(&cfg.typeFilter, "type", "", "only consider named structs matching this comma-separated list of glob patterns (e.g. \"*Request,Config\"); empty means all")
	flag.BoolVar(&cfg.inspect, "inspect", false, "inspect layout: print each field's offset, size, alignment, and padding (no reordering)")
	flag.BoolVar(&cfg.verbose, "verbose", false, "in -inspect mode, show padding on its own _ line instead of folded into the field comment")
	flag.BoolVar(&cfg.tags, "tags", false, "preserve struct field tags in output (default: strip them)")
	flag.BoolVar(&cfg.showVersion, "version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "structalign: print field-aligned struct reorderings (no file changes)\n\n")
		fmt.Fprintf(os.Stderr, "usage: structalign [flags] [packages]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if cfg.showVersion {
		fmt.Println(version)
		return
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	// -width defaults to 0 ("auto") so -h shows no misleading fixed number;
	// resolve it from the terminal here, once the flags are parsed.
	if cfg.width <= 0 {
		cfg.width = resolveWidth()
	}
	typeGlobs := parsePatterns(cfg.typeFilter)
	color := wantColor(cfg.color)

	// go/packages resolves every argument the go tool accepts: ./..., import
	// paths, directories, and (via normalizeArgs) single .go files.
	pkgs, err := loadPackages(normalizeArgs(flag.Args()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "structalign: %v\n", err)
		os.Exit(2)
	}
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].PkgPath < pkgs[j].PkgPath })

	// Load/type errors are reported but not fatal: a partially-resolved package
	// can still yield useful findings.
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			fmt.Fprintf(os.Stderr, "structalign: %s: %v\n", pkg.PkgPath, e)
		}
	}

	var totalFindings int
	for _, pkg := range pkgs {
		if pkg.Types == nil || pkg.TypesSizes == nil {
			continue
		}
		if cfg.inspect {
			totalFindings += inspectStructs(os.Stdout, pkg.Types, pkg.TypesSizes, typeGlobs, color, cfg.verbose, cfg.tags)
		} else {
			totalFindings += diffPackage(os.Stdout, pkg, cfg.diff, cfg.width, color, typeGlobs, cfg.tags)
		}
	}

	if totalFindings == 0 {
		if cfg.inspect {
			fmt.Fprintln(os.Stderr, "no matching structs found")
		} else {
			fmt.Fprintln(os.Stderr, "no struct reorderings found")
		}
	}
	// Exit non-zero only for the diff modes, where a finding means "could be
	// improved" (CI-friendly). Inspect mode is purely informational, so it
	// always exits 0 when it ran successfully.
	if totalFindings > 0 && !cfg.inspect {
		os.Exit(1)
	}
}

// finding is one struct that can be shrunk.
type finding struct {
	fset     *token.FileSet
	pos, end token.Pos
	name     string // enclosing named type, or "" for anonymous structs
	original string // current source of the struct ("struct{...}")
	proposed string // reordered struct from the analyzer's SuggestedFix
	message  string // analyzer diagnostic message (carries size info)
}

// loadPackages resolves the given patterns (./..., import paths, directories,
// or "file=" queries) into typed packages, with enough information for the
// fieldalignment analyzer and the layout inspector: syntax, types, type info,
// and the target's type sizes.
func loadPackages(patterns []string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
			packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo,
		Fset:  token.NewFileSet(),
		Tests: false,
	}
	return packages.Load(cfg, patterns...)
}

// normalizeArgs lets a bare path to a .go file be used as an argument by
// rewriting it to the "file=" query go/packages understands. Everything else
// (./..., import paths, directories) is passed through unchanged.
func normalizeArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if strings.HasSuffix(a, ".go") {
			if fi, err := os.Stat(a); err == nil && !fi.IsDir() {
				a = "file=" + a
			}
		}
		out = append(out, a)
	}
	return out
}

// diffPackage runs the fieldalignment analyzer over one loaded package and
// renders each suggested reordering. If patterns is non-empty, only structs
// whose enclosing named type matches one of the glob patterns are considered.
// It returns the number of findings rendered.
func diffPackage(w io.Writer, pkg *packages.Package, diffStyle string, width int, color bool, patterns []string, keepTags bool) int {
	if len(pkg.Syntax) == 0 {
		return 0
	}

	// Map the position of each `type Name struct {...}` body to its name, so
	// we can label and filter findings. The analyzer reports at the struct-type
	// node's position, which is exactly the StructType.Pos() we record here.
	names := structNameIndex(pkg.Syntax)

	// Satisfy the analyzer's dependency on the inspect pass.
	insp := inspector.New(pkg.Syntax)

	var findings []finding
	pass := &analysis.Pass{
		Analyzer:   fieldalignment.Analyzer,
		Fset:       pkg.Fset,
		Files:      pkg.Syntax,
		Pkg:        pkg.Types,
		TypesInfo:  pkg.TypesInfo,
		TypesSizes: pkg.TypesSizes,
		ResultOf: map[*analysis.Analyzer]any{
			inspect.Analyzer: insp,
		},
		Report: func(d analysis.Diagnostic) {
			f := finding{fset: pkg.Fset, pos: d.Pos, message: d.Message}
			// The analyzer attaches exactly one SuggestedFix with one
			// TextEdit spanning the struct node. Capture it.
			if len(d.SuggestedFixes) > 0 && len(d.SuggestedFixes[0].TextEdits) > 0 {
				e := d.SuggestedFixes[0].TextEdits[0]
				f.pos, f.end = e.Pos, e.End
				f.proposed = string(e.NewText)
				f.original = readSource(pkg.Fset, e.Pos, e.End)
				if !keepTags {
					// Strip tags from both sides so the diff focuses on field
					// order, not tag re-spacing. Best-effort: on a parse error
					// we keep the original text rather than dropping the finding.
					if s, err := stripStructTags(f.original); err == nil {
						f.original = s
					}
					if s, err := stripStructTags(f.proposed); err == nil {
						f.proposed = s
					}
				}
			}
			f.name = names[f.pos]

			// Apply the -type filter. Anonymous structs (no name) are kept
			// only when no filter is set.
			if len(patterns) > 0 && !matchAny(patterns, f.name) {
				return
			}
			findings = append(findings, f)
		},
	}

	if _, err := fieldalignment.Analyzer.Run(pass); err != nil {
		fmt.Fprintf(os.Stderr, "structalign: %s: analyzer: %v\n", pkg.PkgPath, err)
		return 0
	}

	// Stable order: by file, then position within the file.
	sort.Slice(findings, func(i, j int) bool {
		pi, pj := pkg.Fset.Position(findings[i].pos), pkg.Fset.Position(findings[j].pos)
		if pi.Filename != pj.Filename {
			return pi.Filename < pj.Filename
		}
		return pi.Offset < pj.Offset
	})

	for _, f := range findings {
		render(w, f, diffStyle, width, color)
	}
	return len(findings)
}

// parsePatterns splits a comma-separated -type value into trimmed, non-empty
// glob patterns. An empty input yields nil (meaning "match everything").
func parsePatterns(s string) []string {
	var out []string
	for p := range strings.SplitSeq(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// matchAny reports whether name matches any of the glob patterns (path.Match
// syntax: * ? [..]). A name that is empty (anonymous struct) never matches a
// non-empty pattern set. Invalid patterns are treated as non-matching rather
// than aborting the whole run.
func matchAny(patterns []string, name string) bool {
	if name == "" {
		return false
	}
	for _, p := range patterns {
		if ok, err := path.Match(p, name); err == nil && ok {
			return true
		}
	}
	return false
}

// structNameIndex maps the position of each named struct type's StructType node
// to its declared name, for both `type Foo struct{...}` and grouped
// `type ( Foo struct{...}; Bar struct{...} )` declarations. The fieldalignment
// analyzer reports diagnostics at StructType.Pos(), so this lets us label and
// filter findings by type name. Anonymous structs and struct literals are not
// indexed (they have no name).
func structNameIndex(files []*ast.File) map[token.Pos]string {
	index := make(map[token.Pos]string)
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if st, ok := ts.Type.(*ast.StructType); ok {
				index[st.Pos()] = ts.Name.Name
			}
			return true
		})
	}
	return index
}

// readSource returns the raw source text between two positions.
func readSource(fset *token.FileSet, pos, end token.Pos) string {
	pf := fset.File(pos)
	if pf == nil {
		return ""
	}
	data, err := os.ReadFile(pf.Name())
	if err != nil {
		return ""
	}
	start := pf.Offset(pos)
	stop := pf.Offset(end)
	if start < 0 || stop > len(data) || start > stop {
		return ""
	}
	return string(data[start:stop])
}

// stripStructTags removes field tags from a struct type's source text. It
// parses the text (wrapped as a type declaration), clears each field's Tag, and
// reprints it with go/format, which also re-aligns the now-tagless fields. On
// any parse/format error it returns the error so the caller can fall back to
// the original text.
func stripStructTags(src string) (string, error) {
	wrapped := "package p\ntype _ " + src
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wrapped, parser.ParseComments)
	if err != nil {
		return "", err
	}
	var st *ast.StructType
	ast.Inspect(file, func(n ast.Node) bool {
		if s, ok := n.(*ast.StructType); ok && st == nil {
			st = s
			return false
		}
		return true
	})
	if st == nil {
		return "", fmt.Errorf("no struct type found")
	}
	if st.Fields != nil {
		for _, fld := range st.Fields.List {
			fld.Tag = nil
		}
	}
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, st); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// --- inspect mode -----------------------------------------------------------

// layoutField is one field's place in a struct's memory layout.
type layoutField struct {
	name    string
	typ     string
	tag     string // raw struct tag (without surrounding backticks), or ""
	offset  int64
	size    int64
	align   int64
	padding int64 // padding bytes inserted *after* this field (before the next,
	// or trailing padding to the struct's alignment for the last field)
}

// inspectStructs finds every named struct type in pkg (filtered by patterns),
// computes its layout via the type sizes, and prints it. Returns the count of
// structs printed.
func inspectStructs(w io.Writer, pkg *types.Package, sizes types.Sizes, patterns []string, color bool, verbose bool, keepTags bool) int {
	scope := pkg.Scope()
	var names []string
	for _, n := range scope.Names() {
		names = append(names, n)
	}
	sort.Strings(names)

	printed := 0
	for _, n := range names {
		if len(patterns) > 0 && !matchAny(patterns, n) {
			continue
		}
		obj := scope.Lookup(n)
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		// Generic types have no concrete layout until instantiated: a field
		// whose type is a type parameter has no defined size/alignment, and
		// asking go/types for it panics. Skip them with a note.
		if named, ok := tn.Type().(*types.Named); ok && named.TypeParams().Len() > 0 {
			fmt.Fprintf(os.Stderr, "structalign: skipping generic type %s (no concrete layout until instantiated)\n", n)
			continue
		}
		st, ok := tn.Type().Underlying().(*types.Struct)
		if !ok {
			continue
		}
		renderLayout(w, n, st, sizes, color, verbose, keepTags)
		printed++
	}
	return printed
}

// computeLayout builds the per-field layout for a struct, including the padding
// after each field. Padding after field i is the gap between the end of field i
// and the start of field i+1; for the last field it is the gap up to the
// struct's total (aligned) size.
func computeLayout(st *types.Struct, sizes types.Sizes) (fields []layoutField, total, align int64) {
	nf := st.NumFields()
	vars := make([]*types.Var, nf)
	for i := range nf {
		vars[i] = st.Field(i)
	}
	offsets := sizes.Offsetsof(vars)
	total = sizes.Sizeof(st)
	align = sizes.Alignof(st)

	fields = make([]layoutField, nf)
	for i := range nf {
		fsize := sizes.Sizeof(vars[i].Type())
		lf := layoutField{
			name:   vars[i].Name(),
			typ:    types.TypeString(vars[i].Type(), qualifierForLayout(vars[i])),
			tag:    st.Tag(i),
			offset: offsets[i],
			size:   fsize,
			align:  sizes.Alignof(vars[i].Type()),
		}
		end := offsets[i] + fsize
		var next int64
		if i+1 < nf {
			next = offsets[i+1]
		} else {
			next = total
		}
		if next > end {
			lf.padding = next - end
		}
		fields[i] = lf
	}
	return fields, total, align
}

// qualifierForLayout keeps type names short: types from the struct's own
// package are unqualified, others use the package name.
func qualifierForLayout(v *types.Var) types.Qualifier {
	home := v.Pkg()
	return func(p *types.Package) string {
		if p == home {
			return ""
		}
		return p.Name()
	}
}

// renderLayout prints one struct's layout as annotated Go source: the struct
// declaration with per-field `// size: N, align: M` comments, column-aligned so
// the comments line up. Padding is shown either folded onto the preceding
// field's comment (default) or broken onto its own `_` line (verbose).
func renderLayout(w io.Writer, name string, st *types.Struct, sizes types.Sizes, color bool, verbose bool, keepTags bool) {
	fields, total, align := computeLayout(st, sizes)

	var totalPad int64
	for _, f := range fields {
		totalPad += f.padding
	}

	// Field declaration is "<name> <type>" (plus tag when -tags is set); align
	// all comments to the widest declaration.
	decls := make([]string, len(fields))
	declWidth := 0
	for i, f := range fields {
		decls[i] = f.name + " " + f.typ
		if keepTags && f.tag != "" {
			decls[i] += " `" + f.tag + "`"
		}
		if len(decls[i]) > declWidth {
			declWidth = len(decls[i])
		}
	}
	// In verbose mode a lone "_" can also appear; it's never wider than a field.

	// Header: the struct opening line carries size/align/padding.
	header := fmt.Sprintf("type %s struct { // size: %d, align: %d, padding: %d",
		name, total, align, totalPad)
	fmt.Fprintln(w, paint(color, cBold+cCyan, header))

	for i, f := range fields {
		base := fmt.Sprintf("size: %2d, align: %d", f.size, f.align)
		if verbose {
			// Field line carries no padding; padding gets its own `_` line.
			fmt.Fprintf(w, "\t%-*s // %s\n", declWidth, decls[i], base)
			if f.padding > 0 {
				pad := fmt.Sprintf("\t%-*s // %d byte padding", declWidth, "_", f.padding)
				fmt.Fprintln(w, paint(color, cRed, pad))
			}
		} else {
			// Padding folds onto the field's own comment.
			comment := base
			if f.padding > 0 {
				comment = paint(color, cRed, fmt.Sprintf("%s, padding: %d", base, f.padding))
			}
			fmt.Fprintf(w, "\t%-*s // %s\n", declWidth, decls[i], comment)
		}
	}
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)
}

// --- rendering --------------------------------------------------------------

const (
	cReset = "\x1b[0m"
	cRed   = "\x1b[31m"
	cGreen = "\x1b[32m"
	cCyan  = "\x1b[36m"
	cDim   = "\x1b[2m"
	cBold  = "\x1b[1m"
)

func paint(on bool, code, s string) string {
	if !on {
		return s
	}
	return code + s + cReset
}

// relPath shortens an absolute filename (go/packages reports absolute paths) to
// one relative to the current directory, when that doesn't escape upward.
// Otherwise it returns the name unchanged.
func relPath(name string) string {
	wd, err := os.Getwd()
	if err != nil {
		return name
	}
	rel, err := filepath.Rel(wd, name)
	if err != nil || strings.HasPrefix(rel, "..") {
		return name
	}
	return rel
}

func render(w io.Writer, f finding, style string, width int, color bool) {
	loc := f.fset.Position(f.pos)
	file := relPath(loc.Filename)
	var header string
	if f.name != "" {
		header = fmt.Sprintf("%s:%d:%d: %s: %s", file, loc.Line, loc.Column, f.name, f.message)
	} else {
		header = fmt.Sprintf("%s:%d:%d: %s", file, loc.Line, loc.Column, f.message)
	}
	fmt.Fprintln(w, paint(color, cBold+cCyan, header))

	if f.original == "" || f.proposed == "" {
		fmt.Fprintln(w, "  (no suggested fix produced)")
		fmt.Fprintln(w)
		return
	}

	switch style {
	case "none":
		// just the proposed struct
		fmt.Fprintln(w, indent(f.proposed, "  "))
	case "side":
		renderSideBySide(w, f.original, f.proposed, width, color)
	default: // unified
		renderUnified(w, f.original, f.proposed, color)
	}
	fmt.Fprintln(w)
}

func indent(s, pad string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = pad + l
	}
	return strings.Join(lines, "\n")
}

// renderUnified prints a minimal line-based unified diff. We use a simple
// longest-common-subsequence over lines, which is fine for the small,
// reordered struct bodies we deal with here.
func renderUnified(w io.Writer, a, b string, color bool) {
	al := strings.Split(a, "\n")
	bl := strings.Split(b, "\n")
	ops := lcsDiff(al, bl)
	for _, op := range ops {
		switch op.kind {
		case opEqual:
			fmt.Fprintf(w, "  %s\n", op.text)
		case opDel:
			fmt.Fprintln(w, paint(color, cRed, "- "+op.text))
		case opAdd:
			fmt.Fprintln(w, paint(color, cGreen, "+ "+op.text))
		}
	}
}

func renderSideBySide(w io.Writer, a, b string, width int, color bool) {
	al := strings.Split(a, "\n")
	bl := strings.Split(b, "\n")
	ops := lcsDiff(al, bl)

	// Build aligned rows.
	type row struct {
		l, r   string
		lc, rc string
	}
	var rows []row
	var pendDel []string
	flush := func() {
		// pair pending deletions with nothing on the right; handled inline below
	}
	_ = flush
	i := 0
	for i < len(ops) {
		op := ops[i]
		switch op.kind {
		case opEqual:
			rows = append(rows, row{op.text, op.text, "", ""})
			i++
		case opDel:
			// gather a run of deletions, then a run of additions, and zip them
			var dels, adds []string
			for i < len(ops) && ops[i].kind == opDel {
				dels = append(dels, ops[i].text)
				i++
			}
			for i < len(ops) && ops[i].kind == opAdd {
				adds = append(adds, ops[i].text)
				i++
			}
			n := max(len(dels), len(adds))
			for k := range n {
				var l, r string
				var lc, rc string
				if k < len(dels) {
					l, lc = dels[k], cRed
				}
				if k < len(adds) {
					r, rc = adds[k], cGreen
				}
				rows = append(rows, row{l, r, lc, rc})
			}
		case opAdd:
			rows = append(rows, row{"", op.text, "", cGreen})
			i++
		}
		_ = pendDel
	}

	sep := " │ "
	// Header and divider must share the exact column geometry of the data
	// rows: each side is `width` columns, joined by sep (" │ "). The divider
	// mirrors sep as "─┼─" so the ┼ lands directly under every │.
	// Pad the header text manually (not via %-*s) so it stays correct even
	// when paint() wraps it in ANSI escapes, which %-*s would miscount.
	fmt.Fprintf(w, "  %s%s%s\n",
		paint(color, cDim, truncPad("current", width)),
		sep,
		paint(color, cDim, "proposed"))
	fmt.Fprintf(w, "  %s\n", paint(color, cDim,
		strings.Repeat("─", width)+"─┼─"+strings.Repeat("─", width)))
	for _, r := range rows {
		left := truncPad(r.l, width)
		right := truncPad(r.r, width)
		if r.lc != "" {
			left = paint(color, r.lc, left)
		}
		if r.rc != "" {
			right = paint(color, r.rc, right)
		}
		fmt.Fprintf(w, "  %s%s%s\n", left, sep, right)
	}
}

func truncPad(s string, w int) string {
	// expand tabs to 4 spaces for stable columns
	s = strings.ReplaceAll(s, "\t", "    ")
	if len(s) > w {
		if w > 1 {
			return s[:w-1] + "…"
		}
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

// --- tiny LCS line diff -----------------------------------------------------

type opKind int

const (
	opEqual opKind = iota
	opDel
	opAdd
)

type diffOp struct {
	kind opKind
	text string
}

// lcsDiff produces a line-level edit script (equal/del/add) for the two slices
// of lines. It delegates the actual diffing to github.com/aymanbagabas/go-udiff
// (a maintained standalone port of the Myers diff packages gopls uses), via
// udiff.Lines, which yields edits aligned to whole-line boundaries. We then map
// those byte-offset edits back onto line indices to rebuild the equal/del/add
// stream the renderers want.
//
// go-udiff is a real public module (unlike golang.org/x/tools/internal/diff,
// which cannot be imported from outside the x/tools module), so this works from
// any module. It is aliased to `diff` at the import for brevity.
func lcsDiff(a, b []string) []diffOp {
	before := strings.Join(a, "\n")
	after := strings.Join(b, "\n")

	edits := diff.Lines(before, after)
	// diff.Lines should already be sorted by Start, but be defensive.
	diff.SortEdits(edits)

	// Precompute the byte offset at which each before-line starts, so we can
	// translate an edit's [Start,End) byte span into a [first,last) line range.
	lineStart := make([]int, len(a)+1)
	off := 0
	for i, ln := range a {
		lineStart[i] = off
		off += len(ln) + 1 // +1 for the '\n' join separator
	}
	lineStart[len(a)] = off // sentinel: one past the last line

	// offsetToLine maps a byte offset to the index of the line that begins at
	// or before it. Edits from diff.Lines fall on line boundaries, so offsets
	// coincide with entries in lineStart.
	offsetToLine := func(o int) int {
		// linear scan is fine: structs are tiny.
		for i := range a {
			if lineStart[i] == o {
				return i
			}
			if lineStart[i] > o {
				return i // shouldn't happen on a boundary, but stay safe
			}
		}
		return len(a)
	}

	var ops []diffOp
	cur := 0 // next unconsumed before-line index
	for _, e := range edits {
		delStart := offsetToLine(e.Start)
		delEnd := offsetToLine(e.End) // exclusive

		// Emit unchanged lines before this edit.
		for ; cur < delStart; cur++ {
			ops = append(ops, diffOp{opEqual, a[cur]})
		}
		// Emit deletions for the lines this edit replaces.
		for ; cur < delEnd; cur++ {
			ops = append(ops, diffOp{opDel, a[cur]})
		}
		// Emit additions from the edit's replacement text. New ends in "\n"
		// for full-line edits; split and drop the trailing empty element.
		newLines := strings.Split(e.New, "\n")
		if len(newLines) > 0 && newLines[len(newLines)-1] == "" {
			newLines = newLines[:len(newLines)-1]
		}
		for _, nl := range newLines {
			ops = append(ops, diffOp{opAdd, nl})
		}
	}
	// Trailing unchanged lines after the last edit.
	for ; cur < len(a); cur++ {
		ops = append(ops, diffOp{opEqual, a[cur]})
	}
	return ops
}

// --- misc -------------------------------------------------------------------

// resolveWidth returns the default per-side column width for -diff=side,
// derived from the current terminal width. It is used as the flag's default, so
// `structalign -h` shows the real number and an explicit `-width=N` overrides it.
//
// A side-by-side row prints as:
//
//	"  " + <width cols> + " │ " + <width cols>
//	 2   +   width      +   3   +   width        = 2*width + 5 display columns
//
// so width = (terminalCols - 5) / 2. We detect the terminal width directly via
// term.GetSize (an ioctl, equivalent to `tput cols`), which works whether or not
// the shell exports COLUMNS, and which fails cleanly when output is piped — in
// that case we fall back to $COLUMNS, then to a fixed default.
func resolveWidth() int {
	const (
		overhead    = 5  // "  " + " │ "
		fallback    = 80 // when no terminal size is available (e.g. piped)
		minWidth    = 20 // don't let very narrow terminals produce unusable columns
		minTermCols = overhead + 2*minWidth
	)

	fromCols := func(cols int) (int, bool) {
		if cols < minTermCols {
			return 0, false
		}
		return (cols - overhead) / 2, true
	}

	// 1. Ask the terminal attached to stdout directly.
	if cols, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		if w, ok := fromCols(cols); ok {
			return w
		}
		// Terminal exists but is narrower than our minimum: clamp.
		if cols >= overhead+2 {
			return minWidth
		}
	}

	// 2. Fall back to $COLUMNS (only set if the shell exported it).
	if c, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil {
		if w, ok := fromCols(c); ok {
			return w
		}
	}

	// 3. Nothing usable (piped output, unknown terminal): fixed default.
	return fallback
}

func wantColor(mode string) bool {
	switch mode {
	case "always":
		return true
	case "never":
		return false
	default: // auto
		fi, err := os.Stdout.Stat()
		if err != nil {
			return false
		}
		return fi.Mode()&os.ModeCharDevice != 0
	}
}
