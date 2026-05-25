package common

import "go/types"

// Sizes abstracts go/types sizing so the target architecture is injectable:
// a real types.Sizes in production, a fixed types.SizesFor("gc","amd64") in
// tests (making golden output deterministic on any host arch). Its method set
// matches go/types.Sizes, so a common.Sizes value is directly assignable to a
// types.Sizes (e.g. analysis.Pass.TypesSizes).
type Sizes interface {
	Sizeof(t types.Type) int64
	Alignof(t types.Type) int64
	Offsetsof(fields []*types.Var) []int64
}

// Loader resolves Go package patterns (./..., import paths, directories, and
// "file=" queries) into typed Targets.
type Loader interface {
	Load(patterns ...string) ([]Target, error)
}

// Aligner produces the struct-reordering findings for one Target. patterns is
// a set of glob patterns matched against named-type names (nil = all). When
// keepTags is false, field tags are stripped from the rendered struct text.
type Aligner interface {
	Findings(t Target, patterns []string, keepTags bool) ([]Finding, error)
}

// Inspector computes the memory layout of each named struct in a Target,
// filtered by the same glob patterns as Aligner (nil = all).
type Inspector interface {
	Layouts(t Target, patterns []string) []Layout
}
