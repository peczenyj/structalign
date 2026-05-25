package app_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/app"
	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/mocks"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/pkg/common"
)

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
	mockLoader := mocks.NewLoader(t)
	mockLoader.EXPECT().Load(mock.Anything).Return([]common.Target{tgt}, nil)
	a := &app.App{
		Loader:    mockLoader,
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
	mockLoader := mocks.NewLoader(t)
	mockLoader.EXPECT().Load(mock.Anything).Return([]common.Target{tgt}, nil)
	a := &app.App{
		Loader:    mockLoader,
		Aligner:   align.New(),
		Inspector: layout.New(),
		Stdout:    &out,
		Stderr:    &errb,
	}
	code := a.Run([]string{"-inspect", "-type=Mixed", "pkg"})
	assert.Equal(t, 0, code, "inspect mode always exits 0")
	assert.Contains(t, out.String(), "type Mixed struct")
}

func TestRunExcludeSkipsMatchingPackages(t *testing.T) {
	tgt := testutil.Target(t, src) // PkgPath is "sample"
	var out, errb bytes.Buffer
	mockLoader := mocks.NewLoader(t)
	mockLoader.EXPECT().Load(mock.Anything).Return([]common.Target{tgt}, nil)
	a := &app.App{
		Loader:    mockLoader,
		Aligner:   align.New(),
		Inspector: layout.New(),
		Stdout:    &out,
		Stderr:    &errb,
	}
	// -exclude=^sample$ matches the only package, so nothing is analyzed.
	code := a.Run([]string{"-exclude=^sample$", "-type=Mixed", "pkg"})
	assert.Equal(t, 0, code, "all packages excluded → exit 0")
	assert.Empty(t, out.String(), "no diff output expected when all packages are excluded")
	assert.Contains(t, errb.String(), "no struct reorderings found")
}
