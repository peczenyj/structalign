package layout_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/pkg/common"
)

func TestLayoutsInstantiatesGenericAny(t *testing.T) {
	src := "package sample\n\ntype Box[T any] struct {\n\tFlag bool\n\tValue T\n\tCount uint32\n}\n"
	tgt := testutil.Target(t, src)

	got := layout.New().Layouts(tgt, common.Options{Patterns: []string{"Box"}})
	require.Len(t, got, 1)

	l := got[0]
	assert.Equal(t, "Box", l.Name)
	assert.Equal(t, "[T]", l.TypeParams)
	assert.NotEmpty(t, l.Note, "generic layout carries a disclaimer")
	assert.Equal(t, int64(32), l.Total, "T=any (16-byte interface) on amd64")
	require.Len(t, l.Fields, 3)
	assert.Equal(t, "any", l.Fields[1].Type, "type parameter substituted with its representative")
}

func TestLayoutsSkipsNonStructTypes(t *testing.T) {
	// A named non-struct, and a generic non-struct, are both skipped; only the
	// struct is laid out.
	src := "package sample\n\ntype S struct{ A int }\ntype MyInt int\ntype Slice[T any] []T\n"
	tgt := testutil.Target(t, src)

	got := layout.New().Layouts(tgt, common.Options{})
	require.Len(t, got, 1)
	assert.Equal(t, "S", got[0].Name)
}

func TestLayoutsGenericCoreType(t *testing.T) {
	src := "package sample\n\ntype Num[T ~int] struct {\n\tA bool\n\tV T\n}\n"
	tgt := testutil.Target(t, src)

	got := layout.New().Layouts(tgt, common.Options{Patterns: []string{"Num"}})
	require.Len(t, got, 1)
	assert.Equal(t, "int", got[0].Fields[1].Type, "~int constraint resolves to its core type int")
}
