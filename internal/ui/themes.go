package ui

import "github.com/muesli/termenv"

// builtinThemes holds the named palettes. "default" mirrors DefaultTheme().
// The monochrome themes (green/amber) use a single hue and rely on intensity
// (bold vs faint) plus the +/- diff prefixes to distinguish added from removed.
var builtinThemes = map[string]Theme{
	"default": DefaultTheme(),
	// A high-intensity CGA-flavored palette. Magenta is reserved for the header
	// (bold, not reverse-video); the diff body uses cyan (added) and yellow
	// (removed) so it never echoes the header's magenta, while staying visibly
	// distinct from the default. (Strict CGA mode-4 palette 1 is cyan/magenta/
	// white; yellow is a deliberate borrow for a clearer added/removed contrast.)
	"cga": {
		Header:  Style{fg: termenv.ANSIBrightMagenta, bold: true},
		Added:   Style{fg: termenv.ANSIBrightCyan},
		Removed: Style{fg: termenv.ANSIBrightYellow},
		Meta:    Style{fg: termenv.ANSIBrightBlack}, // gray
		Padding: Style{fg: termenv.ANSIBrightWhite},
		Label:   Style{fg: termenv.ANSIBrightWhite, bold: true},
	},
	"green": {
		Header:  Style{fg: termenv.ANSIGreen, bold: true},
		Added:   Style{fg: termenv.ANSIGreen, bold: true},  // distinguished by the "+" prefix
		Removed: Style{fg: termenv.ANSIGreen, faint: true}, // distinguished by the "-" prefix
		Meta:    Style{fg: termenv.ANSIGreen, faint: true},
		Padding: Style{fg: termenv.ANSIGreen, faint: true},
		Label:   Style{fg: termenv.ANSIGreen, bold: true},
	},
	"amber": {
		Header:  Style{fg: termenv.ANSI256Color(214), bold: true},
		Added:   Style{fg: termenv.ANSI256Color(214), bold: true},
		Removed: Style{fg: termenv.ANSI256Color(214), faint: true},
		Meta:    Style{fg: termenv.ANSI256Color(214), faint: true},
		Padding: Style{fg: termenv.ANSI256Color(214), faint: true},
		Label:   Style{fg: termenv.ANSI256Color(214), bold: true},
	},
}

// ThemeByName returns a built-in theme and whether it was found.
func ThemeByName(name string) (Theme, bool) {
	th, ok := builtinThemes[name]
	return th, ok
}
