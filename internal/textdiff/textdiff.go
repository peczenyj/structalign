// Package textdiff produces a line-level edit script (equal/del/add) for two
// slices of lines, delegating the diff to github.com/aymanbagabas/go-udiff.
package textdiff

import (
	"strings"

	diff "github.com/aymanbagabas/go-udiff"
)

// Kind is the kind of a diff Op.
type Kind int

const (
	Equal Kind = iota
	Del
	Add
)

// Op is one line in the edit script.
type Op struct {
	Kind Kind
	Text string
}

// Lines returns the equal/del/add edit script transforming a into b.
func Lines(a, b []string) []Op {
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

	var ops []Op
	cur := 0 // next unconsumed before-line index
	for _, e := range edits {
		delStart := offsetToLine(e.Start)
		delEnd := offsetToLine(e.End) // exclusive

		// Emit unchanged lines before this edit.
		for ; cur < delStart; cur++ {
			ops = append(ops, Op{Equal, a[cur]})
		}
		// Emit deletions for the lines this edit replaces.
		for ; cur < delEnd; cur++ {
			ops = append(ops, Op{Del, a[cur]})
		}
		// Emit additions from the edit's replacement text. New ends in "\n"
		// for full-line edits; split and drop the trailing empty element.
		newLines := strings.Split(e.New, "\n")
		if len(newLines) > 0 && newLines[len(newLines)-1] == "" {
			newLines = newLines[:len(newLines)-1]
		}
		for _, nl := range newLines {
			ops = append(ops, Op{Add, nl})
		}
	}
	// Trailing unchanged lines after the last edit.
	for ; cur < len(a); cur++ {
		ops = append(ops, Op{Equal, a[cur]})
	}
	return ops
}
