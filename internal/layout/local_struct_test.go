package layout_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/pkg/common"
)

func TestLayoutsFindsLocalStructs(t *testing.T) {
	const src = `package sample
func Foo() {
	type Local struct {
		A bool
		B int64
	}
	_ = Local{}
}
`
	tgt := testutil.Target(t, src)
	ls := layout.New().Layouts(tgt, common.Options{})

	found := false
	for _, l := range ls {
		if l.Name == "Local" {
			found = true
		}
	}
	assert.True(t, found, "Local struct inside function should be found in -inspect mode")
}
