package app_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/app"
)

// usageText runs with no args (so Run falls through to the usage dump) and
// returns what landed on stderr.
func usageText(t *testing.T) string {
	t.Helper()
	var out, errb bytes.Buffer
	a := &app.App{Stdout: &out, Stderr: &errb}
	a.Run(nil)
	assert.Empty(t, out.String(), "usage goes to stderr, not stdout")
	return errb.String()
}

func TestUsageHasManPageSections(t *testing.T) {
	u := usageText(t)

	// tagline, usage line, examples block, and the flags header are all present.
	assert.Contains(t, u, "show how a struct's fields could be reordered to use less memory",
		"tagline describes what the tool does")
	assert.Contains(t, u, "usage: structalign [flags] [packages]")
	assert.Contains(t, u, "examples:", "an examples section guides a new user")
	assert.Contains(t, u, "structalign -inspect -type=Config", "an example shows inspect mode")
	assert.Contains(t, u, "flags:", "the flag list is introduced by a header")

	// PrintDefaults still emits the real flags after the header.
	assert.Contains(t, u, "-diff", "the flag list is still printed")
	assert.Contains(t, u, "-inspect")
	assert.Contains(t, u, "-no-rc", "the -no-rc config flag is discoverable in -h")
}

func TestUsageOmitsEasterEggs(t *testing.T) {
	u := usageText(t)
	assert.NotContains(t, u, "-fix", "the -fix egg stays out of help")
	assert.NotContains(t, u, "-cga", "theme eggs stay out of help")
	assert.NotContains(t, u, "-green")
	assert.NotContains(t, u, "-amber")
}
