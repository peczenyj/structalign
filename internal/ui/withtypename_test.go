package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTypeName(t *testing.T) {
	const body = "struct {\n\tA bool\n}"

	assert.Equal(t, body, withTypeName(body, ""), "empty name is a no-op")
	assert.Equal(t, "x int", withTypeName("x int", "Foo"), "non-struct text is a no-op")
	assert.Equal(t, "type Foo "+body, withTypeName(body, "Foo"), "named struct gets the type prefix")
	assert.Equal(t, "type Foo struct {", withTypeName("struct {", "Foo"), "single line with no remainder")
}
