// Package ui renders findings and layouts as terminal output (unified /
// side-by-side / proposed-only diffs and annotated layout). It is the only
// package that produces user-facing text.
package ui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"

	"github.com/peczenyj/structalign/internal/textdiff"
	"github.com/peczenyj/structalign/pkg/common"
)

// styleFactory generates the SGR sequences for every style; termenv owns escape
// generation so we no longer hand-assemble ANSI strings. The profile is fixed at
// ANSI256 (not detected from the terminal) to reproduce the historical output:
// the 16-color roles render as their plain SGR codes and the amber palette keeps
// its 256-color (38;5;214) sequence, which the narrower ANSI profile would
// downsample to bright yellow. The io.Discard writer is irrelevant — with a
// forced profile the factory is used only to render strings, never to write.
var styleFactory = termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.ANSI256))

// Style is the appearance of one semantic role: an optional foreground color
// plus attributes. The zero value is "no styling" and renders text unchanged.
type Style struct {
	fg      termenv.Color // nil = default foreground
	bold    bool
	faint   bool
	reverse bool
}

// render wraps text in the style's SGR sequence (via termenv). Attributes are
// applied before the color so the combined sequence reads bold/faint/reverse
// then foreground (e.g. "\x1b[1;36m" for bold cyan).
func (s Style) render(text string) string {
	if s == (Style{}) {
		return text
	}
	st := styleFactory.String(text)
	if s.bold {
		st = st.Bold()
	}
	if s.faint {
		st = st.Faint()
	}
	if s.reverse {
		st = st.Reverse()
	}
	if s.fg != nil {
		st = st.Foreground(s.fg)
	}
	return st.String()
}

// Theme maps semantic roles to styles. The zero value resolves to DefaultTheme
// (see Printer.theme), which reproduces the historical palette.
type Theme struct {
	Header  Style // finding header / inspect "type X struct {" line
	Added   Style // "+" diff lines / added side cells
	Removed Style // "-" diff lines / removed side cells
	Meta    Style // column titles, divider, layout note, "-- assume" marker
	Padding Style // inspect padding comment / "_" padding line
	Label   Style // the -summary "Summary:" label
}

// DefaultTheme is the historical palette: bold cyan header, green added, red
// removed/padding, dim meta, bold label. termenv renders these as the same
// visual output the hand-rolled SGR constants produced.
func DefaultTheme() Theme {
	return Theme{
		Header:  Style{fg: termenv.ANSICyan, bold: true},
		Added:   Style{fg: termenv.ANSIGreen},
		Removed: Style{fg: termenv.ANSIRed},
		Meta:    Style{faint: true},
		Padding: Style{fg: termenv.ANSIRed},
		Label:   Style{bold: true},
	}
}

// Printer renders to Out using the given color/width settings.
type Printer struct {
	Out   io.Writer
	Err   io.Writer // error stream (defaults to os.Stderr if nil)
	Color bool
	Width int   // per-side column width for side-by-side diffs
	Theme Theme // zero value resolves to DefaultTheme
}

// err returns the configured error stream, or os.Stderr when unset.
func (p *Printer) err() io.Writer {
	if p.Err != nil {
		return p.Err
	}
	return os.Stderr
}

// theme returns the configured theme, or DefaultTheme when unset.
func (p *Printer) theme() Theme {
	if (p.Theme == Theme{}) {
		return DefaultTheme()
	}
	return p.Theme
}

// RenderFindings renders each finding in the chosen diff style. Returns the count.
func (p *Printer) RenderFindings(findings []common.Finding, style common.DiffStyle) int {
	for _, f := range findings {
		p.renderFinding(f, style)
	}
	return len(findings)
}

// RenderLayouts renders each struct layout. Returns the count.
func (p *Printer) RenderLayouts(layouts []common.Layout, verbose, keepTags bool) int {
	for _, l := range layouts {
		p.renderLayout(l, verbose, keepTags)
	}
	return len(layouts)
}

// RenderSummary writes a one-line diff-mode summary to Out. The "Summary:"
// label is bold when color is on; counts are pluralized.
func (p *Printer) RenderSummary(structs int, bytesSaved int64) {
	fmt.Fprintf(p.Out, "%s %d %s affected, %d %s saved total\n", //nolint:errcheck
		p.paint(p.theme().Label, "Summary:"),
		structs, plural(int64(structs), "struct", "structs"),
		bytesSaved, plural(bytesSaved, "byte", "bytes"))
}

