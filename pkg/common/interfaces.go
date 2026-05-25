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

// Aligner produces the struct-reordering findings for one Target, controlled
// by opts.
type Aligner interface {
	Findings(t Target, opts Options) ([]Finding, error)
}

// Inspector computes the memory layout of each named struct in a Target,
// controlled by opts.
type Inspector interface {
	Layouts(t Target, opts Options) []Layout
}
