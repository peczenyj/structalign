package app_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/app"
)

func TestNew(t *testing.T) {
	a := app.New(os.Stdout, os.Stderr)
	require.NotNil(t, a.Loader)
	require.NotNil(t, a.Aligner)
	require.NotNil(t, a.Inspector)
}

func TestRunNoArgsExitsTwo(t *testing.T) {
	var out, errb bytes.Buffer
	a := &app.App{Stdout: &out, Stderr: &errb}
	assert.Equal(t, 2, a.Run(nil), "no packages -> usage error, exit 2")
	assert.Contains(t, errb.String(), "usage:")
}
