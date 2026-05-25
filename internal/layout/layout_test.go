package layout_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/testutil"
)

const src = `package sample

type Mixed struct {
	A bool
	B int64
	C bool
}
`

func TestLayoutsComputesPadding(t *testing.T) {
	tgt := testutil.Target(t, src)
	got := layout.New().Layouts(tgt, []string{"Mixed"})
	require.Len(t, got, 1)

	l := got[0]
	assert.Equal(t, "Mixed", l.Name)
	assert.Equal(t, int64(24), l.Total)
	assert.Equal(t, int64(8), l.Align)
	require.Len(t, l.Fields, 3)
	// A bool at offset 0, size 1, then 7 bytes padding before B int64.
	assert.Equal(t, int64(7), l.Fields[0].Padding)
}
