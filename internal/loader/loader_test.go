package loader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/loader"
)

func TestLoadSingleFile(t *testing.T) {
	dir := t.TempDir()
	src := "package sample\n\ntype Mixed struct {\n\tA bool\n\tB int64\n\tC bool\n}\n"
	file := filepath.Join(dir, "types.go")
	require.NoError(t, os.WriteFile(file, []byte(src), 0o644))

	// A bare .go path is rewritten to a file= query by normalizeArgs.
	targets, err := loader.New().Load(file)
	require.NoError(t, err)
	require.NotEmpty(t, targets)

	tgt := targets[0]
	require.NotNil(t, tgt.Types, "target should carry a typed package")
	require.NotNil(t, tgt.Sizes, "target should carry sizes")
	assert.NotNil(t, tgt.Types.Scope().Lookup("Mixed"), "loaded package should define type Mixed")
}
