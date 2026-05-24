package match_test

import (
	"testing"

	"github.com/peczenyj/structalign/internal/match"
)

func TestParsePatterns(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"Config", []string{"Config"}},
		{"*Request,Config", []string{"*Request", "Config"}},
		{" A , B ,", []string{"A", "B"}},
	}
	for _, tc := range tests {
		got := match.ParsePatterns(tc.in)
		if len(got) != len(tc.want) {
			t.Errorf("ParsePatterns(%q) = %v, want %v", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("ParsePatterns(%q) = %v, want %v", tc.in, got, tc.want)
				break
			}
		}
	}
}

func TestMatchAny(t *testing.T) {
	tests := []struct {
		patterns []string
		name     string
		want     bool
	}{
		{[]string{"Config"}, "Config", true},
		{[]string{"*Request"}, "HTTPRequest", true},
		{[]string{"*Request"}, "Requester", false},
		{[]string{"*"}, "", false},
		{nil, "Anything", false},
		{[]string{"["}, "[", false},
	}
	for _, tc := range tests {
		if got := match.MatchAny(tc.patterns, tc.name); got != tc.want {
			t.Errorf("MatchAny(%v, %q) = %v, want %v", tc.patterns, tc.name, got, tc.want)
		}
	}
}
