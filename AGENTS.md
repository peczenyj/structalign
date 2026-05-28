# AGENTS.md

This is the single source of truth for both human contributors and coding agents
(Claude Code and others) working in this repository: the architectural overview,
development workflows, and coding conventions.

## What this is

`structalign` is a single-binary Go CLI that shows how a struct's fields could be
reordered to use less memory, as **human-readable** output: it prints the
reordered struct plus a diff for review, rather than editing files or emitting a
machine-applicable patch the way `fieldalignment -fix` / `-fix -diff` do. It also
has an `-inspect` mode that prints a struct's memory layout (offset/size/align/
padding per field).

The program is split into small, decoupled packages:

- `main.go` (module root) — a thin entrypoint: `os.Exit(app.New(os.Stdout, os.Stderr).Run(os.Args[1:]))`.
- `pkg/common` — the public **contracts**: data types (`Target`, `Finding`,
  `Layout`, `LayoutField`, `DiffStyle`, `Colorize`) and interfaces (`Loader`, `Aligner`,
  `Inspector`, `Sizes`). Kept out of `internal/` so mockery can generate mocks
  from a non-internal source.
- `internal/` — the implementations: `loader` (go/packages adapter), `align`
  (runs the analyzer → findings), `layout` (computes struct layouts), `sizes`
  (`go/types` sizing adapter), `textdiff` (go-udiff line diff), `match` (glob
  filtering), `structfilter` (generated-file and `cpu.CacheLinePad` predicates),
  `config` (.structalignrc and env var mapping),
  `ui` (the `Printer` — all rendering + color/width helpers), `app` (flag parsing
  + wiring). Plus `testutil` (in-process `Target` builder for tests) and `mocks`
  (mockery-generated, test-only).

`_example/types.go` is sample input used for manual testing; the leading
underscore makes the Go tool skip the directory, so it never enters `./...`.
The module path is `github.com/peczenyj/structalign`, and `main` is at the module
root, so the install target is `github.com/peczenyj/structalign@latest` → binary
`structalign`.

## Commands

The repo uses [Task](https://taskfile.dev); the `Makefile` is a thin delegator
(`make X` runs `task X`). Run `task --list` to see everything.

```
task build                 # -> ./structalign
task lint                  # golangci-lint v2 (lint + formatters: gofumpt/goimports/gci)
task test                  # gotestsum over all packages
task test -- -update       # regenerate golden fixtures (internal/ui/testdata/*.golden)
task smoke                 # run both modes against ./_example
task generate              # regenerate code (go generate) + mocks (mockery); subtasks generate:code / generate:mocks
task ci                    # full pre-push gate: tidy:check, lint, go-consistent, build, test, smoke
go run . [flags] [packages]                   # packages: ./..., import paths, dirs, files
```

`enumer` and `mockery` are code generators; `DiffStyle` and `Colorize`
(`pkg/common`) are enumer-generated `uint8` enums (`go generate ./pkg/common`
after changing their constants), and mocks come from `mockery`. Generated files (`*_enumer.go`,
`internal/mocks/*`) are committed — regenerate, never hand-edit.

Exit code is meaningful: diff modes exit **1** when any reordering is found
(CI-friendly), **0** when none; `-inspect` always exits 0.

## Core architecture

The key design decision (see the README "How it works"): this tool **does not
reimplement** the field-alignment algorithm. `internal/align` runs the unmodified
upstream `fieldalignment.Analyzer` and intercepts the `SuggestedFix` it already
produces. The pipeline, orchestrated by `app.Run`:

1. **`loader.Load`** resolves CLI args via `golang.org/x/tools/go/packages`
   (`./...`, import paths, dirs, and — via `normalizeArgs` — single `.go` files)
   into `[]common.Target`. A `Target` is a loader-agnostic view of one typed
   package (syntax, types, type info, sizes) — it hides `go/packages.Package`.
2. **`align.Findings`** runs the analyzer over a `Target` (wiring an
   `analysis.Pass` and satisfying the `inspect` pass with `inspector.New`) and
   returns `[]common.Finding` — plain data (original + proposed struct text +
   message), not rendered output. **`layout.Layouts`** is the parallel inspect
   path: it reads `Sizes.Offsetsof`/`Sizeof`/`Alignof` to produce `[]common.Layout`.
3. **`app.Run` collects** all findings (or layouts) across the scanned targets
   into one slice, then post-processes that slice in `app`: **filter**
   (`-threshold`, by absolute bytes saved), **sort** (`-sort`, largest-first —
   findings by savings, layouts by `Layout.Total`), and hand the result to
   **`ui.Printer`**, which renders (unified / side-by-side / proposed-only diff
   via `textdiff`, or annotated layout) to an `io.Writer`. With `-summary` (diff
   only) it then prints a one-line `Summary: N structs affected, M bytes saved total`.
   The savings metric is the shared `app.savings(common.Finding) int64` helper
   (used by sort, threshold, and summary). Because the logic packages return data
   and `ui` consumes it, rendering is testable by injecting findings — no
   analyzer, no toolchain.

Two **injectable wrappers** are the crux of the decoupling and testability:

- **`common.Sizes`** abstracts `go/types` sizing. Its method set matches
  `go/types.Sizes`, so a `common.Sizes` is assignable directly to
  `analysis.Pass.TypesSizes`. Production uses the loaded package's sizes (host
  `GOOS`/`GOARCH`); tests inject `sizes.ForArch("amd64")`, making golden output
  deterministic on any host (no arch `t.Skip`).
