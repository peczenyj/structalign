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
