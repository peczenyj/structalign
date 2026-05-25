// Package loader resolves Go package patterns into typed common.Targets using
// golang.org/x/tools/go/packages.
package loader

import (
	"go/token"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/peczenyj/structalign/internal/sizes"
	"github.com/peczenyj/structalign/pkg/common"
)

// Loader implements common.Loader over go/packages.
type Loader struct{ tests bool }

// New returns a Loader. When tests is true, _test.go files (and test-variant
// packages) are loaded too.
func New(tests bool) *Loader { return &Loader{tests: tests} }

// Load resolves the patterns (./..., import paths, directories, "file=" queries,
// or bare .go file paths) into typed Targets.
func (l *Loader) Load(patterns ...string) ([]common.Target, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
			packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo,
		Fset:  token.NewFileSet(),
		Tests: l.tests,
	}
	pkgs, err := packages.Load(cfg, normalizeArgs(patterns)...)
	if err != nil {
		return nil, err
	}
	out := make([]common.Target, 0, len(pkgs))
	for _, p := range pkgs {
		t := common.Target{
			PkgPath:   p.PkgPath,
			Fset:      p.Fset,
			Syntax:    p.Syntax,
			Types:     p.Types,
			TypesInfo: p.TypesInfo,
		}
		if p.TypesSizes != nil {
			t.Sizes = sizes.New(p.TypesSizes)
		}
		for _, e := range p.Errors {
			t.Errors = append(t.Errors, e)
		}
		out = append(out, t)
	}
	return out, nil
}

// normalizeArgs rewrites a bare path to an existing .go file into the "file="
// query go/packages understands. Everything else passes through unchanged.
func normalizeArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if strings.HasSuffix(a, ".go") {
			if fi, err := os.Stat(a); err == nil && !fi.IsDir() {
				a = "file=" + a
			}
		}
		out = append(out, a)
	}
	return out
}
