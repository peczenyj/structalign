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

	res2 := truncPad("short", 10)
	assert.Equal(t, "short     ", res2)
	assert.Equal(t, 10, len(res2))
}
