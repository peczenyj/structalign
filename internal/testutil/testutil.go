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
	"os"
	"testing"

	"github.com/peczenyj/structalign/internal/sizes"
	"github.com/peczenyj/structalign/pkg/common"
)

// srcName is the (relative) filename used for the in-memory source. Running from
// a temp dir and using a relative name keeps the analyzer's recorded path a
// stable "src.go" — deterministic for golden tests — while still letting
// readSource find the bytes on disk via that relative path.
const srcName = "src.go"

// Target parses and type-checks src (a complete .go file, including its package
// clause) and returns a common.Target with deterministic amd64 sizes. It runs
// the test from a temp dir (tb.Chdir, auto-restored) and writes the source to
// "src.go" there, so the recorded filename is stable and exists on disk.
func Target(tb testing.TB, src string) common.Target {
	tb.Helper()
	tb.Chdir(tb.TempDir())
	if err := os.WriteFile(srcName, []byte(src), 0o600); err != nil {
		tb.Fatalf("write source: %v", err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, srcName, src, parser.ParseComments)
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
