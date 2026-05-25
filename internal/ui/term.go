package ui

import (
	"os"
	"strconv"

	"golang.org/x/term"
)

// WantColor resolves the -color mode ("auto"|"always"|"never") against out.
func WantColor(mode string, out *os.File) bool {
	switch mode {
	case "always":
		return true
	case "never":
		return false
	default:
		fi, err := out.Stat()
		if err != nil {
			return false
		}
		return fi.Mode()&os.ModeCharDevice != 0
	}
}

// ResolveWidth returns the default per-side column width for side-by-side diff,
// derived from the terminal attached to out (falling back to $COLUMNS, then 80).
func ResolveWidth(out *os.File) int {
	const (
		overhead    = 5
		fallback    = 80
		minWidth    = 20
		minTermCols = overhead + 2*minWidth
	)
	fromCols := func(cols int) (int, bool) {
		if cols < minTermCols {
			return 0, false
		}
		return (cols - overhead) / 2, true
	}
	if cols, _, err := term.GetSize(int(out.Fd())); err == nil {
		if w, ok := fromCols(cols); ok {
			return w
		}
		if cols >= overhead+2 {
			return minWidth
		}
	}
	if c, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil {
		if w, ok := fromCols(c); ok {
			return w
		}
	}
	return fallback
}
