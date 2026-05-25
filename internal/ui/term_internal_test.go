package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWantColorDecision(t *testing.T) {
	cases := []struct {
		name           string
		mode           string
		noColor, isTTY bool
		want           bool
	}{
		{"always forces on, even with NO_COLOR", "always", true, false, true},
		{"always forces on without a tty", "always", false, false, true},
		{"never forces off", "never", false, true, false},
		{"never forces off even with NO_COLOR", "never", true, true, false},
		{"auto on a tty without NO_COLOR colors", "auto", false, true, true},
		{"auto on a tty with NO_COLOR is suppressed", "auto", true, true, false},
		{"auto off a tty never colors", "auto", false, false, false},
		{"unknown mode behaves like auto", "weird", false, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, wantColor(c.mode, c.noColor, c.isTTY))
		})
	}
}
