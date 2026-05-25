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
	targets, err := loader.New(false).Load(file)
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
	_, err := loader.New(false).Load("bogus=x")
	require.Error(t, err)
}

func TestLoadCapturesPackageErrors(t *testing.T) {
	dir := t.TempDir()
	// A type error (undefined type) surfaces as a package-level error, which the
	// loader maps onto Target.Errors rather than failing the whole load.
	src := "package sample\n\ntype Bad struct {\n\tX NoSuchType\n}\n"
	file := filepath.Join(dir, "bad.go")
	require.NoError(t, os.WriteFile(file, []byte(src), 0o644))

	targets, err := loader.New(false).Load(file)
	require.NoError(t, err)
	require.NotEmpty(t, targets)
	assert.NotEmpty(t, targets[0].Errors, "type error should be captured in Target.Errors")
}

func TestLoadTestsFlag(t *testing.T) {
	dir := t.TempDir()

	// A go.mod is required so that go/packages can resolve test-variant packages.
	gomod := "module example.com/sample\n\ngo 1.21\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644))

	mainSrc := "package sample\n\ntype Mixed struct {\n\tA bool\n\tB int64\n\tC bool\n}\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "types.go"), []byte(mainSrc), 0o644))

	testSrc := "package sample\n\ntype InTest struct {\n\tX bool\n\tY int64\n\tZ bool\n}\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "types_test.go"), []byte(testSrc), 0o644))

	// Change into dir so that ./... resolves against the temp module.
	t.Chdir(dir)

	// With tests=true, InTest defined in the _test.go file should be visible in
	// the test-augmented package variant (the one with both types.go and types_test.go).
	targetsWithTests, err := loader.New(true).Load("./...")
	require.NoError(t, err)
	require.NotEmpty(t, targetsWithTests)
	foundInTest := false
	for _, tgt := range targetsWithTests {
		if tgt.Types != nil && tgt.Types.Scope().Lookup("InTest") != nil {
			foundInTest = true
			break
		}
	}
	assert.True(t, foundInTest, "InTest should be visible when tests=true")

	// With tests=false, InTest should not appear in any target.
	targetsWithoutTests, err := loader.New(false).Load("./...")
	require.NoError(t, err)
	for _, tgt := range targetsWithoutTests {
		if tgt.Types != nil {
			assert.Nil(t, tgt.Types.Scope().Lookup("InTest"), "InTest should not be visible when tests=false")
		}
	}
}
