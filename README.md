# structalign

A standalone CLI that shows how a Go struct's fields could be **reordered to use
less memory** — and prints the reordered struct plus a **diff** (unified or
side-by-side) instead of rewriting your files the way `fieldalignment -fix` does.

## Why this exists

`golang.org/x/tools/.../fieldalignment` has exactly two modes:

- **report** — prints a terse message like `struct of size 24 could be 16` and
  nothing else;
- **`-fix`** — silently rewrites your source.

There is no "show me the proposed struct / show me the diff" mode. `structalign`
fills that gap.

## How it works (reuse, not reimplementation)

It does **not** reimplement the alignment algorithm. It runs the **unmodified**
`fieldalignment.Analyzer` and intercepts the `analysis.SuggestedFix` the analyzer
already produces. That fix is a single `analysis.TextEdit` whose range spans the
whole struct node (`Pos … End`) and whose `NewText` is the optimally-ordered,
gofmt'd struct. `structalign`:

1. Type-checks the target package with the stdlib (`go/types`), supplying
   `types.SizesFor("gc", "amd64")` so the analyzer's size math is correct.
2. Satisfies the analyzer's only dependency — the `inspect` pass — by building an
   `inspector.New(files)` and placing it in `Pass.ResultOf`.
3. Provides a custom `Pass.Report` that captures each diagnostic's `NewText`
   (the proposed struct) and reads the original source slice between `Pos` and
   `End` (the current struct).
4. Diffs the two using `github.com/aymanbagabas/go-udiff` (a maintained
   standalone port of the Myers diff packages gopls uses, via `udiff.Lines`)
   and renders the result as a unified or side-by-side diff, or just prints the
   reordered struct.

Because all the alignment logic — including the GC pointer-bytes optimization and
size calculations — comes straight from upstream, the results match
`fieldalignment` exactly. Only the *presentation* is new.

### Dependencies and the internal-package rule

This tool lives in its **own standalone module** (`github.com/peczenyj/structalign`)
and pulls two dependencies as ordinary `go get`-able modules:

- `golang.org/x/tools` — for the public `.../passes/fieldalignment` analyzer.
- `github.com/aymanbagabas/go-udiff` — for line diffing.

A note on Go's internal-package rule, because it is easy to get wrong (I did, in
an earlier draft of this README). The rule: a package may import
`<prefix>/internal/...` only if the **importing package's own path** is rooted at
`<prefix>/`.

- `fieldalignment` imports `golang.org/x/tools/internal/astutil`. That is fine —
  but the importer there is `fieldalignment` itself, whose path *is* under
  `golang.org/x/tools/`. This tool's code only touches `fieldalignment`'s public
  API, so importing the analyzer from any module works.
- `golang.org/x/tools/internal/diff`, by contrast, **cannot** be imported from
  `github.com/peczenyj/structalign`: that path is not under `golang.org/x/tools/`, so the
  compiler rejects it (`use of internal package ... not allowed`). That is exactly
  why this tool uses `go-udiff` instead — a real public module with a stable API,
  importable from anywhere. It is a port of the same gopls diff code, so the
  results are equivalent.

## Usage

```
structalign [flags] <file.go | package-dir> [...]

  -diff string    diff style: unified | side | none   (default "unified")
  -width int      column width per side for -diff=side (default: auto from terminal)
  -color string   auto | always | never               (default "auto")
  -type string    only consider named structs matching these comma-separated
                  glob patterns (e.g. "*Request,Config"); empty means all
  -inspect        inspect layout instead of diffing: print each struct as
                  annotated Go source with size/align/padding comments
  -verbose        in -inspect mode, show padding on its own `_` line
  -tags           preserve struct field tags in output (default: strip them)
```

Exit code is **1 when reorderings are found**, **0 when none** — so it drops into
CI as a check. Inspect mode is informational and always exits 0.

### Filtering by type name

`-type` takes a comma-separated list of glob patterns (`path.Match` syntax:
`*`, `?`, `[...]`) matched against the *declared* name of each struct type.
Anonymous structs and struct literals are never matched by a non-empty filter.
It applies to every mode:

```
$ structalign -type='*Request' ./...        # only structs ending in Request
$ structalign -type='Record,Config' ./pkg   # exact names
$ structalign -inspect -type='*ID*' ./pkg   # inspect just ID-related structs
```

