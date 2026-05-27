package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The default theme must render the same visual output as the historical
// hand-rolled SGR palette. termenv emits combined sequences (e.g. "\x1b[1;36m"
// for bold cyan rather than "\x1b[1m\x1b[36m"), which are visually identical;
// the single-attribute roles are byte-for-byte unchanged.
func TestDefaultThemeRendersHistoricalVisuals(t *testing.T) {
	th := DefaultTheme()
	cases := []struct {
		name  string
		style Style
		want  string
	}{
		{"header bold cyan", th.Header, "\x1b[1;36mX\x1b[0m"},
		{"added green", th.Added, "\x1b[32mX\x1b[0m"},
		{"removed red", th.Removed, "\x1b[31mX\x1b[0m"},
		{"meta dim", th.Meta, "\x1b[2mX\x1b[0m"},
		{"padding red", th.Padding, "\x1b[31mX\x1b[0m"},
		{"label bold", th.Label, "\x1b[1mX\x1b[0m"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, c.style.render("X"))
		})
	}
}

// A zero Style adds no escapes.
func TestZeroStyleRendersPlain(t *testing.T) {
	assert.Equal(t, "X", Style{}.render("X"))
}

// CGA must be a visibly distinct palette, not a brightened default. Magenta (95)
// is reserved for the header bar; the diff body uses cyan (added) and yellow
// (removed), never the default's red (31) and never echoing the header's magenta.
func TestCgaThemeIsDistinctFromDefault(t *testing.T) {
	def := DefaultTheme()
	cga := builtinThemes["cga"]
	assert.NotEqual(t, def.Header.render("X"), cga.Header.render("X"), "cga header must differ from default")
	assert.Contains(t, cga.Header.render("X"), "95", "cga header is magenta")
	assert.Contains(t, cga.Added.render("X"), "96", "cga added is bright cyan")
	assert.Contains(t, cga.Removed.render("X"), "93", "cga removed is bright yellow")
	assert.NotContains(t, cga.Removed.render("X"), "31", "cga must not reuse the default red")
	assert.NotContains(t, cga.Removed.render("X"), "95", "magenta is reserved for the header bar")
	// Inspect padding shares the removed yellow rather than flat white.
	assert.Contains(t, cga.Padding.render("X"), "93", "cga inspect padding is bright yellow, matching removed")
}

// The green (P1 phosphor) theme is monochrome: it must never use red (31),
// since add/removed are distinguished by intensity + the +/- prefixes.
func TestGreenThemeIsMonochrome(t *testing.T) {
	green := builtinThemes["green"]
	for _, st := range []Style{green.Header, green.Added, green.Removed, green.Meta, green.Padding, green.Label} {
		seq := st.render("X")
		assert.NotContains(t, seq, "31", "green theme must not use red")
		assert.True(t, strings.Contains(seq, "32"), "green theme uses the green family: %q", seq)
	}
}
