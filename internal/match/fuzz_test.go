package match_test

import (
	"testing"

	"github.com/peczenyj/structalign/internal/match"
)

func FuzzMatchAny(f *testing.F) {
	f.Add("Mixed", "Mixed")
	f.Add("*", "Mixed")
	f.Add("", "Mixed")

	f.Fuzz(func(t *testing.T, pattern string, name string) {
		_ = match.MatchAny([]string{pattern}, name)
	})
}

func FuzzSplitCSV(f *testing.F) {
	f.Add("a,b,c")
	f.Add(`"a,b",c`)
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		_ = match.SplitCSV(s)
	})
}
