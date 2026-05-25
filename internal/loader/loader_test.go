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

func TestLoadReturnsErrorForInvalidQuery(t *testing.T) {
	// An unrecognized "name=value" query makes go/packages fail at the driver
	// level (a top-level error, not a per-package one).
	_, err := loader.New().Load("bogus=x")
	require.Error(t, err)
}

func TestLoadCapturesPackageErrors(t *testing.T) {
	dir := t.TempDir()
	// A type error (undefined type) surfaces as a package-level error, which the
	// loader maps onto Target.Errors rather than failing the whole load.
	src := "package sample\n\ntype Bad struct {\n\tX NoSuchType\n}\n"
	file := filepath.Join(dir, "bad.go")
	require.NoError(t, os.WriteFile(file, []byte(src), 0o644))

	targets, err := loader.New().Load(file)
	require.NoError(t, err)
	require.NotEmpty(t, targets)
	assert.NotEmpty(t, targets[0].Errors, "type error should be captured in Target.Errors")
}
