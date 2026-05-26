package app_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/peczenyj/structalign/internal/align"
	"github.com/peczenyj/structalign/internal/app"
	"github.com/peczenyj/structalign/internal/layout"
	"github.com/peczenyj/structalign/internal/mocks"
	"github.com/peczenyj/structalign/internal/testutil"
	"github.com/peczenyj/structalign/pkg/common"
)

func summaryApp(t *testing.T, tgts []common.Target, out, errb *bytes.Buffer) *app.App {
	t.Helper()
	ml := mocks.NewLoader(t)
	ml.EXPECT().Load(mock.Anything).Return(tgts, nil)
	return &app.App{Loader: ml, Aligner: align.New(), Inspector: layout.New(), Stdout: out, Stderr: errb}
}

func TestRunSummaryPrintsLine(t *testing.T) {
	tgt := testutil.Target(t, src)
	var out, errb bytes.Buffer
	a := summaryApp(t, []common.Target{tgt}, &out, &errb)
	code := a.Run([]string{"-summary", "-type=Mixed", "pkg"})
	assert.Equal(t, 1, code)
	assert.Contains(t, out.String(), "Summary: 1 struct affected,")
	assert.Contains(t, out.String(), "bytes saved")
}

func TestRunSummaryZeroFindings(t *testing.T) {
	const aligned = "package sample\n\ntype Good struct {\n\tB int64\n\tA bool\n}\n"
	tgt := testutil.Target(t, aligned)
	var out, errb bytes.Buffer
	a := summaryApp(t, []common.Target{tgt}, &out, &errb)
	code := a.Run([]string{"-summary", "-type=Good", "pkg"})
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "Summary: 0 structs affected, 0 bytes saved")
	assert.NotContains(t, errb.String(), "no struct reorderings found")
}

func TestRunNoSummaryByDefault(t *testing.T) {
	tgt := testutil.Target(t, src)
	var out, errb bytes.Buffer
	a := summaryApp(t, []common.Target{tgt}, &out, &errb)
	_ = a.Run([]string{"-type=Mixed", "pkg"})
	assert.NotContains(t, out.String(), "Summary:")
}

func TestRunInspectIgnoresSummary(t *testing.T) {
	tgt := testutil.Target(t, src)
	var out, errb bytes.Buffer
	a := summaryApp(t, []common.Target{tgt}, &out, &errb)
	code := a.Run([]string{"-inspect", "-summary", "-type=Mixed", "pkg"})
	assert.Equal(t, 0, code)
	assert.NotContains(t, out.String(), "Summary:")
}
