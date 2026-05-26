package align_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/pkg/common"
)

func TestFindingsNolintBlockComment(t *testing.T) {
	const src = `package sample
/* nolint */
type Block struct {
	A bool
	B int64
	C bool
}
`
	tgt := testutil.Target(t, src)

	fs, err := align.New().Findings(tgt, common.Options{
		RespectNolint: true,
		NolintLinters: []string{"fieldalignment"},
	})
	require.NoError(t, err)

	found := false
	for _, f := range fs {
		if f.Name == "Block" {
			found = true
		}
	}
	assert.False(t, found, "nolint in block comment should suppress")
}
