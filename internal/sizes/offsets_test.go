package sizes_test

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/sizes"
)

func TestOffsetsof(t *testing.T) {
	s := sizes.ForArch("amd64")
	// For struct{ A bool; B int64 } on amd64, B aligns to offset 8.
	fields := []*types.Var{
		types.NewField(token.NoPos, nil, "A", types.Typ[types.Bool], false),
		types.NewField(token.NoPos, nil, "B", types.Typ[types.Int64], false),
	}
	require.Equal(t, []int64{0, 8}, s.Offsetsof(fields))
}
