package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/tools/go/packages"
)

// update regenerates the .golden fixtures: `go test ./... -update`.
var update = flag.Bool("update", false, "update .golden files in testdata")

// --- pure-helper unit tests -------------------------------------------------

func TestParsePatterns(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"Config", []string{"Config"}},
		{"*Request,Config", []string{"*Request", "Config"}},
		{" A , B ,", []string{"A", "B"}},
		{",,", nil},
	}
	for _, tc := range tests {
		got := parsePatterns(tc.in)
		if !equalStrings(got, tc.want) {
			t.Errorf("parsePatterns(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestMatchAny(t *testing.T) {
	tests := []struct {
		patterns []string
		name     string
		want     bool
	}{
		{[]string{"Config"}, "Config", true},
		{[]string{"Config"}, "Other", false},
		{[]string{"*Request"}, "HTTPRequest", true},
		{[]string{"*Request"}, "Requester", false},
		{[]string{"Foo", "*ID*"}, "UserIDList", true},
		{[]string{"?at"}, "cat", true},
		{[]string{"*"}, "", false}, // anonymous never matches a non-empty set
		{nil, "Anything", false},
		{[]string{"["}, "[", false}, // invalid pattern is non-matching, not fatal
	}
	for _, tc := range tests {
		if got := matchAny(tc.patterns, tc.name); got != tc.want {
			t.Errorf("matchAny(%v, %q) = %v, want %v", tc.patterns, tc.name, got, tc.want)
		}
	}
}

func TestNormalizeArgs(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "real.go")
	if err := os.WriteFile(goFile, []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	in := []string{
		goFile,                      // existing .go file -> file= query
		"./...",                     // pattern, untouched
		"github.com/foo/bar",        // import path, untouched
		dir,                         // directory, untouched
		filepath.Join(dir, "no.go"), // .go suffix but does not exist -> untouched
	}
	want := []string{
		"file=" + goFile,
		"./...",
		"github.com/foo/bar",
		dir,
		filepath.Join(dir, "no.go"),
	}
	got := normalizeArgs(in)
	if !equalStrings(got, want) {
		t.Errorf("normalizeArgs:\n  got  %v\n  want %v", got, want)
	}
}

func TestStripStructTags(t *testing.T) {
	src := "struct {\n\tA bool `json:\"a\"`\n\tB int64 `json:\"b\" db:\"b\"`\n}"
	out, err := stripStructTags(src)
	if err != nil {
		t.Fatalf("stripStructTags returned error: %v", err)
	}
	if strings.Contains(out, "json") || strings.Contains(out, "`") {
		t.Errorf("tags not stripped:\n%s", out)
	}
	for _, field := range []string{"A bool", "B int64"} {
		if !strings.Contains(out, field) {
			t.Errorf("field %q missing from output:\n%s", field, out)
		}
	}

	// A struct without tags round-trips through gofmt unchanged in substance.
	if _, err := stripStructTags("struct {\n\tA bool\n}"); err != nil {
		t.Errorf("stripStructTags on tagless struct: %v", err)
	}

	// Unparseable input returns an error so the caller can fall back.
	if _, err := stripStructTags("not a struct"); err == nil {
		t.Error("stripStructTags on garbage: want error, got nil")
	}
}

func TestRelPath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	under := filepath.Join(wd, "sub", "file.go")
	if got, want := relPath(under), filepath.Join("sub", "file.go"); got != want {
		t.Errorf("relPath(%q) = %q, want %q", under, got, want)
	}

	// A path that escapes upward is returned unchanged (absolute).
	outside := filepath.Join(filepath.Dir(wd), "sibling.go")
	if got := relPath(outside); got != outside {
		t.Errorf("relPath(%q) = %q, want it unchanged", outside, got)
	}
}

func TestLcsDiffReconstructs(t *testing.T) {
	a := []string{"struct {", "\tA bool", "\tB int64", "\tC bool", "}"}
	b := []string{"struct {", "\tB int64", "\tA bool", "\tC bool", "}"}

	ops := lcsDiff(a, b)

	// Equal+deleted lines, in order, must reconstruct the original; equal+added
	// lines must reconstruct the proposed. This holds for any valid edit script,
	// so it does not depend on the diff algorithm's exact choices.
	var gotA, gotB []string
	for _, op := range ops {
		switch op.kind {
		case opEqual:
			gotA = append(gotA, op.text)
			gotB = append(gotB, op.text)
		case opDel:
			gotA = append(gotA, op.text)
		case opAdd:
			gotB = append(gotB, op.text)
		}
	}
	if !equalStrings(gotA, a) {
		t.Errorf("equal+del did not reconstruct original:\n  got  %v\n  want %v", gotA, a)
	}
	if !equalStrings(gotB, b) {
		t.Errorf("equal+add did not reconstruct proposed:\n  got  %v\n  want %v", gotB, b)
	}
}

func TestRenderUnified(t *testing.T) {
	a := "struct {\n\tA bool\n\tB int64\n}"
	b := "struct {\n\tB int64\n\tA bool\n}"

	var plain bytes.Buffer
	renderUnified(&plain, a, b, false)
	out := plain.String()
	if !strings.Contains(out, "+ ") || !strings.Contains(out, "- ") {
		t.Errorf("unified diff missing +/- lines:\n%s", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("color=false output contains ANSI escapes:\n%q", out)
	}

	var colored bytes.Buffer
	renderUnified(&colored, a, b, true)
	if !strings.Contains(colored.String(), "\x1b[") {
		t.Errorf("color=true output has no ANSI escapes:\n%q", colored.String())
	}
}

// --- golden integration tests against ./_example ----------------------------

func TestGoldenExample(t *testing.T) {
	requireSixtyFourBit(t)

	// Resolve the golden dir before chdir, so it stays valid afterwards.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	goldenDir := filepath.Join(wd, "testdata")

	// Run from the repo root so go/packages resolves "./_example" and relPath
	// reports "_example/types.go" (deterministic, matching the README).
	t.Chdir(filepath.Join("..", ".."))
	pkg := loadExamplePackage(t)

	cases := []struct {
		name   string
		golden string
		run    func(w io.Writer) int
	}{
		{"diff_unified_mixed", "diff_unified_mixed.golden", func(w io.Writer) int {
			return diffPackage(w, pkg, "unified", 0, false, []string{"Mixed"}, false)
		}},
		{"diff_side_mixed", "diff_side_mixed.golden", func(w io.Writer) int {
			return diffPackage(w, pkg, "side", 28, false, []string{"Mixed"}, false)
		}},
		{"diff_none_record", "diff_none_record.golden", func(w io.Writer) int {
			return diffPackage(w, pkg, "none", 0, false, []string{"Record"}, false)
		}},
		{"diff_tags_tagged", "diff_tags_tagged.golden", func(w io.Writer) int {
			return diffPackage(w, pkg, "unified", 0, false, []string{"Tagged"}, true)
		}},
		{"inspect_mixed", "inspect_mixed.golden", func(w io.Writer) int {
			return inspectStructs(w, pkg.Types, pkg.TypesSizes, []string{"Mixed"}, false, false, false)
		}},
		{"inspect_verbose_mixed", "inspect_verbose_mixed.golden", func(w io.Writer) int {
			return inspectStructs(w, pkg.Types, pkg.TypesSizes, []string{"Mixed"}, false, true, false)
		}},
		{"inspect_tags_tagged", "inspect_tags_tagged.golden", func(w io.Writer) int {
			return inspectStructs(w, pkg.Types, pkg.TypesSizes, []string{"Tagged"}, false, false, true)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if n := tc.run(&buf); n == 0 {
				t.Fatalf("%s: expected at least one finding, got 0", tc.name)
			}
			compareGolden(t, filepath.Join(goldenDir, tc.golden), buf.Bytes())
		})
	}

	// An already-optimal struct yields no findings and no output.
	t.Run("diff_good_optimal", func(t *testing.T) {
		var buf bytes.Buffer
		if n := diffPackage(&buf, pkg, "unified", 0, false, []string{"Good"}, false); n != 0 {
			t.Errorf("Good is already optimal: want 0 findings, got %d:\n%s", n, buf.String())
		}
		if buf.Len() != 0 {
			t.Errorf("Good produced output:\n%s", buf.String())
		}
	})
}

// --- helpers ----------------------------------------------------------------

func requireSixtyFourBit(t *testing.T) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("golden layout assumes a 64-bit target (pointer size 8)")
	}
}

func loadExamplePackage(t *testing.T) *packages.Package {
	t.Helper()
	pkgs, err := loadPackages([]string{"./_example"})
	if err != nil {
		t.Fatalf("loadPackages(./_example): %v", err)
	}
	for _, p := range pkgs {
		if strings.HasSuffix(p.PkgPath, "_example") && p.Types != nil && p.TypesSizes != nil {
			return p
		}
	}
	t.Fatalf("could not load ./_example as a typed package (got %d packages)", len(pkgs))
	return nil
}

func compareGolden(t *testing.T, path string, got []byte) {
	t.Helper()
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden %s (run `go test ./... -update` to create it): %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output does not match %s\n--- got ---\n%s\n--- want ---\n%s",
			filepath.Base(path), got, want)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
