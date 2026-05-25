package app_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/app"
)

func TestRunInvalidExcludeRegexp(t *testing.T) {
	var out, errb bytes.Buffer
	a := &app.App{Stdout: &out, Stderr: &errb}
	// A malformed -exclude regexp fails fast (before loading), exit 2.
	code := a.Run([]string{"-exclude=[", "pkg"})
	assert.Equal(t, 2, code)
	assert.Contains(t, errb.String(), "invalid -exclude")
}
