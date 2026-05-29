package align

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/pkg/common"
)

func TestIsCachePaddedFallback(t *testing.T) {
	pkg := types.NewPackage("golang.org/x/sys/cpu", "cpu")
	padType := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "CacheLinePad", nil), types.Universe.Lookup("int").Type(), nil)
	field := types.NewField(token.NoPos, pkg, "_", padType, false)
	st := types.NewStruct([]*types.Var{field}, nil)

	samplePkg := types.NewPackage("sample", "sample")
	samplePkg.Scope().Insert(types.NewTypeName(token.NoPos, samplePkg, "WithPad", types.NewNamed(types.NewTypeName(token.NoPos, samplePkg, "WithPad", nil), st, nil)))

	tgt := common.Target{Types: samplePkg}
	structs := make(map[token.Pos]*types.Struct)
	// Empty map ensures we hit the fallback.

	// Named type that exists
	assert.True(t, isCachePadded(tgt, token.NoPos, "WithPad", structs))

	// Named type that doesn't exist
	assert.False(t, isCachePadded(tgt, token.NoPos, "NoSuchType", structs))

	// Anonymous fallback
	assert.False(t, isCachePadded(tgt, token.NoPos, "", structs))
}

func TestNormalizeStructError(t *testing.T) {
	_, err := normalizeStruct("invalid struct {", false)
	assert.Error(t, err)

	_, err = normalizeStruct("int", false) // not a struct
	assert.Error(t, err)
}

func TestNormalizeStructStripsComments(t *testing.T) {
	src := "struct {\n\tA string // trailing\n\t// leading\n\tB int64\n}"

	got, err := normalizeStruct(src, false)
	assert.NoError(t, err)
	assert.NotContains(t, got, "// trailing")
	assert.NotContains(t, got, "// leading")
	assert.Contains(t, got, "A string")
	assert.Contains(t, got, "B int64")
}

func TestFindingsNoSyntax(t *testing.T) {
	a := Aligner{}
	f, err := a.Findings(common.Target{}, common.Options{})
	assert.NoError(t, err)
	assert.Nil(t, f)
}

func TestReadSourceErrors(t *testing.T) {
	fset := token.NewFileSet()
	// pf == nil
	assert.Empty(t, readSource(fset, token.NoPos, token.NoPos))

	f := fset.AddFile("missing.go", -1, 10)
	// os.ReadFile error
	assert.Empty(t, readSource(fset, f.Pos(0), f.Pos(5)))

	// The remaining bounds guards (start < 0, start > stop, stop > len(data))
	// are intentionally not covered here: readSource derives both offsets from
	// valid token.Pos values off the same file, so they can't be provoked
	// without a corrupt FileSet. The guards stay as defense in depth.
}
