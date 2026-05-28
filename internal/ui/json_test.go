package ui_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/internal/ui"
	"github.com/peczenyj/structalign/pkg/common"
)

func TestRenderJSON(t *testing.T) {
	tgt := testutil.Target(t, sampleSrc)
	al := align.New()
	in := layout.New()

	cases := []struct {
		name   string
		golden string
		render func(p *ui.Printer)
	}{
		{"json_diff_mixed", "diff_mixed.json.golden", func(p *ui.Printer) {
			f, _ := al.Findings(tgt, common.Options{Patterns: []string{"Mixed"}})
			p.RenderJSON("v0.7.0-test", f, nil)
		}},
		{"json_inspect_mixed", "inspect_mixed.json.golden", func(p *ui.Printer) {
			l := in.Layouts(tgt, common.Options{Patterns: []string{"Mixed"}})
			p.RenderJSON("v0.7.0-test", nil, l)
		}},
		{"json_empty", "empty.json.golden", func(p *ui.Printer) {
			p.RenderJSON("v0.7.0-test", nil, nil)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := &ui.Printer{Out: &buf}
			tc.render(p)
			compareGolden(t, filepath.Join(testdataDir, tc.golden), buf.Bytes())
		})
	}
}
