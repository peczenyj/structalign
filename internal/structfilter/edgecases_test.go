package structfilter_test

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/structfilter"
	"github.com/peczenyj/structalign/internal/testutil"
)

func TestHasCacheLinePadPlainStruct(t *testing.T) {
	tgt := testutil.Target(t, "package sample\n\ntype Plain struct {\n\tA bool\n\tB int64\n}\n")
	st := tgt.Types.Scope().Lookup("Plain").(*types.TypeName).Type().Underlying().(*types.Struct)
	require.False(t, structfilter.HasCacheLinePad(st), "a struct of basic fields has no cache-line pad")
}

func TestInGeneratedFileOutOfRange(t *testing.T) {
	tgt := testutil.Target(t, "package sample\n\ntype X struct{ A bool }\n")
	// A position that belongs to no syntax file is treated as not generated.
	require.False(t, structfilter.InGeneratedFile(tgt, token.NoPos))
}
