package ui

import (
	"os"
	"strconv"

	"golang.org/x/term"

	"github.com/peczenyj/structalign/pkg/common"
)

// WantColor resolves the -color mode ("auto"|"always"|"never") against out and
// the environment. In "auto" mode it honors NO_COLOR (https://no-color.org); an
// explicit -color=always still wins, per that convention.
func WantColor(colorize common.Colorize, out *os.File) bool {
	return wantColor(colorize, os.Getenv("NO_COLOR") != "", isCharDevice(out))
}

// wantColor is the pure decision: "always" forces color on, "never" forces it
// off, and "auto" emits color only on a terminal and only when NO_COLOR is unset.
func wantColor(colorize common.Colorize, noColor, isTTY bool) bool {
	switch colorize {
	case common.ColorizeAlways:
		return true
	case common.ColorizeNever:
		return false
	default:
		return isTTY && !noColor
	}
}

func isCharDevice(out *os.File) bool {
	fi, err := out.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
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
