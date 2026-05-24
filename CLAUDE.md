# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`structalign` is a single-binary Go CLI that shows how a struct's fields could be
reordered to use less memory — printing the reordered struct plus a diff, instead
of rewriting files the way `fieldalignment -fix` does. It also has an `-inspect`
mode that prints a struct's memory layout (offset/size/align/padding per field).

The entire program lives in `cmd/structalign/main.go` (one `package main`).
`_example/types.go` is sample input used for manual testing; the leading
underscore makes the Go tool skip the directory, so it never enters `./...`.
The module path is `github.com/peczenyj/structalign`, so the install target is
`github.com/peczenyj/structalign/cmd/structalign@latest` → binary `structalign`.

## Commands

```
go build -o structalign ./cmd/structalign     # produces ./structalign
go vet ./...
go run ./cmd/structalign [flags] <file.go | dir> [...]
go run ./cmd/structalign ./_example            # exercise diff mode
go run ./cmd/structalign -inspect ./_example   # exercise inspect mode
```

There are **no tests** in this repo. Verify changes by running the binary against
`./_example` and eyeballing output. Exit code is meaningful: diff modes exit **1**
when any reordering is found (CI-friendly), **0** when none; `-inspect` always
exits 0.

## Core architecture

The key design decision (see the package doc and README "How it works"): this tool
**does not reimplement** the field-alignment algorithm. It runs the unmodified
upstream `golang.org/x/tools/go/analysis/passes/fieldalignment.Analyzer` and
intercepts the `SuggestedFix` it already produces. The pipeline in `process()`:

1. `loadFiles` parses a `.go` file or every non-test `.go` in a directory as one package.
2. Type-check with `go/types`, supplying `types.SizesFor("gc", "amd64")` so the
   analyzer's size math is correct. The `types.Config.Error` callback swallows
   errors so partially-resolving packages (missing deps) are still analyzed.
3. Run the analyzer manually via an `analysis.Pass`, satisfying its only
   dependency (the `inspect` pass) with `inspector.New(files)` in `Pass.ResultOf`.
4. A custom `Pass.Report` captures each diagnostic's single `TextEdit`: `NewText`
   is the proposed reordered struct, and `readSource` reads the original source
   slice between the edit's `Pos`/`End`.
5. `lcsDiff` diffs the two via `github.com/aymanbagabas/go-udiff` (`udiff.Lines`),
   and `render` emits unified / side-by-side / proposed-only output.

**Why go-udiff and not x/tools' own diff:** Go's internal-package rule forbids
importing `golang.org/x/tools/internal/diff` from a module not rooted under
`golang.org/x/tools/`. `fieldalignment`'s *own* internal imports are fine because
the importer is inside x/tools and this tool only touches its public API. go-udiff
is a public port of the same gopls diff code, so results are equivalent. Don't try
to swap it back for the internal package — it won't compile from this module.

`-inspect` mode is a separate path (`inspectStructs` / `computeLayout` /
`renderLayout`) that never runs the analyzer; it reads `types.Sizes.Offsetsof` /
`Sizeof` / `Alignof` directly to compute per-field padding.

## Things to keep consistent when editing

- **Architecture target is hardcoded** to `("gc", "amd64")` in `cmd/structalign/main.go`'s `process` and
  reused for inspect. The README documents changing it for 32-bit targets — keep
  both call sites in sync if you parameterize it.
- **Struct name labeling** depends on `structNameIndex` mapping `StructType.Pos()`
  to the declared type name, because the analyzer reports diagnostics at exactly
  that position. Anonymous structs have no name and are filtered out by any
  non-empty `-type` glob (`matchAny`).
- **Tag stripping** (`stripStructTags`, on by default) removes diff noise from
  gofmt re-aligning tags when columns shift; it is applied to both original and
  proposed text and is best-effort (falls back to original on parse error). Tags
  never affect layout numbers.
- Color output is gated on `wantColor` (auto = stdout is a TTY); `resolveWidth`
  derives the side-by-side column width from the terminal size.
