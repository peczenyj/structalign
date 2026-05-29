// Package config handles persistent defaults via environment variables and
// .structalignrc files.
package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// EnvName derives the environment variable name for a flag name,
// e.g. "skip-cache-padded" -> "STRUCTALIGN_SKIP_CACHE_PADDED".
func EnvName(flagName string) string {
	name := strings.ReplaceAll(flagName, "-", "_")
	return "STRUCTALIGN_" + strings.ToUpper(name)
}

// Load reads and merges .structalignrc files from the home directory and the
// current working directory. CWD settings override home settings.
// Returns the merged key-value map.
func Load(home, cwd string) map[string]string {
	merged := make(map[string]string)

	// Home directory rc (personal base)
	if home != "" {
		merge(merged, parseRC(filepath.Join(home, ".structalignrc")))
	}

	// CWD directory rc (project overrides)
	if cwd != "" {
		merge(merged, parseRC(filepath.Join(cwd, ".structalignrc")))
	}

	return merged
}

func merge(dst, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

// parseRC reads a key = value file, ignoring # comments and blank lines.
func parseRC(path string) map[string]string {
	kv := make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		return kv
	}
	defer f.Close() //nolint:errcheck

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split at the first '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if k != "" {
			kv[k] = v
		}
	}
	return kv
}
