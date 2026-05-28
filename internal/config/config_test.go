package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peczenyj/structalign/internal/config"
)

func TestEnvName(t *testing.T) {
	assert.Equal(t, "STRUCTALIGN_SORT", config.EnvName("sort"))
	assert.Equal(t, "STRUCTALIGN_SKIP_CACHE_PADDED", config.EnvName("skip-cache-padded"))
}

func TestLoad(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	cwd := filepath.Join(tmp, "cwd")
	require.NoError(t, os.Mkdir(home, 0o755))
	require.NoError(t, os.Mkdir(cwd, 0o755))

	homeRC := `
# Personal base
sort = true
threshold = 8
`
	cwdRC := `
# Project override
threshold = 16
skip-cache-padded = true
`
	require.NoError(t, os.WriteFile(filepath.Join(home, ".structalignrc"), []byte(homeRC), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(cwd, ".structalignrc"), []byte(cwdRC), 0o644))

	got := config.Load(home, cwd)
	assert.Equal(t, "true", got["sort"], "home setting preserved")
	assert.Equal(t, "16", got["threshold"], "cwd setting overrides home")
	assert.Equal(t, "true", got["skip-cache-padded"], "cwd setting added")
}

func TestLoad_MissingFiles(t *testing.T) {
	got := config.Load("/no/such/dir", "/no/such/cwd")
	assert.Empty(t, got)
}

func TestLoad_MalformedLines(t *testing.T) {
	tmp := t.TempDir()
	rc := `
valid = yes
missing-value
= missing-key
# comment
   
`
	require.NoError(t, os.WriteFile(filepath.Join(tmp, ".structalignrc"), []byte(rc), 0o644))
	got := config.Load("", tmp)
	assert.Len(t, got, 1)
	assert.Equal(t, "yes", got["valid"])
}
