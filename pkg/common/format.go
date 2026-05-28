package common

//go:generate go tool enumer -type=Format -text -transform=lower -trimprefix=Format -pflag.value

// Format selects the output presentation style. enumer generates its
// String/parse/text-marshal helpers, a flag.Value implementation (Set), and a
// Type method for usage strings, see format_enumer.go; the names map to
// "text"/"json" via -trimprefix=Format -transform=lower.
type Format uint8

const (
	FormatText Format = iota
	FormatJSON
)
