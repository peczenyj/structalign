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

func TestStripStructTagsError(t *testing.T) {
	_, err := stripStructTags("invalid struct {")
	assert.Error(t, err)

	_, err = stripStructTags("int") // not a struct
	assert.Error(t, err)
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

	// Bounds check: stop > len(data)
	// We need a file that exists but we use offsets past its length.
	// But readSource reads the file from disk using pf.Name().
	// It's hard to trigger start < 0 or start > stop with valid token.Pos.
}
