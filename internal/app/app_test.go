package app_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/app"
	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/pkg/common"
)

// stubLoader returns fixed targets, ignoring patterns — so app tests never shell
// out to the go toolchain. (A later PR replaces it with a generated mock.)
type stubLoader struct{ targets []common.Target }

func (s stubLoader) Load(...string) ([]common.Target, error) { return s.targets, nil }

const src = `package sample

type Mixed struct {
	A bool
	B int64
	C bool
}
`

func TestRunDiffExitsOneOnFindings(t *testing.T) {
	tgt := testutil.Target(t, src)
	var out, errb bytes.Buffer
	a := &app.App{
		Loader:    stubLoader{targets: []common.Target{tgt}},
		Aligner:   align.New(),
		Inspector: layout.New(),
		Stdout:    &out,
		Stderr:    &errb,
	}
	code := a.Run([]string{"-type=Mixed", "pkg"})
	assert.Equal(t, 1, code, "diff with findings should exit 1")
	assert.NotEmpty(t, out.String(), "expected diff output on stdout")
}

func TestRunVersion(t *testing.T) {
	var out, errb bytes.Buffer
	a := &app.App{Stdout: &out, Stderr: &errb}
	require.Equal(t, 0, a.Run([]string{"-version"}))
	assert.NotEmpty(t, out.String(), "expected version on stdout")
}

func TestRunInspectExitsZero(t *testing.T) {
	tgt := testutil.Target(t, src)
	var out, errb bytes.Buffer
	a := &app.App{
		Loader:    stubLoader{targets: []common.Target{tgt}},
		Aligner:   align.New(),
		Inspector: layout.New(),
		Stdout:    &out,
		Stderr:    &errb,
	}
	code := a.Run([]string{"-inspect", "-type=Mixed", "pkg"})
	assert.Equal(t, 0, code, "inspect mode always exits 0")
	assert.Contains(t, out.String(), "type Mixed struct")
}
