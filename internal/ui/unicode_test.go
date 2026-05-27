package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncPadUnicode(t *testing.T) {
	// "世" is 3 bytes but 1 rune.
	s := "Hello 世界" // 8 runes
	res := truncPad(s, 7)
	assert.Equal(t, "Hello …", res)
	assert.Equal(t, 7, len([]rune(res)))

	assert.Equal(t, "…", truncPad(s, 1))
	assert.Equal(t, "", truncPad(s, 0))

	res2 := truncPad("short", 10)
	assert.Equal(t, "short     ", res2)
	assert.Equal(t, 10, len(res2))
}

// truncPad must measure display width, not rune count: CJK ideographs are two
// cells wide. A rune-count implementation would mis-pad and overflow the
// column; these cases lock in the width-aware behavior.
func TestTruncPadWideRunes(t *testing.T) {
	// "世界" is 2 runes but 4 display cells. No truncation needed at width 6,
	// so it pads by the remaining 2 cells (not 4 as a rune count would give).
	assert.Equal(t, "世界  ", truncPad("世界", 6))

	// "世界世" is 6 cells; at width 5 it truncates after the second wide rune,
	// landing exactly on the column with the ellipsis ("世界" = 4 + "…" = 5).
	assert.Equal(t, "世界…", truncPad("世界世", 5))
}
