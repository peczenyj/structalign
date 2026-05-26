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

func themeApp(t *testing.T, out, errb *bytes.Buffer) *app.App {
	t.Helper()
	tgt := testutil.Target(t, src)
	ml := mocks.NewLoader(t)
	ml.EXPECT().Load(mock.Anything).Return([]common.Target{tgt}, nil)
	return &app.App{Loader: ml, Aligner: align.New(), Inspector: layout.New(), Stdout: out, Stderr: errb}
}

func TestRunCgaEggAccepted(t *testing.T) {
	var out, errb bytes.Buffer
	a := themeApp(t, &out, &errb)
	// The egg flag is stripped before flag parsing, so the run proceeds normally
	// (exit 1 on findings) rather than failing with "flag provided but not defined".
	code := a.Run([]string{"-cga", "-type=Mixed", "pkg"})
	assert.Equal(t, 1, code)
	assert.NotContains(t, errb.String(), "not defined")
}

func TestRunGreenAndAmberEggsAccepted(t *testing.T) {
	for _, egg := range []string{"-green", "-amber"} {
		var out, errb bytes.Buffer
		a := themeApp(t, &out, &errb)
		code := a.Run([]string{egg, "-type=Mixed", "pkg"})
		assert.Equal(t, 1, code, "%s should be stripped and the run proceed", egg)
		assert.NotContains(t, errb.String(), "not defined")
	}
}

func TestRunValidThemeEnvNoWarning(t *testing.T) {
	var out, errb bytes.Buffer
	a := themeApp(t, &out, &errb)
	t.Setenv("STRUCTALIGN_THEME", "green")
	a.Run([]string{"-type=Mixed", "pkg"})
	assert.NotContains(t, errb.String(), "unknown theme", "a valid theme must not warn")
}

func TestRunDoubleDashStopsEggStripping(t *testing.T) {
	var out, errb bytes.Buffer
	a := themeApp(t, &out, &errb)
	// After "--", args are positional (package) args; the loop's afterDD branch
	// passes them through untouched.
	code := a.Run([]string{"-type=Mixed", "--", "pkg"})
	assert.Equal(t, 1, code)
}

func TestRunUnknownThemeEnvWarns(t *testing.T) {
	var out, errb bytes.Buffer
	a := themeApp(t, &out, &errb)
	t.Setenv("STRUCTALIGN_THEME", "bogus")
	a.Run([]string{"-type=Mixed", "pkg"})
	assert.Contains(t, errb.String(), `unknown theme "bogus"`)
}

func TestEggFlagNotInUsage(t *testing.T) {
	var out, errb bytes.Buffer
	a := &app.App{Stdout: &out, Stderr: &errb}
	a.Run(nil) // no args => prints usage to stderr
	assert.NotContains(t, errb.String(), "-cga", "easter-egg flags stay invisible in usage")
	assert.NotContains(t, errb.String(), "-green")
	assert.NotContains(t, errb.String(), "-amber")
}
