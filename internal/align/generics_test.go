package align_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/pkg/common"
)

func TestFindingsCapturesTypeParams(t *testing.T) {
	src := "package sample\n\ntype Generic[T any] struct {\n\tFlag bool\n\tValue T\n\tCount uint32\n}\n"
	tgt := testutil.Target(t, src)

	findings, err := align.New().Findings(tgt, common.Options{Patterns: []string{"Generic"}})
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "[T]", findings[0].TypeParams, "generic type-parameter names are captured")
}

func TestFindingsNoTypeParamsForPlainStruct(t *testing.T) {
	src := "package sample\n\ntype Mixed struct {\n\tA bool\n\tB int64\n\tC bool\n}\n"
	tgt := testutil.Target(t, src)

	findings, err := align.New().Findings(tgt, common.Options{Patterns: []string{"Mixed"}})
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Empty(t, findings[0].TypeParams, "non-generic types have no type parameters")
}
