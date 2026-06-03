#!/bin/bash -eu
# Build the repo's native Go fuzz targets (go test -fuzz style) as libFuzzer
# binaries via the OSS-Fuzz toolchain's compile_native_go_fuzzer helper.
# Usage: compile_native_go_fuzzer <import path> <FuzzFunc> <output binary>

# The rewrite shim used by compile_native_go_fuzzer; added to go.mod only
# inside the build container, never in the repo.
go get github.com/AdamKorcz/go-118-fuzz-build/testing

compile_native_go_fuzzer github.com/peczenyj/structalign/internal/match FuzzMatchAny fuzz_match_any
compile_native_go_fuzzer github.com/peczenyj/structalign/internal/match FuzzSplitCSV fuzz_split_csv
compile_native_go_fuzzer github.com/peczenyj/structalign/internal/ui FuzzTruncPad fuzz_trunc_pad
