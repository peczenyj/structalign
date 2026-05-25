package ui_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/internal/ui"
	"github.com/peczenyj/structalign/pkg/common"
)

// sampleSrc is defined in printer_test.go (same package).
func TestRenderNoneStyle(t *testing.T) {
	tgt := testutil.Target(t, sampleSrc)
	findings, err := align.New().Findings(tgt, common.Options{Patterns: []string{"Mixed"}})
	require.NoError(t, err)

	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: false, Width: 28}
	require.Equal(t, 1, p.RenderFindings(findings, common.DiffNone))

	out := buf.String()
	assert.Contains(t, out, "B int64", "none style prints the proposed struct (indented)")
	assert.NotContains(t, out, "+ ")
	assert.NotContains(t, out, "- ")
}