// plural returns one when n == 1, else many.
func plural(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func (p *Printer) renderFinding(f common.Finding, style common.DiffStyle) {
	loc := f.Fset.Position(f.Pos)
	file := relPath(loc.Filename)
	var header string
	if f.Name != "" {
		header = fmt.Sprintf("%s:%d:%d: %s: %s", file, loc.Line, loc.Column, f.Name, f.Message)
	} else {
		header = fmt.Sprintf("%s:%d:%d: %s", file, loc.Line, loc.Column, f.Message)
	}
	if f.OldSize > 0 && f.NewSize > 0 && f.NewSize < f.OldSize {
		saved := f.OldSize - f.NewSize
		pct := float64(saved) / float64(f.OldSize) * 100
		// Surface the absolute bytes saved (the -threshold unit) next to the percentage.
		header += fmt.Sprintf(", saving %d %s (%02.2f%% smaller)", saved, plural(saved, "byte", "bytes"), pct)
	}
	fmt.Fprintln(p.Out, p.paint(p.theme().Header, header)) //nolint:errcheck

	if f.Original == "" || f.Proposed == "" {
		fmt.Fprintln(p.Out, "  (no suggested fix produced)") //nolint:errcheck
		fmt.Fprintln(p.Out)                                  //nolint:errcheck
		return
	}

	// Name plus any generic type parameters, e.g. "Generic[T]".
	decl := f.Name + f.TypeParams
	orig := withTypeName(f.Original, decl)
	prop := withTypeName(f.Proposed, decl)

	switch style {
	case common.DiffNone:
		// just the proposed struct
		fmt.Fprintln(p.Out, indent(prop, "  ")) //nolint:errcheck
	case common.DiffSide:
		p.renderSideBySide(orig, prop)
	default: // unified
		p.renderUnified(orig, prop)
	}
	fmt.Fprintln(p.Out) //nolint:errcheck
}

func (p *Printer) renderUnified(a, b string) {
	th := p.theme()
	al := strings.Split(a, "\n")
	bl := strings.Split(b, "\n")
	ops := textdiff.Lines(al, bl)
	for _, op := range ops {
		switch op.Kind {
		case textdiff.Equal:
			fmt.Fprintf(p.Out, "  %s\n", op.Text) //nolint:errcheck
		case textdiff.Del:
			fmt.Fprintln(p.Out, p.paint(th.Removed, "- "+op.Text)) //nolint:errcheck
		case textdiff.Add:
			fmt.Fprintln(p.Out, p.paint(th.Added, "+ "+op.Text)) //nolint:errcheck
		}
	}
}

func (p *Printer) renderSideBySide(a, b string) {
	th := p.theme()
	al := strings.Split(a, "\n")
	bl := strings.Split(b, "\n")
	ops := textdiff.Lines(al, bl)

	// Build aligned rows. lc/rc are the per-side styles (zero Style = no color).
	type row struct {
		l, r   string
		lc, rc Style
	}
	var rows []row
	i := 0
	for i < len(ops) {
		op := ops[i]
		switch op.Kind {
		case textdiff.Equal:
			rows = append(rows, row{l: op.Text, r: op.Text})
			i++
		case textdiff.Del:
			// gather a run of deletions, then a run of additions, and zip them
			var dels, adds []string
			for i < len(ops) && ops[i].Kind == textdiff.Del {
				dels = append(dels, ops[i].Text)
				i++
			}
			for i < len(ops) && ops[i].Kind == textdiff.Add {
				adds = append(adds, ops[i].Text)
				i++
			}
			n := max(len(dels), len(adds))
			for k := range n {
				var l, r string
				var lc, rc Style
				if k < len(dels) {
					l, lc = dels[k], th.Removed
				}
				if k < len(adds) {
					r, rc = adds[k], th.Added
				}
				rows = append(rows, row{l, r, lc, rc})
			}
		case textdiff.Add:
			rows = append(rows, row{r: op.Text, rc: th.Added})
			i++
		}
	}

	sep := " │ "
	// Header and divider must share the exact column geometry of the data
	// rows: each side is `width` columns, joined by sep (" │ "). The divider
	// mirrors sep as "─┼─" so the ┼ lands directly under every │.
	// Pad the header text manually (not via %-*s) so it stays correct even
	// when paint wraps it in ANSI escapes, which %-*s would miscount.
	fmt.Fprintf(p.Out, "  %s%s%s\n", //nolint:errcheck
		p.paint(th.Meta, truncPad("current", p.Width)),
		sep,
		p.paint(th.Meta, "proposed"))
	fmt.Fprintf(p.Out, "  %s\n", p.paint(th.Meta, //nolint:errcheck
		strings.Repeat("─", p.Width)+"─┼─"+strings.Repeat("─", p.Width)))
	for _, r := range rows {
		left := truncPad(r.l, p.Width)
		right := truncPad(r.r, p.Width)
		if r.lc != (Style{}) {
			left = p.paint(r.lc, left)
		}
		if r.rc != (Style{}) {
			right = p.paint(r.rc, right)
		}
		fmt.Fprintf(p.Out, "  %s%s%s\n", left, sep, right) //nolint:errcheck
	}
}

func (p *Printer) renderLayout(l common.Layout, verbose, keepTags bool) {
	th := p.theme()
	// Field declaration is "<name> <type>" (plus tag when -tags is set); align
	// all comments to the widest declaration.
	decls := make([]string, len(l.Fields))
	declWidth := 0
	for i, f := range l.Fields {
		decls[i] = f.Name + " " + f.Type
		if keepTags && f.Tag != "" {
			decls[i] += " `" + f.Tag + "`"
		}
		if len(decls[i]) > declWidth {
			declWidth = len(decls[i])
		}
	}
	// In verbose mode a lone "_" can also appear; it's never wider than a field.

	// Optional caveat (e.g. the generic-type disclaimer) above the declaration.
	if l.Note != "" {
		fmt.Fprintln(p.Out, p.paint(th.Meta, "// "+l.Note)) //nolint:errcheck
	}

	// Header: the struct opening line carries size/align/padding.
	header := fmt.Sprintf("type %s%s struct { // size: %d, align: %d, padding: %d",
		l.Name, l.TypeParams, l.Total, l.Align, l.Padding)
	fmt.Fprintln(p.Out, p.paint(th.Header, header)) //nolint:errcheck

	comments, commentWidth := layoutComments(l.Fields, verbose)

	for i, f := range l.Fields {
		comment := comments[i]
		rendered := comment
		if !verbose && f.Padding > 0 {
			rendered = p.paint(th.Padding, comment)
		}
		line := fmt.Sprintf("\t%-*s // %s", declWidth, decls[i], rendered)
		if f.Assume != "" {
			pad := strings.Repeat(" ", commentWidth-len(comment))
			line += pad + "   " + p.paint(th.Meta, "-- assume "+f.Assume)
		}
		fmt.Fprintln(p.Out, line) //nolint:errcheck
		if verbose && f.Padding > 0 {
			// Field line carries no padding; padding gets its own `_` line.
			pad := fmt.Sprintf("\t%-*s // %d byte padding", declWidth, "_", f.Padding)
			fmt.Fprintln(p.Out, p.paint(th.Padding, pad)) //nolint:errcheck
		}
	}
	fmt.Fprintln(p.Out, "}") //nolint:errcheck
	fmt.Fprintln(p.Out)      //nolint:errcheck
}

// layoutComments builds the plain (un-colored) comment text for each field line
// and the widest of them, so a generic field's "-- assume T=any" marker can be
// aligned in a column past every comment. In non-verbose mode the trailing
// padding folds onto the comment; in verbose mode it gets its own line.
func layoutComments(fields []common.LayoutField, verbose bool) (comments []string, width int) {
	comments = make([]string, len(fields))
	for i, f := range fields {
		c := fmt.Sprintf("size: %2d, align: %d", f.Size, f.Align)
		if !verbose && f.Padding > 0 {
			c += fmt.Sprintf(", padding: %d", f.Padding)
		}
		comments[i] = c
		if len(c) > width {
			width = len(c)
		}
	}
	return comments, width
}

func indent(s, pad string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if i == len(lines)-1 && l == "" {
			break
		}
		lines[i] = pad + l
	}
	return strings.Join(lines, "\n")
}

