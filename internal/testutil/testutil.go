// Package testutil provides shared test helpers. Target type-checks a single
// source file in-process and returns a common.Target with amd64 sizes, so
// package tests get a real typed package without shelling out to `go list`.
package testutil

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/peczenyj/structalign/internal/sizes"
	"github.com/peczenyj/structalign/pkg/common"
)

// Target parses and type-checks src (a complete .go file, including its package
// clause) and returns a common.Target with deterministic amd64 sizes.
func Target(tb testing.TB, src string) common.Target {
	tb.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src.go", src, parser.ParseComments)
	if err != nil {
		tb.Fatalf("parse: %v", err)
	}
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("sample", fset, []*ast.File{file}, info)
	if err != nil {
		tb.Fatalf("typecheck: %v", err)
	}
	return common.Target{
		PkgPath:   "sample",
		Fset:      fset,
		Syntax:    []*ast.File{file},
		Types:     pkg,
		TypesInfo: info,
		Sizes:     sizes.ForArch("amd64"),
	}
}
