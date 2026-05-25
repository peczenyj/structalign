package ui_test

import (
	"bytes"
	"strings"
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
		Fields:     []common.LayoutField{{Name: "V", Type: "T", Assume: "T=any", Size: 16, Align: 8}},
	}}, false, false)

	assert.Equal(t, 1, n)
	out := buf.String()
	assert.Contains(t, out, "// generic type — layout assumes T=any", "disclaimer printed above the struct")
	assert.Contains(t, out, "type Box[T] struct {", "header includes type parameters")
	assert.Contains(t, out, "V T", "field stays source-faithful (not substituted)")
	assert.Contains(t, out, "-- assume T=any", "per-field assumption marker")
}

func TestRenderLayoutAlignsAssumeMarker(t *testing.T) {
	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: false}
	p.RenderLayouts([]common.Layout{{
		Name:       "Generic",
		TypeParams: "[T]",
		Note:       "generic type",
		Total:      32,
		Align:      8,
		Padding:    11,
		Fields: []common.LayoutField{
			{Name: "Flag", Type: "bool", Size: 1, Align: 1, Padding: 7},
			{Name: "Value", Type: "T", Assume: "T=any", Size: 16, Align: 8},
			{Name: "Count", Type: "uint32", Size: 4, Align: 4, Padding: 4},
		},
	}}, false, false)

	out := buf.String()
	// The marker is pushed past the widest comment, so the "--" sits to the
	// right of even the longest "..., padding: N" line.
	markerCol := strings.Index(out, "-- assume")
	lineStart := strings.LastIndex(out[:markerCol], "\n") + 1
	flagComment := strings.Index(out, "padding: 7")
	assert.Greater(t, markerCol-lineStart, flagComment-strings.LastIndex(out[:flagComment], "\n")-1,
		"marker aligns past the widest comment")
}
