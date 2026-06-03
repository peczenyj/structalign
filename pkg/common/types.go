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
	// PkgPath is the full Go package import path (e.g., "github.com/peczenyj/structalign/pkg/common").
	PkgPath string
	// Fset is the FileSet used to resolve AST positions within this package.
	Fset *token.FileSet
	// Syntax is the parsed AST slice representing all Go source files in the package.
	Syntax []*ast.File
	// Types is the type system package representation containing scopes and type definitions.
	Types *types.Package
	// TypesInfo contains the type-checker's AST-to-type associations, definitions, and uses.
	TypesInfo *types.Info
	// Sizes provides the target-architecture specific type sizing logic.
	Sizes Sizes
	// Errors lists any syntax or type-checking errors encountered during package loading.
	Errors []error
}

// Finding is one struct whose fields could be reordered to use less memory.
type Finding struct {
	// Package is the name of the package containing the finding.
	Package string
	// Fset is the FileSet used to decode the Pos location into file, line, and column.
	Fset *token.FileSet
	// Pos is the position of the struct declaration.
	Pos token.Pos
	// Name is the enclosing named type name, or an empty string for an anonymous struct.
	Name string
	// TypeParams represents the type-parameter names for a generic type, e.g. "[T]" (else empty).
	TypeParams string
	// Message is the high-level human-readable size comparison diagnostic.
	Message string
	// Original is the original struct source code representation (e.g. "struct{...}").
	Original string
	// Proposed is the reordered struct source code from the analyzer's suggested fix.
	Proposed string
	// OldSize is the original struct size in bytes.
	OldSize int64
	// NewSize is the proposed (reordered) struct size in bytes.
	NewSize int64
}

// LayoutField is one field's place in a struct's memory layout.
type LayoutField struct {
	// Name is the name of the struct field.
	Name string
	// Type is the type of the field represented as a string.
	Type string
	// Tag is the raw struct field tag (excluding surrounding backticks), or empty if none exists.
	Tag string
	// Assume is the assumed type parameter(s) for a generic field, e.g. "T=any", or empty if none exists.
	Assume string
	// Offset is the byte offset of the field within the struct layout.
	Offset int64
	// Size is the size of the field's type in bytes.
	Size int64
	// Align is the alignment of the field's type in bytes.
	Align int64
	// Padding is the trailing padding in bytes inserted after this field.
	Padding int64
}

// Layout is one named struct's computed memory layout.
type Layout struct {
	// Package is the name of the package declaring the struct.
	Package string
	// Name is the name of the struct type.
	Name string
	// TypeParams represents the type parameters for a generic type, e.g. "[T]" (else empty).
	TypeParams string
	// Note is an optional message warning or disclaimer regarding the layout details.
	Note string
	// Total is the total size of the struct in bytes, including padding.
	Total int64
	// Align is the alignment requirement of the struct in bytes.
	Align int64
	// Padding is the sum of all padding bytes within the struct.
	Padding int64
	// Fields lists the struct's fields in memory-layout order.
	Fields []LayoutField
}
