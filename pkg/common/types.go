// Package common holds structalign's public contracts: the interfaces that
// decouple the CLI's layers and the plain data types those interfaces traffic
// in. Implementations live under internal/. Keeping the contracts here (rather
// than internal/) lets mockery generate mocks from a non-internal source.
package common

import (
	"go/ast"
	"go/token"
	"go/types"
)

// Target is a loader-agnostic view of one typed Go package: everything the
// analyzer and the layout inspector need, without exposing go/packages.Package.
type Target struct {
	PkgPath   string
	Fset      *token.FileSet
	Syntax    []*ast.File
	Types     *types.Package
	TypesInfo *types.Info
	Sizes     Sizes
	Errors    []error
}

// Finding is one struct whose fields could be reordered to use less memory.
type Finding struct {
	Package    string
	Fset       *token.FileSet
	Pos        token.Pos
	Name       string // enclosing named type, or "" for an anonymous struct
	TypeParams string // type-parameter names for a generic type, e.g. "[T]" (else "")
	Message    string // analyzer diagnostic (carries the size info)
	Original   string // current struct source ("struct{...}")
	Proposed   string // reordered struct from the analyzer's SuggestedFix
	OldSize    int64  // current struct size from the analyzer message (0 if unknown)
	NewSize    int64  // proposed (optimal) size from the analyzer message (0 if unknown)
}

// LayoutField is one field's place in a struct's memory layout.
type LayoutField struct {
	Name    string
	Type    string
	Tag     string // raw struct tag without backticks, or ""
	Assume  string // for a generic field, the assumed type param(s), e.g. "T=any" (else "")
	Offset  int64
	Size    int64
	Align   int64
	Padding int64 // padding inserted after this field
}

// Layout is one named struct's computed memory layout.
type Layout struct {
	Package    string
	Name       string
	TypeParams string // for a generic type, e.g. "[T]" (else "")
	Note       string // optional caveat shown above the struct (e.g. the generic disclaimer)
	Total      int64
	Align      int64
	Padding    int64 // total padding across all fields
	Fields     []LayoutField
}
