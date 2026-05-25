package app_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/app"
)

func TestFixEasterEgg(t *testing.T) {
	var out, errb bytes.Buffer
	a := &app.App{Stdout: &out, Stderr: &errb}

	assert.Equal(t, 2, a.Run([]string{"-fix", "./..."}), "-fix is a usage error")
	assert.Contains(t, errb.String(), "fieldalignment", "redirects to fieldalignment")
	assert.Empty(t, out.String(), "nothing on stdout")
}

func TestFixNotAdvertisedInUsage(t *testing.T) {
	var out, errb bytes.Buffer
	a := &app.App{Stdout: &out, Stderr: &errb}

	a.Run(nil) // no args -> usage dump
	assert.NotContains(t, errb.String(), "-fix", "the easter egg stays out of -help")
}
