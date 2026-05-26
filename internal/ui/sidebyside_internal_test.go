package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderSideBySideComplex(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Out: &buf, Width: 20}
	a := "line1\nline2\nline3"
	b := "line1\nline2alt\nline4"
	// line2 replaced by line2alt, line3 replaced by line4
	p.renderSideBySide(a, b)
	out := buf.String()
	assert.Contains(t, out, "line2")
	assert.Contains(t, out, "line2alt")
}