// withTypeName rewrites a leading "struct {" line into "type <name> struct {"
// so the rendered diff reads as a full type declaration. Only the first line is
// touched (nested "struct {" fields are left alone). A no-op when name == "" or
// the text doesn't begin with a struct line.
func withTypeName(src, name string) string {
	if name == "" {
		return src
	}
	first, rest, _ := strings.Cut(src, "\n")
	if !strings.HasPrefix(first, "struct {") {
		return src
	}
	first = "type " + name + " " + first
	if rest == "" {
		return first
	}
	return first + "\n" + rest
}

// truncPad fits s into exactly w display cells: right-padded with spaces when
// shorter, or truncated with a trailing "…" when longer. Width is measured in
// terminal cells via runewidth, so CJK and other wide runes count as two.
func truncPad(s string, w int) string {
	if w <= 0 {
		return ""
	}
	// expand tabs to 4 spaces for stable columns
	s = strings.ReplaceAll(s, "\t", "    ")
	// runewidth.Truncate fits s into w cells, appending "…" (and reserving its
	// cell) when it has to cut — including the wide-rune-straddling-the-boundary
	// case. FillRight then right-pads the result to exactly w cells. Both measure
	// in terminal cells, so CJK and other wide runes count as two.
	return runewidth.FillRight(runewidth.Truncate(s, w, "…"), w)
}

// paint renders s in role's style when color is enabled, else returns s plain.
func (p *Printer) paint(role Style, s string) string {
	if !p.Color {
		return s
	}
	return role.render(s)
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
