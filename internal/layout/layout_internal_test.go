package layout

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/pkg/common"
)

func TestBuildLayoutInternal(t *testing.T) {
	i := &Inspector{}
	// Test st == nil branch: use a named int
	tn := types.NewTypeName(token.NoPos, nil, "Foo", types.Typ[types.Int])
	_, ok := i.buildLayout(common.Target{}, "Foo", tn, common.Options{})
	assert.False(t, ok)
}
