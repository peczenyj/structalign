// Package structfilter holds predicates for excluding structs from analysis:
// generated files and cache-line-padded structs.
package structfilter

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/peczenyj/structalign/pkg/common"
)

// InGeneratedFile reports whether pos falls in one of t's syntax files that
// carries the standard "// Code generated ... DO NOT EDIT." marker.
func InGeneratedFile(t common.Target, pos token.Pos) bool {
	for _, f := range t.Syntax {
		if f.FileStart <= pos && pos < f.FileEnd {
			return ast.IsGenerated(f)
		}
	}
	return false
}

// HasCacheLinePad reports whether st has a field of type
// golang.org/x/sys/cpu.CacheLinePad (the false-sharing guard).
func HasCacheLinePad(st *types.Struct) bool {
	for i := range st.NumFields() {
		named, ok := st.Field(i).Type().(*types.Named)
		if !ok {
			continue
		}
		obj := named.Obj()
		if obj.Pkg() != nil && obj.Pkg().Path() == "golang.org/x/sys/cpu" && obj.Name() == "CacheLinePad" {
			return true
		}
	}
	return false
}
