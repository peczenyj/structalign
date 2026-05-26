package app_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These reuse sortApp / sortSrc from sort_test.go, where Small saves 8 bytes
// (24->16) and Big saves 16 (40->24).

func TestRunThresholdFiltersSmallSavings(t *testing.T) {
	var out, errb bytes.Buffer
	a := sortApp(t, &out, &errb)
	// 16 is inclusive: keeps Big (saves 16), drops Small (saves 8).
	code := a.Run([]string{"-threshold=16", "pkg"})
	assert.Equal(t, 1, code)
	s := out.String()
	assert.Contains(t, s, "Big")
	assert.NotContains(t, s, "Small")
}

func TestRunThresholdZeroShowsAll(t *testing.T) {
	var out, errb bytes.Buffer
	a := sortApp(t, &out, &errb)
	_ = a.Run([]string{"-threshold=0", "pkg"})
	s := out.String()
	assert.Contains(t, s, "Big")
	assert.Contains(t, s, "Small")
}

func TestRunThresholdNegativeIsZero(t *testing.T) {
	var out, errb bytes.Buffer
	a := sortApp(t, &out, &errb)
	_ = a.Run([]string{"-threshold=-5", "pkg"})
	assert.Contains(t, out.String(), "Small", "a negative threshold behaves like 0 (no filtering)")
}

func TestRunThresholdExitsZeroWhenAllFiltered(t *testing.T) {
	var out, errb bytes.Buffer
	a := sortApp(t, &out, &errb)
	code := a.Run([]string{"-threshold=10000", "pkg"})
	assert.Equal(t, 0, code, "all filtered out => no findings => exit 0")
	assert.Equal(t, "", strings.TrimSpace(out.String()))
}

func TestRunThresholdSummaryReflectsFiltered(t *testing.T) {
	var out, errb bytes.Buffer
	a := sortApp(t, &out, &errb)
	_ = a.Run([]string{"-threshold=16", "-summary", "pkg"})
	assert.Contains(t, out.String(), "Summary: 1 struct affected, 16 bytes saved",
		"summary counts only the structs that pass the threshold")
}
