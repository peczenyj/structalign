package ui_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/ui"
	"github.com/peczenyj/structalign/pkg/common"
)

func TestWantColor(t *testing.T) {
	assert.True(t, ui.WantColor(common.ColorizeAlways, os.Stdout))
	assert.False(t, ui.WantColor(common.ColorizeNever, os.Stdout))

	// common.ColorizeAuto against a pipe (not a terminal) is off.
	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { r.Close(); w.Close() })
	assert.False(t, ui.WantColor(common.ColorizeAuto, w))
}

func TestWantColorStatErrorIsNotATerminal(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	r.Close()
	w.Close()
	// Stat on a closed file errors; auto treats that as "not a terminal".
	assert.False(t, ui.WantColor(common.ColorizeAuto, w))
}

func TestWantColorHonorsNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// -color=always still wins over NO_COLOR (the convention defers to an
	// explicit color flag); auto stays off.
	assert.True(t, ui.WantColor(common.ColorizeAlways, os.Stdout), "explicit -color=always overrides NO_COLOR")

	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { r.Close(); w.Close() })
	assert.False(t, ui.WantColor(common.ColorizeAuto, w), "auto + NO_COLOR is off")
}

func TestResolveWidth(t *testing.T) {
	// A pipe is not a terminal, so ResolveWidth falls back to $COLUMNS then 80.
	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { r.Close(); w.Close() })

	t.Setenv("COLUMNS", "")
	assert.Equal(t, 80, ui.ResolveWidth(w), "no terminal and no COLUMNS -> fallback 80")

	t.Setenv("COLUMNS", "120")
	assert.Equal(t, 57, ui.ResolveWidth(w), "(120-5)/2 = 57")
}