- **`common.Target`** hides `go/packages.Package`. `testutil.Target(tb, src)`
  builds one from a source string in-process (`go/parser` + `go/types`, no
  `go list` shell-out) — fast and hermetic. It runs the test from a temp dir
  (`tb.Chdir`) and writes a relative `src.go`, so the analyzer's recorded
  filename is a stable `"src.go"` (deterministic golden output) while
  `align.readSource` can still read the bytes off disk.

**Testing:** each package has black-box `_test` tests using
`github.com/stretchr/testify` (`require`/`assert`). The golden tests live in
`internal/ui` (build findings/layouts via `align`/`layout` against
`testutil.Target`, compare to `testdata/*.golden`; regenerate with
`task test -- -update`). mockery generates `Loader`/`Aligner`/`Inspector` mocks
into `internal/mocks` (test-only, excluded from lint/coverage via the Taskfile's
`PKG_LIST`); `Sizes` is intentionally **not** mocked — it has a real
deterministic implementation.

Package load/type errors are surfaced on each `Target.Errors` and printed to
stderr but are non-fatal — a partially-resolved package can still produce findings.

**Layered Configuration:** defaults are loaded from four layers before parsing
CLI arguments (highest precedence wins):
1. **CLI flags** (e.g. `-sort`)
2. **Environment variables** (`STRUCTALIGN_<FLAG>`, e.g. `STRUCTALIGN_SORT=true`)
3. **CWD RC file** (`./.structalignrc`, key=value format)
4. **Home RC file** (`~/.structalignrc`)
5. **Built-in defaults**

The `-no-rc` flag (detected early) disables loading both `.structalignrc` files.
RC files use `key = value` lines; `#` comments and blank lines are ignored.
Keys map directly to flag names. **theme** is not an RC key (use
`STRUCTALIGN_THEME`).

**Why go-udiff and not x/tools' own diff:** Go's internal-package rule forbids
importing `golang.org/x/tools/internal/diff` from a module not rooted under
`golang.org/x/tools/`. `fieldalignment`'s *own* internal imports are fine because
the importer is inside x/tools and this tool only touches its public API. go-udiff
is a public port of the same gopls diff code, so results are equivalent. Don't try
to swap it back for the internal package — it won't compile from this module.

## Things to keep consistent when editing

- **Type sizes flow through `common.Sizes`.** Production wraps the toolchain's
  real target sizes (`pkg.TypesSizes`); tests inject `sizes.ForArch("amd64")`.
  Don't reach for a hardcoded arch or a mock — use the interface.
- **`align` and `layout` return data, `ui` renders it.** Keep that split: no
  printing in the logic packages, no analysis in `ui`. New output formatting goes
  in `ui`; new analysis/derived fields go on the `common` types.
