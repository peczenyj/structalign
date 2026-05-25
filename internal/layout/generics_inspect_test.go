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
	// The field stays source-faithful (Value T); the assumed substitution that
	// drives its size moves to the per-field marker.
	assert.Equal(t, "T", l.Fields[1].Type, "field type keeps the type-parameter name")
	assert.Equal(t, "T=any", l.Fields[1].Assume, "field carries the assumed substitution")
	assert.Empty(t, l.Fields[0].Assume, "non-generic field has no marker")
	assert.Empty(t, l.Fields[2].Assume, "non-generic field has no marker")
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
	assert.Equal(t, "T", got[0].Fields[1].Type, "field type keeps the type-parameter name")
	assert.Equal(t, "T=int", got[0].Fields[1].Assume, "~int constraint resolves to its core type int")
}

func TestLayoutsGenericOneParam(t *testing.T) {
	// One type parameter: the field that uses it carries its assumption, the
	// others stay clean.
	src := "package sample\n\ntype Foo[T any] struct {\n\tFlag  bool\n\tValue T\n\tCount uint32\n}\n"
	tgt := testutil.Target(t, src)

	got := layout.New().Layouts(tgt, common.Options{Patterns: []string{"Foo"}})
	require.Len(t, got, 1)
	assert.Equal(t, "[T]", got[0].TypeParams)
	require.Len(t, got[0].Fields, 3)
	assert.Empty(t, got[0].Fields[0].Assume, "bool field has no marker")
	assert.Equal(t, "T", got[0].Fields[1].Type)
	assert.Equal(t, "T=any", got[0].Fields[1].Assume)
	assert.Empty(t, got[0].Fields[2].Assume, "uint32 field has no marker")
}

func TestLayoutsGenericTwoParams(t *testing.T) {
	// Two type parameters used in separate fields: each field reports only the
	// parameter it actually references.
	src := "package sample\n\ntype Foo[K comparable, V any] struct {\n\tKey   K\n\tVal   V\n\tCount int\n}\n"
	tgt := testutil.Target(t, src)

	got := layout.New().Layouts(tgt, common.Options{Patterns: []string{"Foo"}})
	require.Len(t, got, 1)
	assert.Equal(t, "[K, V]", got[0].TypeParams)
	require.Len(t, got[0].Fields, 3)
	assert.Equal(t, "K", got[0].Fields[0].Type)
	assert.Equal(t, "K=any", got[0].Fields[0].Assume, "first field references only K")
	assert.Equal(t, "V", got[0].Fields[1].Type)
	assert.Equal(t, "V=any", got[0].Fields[1].Assume, "second field references only V")
	assert.Empty(t, got[0].Fields[2].Assume, "int field references neither")
}

func TestLayoutsGenericNestedGenericMember(t *testing.T) {
	// A member whose type is neither K nor V, but a generic type instantiated
	// with one of them (Inner[V]), still reports the parameter it depends on.
	src := "package sample\n\ntype Inner[T any] struct{ V T }\n" +
		"type Foo[K comparable, V any] struct {\n\tKey   K\n\tInner Inner[V]\n}\n"
	tgt := testutil.Target(t, src)

	got := layout.New().Layouts(tgt, common.Options{Patterns: []string{"Foo"}})
	require.Len(t, got, 1)
	require.Len(t, got[0].Fields, 2)
	assert.Equal(t, "K", got[0].Fields[0].Type)
	assert.Equal(t, "K=any", got[0].Fields[0].Assume)
	assert.Equal(t, "Inner[V]", got[0].Fields[1].Type, "nested generic rendered with its type argument")
	assert.Equal(t, "V=any", got[0].Fields[1].Assume, "depends on V through Inner[V], not K")
}

func TestLayoutsGenericMultiParamMarker(t *testing.T) {
	// A field referencing several type parameters lists them all, in
	// declaration order; a field referencing none carries no marker.
	src := "package sample\n\ntype Pair[K comparable, V any] struct {\n\tEntries map[K]V\n\tCount   int\n}\n"
	tgt := testutil.Target(t, src)

	got := layout.New().Layouts(tgt, common.Options{Patterns: []string{"Pair"}})
	require.Len(t, got, 1)
	assert.Equal(t, "[K, V]", got[0].TypeParams)
	assert.Equal(t, "map[K]V", got[0].Fields[0].Type)
	assert.Equal(t, "K=any, V=any", got[0].Fields[0].Assume, "both params listed in declaration order")
	assert.Empty(t, got[0].Fields[1].Assume, "non-generic field has no marker")
}
