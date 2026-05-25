package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/pkg/common"
)

func TestWantColorDecision(t *testing.T) {
	cases := []struct {
		name           string
		mode           common.Colorize
		noColor, isTTY bool
		want           bool
	}{
		{"always forces on, even with NO_COLOR", common.ColorizeAlways, true, false, true},
		{"always forces on without a tty", common.ColorizeAlways, false, false, true},
		{"never forces off", common.ColorizeNever, false, true, false},
		{"never forces off even with NO_COLOR", common.ColorizeNever, true, true, false},
		{"auto on a tty without NO_COLOR colors", common.ColorizeAuto, false, true, true},
		{"auto on a tty with NO_COLOR is suppressed", common.ColorizeAuto, true, true, false},
		{"auto off a tty never colors", common.ColorizeAuto, false, false, false},
		{"unknown mode behaves like auto", common.Colorize(64), false, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, wantColor(c.mode, c.noColor, c.isTTY))
		})
	}
}
