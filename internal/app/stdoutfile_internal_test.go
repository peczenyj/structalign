package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stdoutFile returns the *os.File behind w (the production path, where Stdout is
// os.Stdout) or nil for any other writer (e.g. a test buffer), so the terminal
// queries in WantColor/ResolveWidth can fall back to safe defaults.
func TestStdoutFile(t *testing.T) {
	assert.Same(t, os.Stdout, stdoutFile(os.Stdout), "a real *os.File is returned as-is")
	assert.Nil(t, stdoutFile(&bytes.Buffer{}), "a non-file writer yields nil")
}
