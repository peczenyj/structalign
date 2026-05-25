package align_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/testutil"
)

const mixedSrc = `package sample

type Mixed struct {
	A bool
	B int64
	C bool
}

type Good struct {
	B int64
	A bool
	C bool
}
`

func TestFindingsReportsSuboptimalStruct(t *testing.T) {
	tgt := testutil.Target(t, mixedSrc)
	findings, err := align.New().Findings(tgt, []string{"Mixed"}, false)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "Mixed", findings[0].Name)
	assert.NotEmpty(t, findings[0].Original)
	assert.NotEmpty(t, findings[0].Proposed)
}

func TestFindingsSkipsOptimalStruct(t *testing.T) {
	tgt := testutil.Target(t, mixedSrc)
	findings, err := align.New().Findings(tgt, []string{"Good"}, false)
	require.NoError(t, err)
	assert.Empty(t, findings)
}
