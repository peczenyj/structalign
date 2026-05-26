package ui_test

import (
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
