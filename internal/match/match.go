// Package match implements the -type glob filtering: parsing the comma-separated
// pattern list and testing a type name against it.
package match

import (
	"path"
	"strings"
)

// SplitCSV splits a comma-separated list, trimming spaces and dropping empties.
func SplitCSV(s string) []string {
	var out []string
	for p := range strings.SplitSeq(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ParsePatterns splits a comma-separated value into trimmed, non-empty glob
// patterns. Empty input yields nil (meaning "match everything").
func ParsePatterns(s string) []string {
	return SplitCSV(s)
}

// MatchAny reports whether name matches any glob pattern (path.Match syntax).
// An empty name (anonymous struct) never matches a non-empty pattern set;
// invalid patterns are treated as non-matching rather than fatal.
func MatchAny(patterns []string, name string) bool {
	if name == "" {
		return false
	}
	for _, p := range patterns {
		if ok, err := path.Match(p, name); err == nil && ok {
			return true
		}
	}
	return false
}
