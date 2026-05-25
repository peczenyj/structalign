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

	"github.com/peczenyj/structalign/internal/textdiff"
	"github.com/peczenyj/structalign/pkg/common"
)

const (
	cReset = "\x1b[0m"
	cRed   = "\x1b[31m"
	cGreen = "\x1b[32m"
	cCyan  = "\x1b[36m"
	cDim   = "\x1b[2m"
	cBold  = "\x1b[1m"
)

// Printer renders to Out using the given color/width settings.
type Printer struct {
	Out   io.Writer
	Color bool
	Width int // per-side column width for side-by-side diffs
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
		pct := float64(f.OldSize-f.NewSize) / float64(f.OldSize) * 100
		header += fmt.Sprintf(" (%02.2f%% smaller)", pct)
	}
	fmt.Fprintln(p.Out, paint(p.Color, cBold+cCyan, header)) //nolint:errcheck

	if f.Original == "" || f.Proposed == "" {
		fmt.Fprintln(p.Out, "  (no suggested fix produced)") //nolint:errcheck
		fmt.Fprintln(p.Out)                                  //nolint:errcheck
		return
	}

	orig := withTypeName(f.Original, f.Name)
	prop := withTypeName(f.Proposed, f.Name)

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
	al := strings.Split(a, "\n")
	bl := strings.Split(b, "\n")
	ops := textdiff.Lines(al, bl)
	for _, op := range ops {
		switch op.Kind {
		case textdiff.Equal:
			fmt.Fprintf(p.Out, "  %s\n", op.Text) //nolint:errcheck
		case textdiff.Del:
			fmt.Fprintln(p.Out, paint(p.Color, cRed, "- "+op.Text)) //nolint:errcheck
		case textdiff.Add:
			fmt.Fprintln(p.Out, paint(p.Color, cGreen, "+ "+op.Text)) //nolint:errcheck
		}
	}
}

func (p *Printer) renderSideBySide(a, b string) {
	al := strings.Split(a, "\n")
	bl := strings.Split(b, "\n")
	ops := textdiff.Lines(al, bl)

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
		switch op.Kind {
		case textdiff.Equal:
			rows = append(rows, row{op.Text, op.Text, "", ""})
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
				var lc, rc string
				if k < len(dels) {
					l, lc = dels[k], cRed
				}
				if k < len(adds) {
					r, rc = adds[k], cGreen
				}
				rows = append(rows, row{l, r, lc, rc})
			}
		case textdiff.Add:
			rows = append(rows, row{"", op.Text, "", cGreen})
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
	fmt.Fprintf(p.Out, "  %s%s%s\n", //nolint:errcheck
		paint(p.Color, cDim, truncPad("current", p.Width)),
		sep,
		paint(p.Color, cDim, "proposed"))
	fmt.Fprintf(p.Out, "  %s\n", paint(p.Color, cDim, //nolint:errcheck
		strings.Repeat("─", p.Width)+"─┼─"+strings.Repeat("─", p.Width)))
	for _, r := range rows {
		left := truncPad(r.l, p.Width)
		right := truncPad(r.r, p.Width)
		if r.lc != "" {
			left = paint(p.Color, r.lc, left)
		}
		if r.rc != "" {
			right = paint(p.Color, r.rc, right)
		}
		fmt.Fprintf(p.Out, "  %s%s%s\n", left, sep, right) //nolint:errcheck
	}
}

func (p *Printer) renderLayout(l common.Layout, verbose, keepTags bool) {
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

	// Header: the struct opening line carries size/align/padding.
	header := fmt.Sprintf("type %s struct { // size: %d, align: %d, padding: %d",
		l.Name, l.Total, l.Align, l.Padding)
	fmt.Fprintln(p.Out, paint(p.Color, cBold+cCyan, header)) //nolint:errcheck

	for i, f := range l.Fields {
		base := fmt.Sprintf("size: %2d, align: %d", f.Size, f.Align)
		if verbose {
			// Field line carries no padding; padding gets its own `_` line.
			fmt.Fprintf(p.Out, "\t%-*s // %s\n", declWidth, decls[i], base) //nolint:errcheck
			if f.Padding > 0 {
				pad := fmt.Sprintf("\t%-*s // %d byte padding", declWidth, "_", f.Padding)
				fmt.Fprintln(p.Out, paint(p.Color, cRed, pad)) //nolint:errcheck
			}
		} else {
			// Padding folds onto the field's own comment.
			comment := base
			if f.Padding > 0 {
				comment = paint(p.Color, cRed, fmt.Sprintf("%s, padding: %d", base, f.Padding))
			}
			fmt.Fprintf(p.Out, "\t%-*s // %s\n", declWidth, decls[i], comment) //nolint:errcheck
		}
	}
	fmt.Fprintln(p.Out, "}") //nolint:errcheck
	fmt.Fprintln(p.Out)      //nolint:errcheck
}

func indent(s, pad string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
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
