package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/pkg/common"
)

func TestSavings(t *testing.T) {
	assert.Equal(t, int64(8), savings(common.Finding{OldSize: 24, NewSize: 16}))
	assert.Equal(t, int64(0), savings(common.Finding{OldSize: 16, NewSize: 24}))
	assert.Equal(t, int64(0), savings(common.Finding{OldSize: 0, NewSize: 16}))
}

func TestCmp(t *testing.T) {
	assert.Equal(t, -1, cmp(10, 20))
	assert.Equal(t, 1, cmp(20, 10))
	assert.Equal(t, 0, cmp(10, 10))
}
