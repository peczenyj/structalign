# structalign

[![Go Reference](https://pkg.go.dev/badge/github.com/peczenyj/structalign.svg)](https://pkg.go.dev/github.com/peczenyj/structalign)
[![Go Report Card](https://goreportcard.com/badge/github.com/peczenyj/structalign)](https://goreportcard.com/report/github.com/peczenyj/structalign)
[![CI](https://github.com/peczenyj/structalign/actions/workflows/ci.yml/badge.svg)](https://github.com/peczenyj/structalign/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> See how reordering a Go struct's fields could save memory — as a **diff**, not a
> rewrite — plus a per-field **layout inspector**.

A read-only companion to `golang.org/x/tools`'s `fieldalignment`: it shows the
memory-optimal struct as a unified or side-by-side diff instead of silently
rewriting your files, and can also print any struct's offset/size/align/padding
layout. The analysis comes straight from the upstream analyzer, so results match
`fieldalignment` exactly — only the presentation is new.

## Quick start

```sh
go install github.com/peczenyj/structalign/cmd/structalign@latest

structalign ./...            # scan a tree, print a diff per shrinkable struct
```

Pointed at the bundled sample (`./_example`), it reports the reordering and exits
non-zero so it can gate CI:

```
$ structalign -type=Mixed ./_example
_example/types.go:4:12: Mixed: struct of size 24 could be 16
  struct {
+ 	B int64
  	A bool
- 	B int64
  	C bool
  }
$ echo $?
1
```

## Why it exists

`golang.org/x/tools/.../fieldalignment` has exactly two modes:

- **report** — prints a terse message like `struct of size 24 could be 16` and nothing else;
- **`-fix`** — silently rewrites your source.

There is no "show me the proposed struct / show me the diff" mode, and no way to
inspect a struct's layout. `structalign` fills both gaps.

| | report a problem | show the diff | rewrite files | inspect layout | CI-friendly exit code |
|---|:---:|:---:|:---:|:---:|:---:|
| `fieldalignment`        | ✅ | ❌ | ❌ | ❌ | ✅ |
| `fieldalignment -fix`   | ❌ | ❌ | ✅ | ❌ | ❌ |
| `structlayout`          | ❌ | ❌ | ❌ | ✅ | ❌ |
| **structalign**         | ✅ | ✅ | ❌ | ✅ | ✅ |

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

## Modes

### Diff (default)

Unified diff:

```
$ structalign -type=Mixed ./_example
_example/types.go:4:12: Mixed: struct of size 24 could be 16
  struct {
+ 	B int64
  	A bool
- 	B int64
  	C bool
  }
```

Side-by-side:

```
$ structalign -diff=side -width=28 -type=Mixed ./_example
_example/types.go:4:12: Mixed: struct of size 24 could be 16
  current                      │ proposed
  ─────────────────────────────┼─────────────────────────────
  struct {                     │ struct {
                               │     B int64
      A bool                   │     A bool
      B int64                  │
      C bool                   │     C bool
  }                            │ }
```

Print the reordered struct only (no diff): `structalign -diff=none ./_example`.

### Inspect layout

`-inspect` skips the alignment analyzer entirely and prints each (filtered) named
struct as annotated Go source: the declaration with per-field `// size: N, align: M`
comments, column-aligned, plus a size/align/padding summary on the opening line.
Padding is folded onto the field comment by default:

```
$ structalign -inspect -type=Mixed ./_example
type Mixed struct { // size: 24, align: 8, padding: 14
	A bool  // size:  1, align: 1, padding: 7
	B int64 // size:  8, align: 8
	C bool  // size:  1, align: 1, padding: 7
}
```

With `-verbose`, padding moves onto its own `_` line:

```
$ structalign -inspect -verbose -type=Mixed ./_example
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
target. This is similar to `honnef.co/go/tools/cmd/structlayout`, but stays inside
this one tool and honors the same `-type` filter.

### Filtering by type name

`-type` takes a comma-separated list of glob patterns (`path.Match` syntax: `*`,
`?`, `[...]`) matched against the *declared* name of each struct type. Anonymous
structs and struct literals are never matched by a non-empty filter. It applies to
every mode:

```sh
structalign -type='*Request' ./...          # only structs ending in Request
structalign -type='Record,Config' ./pkg     # exact names
structalign -inspect -type='*ID*' ./pkg     # inspect just ID-related structs
```

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
$ structalign -inspect -tags -type=Tagged ./_example
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

## How it works

`structalign` does **not** reimplement the alignment algorithm. It runs the
**unmodified** `fieldalignment.Analyzer`, intercepts the `analysis.SuggestedFix`
it already produces (a single `TextEdit` replacing the whole struct node with the
optimally-ordered, gofmt'd version), and diffs that against your original source.
Because all the alignment logic — including the GC pointer-bytes optimization and
size calculations — comes straight from upstream, results match `fieldalignment`
exactly. Only the *presentation* is new.

## Building from source

```sh
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
  differ; change `types.SizesFor("gc", "...")` in `cmd/structalign/main.go` if needed.

## Design notes

### Pipeline

1. Type-check the target package with the stdlib (`go/types`), supplying
   `types.SizesFor("gc", "amd64")` so the analyzer's size math is correct.
2. Satisfy the analyzer's only dependency — the `inspect` pass — by building an
   `inspector.New(files)` and placing it in `Pass.ResultOf`.
3. Provide a custom `Pass.Report` that captures each diagnostic's `NewText` (the
   proposed struct) and reads the original source slice between `Pos` and `End`.
4. Diff the two with `github.com/aymanbagabas/go-udiff` (a maintained standalone
   port of the Myers diff packages gopls uses, via `udiff.Lines`) and render the
   result as a unified or side-by-side diff, or just print the reordered struct.

### Dependencies and the internal-package rule

This tool lives in its own standalone module (`github.com/peczenyj/structalign`)
and pulls two dependencies as ordinary `go get`-able modules:

- `golang.org/x/tools` — for the public `.../passes/fieldalignment` analyzer.
- `github.com/aymanbagabas/go-udiff` — for line diffing.

Go's internal-package rule says a package may import `<prefix>/internal/...` only
if the **importing package's own path** is rooted at `<prefix>/`. That is why
diffing uses `go-udiff` rather than x/tools' own diff package:

- `fieldalignment` imports `golang.org/x/tools/internal/astutil` — fine, because
  the importer is itself under `golang.org/x/tools/`. This tool only touches
  `fieldalignment`'s public API, so importing the analyzer from any module works.
- `golang.org/x/tools/internal/diff`, by contrast, **cannot** be imported from
  `github.com/peczenyj/structalign` (not under `golang.org/x/tools/`), so the
  compiler rejects it. `go-udiff` is a public port of the same gopls diff code,
  so the results are equivalent.

## License

[MIT](LICENSE) © Tiago Peczenyj
