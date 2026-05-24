package common

//go:generate go tool enumer -type=DiffStyle -text -transform=lower -trimprefix=Diff -flag.value

// DiffStyle selects how a Finding is rendered. enumer generates its
// String/parse/text-marshal helpers and a flag.Value implementation (Set), see
// diffstyle_enumer.go; the names map to "unified"/"side"/"none" via
// -trimprefix=Diff -transform=lower.
type DiffStyle uint8

const (
	DiffUnified DiffStyle = iota
	DiffSide
	DiffNone
)
