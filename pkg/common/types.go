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
	Fset     *token.FileSet
	Pos      token.Pos
	Name     string // enclosing named type, or "" for an anonymous struct
	Message  string // analyzer diagnostic (carries the size info)
	Original string // current struct source ("struct{...}")
	Proposed string // reordered struct from the analyzer's SuggestedFix
}

// LayoutField is one field's place in a struct's memory layout.
type LayoutField struct {
	Name    string
	Type    string
	Tag     string // raw struct tag without backticks, or ""
	Offset  int64
	Size    int64
	Align   int64
	Padding int64 // padding inserted after this field
}

// Layout is one named struct's computed memory layout.
type Layout struct {
	Name    string
	Total   int64
	Align   int64
	Padding int64 // total padding across all fields
	Fields  []LayoutField
}
