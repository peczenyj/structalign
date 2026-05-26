package app_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/app"
	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/mocks"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/pkg/common"
)

const appNolintSrc = `package sample

//nolint:fieldalignment
type Hidden struct {
	A bool
	B int64
	C bool
}
`

func nolintApp(t *testing.T, out, errb *bytes.Buffer) *app.App {
	t.Helper()
	tgt := testutil.Target(t, appNolintSrc)
	ml := mocks.NewLoader(t)
	ml.EXPECT().Load(mock.Anything).Return([]common.Target{tgt}, nil)
	return &app.App{Loader: ml, Aligner: align.New(), Inspector: layout.New(), Stdout: out, Stderr: errb}
}

func TestRunHidesNolintByDefault(t *testing.T) {
	var out, errb bytes.Buffer
	a := nolintApp(t, &out, &errb)
	code := a.Run([]string{"pkg"})
	assert.Equal(t, 0, code, "only struct is suppressed => no findings => exit 0")
	assert.NotContains(t, out.String(), "Hidden")
}

func TestRunShowNolintReveals(t *testing.T) {
	var out, errb bytes.Buffer
	a := nolintApp(t, &out, &errb)
	code := a.Run([]string{"-show-nolint", "pkg"})
	assert.Equal(t, 1, code)
	assert.Contains(t, out.String(), "Hidden")
}

func TestRunNolintLintersOptOut(t *testing.T) {
	var out, errb bytes.Buffer
	a := nolintApp(t, &out, &errb)
	// Honor only betteralign => :fieldalignment no longer suppresses Hidden.
	code := a.Run([]string{"-nolint-linters=betteralign", "pkg"})
	assert.Equal(t, 1, code)
	assert.Contains(t, out.String(), "Hidden")
}

func TestRunInspectIgnoresNolint(t *testing.T) {
	var out, errb bytes.Buffer
	a := nolintApp(t, &out, &errb)
	_ = a.Run([]string{"-inspect", "pkg"})
	assert.Contains(t, out.String(), "type Hidden struct", "inspect ignores //nolint")
}
