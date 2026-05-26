package app_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/peczenyj/structalign/internal/app"
	"github.com/peczenyj/structalign/internal/mocks"
	"github.com/peczenyj/structalign/pkg/common"
)

func TestRunEmptyExcludeDoesNotFilterEverything(t *testing.T) {
	var out, errb bytes.Buffer
	ml := mocks.NewLoader(t)
	// Mock a target that should NOT be filtered
	ml.EXPECT().Load(mock.Anything).Return([]common.Target{{PkgPath: "foo"}}, nil)

	a := &app.App{Loader: ml, Stdout: &out, Stderr: &errb}

	// Run with empty exclude
	_ = a.Run([]string{"-exclude=", "foo"})

	// If it filtered everything, we'd see "no matching structs found" in stderr
	assert.NotContains(t, errb.String(), "no matching structs found")
}

func TestRunEggFlagWithValue(t *testing.T) {
	var out, errb bytes.Buffer
	// Loader should be called because egg flag was stripped
	ml := mocks.NewLoader(t)
	ml.EXPECT().Load(mock.Anything).Return(nil, nil)

	a := &app.App{Loader: ml, Stdout: &out, Stderr: &errb}

	code := a.Run([]string{"-cga=true", "pkg"})
	assert.NotEqual(t, 2, code, "should not fail with 'flag provided but not defined'")
}
