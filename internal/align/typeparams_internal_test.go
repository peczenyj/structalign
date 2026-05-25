package align

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/testutil"
)

func TestTypeParamNames(t *testing.T) {
	src := "package sample\n\n" +
		"type Generic[T any] struct{ V T }\n" +
		"type Pair[K comparable, V any] struct{ K K; V V }\n" +
		"type Plain struct{ A int }\n" +
		"type Alias = int\n"
	tgt := testutil.Target(t, src)

	assert.Equal(t, "[T]", typeParamNames(tgt, "Generic"), "single type parameter")
	assert.Equal(t, "[K, V]", typeParamNames(tgt, "Pair"), "multiple type parameters")
	assert.Empty(t, typeParamNames(tgt, "Plain"), "non-generic named type")
	assert.Empty(t, typeParamNames(tgt, ""), "anonymous (no name)")
	assert.Empty(t, typeParamNames(tgt, "Missing"), "name not in scope")
	assert.Empty(t, typeParamNames(tgt, "Alias"), "alias resolves to a non-Named type")
}
