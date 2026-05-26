package ui_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peczenyj/structalign/internal/ui"
)

func TestRenderSummaryPlural(t *testing.T) {
	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: false}
	p.RenderSummary(3, 40)
	assert.Equal(t, "Summary: 3 structs affected, 40 bytes saved total\n", buf.String())
}

func TestRenderSummarySingular(t *testing.T) {
	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: false}
	p.RenderSummary(1, 1)
	assert.Equal(t, "Summary: 1 struct affected, 1 byte saved total\n", buf.String())
}

func TestRenderSummaryZero(t *testing.T) {
	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: false}
	p.RenderSummary(0, 0)
	assert.Equal(t, "Summary: 0 structs affected, 0 bytes saved total\n", buf.String())
}

func TestRenderSummaryColorBoldsLabel(t *testing.T) {
	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: true}
	p.RenderSummary(2, 16)
	out := buf.String()
	assert.Contains(t, out, "\x1b[1m")
	assert.Contains(t, out, "2 structs affected, 16 bytes saved total")
}
