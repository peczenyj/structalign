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

func TestFixScanStopsAtDoubleDash(t *testing.T) {
	var out, errb bytes.Buffer
	a := &app.App{Stdout: &out, Stderr: &errb}

	// "--" ends flag scanning, so anything after it is never matched as -fix;
	// here it leaves no packages, so Run falls through to the usage error.
	assert.Equal(t, 2, a.Run([]string{"--"}))
	assert.NotContains(t, errb.String(), "sorry", "easter egg not triggered past --")
	assert.Contains(t, errb.String(), "usage:")
}
