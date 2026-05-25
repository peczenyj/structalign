package textdiff_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/textdiff"
)

func TestLinesReconstructs(t *testing.T) {
	a := []string{"struct {", "\tA bool", "\tB int64", "}"}
	b := []string{"struct {", "\tB int64", "\tA bool", "}"}

	ops := textdiff.Lines(a, b)

	var gotA, gotB []string
	for _, op := range ops {
		switch op.Kind {
		case textdiff.Equal:
			gotA = append(gotA, op.Text)
			gotB = append(gotB, op.Text)
		case textdiff.Del:
			gotA = append(gotA, op.Text)
		case textdiff.Add:
			gotB = append(gotB, op.Text)
		}
	}
	assert.Equal(t, a, gotA, "equal+del should reconstruct the original")
	assert.Equal(t, b, gotB, "equal+add should reconstruct the proposed")
}
