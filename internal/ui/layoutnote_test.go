package ui_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/ui"
	"github.com/peczenyj/structalign/pkg/common"
)

func TestRenderLayoutNoteAndTypeParams(t *testing.T) {
	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: false}
	n := p.RenderLayouts([]common.Layout{{
		Name:       "Box",
		TypeParams: "[T]",
		Note:       "generic type — layout assumes T=any",
		Total:      16,
		Align:      8,
		Fields:     []common.LayoutField{{Name: "V", Type: "any", Size: 16, Align: 8}},
	}}, false, false)

	assert.Equal(t, 1, n)
	out := buf.String()
	assert.Contains(t, out, "// generic type — layout assumes T=any", "disclaimer printed above the struct")
	assert.Contains(t, out, "type Box[T] struct {", "header includes type parameters")
}
