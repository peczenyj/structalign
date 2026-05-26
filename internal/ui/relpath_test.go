package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelPath(t *testing.T) {
	// Relative path remains unchanged
	assert.Equal(t, "foo.go", relPath("foo.go"))

	// Path outside current working directory returns absolute path (or original)
	// /tmp/foo.go is likely outside this repo's wd
	assert.Equal(t, "/tmp/foo.go", relPath("/tmp/foo.go"))
}
