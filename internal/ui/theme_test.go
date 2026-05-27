package ui_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/ui"
)

func TestThemeByNameKnown(t *testing.T) {
	for _, name := range []string{"default", "cga", "green", "amber"} {
		_, ok := ui.ThemeByName(name)
		assert.True(t, ok, "theme %q should exist", name)
	}
}

func TestThemeByNameUnknown(t *testing.T) {
	_, ok := ui.ThemeByName("nope")
	assert.False(t, ok)
}

// Rendering with a non-default theme set on the Printer must use that theme's
// codes (exercises Printer.theme()'s "theme is set" branch). cga's Label is
// bold bright white, which termenv renders as the combined "\x1b[1;97m".
func TestPrinterUsesSetTheme(t *testing.T) {
	cga, ok := ui.ThemeByName("cga")
	assert.True(t, ok)
	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: true, Theme: cga}
	p.RenderSummary(1, 8)
	assert.Contains(t, buf.String(), "\x1b[1;97m", "cga Label (bold bright white) should be used")
}
