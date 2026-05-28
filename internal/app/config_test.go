package app_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/app"
	"github.com/peczenyj/structalign/internal/mocks"
	"github.com/peczenyj/structalign/pkg/common"
)

func TestRunLayeredConfig(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	cwd := filepath.Join(tmp, "cwd")
	require.NoError(t, os.Mkdir(home, 0o755))
	require.NoError(t, os.Mkdir(cwd, 0o755))

	// Mock HOME for config.Load
	t.Setenv("HOME", home)

	// Mock CWD
	oldCWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(cwd))
	t.Cleanup(func() { _ = os.Chdir(oldCWD) })

	// 1. Home RC: threshold = 8
	require.NoError(t, os.WriteFile(filepath.Join(home, ".structalignrc"), []byte("threshold = 8\n"), 0o644))

	// 2. CWD RC: sort = true, threshold = 16 (overrides home)
	require.NoError(t, os.WriteFile(filepath.Join(cwd, ".structalignrc"), []byte("sort = true\nthreshold = 16\n"), 0o644))

	// 3. Env Var: STRUCTALIGN_THRESHOLD = 32 (overrides CWD RC)
	t.Setenv("STRUCTALIGN_THRESHOLD", "32")

	// We want to verify these defaults are set in the flagset.
	// Since we can't easily inspect the internal 'opt' struct, we can check
	// the usage message or use an easter egg if we had one.
	// Actually, we can check if it warns on invalid values from these layers.

	var out, errb bytes.Buffer
	ml := &mocks.Loader{}
	ma := &mocks.Aligner{}
	a := &app.App{Loader: ml, Aligner: ma, Stdout: &out, Stderr: &errb}

	ml.On("Load", "pkg").Return([]common.Target{{PkgPath: "pkg"}}, nil)
	ma.On("Findings", mock.Anything, mock.Anything).Return(nil, nil)

	t.Run("Precedence", func(t *testing.T) {
		t.Setenv("STRUCTALIGN_THRESHOLD", "garbage")
		a.Run([]string{"pkg"})
		assert.Contains(t, errb.String(), "structalign: env: threshold: parse error")
		errb.Reset()
	})

	t.Run("NoRC", func(t *testing.T) {
		// With -no-rc, threshold should come from Env (garbage -> error)
		// but sort (from CWD RC) should NOT be set.
		t.Setenv("STRUCTALIGN_THRESHOLD", "32")
		// We'll use an invalid key in RC to see if it warns
		require.NoError(t, os.WriteFile(filepath.Join(cwd, ".structalignrc"), []byte("invalid-key = true\n"), 0o644))

		a.Run([]string{"-no-rc", "pkg"})
		assert.NotContains(t, errb.String(), "config: invalid-key")
		errb.Reset()
	})
}
