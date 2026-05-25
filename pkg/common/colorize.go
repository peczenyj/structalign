package common

//go:generate go tool enumer -type=Colorize -text -transform=lower -trimprefix=Colorize -pflag.value

// Colorize selects when colored output is emitted. enumer generates its
// String/parse/text-marshal helpers, a flag.Value implementation (Set), and a
// Type method for usage strings, see colorize_enumer.go; the names map to
// "auto"/"always"/"never" via -trimprefix=Colorize -transform=lower.
type Colorize uint8

const (
	ColorizeAuto   Colorize = iota // color only on a TTY with NO_COLOR unset
	ColorizeAlways                 // always color, overriding NO_COLOR
	ColorizeNever                  // never color
)