### Inspecting layout

`-inspect` skips the alignment analyzer entirely and prints each (filtered)
named struct as annotated Go source: the declaration with per-field
`// size: N, align: M` comments, column-aligned, plus a size/align/padding
summary on the opening line. Padding is folded onto the field comment by
default:

```
$ structalign -inspect -type=Mixed ./pkg
type Mixed struct { // size: 24, align: 8, padding: 14
	A bool  // size:  1, align: 1, padding: 7
	B int64 // size:  8, align: 8
	C bool  // size:  1, align: 1, padding: 7
}
```

With `-verbose`, padding moves onto its own `_` line:

```
$ structalign -inspect -verbose -type=Mixed ./pkg
type Mixed struct { // size: 24, align: 8, padding: 14
	A bool  // size:  1, align: 1
	_       // 7 byte padding
	B int64 // size:  8, align: 8
	C bool  // size:  1, align: 1
	_       // 7 byte padding
}
```

The layout comes from the same `go/types` sizing the diff modes use
(`types.Sizes.Offsetsof` / `Sizeof` / `Alignof`), so it respects the `gc`/`amd64`
target. This is similar to `honnef.co/go/tools/cmd/structlayout` but stays inside
this one tool and honors the same `-type` filter.

### Field tags

By default the tool **strips struct field tags** from all output, so the focus
stays on field order and layout rather than tag text. This matters most in diff
mode: reordering changes column widths, which makes `gofmt` re-align tags, and
those re-spacing changes would otherwise show up as diff noise unrelated to the
actual reorder. Stripping tags from both sides removes that distraction.

Pass `-tags` to keep tags. In diff mode they stay bound to their fields as the
fields move; in inspect mode they are appended to each field declaration (with
comments still column-aligned):

```
$ structalign -inspect -tags -type=Tagged ./pkg
type Tagged struct { // size: 48, align: 8, padding: 18
	Flag bool `json:"flag"`       // size:  1, align: 1, padding: 7
	ID string `json:"id" db:"id"` // size: 16, align: 8
	Count uint32 `json:"count"`   // size:  4, align: 4, padding: 4
	Ptr *uint64                   // size:  8, align: 8
	Enabled bool `json:"enabled"` // size:  1, align: 1, padding: 7
}
```

Tags never affect the layout numbers (size/offset/alignment are independent of
tags), so stripping them changes only the display, never the analysis.

### Examples

Unified diff (default):

```
$ structalign ./types.go
types.go:4:12: struct of size 24 could be 16
  struct {
- 	A bool
  	B int64
+ 	A bool
  	C bool
  }
```

Side-by-side:

```
$ structalign -diff=side -width=28 ./types.go
types.go:4:12: struct of size 24 could be 16
  current                      │ proposed
  ─────────────────────────────┼─────────────────────────────
  struct {                     │ struct {
      A bool                   │
      B int64                  │     B int64
                               │     A bool
      C bool                   │     C bool
  }                            │ }
```

Print the reordered struct only (no diff):

```
$ structalign -diff=none ./types.go
```

## Installing

```
go install github.com/peczenyj/structalign/cmd/structalign@latest
```

This puts a `structalign` binary on your `$GOPATH/bin` (or `$GOBIN`). The tool is a
**normal standalone module** — it does *not* need to live inside the
`golang.org/x/tools` tree. `fieldalignment`'s own `internal/astutil` import is
legal because the importer is inside x/tools; this code only uses the analyzer's
public API.

## Building from source

```
git clone https://github.com/peczenyj/structalign
cd structalign
go build -o structalign ./cmd/structalign
go run ./cmd/structalign ./_example         # quick smoke test against the sample
```

The program is a single file at `cmd/structalign/main.go`; `_example/` holds sample
structs for manual testing. The leading underscore keeps the Go tool from treating
it as a package, so it stays out of `go build ./...` and friends.

## Caveats inherited from fieldalignment

- The most compact order is not always the most efficient — packing fields tightly
  can occasionally induce false sharing between goroutines.
- Reordering can hurt logical grouping/readability; treat the output as advice,
  most valuable for hot, frequently-allocated structs.
- Sizes are computed for a target architecture (`amd64` here). 32-bit targets can
  differ; change `types.SizesFor("gc", "...")` in `main.go` if needed.
