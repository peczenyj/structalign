// Package-internal (not match_test) so ClusterFuzzLite's
// compile_native_go_fuzzer can rewrite this file: an external test package
// would clash with the non-test package during the go-118-fuzz-build rewrite.
package match

import "testing"

func FuzzMatchAny(f *testing.F) {
	f.Add("Mixed", "Mixed")
	f.Add("*", "Mixed")
	f.Add("", "Mixed")

	f.Fuzz(func(t *testing.T, pattern string, name string) {
		_ = MatchAny([]string{pattern}, name)
	})
}

func FuzzSplitCSV(f *testing.F) {
	f.Add("a,b,c")
	f.Add(`"a,b",c`)
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		_ = SplitCSV(s)
	})
}
