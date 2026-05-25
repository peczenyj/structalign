package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveVersion(t *testing.T) {
	orig := version
	t.Cleanup(func() { version = orig })

	// A GoReleaser-stamped ldflags version is used verbatim.
	version = "v1.2.3"
	assert.Equal(t, "v1.2.3", resolveVersion())

	// Without a stamp, fall back to build info (never the bare "(devel)" sentinel).
	version = "dev"
	got := resolveVersion()
	assert.NotEmpty(t, got)
	assert.NotEqual(t, "(devel)", got)
}
