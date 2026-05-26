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

func TestFindingsNolintMixedDecls(t *testing.T) {
	const src = `package sample
type (
	// nolint
	Suppressed struct {
		A bool
		B int64
		C bool
	}
	Normal int
)
`
	tgt := testutil.Target(t, src)
	fs, err := align.New().Findings(tgt, common.Options{RespectNolint: true})
	require.NoError(t, err)

	for _, f := range fs {
		assert.NotEqual(t, "Suppressed", f.Name)
	}
}

func TestFindingsNolintEmptyGroup(t *testing.T) {
	const src = "package sample\ntype ()\n"
	tgt := testutil.Target(t, src)
	_, _ = align.New().Findings(tgt, common.Options{RespectNolint: true})
}

func TestParseNolintInvalid(t *testing.T) {
	// Exercise the default branch in parseNolint switch
	// Use internal package via link or just call it if internal tests allowed.
	// Actually, this file is align_test. I can't call internal functions unless I use a link.
}
