package ui_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/internal/ui"
	"github.com/peczenyj/structalign/pkg/common"
)

// testdataDir is the absolute path to this package's testdata directory.
// It is computed once at init time (before any test can Chdir) so that
// testutil.Target's tb.Chdir call doesn't break relative golden paths.
var testdataDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	testdataDir = filepath.Join(filepath.Dir(file), "testdata")
}

var update = flag.Bool("update", false, "update .golden files")

const sampleSrc = `package sample

type Mixed struct {
	A bool
	B int64
	C bool
}

type Tagged struct {
	Flag    bool   ` + "`json:\"flag\"`" + `
	ID      string ` + "`json:\"id\" db:\"id\"`" + `
	Count   uint32 ` + "`json:\"count\"`" + `
	Ptr     *uint64
	Enabled bool ` + "`json:\"enabled\"`" + `
}
`

func TestRenderUnified(t *testing.T) {
	tgt := testutil.Target(t, sampleSrc)
	findings, err := align.New().Findings(tgt, []string{"Mixed"}, false)
	require.NoError(t, err)

	var buf bytes.Buffer
	p := &ui.Printer{Out: &buf, Color: false, Width: 28}
	require.Equal(t, 1, p.RenderFindings(findings, common.DiffUnified))

	out := buf.String()
	assert.Contains(t, out, "+ ")
	assert.Contains(t, out, "- ")
	assert.NotContains(t, out, "\x1b[", "color=false output must have no ANSI escapes")
}

func TestGoldenRendering(t *testing.T) {
	tgt := testutil.Target(t, sampleSrc)
	al := align.New()
	in := layout.New()

	cases := []struct {
		name   string
		golden string
		render func(p *ui.Printer) int
	}{
		{"diff_unified_mixed", "diff_unified_mixed.golden", func(p *ui.Printer) int {
			f, _ := al.Findings(tgt, []string{"Mixed"}, false)
			return p.RenderFindings(f, common.DiffUnified)
		}},
		{"diff_side_mixed", "diff_side_mixed.golden", func(p *ui.Printer) int {
			f, _ := al.Findings(tgt, []string{"Mixed"}, false)
			return p.RenderFindings(f, common.DiffSide)
		}},
		{"inspect_mixed", "inspect_mixed.golden", func(p *ui.Printer) int {
			return p.RenderLayouts(in.Layouts(tgt, []string{"Mixed"}), false, false)
		}},
		{"inspect_verbose_mixed", "inspect_verbose_mixed.golden", func(p *ui.Printer) int {
			return p.RenderLayouts(in.Layouts(tgt, []string{"Mixed"}), true, false)
		}},
		{"inspect_tags_tagged", "inspect_tags_tagged.golden", func(p *ui.Printer) int {
			return p.RenderLayouts(in.Layouts(tgt, []string{"Tagged"}), false, true)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := &ui.Printer{Out: &buf, Color: false, Width: 28}
			require.NotZero(t, tc.render(p), "expected output")
			compareGolden(t, filepath.Join(testdataDir, tc.golden), buf.Bytes())
		})
	}
}

func compareGolden(t *testing.T, path string, got []byte) {
	t.Helper()
	if *update {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, got, 0o644))
		return
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err, "missing golden (run: go test ./internal/ui/ -update)")
	assert.Equal(t, string(want), string(got))
}
