package common

//go:generate go tool enumer -type=DiffStyle -text -transform=lower -trimprefix=Diff -pflag.value

// DiffStyle selects how a Finding is rendered. enumer generates its
// String/parse/text-marshal helpers, a flag.Value implementation (Set), and a
// Type method for usage strings, see diffstyle_enumer.go; the names map to
// "unified"/"side"/"none" via -trimprefix=Diff -transform=lower.
type DiffStyle uint8

const (
	// DiffUnified displays changes in the standard unified diff format.
	DiffUnified DiffStyle = iota
	// DiffSide displays changes in side-by-side columns.
	DiffSide
	// DiffNone suppresses diff rendering.
	DiffNone
)