- **Scan options travel in `common.Options`** (`Patterns`, `KeepTags`,
  `IncludeGenerated`, `SkipCachePadded`, `RespectNolint`, `NolintLinters`), passed
  to `Aligner.Findings` / `Inspector.Layouts`. `align`/`layout` apply the filters
  via `internal/structfilter`
  (`InGeneratedFile` uses `go/ast.IsGenerated`; `HasCacheLinePad` checks for a
  `golang.org/x/sys/cpu.CacheLinePad` field, skipped via `-skip-cache-padded`). **Generated files are skipped by
  default** (`-generated` opts in); `_test.go` is loaded only with `-tests`
  (`loader.New(tests)`); `-exclude` drops packages by import-path regexp in `app`.
  Add a new scan knob to `Options`, not as another positional arg.
  - **Config discovery lives in `internal/config`.** It handles `.structalignrc`
  parsing and env-name derivation (`-skip-cache-padded` →
  `STRUCTALIGN_SKIP_CACHE_PADDED`). `app.Run` wires these as defaults via
  `fs.Set` before calling `fs.Parse`.
  - **//nolint is respected by default (diff only).** `align.nolintIndex` maps
  `StructType.Pos()` to the directive parsed from the type's doc comment
  (`TypeSpec.Doc` / grouped `GenDecl.Doc`) **and** any comment on the type's
  opening line (a trailing `type T struct { //nolint`, matched by line since the
  AST doesn't attach it to `TypeSpec.Comment`). `buildFinding` drops a finding
  when `Options.RespectNolint` and the directive is bare `//nolint` or names a
  token in `Options.NolintLinters` (default `["fieldalignment"]`). `app` wires
  `-show-nolint` (→ `RespectNolint = !showNolint`) and `-nolint-linters`. Inspect
  ignores `//nolint` (`layout` doesn't read these fields).
- **Diff presentation extras** live on `common.Finding`: `OldSize`/`NewSize` (parsed
  from the analyzer message) drive the `(NN.NN% smaller)` suffix, and `TypeParams`
  (e.g. `"[T]"`) lets `ui` render `type Name[T] struct {` for generics. Generic
  diffs use the type params' assumed sizes; inspect instantiates a generic with a
  representative type per parameter (`layout.representativeType`: constraint core
  type, else `interface{}`) for sizing, but renders fields from the **origin**
  struct so they stay source-faithful (`Value T`, not `Value any`). Each field
  carries `LayoutField.Assume` (e.g. `"T=any"`, or `"K=any, V=any"`), computed by
  walking the field's origin type for referenced type params (`layout.fieldAssume`
  / `collectTypeParams`, which follows pointers/slices/maps/nested generics); `ui`
  renders it as an aligned `-- assume …` marker, and `Layout.Note` carries the
  top-line disclaimer.
- **Struct name labeling** depends on `structNameIndex` (in `align`) mapping
  `StructType.Pos()` to the declared type name, because the analyzer reports at
  that position. Anonymous structs have no name and are filtered out by any
  non-empty `-type` glob (`match.MatchAny`).
- **Tag stripping** (`stripStructTags` in `align`, on by default; `-tags` preserves them) removes diff
  noise from gofmt re-aligning tags when columns shift; best-effort (falls back to
  original on parse error). Tags never affect layout numbers.
- **`DiffStyle` and `Colorize` are enumer-generated `uint8` enums** that implement
  `flag.Value` (the `-diff` and `-color` flags bind via `flag.Var`; their `Type()`
  method feeds the usage strings). Change the constants in
  `pkg/common/diffstyle.go` / `pkg/common/colorize.go`, then `go generate ./pkg/common`.
- Color, width, and padding verbosity live in `ui`: `ui.WantColor(colorize, out)` takes a
  `common.Colorize` (auto = stdout is a TTY and
  `NO_COLOR` is unset; `-color=always` overrides `NO_COLOR`, per no-color.org),
  `ui.ResolveWidth(out)` (side-by-side column width from the terminal size), and
  the `-verbose` flag (whether padding gets its own `_` line in inspect mode).
- **Themes** route color through `ui.Theme` (semantic roles `Header/Added/Removed/
  Meta/Padding/Label`); `Printer.Theme` zero value resolves to `ui.DefaultTheme()`,
  which is **byte-for-byte the historical palette** (golden fixtures must stay
  unchanged — never `-update` them for a theme change). Built-ins
  (`default`/`cga`/`green`/`amber`) live in `internal/ui/themes.go`
  (`ui.ThemeByName`). `app` selects one — hidden easter-egg flags
  `-cga`/`-green`/`-amber` (caught in the pre-parse scan beside `-fix` and
  stripped from args, so invisible in `-help`) win over the documented
  `STRUCTALIGN_THEME` env var, else default; an unknown name warns to stderr.
  Theme is orthogonal to `-color` (palette only applies when color is on). This is
  not a full theming system; per-role/custom themes are a planned later feature.
