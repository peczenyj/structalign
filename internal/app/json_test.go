package app_test

import (
	"bytes"
	"encoding/json"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/app"
	"github.com/peczenyj/structalign/internal/mocks"
	"github.com/peczenyj/structalign/pkg/common"
)

type dummySizes struct{ common.Sizes }

func TestRunJSONFormat(t *testing.T) {
	var out, errb bytes.Buffer
	ml := &mocks.Loader{}
	ma := &mocks.Aligner{}
	a := &app.App{Loader: ml, Aligner: ma, Stdout: &out, Stderr: &errb}

	tgt := common.Target{
		PkgPath: "pkg",
		Types:   &types.Package{},
		Sizes:   dummySizes{},
	}
	finding := common.Finding{
		Package: "pkg",
		Name:    "Mixed",
		OldSize: 24,
		NewSize: 16,
	}

	ml.On("Load", "pkg").Return([]common.Target{tgt}, nil)
	ma.On("Findings", mock.Anything, mock.Anything).Return([]common.Finding{finding}, nil)

	code := a.Run([]string{"-format=json", "pkg"})
	assert.Equal(t, 1, code, "exit 1 when findings exist")
	assert.Empty(t, errb.String(), "no 'no reorderings' message in JSON mode")

	var doc struct {
		Mode     string `json:"mode"`
		Findings []any  `json:"findings"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &doc))
	assert.Equal(t, "diff", doc.Mode)
	assert.Len(t, doc.Findings, 1)
}

func TestRunJSONInspect(t *testing.T) {
	var out, errb bytes.Buffer
	ml := &mocks.Loader{}
	mi := &mocks.Inspector{}
	a := &app.App{Loader: ml, Inspector: mi, Stdout: &out, Stderr: &errb}

	tgt := common.Target{
		PkgPath: "pkg",
		Types:   &types.Package{},
		Sizes:   dummySizes{},
	}
	layout := common.Layout{
		Package: "pkg",
		Name:    "Mixed",
		Total:   24,
	}

	ml.On("Load", "pkg").Return([]common.Target{tgt}, nil)
	mi.On("Layouts", mock.Anything, mock.Anything).Return([]common.Layout{layout})

	code := a.Run([]string{"-inspect", "-format=json", "pkg"})
	assert.Equal(t, 0, code, "inspect always exits 0")

	var doc struct {
		Mode    string `json:"mode"`
		Layouts []any  `json:"layouts"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &doc))
	assert.Equal(t, "inspect", doc.Mode)
	assert.Len(t, doc.Layouts, 1)
}
