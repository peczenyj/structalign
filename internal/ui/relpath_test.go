package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelPath(t *testing.T) {
	// Relative path remains unchanged
	assert.Equal(t, "foo.go", relPath("foo.go"))

	// Path outside current working directory returns absolute path (or original)
	wd, err := os.Getwd()
	require.NoError(t, err)
	parent := filepath.Dir(wd)
	outsideFile := filepath.Join(parent, "some_outside_file.go")
	assert.Equal(t, outsideFile, relPath(outsideFile))
}
