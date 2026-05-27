package align

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNolintInternal(t *testing.T) {
	var info nolintInfo
	// Case default: not a comment
	parseNolint("not a comment", &info)
	assert.False(t, info.bare)
}

// Block comments may be single-line (/* nolint */) or, in their canonical
// gofmt'd multiline form, carry a leading "*" on each continuation line
// (/*\n * nolint\n */). parseNolint must strip those per-line stars so the
// directive is recognized in either shape.
func TestParseNolintBlockComment(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		bare  bool
		token string // "" means no named token expected
	}{
		{"single-line bare", "/* nolint */", true, ""},
		{"single-line with token", "/* nolint:fieldalignment */", false, "fieldalignment"},
		{"multiline star-prefixed bare", "/*\n * nolint\n */", true, ""},
		{"multiline star-prefixed token", "/*\n * nolint:fieldalignment\n */", false, "fieldalignment"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var info nolintInfo
			parseNolint(c.text, &info)
			assert.Equal(t, c.bare, info.bare, "bare")
			if c.token == "" {
				assert.Empty(t, info.tokens, "no named tokens expected")
			} else {
				_, ok := info.tokens[c.token]
				assert.True(t, ok, "expected token %q in %v", c.token, info.tokens)
			}
		})
	}
}
