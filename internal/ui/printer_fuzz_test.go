package ui

import "testing"

func FuzzTruncPad(f *testing.F) {
	f.Add("Hello", 10)
	f.Add("世界", 10)
	f.Add("世", 1)
	f.Add("", 0)
	f.Add("VeryLongString", 5)

	f.Fuzz(func(t *testing.T, s string, w int) {
		_ = truncPad(s, w)
	})
}
