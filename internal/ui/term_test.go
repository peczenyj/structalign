package ui_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/ui"
)

func TestWantColor(t *testing.T) {
	assert.True(t, ui.WantColor("always", os.Stdout))
	assert.False(t, ui.WantColor("never", os.Stdout))

	// "auto" against a pipe (not a terminal) is off.
	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { r.Close(); w.Close() })
	assert.False(t, ui.WantColor("auto", w))
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
