package ui

// builtinThemes holds the named palettes. "default" mirrors DefaultTheme().
// The monochrome themes (green/amber) use a single hue and rely on intensity
// (bold vs dim) plus the +/- diff prefixes to distinguish added from removed.
var builtinThemes = map[string]Theme{
	"default": DefaultTheme(),
	// The iconic CGA mode-4 palette 1 (high intensity): cyan, magenta, white on
	// black. Using magenta for removed/padding instead of the default's red makes
	// the palette visibly distinct from the default rather than a mere brightening.
	"cga": {
		Header:  "\x1b[1m\x1b[7m\x1b[95m", // bold + reverse-video magenta (a header bar)
		Added:   "\x1b[96m",               // bright cyan
		Removed: "\x1b[95m",               // bright magenta
		Meta:    "\x1b[90m",               // bright black (gray)
		Padding: "\x1b[95m",               // bright magenta
		Label:   "\x1b[1m\x1b[97m",        // bright white
	},
	"green": {
		Header:  "\x1b[1m\x1b[32m", // bold green
		Added:   "\x1b[1m\x1b[32m", // bold green (distinguished by the "+" prefix)
		Removed: "\x1b[2m\x1b[32m", // dim green (distinguished by the "-" prefix)
		Meta:    "\x1b[2m\x1b[32m", // dim green
		Padding: "\x1b[2m\x1b[32m", // dim green
		Label:   "\x1b[1m\x1b[32m", // bold green
	},
	"amber": {
		Header:  "\x1b[1m\x1b[38;5;214m", // bold amber (256-color)
		Added:   "\x1b[1m\x1b[38;5;214m",
		Removed: "\x1b[2m\x1b[38;5;214m", // dim amber
		Meta:    "\x1b[2m\x1b[38;5;214m",
		Padding: "\x1b[2m\x1b[38;5;214m",
		Label:   "\x1b[1m\x1b[38;5;214m",
	},
}

// ThemeByName returns a built-in theme and whether it was found.
func ThemeByName(name string) (Theme, bool) {
	th, ok := builtinThemes[name]
	return th, ok
}
