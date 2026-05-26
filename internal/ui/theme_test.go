package ui_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/ui"
)

// The default theme must reproduce the historical hardcoded ANSI palette, so
// existing golden output is byte-for-byte unchanged.
func TestDefaultThemeMatchesLegacyConstants(t *testing.T) {
	th := ui.DefaultTheme()
	assert.Equal(t, "\x1b[1m\x1b[36m", th.Header, "header was bold+cyan")
	assert.Equal(t, "\x1b[32m", th.Added, "added was green")
	assert.Equal(t, "\x1b[31m", th.Removed, "removed was red")
	assert.Equal(t, "\x1b[2m", th.Meta, "meta was dim")
	assert.Equal(t, "\x1b[31m", th.Padding, "padding was red")
	assert.Equal(t, "\x1b[1m", th.Label, "label was bold")
}

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
// codes (exercises Printer.theme()'s "theme is set" branch).
func TestPrinterUsesSetTheme(t *testing.T) {
	cga, ok := ui.ThemeByName("cga")
	assert.True(t, ok)
	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: true, Theme: cga}
	p.RenderSummary(1, 8)
	assert.Contains(t, buf.String(), "\x1b[1m\x1b[97m", "cga Label (bold bright white) should be used")
}

// The green (P1 phosphor) theme is monochrome: it must never use red (31),
// since add/removed are distinguished by intensity + the +/- prefixes.
func TestGreenThemeIsMonochrome(t *testing.T) {
	th, ok := ui.ThemeByName("green")
	assert.True(t, ok)
	for _, sgr := range []string{th.Header, th.Added, th.Removed, th.Meta, th.Padding, th.Label} {
		assert.NotContains(t, sgr, "31", "green theme must not use red")
		assert.Contains(t, sgr, "32", "green theme uses the green family")
	}
}
