# Github Issues Report

## Issue #1: [BUG] Unicode-unsafe truncation in side-by-side diffs
**Problem:** `truncPad` in `internal/ui/printer.go` slices strings by bytes, which breaks multi-byte Unicode characters and misaligns columns.
**Assignee:** @peczenyj

## Issue #2: [BUG] Block comments ignored by //nolint
**Problem:** `parseNolint` only handles `//` prefixes. `/* nolint */` block comments are ignored.
**Assignee:** @peczenyj

## Issue #3: [BUG] Empty -exclude pattern excludes all packages
**Problem:** Providing `-exclude=""` results in a regex that matches every string, suppressing all output.
**Assignee:** @peczenyj

## Issue #4: [BUG] Incomplete easter-egg flag stripping
**Problem:** The manual flag stripping for `-cga`, `-green`, etc., only handles exact matches and fails on `-flag=value` syntax.
**Assignee:** @peczenyj

## Issue #5: [FEAT] Inspect mode misses structs declared inside functions
**Problem:** `layout.Layouts` only scans the package scope. Local structs inside function bodies are not shown in `-inspect` mode.
**Assignee:** @peczenyj
