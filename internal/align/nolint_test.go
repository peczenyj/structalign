package align_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/pkg/common"
)

const nolintSrc = `package sample

//nolint:fieldalignment
type Suppressed struct {
	A bool
	B int64
	C bool
}

type Visible struct {
	A bool
	B int64
	C bool
}

//nolint:errcheck
type Unrelated struct {
	A bool
	B int64
	C bool
}

type Trailing struct { //nolint
	A bool
	B int64
	C bool
}
`

func findingNames(fs []common.Finding) map[string]bool {
	m := make(map[string]bool)
	for _, f := range fs {
		m[f.Name] = true
	}
	return m
}

func TestFindingsRespectNolintByDefault(t *testing.T) {
	tgt := testutil.Target(t, nolintSrc)
	fs, err := align.New().Findings(tgt, common.Options{
		RespectNolint: true,
		NolintLinters: []string{"fieldalignment"},
	})
	require.NoError(t, err)
	names := findingNames(fs)
	assert.False(t, names["Suppressed"], "//nolint:fieldalignment must suppress")
	assert.True(t, names["Visible"], "unmarked struct stays")
	assert.True(t, names["Unrelated"], "//nolint:errcheck must NOT suppress fieldalignment")
	assert.False(t, names["Trailing"], "a bare //nolint (trailing) must suppress")
}

func TestFindingsShowNolint(t *testing.T) {
	tgt := testutil.Target(t, nolintSrc)
	fs, err := align.New().Findings(tgt, common.Options{RespectNolint: false})
	require.NoError(t, err)
	names := findingNames(fs)
	assert.True(t, names["Suppressed"], "respect off => suppressed struct reappears")
	assert.True(t, names["Trailing"], "respect off => bare-nolint struct reappears")
}

func TestFindingsNolintLintersConfigurable(t *testing.T) {
	tgt := testutil.Target(t, nolintSrc)
	// Only honoring "betteralign" => :fieldalignment no longer suppresses.
	fs, err := align.New().Findings(tgt, common.Options{
		RespectNolint: true,
		NolintLinters: []string{"betteralign"},
	})
	require.NoError(t, err)
	names := findingNames(fs)
	assert.True(t, names["Suppressed"], "fieldalignment token not in the configured set => not suppressed")
	assert.False(t, names["Trailing"], "bare //nolint is always honored regardless of the configured set")
}
