package app_test

import (
	"bytes"
	"strings"
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

// Small saves 8 bytes (24->16); Big saves 16 (40->24) — more interleaved fields.
const sortSrc = `package sample

type Small struct {
	A bool
	B int64
	C bool
}

type Big struct {
	A bool
	B int64
	C bool
	D int64
	E bool
}
`

func sortApp(t *testing.T, out, errb *bytes.Buffer) *app.App {
	t.Helper()
	tgt := testutil.Target(t, sortSrc)
	ml := mocks.NewLoader(t)
	ml.EXPECT().Load(mock.Anything).Return([]common.Target{tgt}, nil)
	return &app.App{Loader: ml, Aligner: align.New(), Inspector: layout.New(), Stdout: out, Stderr: errb}
}

func TestRunSortDiffOrdersBySavings(t *testing.T) {
	var out, errb bytes.Buffer
	a := sortApp(t, &out, &errb)
	code := a.Run([]string{"-sort", "pkg"})
	assert.Equal(t, 1, code)
	s := out.String()
	assert.Less(t, strings.Index(s, "Big"), strings.Index(s, "Small"),
		"with -sort, the bigger-saving struct must render first")
}

func TestRunDiffDefaultOrderUnchanged(t *testing.T) {
	var out, errb bytes.Buffer
	a := sortApp(t, &out, &errb)
	_ = a.Run([]string{"pkg"})
	s := out.String()
	assert.Less(t, strings.Index(s, "Small"), strings.Index(s, "Big"),
		"without -sort, source order (Small before Big) is preserved")
}
