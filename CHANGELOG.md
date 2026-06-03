# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.1] - 2026-06-03

### Changed

- *(ui)* Simplify truncPad via runewidth Truncate/FillRight (#94)

### Fixed

- Correct .bestpractices.json criterion field names (#90)
- Set dynamic_analysis to Met (criterion forbids N/A) (#91)
- *(ui)* Clamp column width to avoid makeslice panic on huge -width (#101)

### Documentation

- *(pkg/common)* Add comprehensive GoDoc comments for public contracts (#97)

## [0.7.0] - 2026-05-29

### Changed

- Pre-v0.7.0 Cleanup: Sorting, Efficiency, and Bug Fixes (#68)
- Layered Configuration (Env Vars + .structalignrc) (#69)
- JSON Output Support (-format=json) (#71)

### Fixed

- *(ci)* Pin codeql action to its commit SHA, not the tag object (#63)
- *(usage)* List -no-rc in -h and document the theme eggs (#74)
- Address major bugs, unhandled AST aliases, and config issues on devel (#82)
- Graceful RC keys, JSON encode-error stream, and docs (#87)
- *(align)* Strip field comments from both diff sides (#89)
- *(ci)* Correct cosign-installer pinned commit SHA
- *(ci)* Correct cosign-installer pinned commit SHA
- *(ci)* Bump cosign-installer to v4.1.2 (cosign v3.0.6)

### Documentation

- Refresh diff screenshot and drop redundant tag badge (#64)

## [0.6.1] - 2026-05-27

### Added

- Enrich -h into man-page-style help (#52)
- Report bytes saved in the diff header (#57)

### Changed

- Update README.md
- Update README.md
- Update .gitignore
- Replace hand-rolled ANSI with termenv (color engine) (#54)

### Fixed

- Address bugs in nolint parsing, unicode alignment, and flag stripping (#50)
- *(cga)* Inspect padding shares the removed yellow, not flat white (#59)

### Documentation

- Clarify fieldalignment can diff but structalign is human-readable (#45)

## [0.6.0] - 2026-05-26

### Added

- Add -summary flag for diff-mode summary line
- -sort orders diff findings by bytes saved (#36)
- Extend -sort to order inspect layouts by struct size (#37)
- -threshold filters diff findings by bytes saved (#38)
- Respect //nolint directives (diff mode) (#39)

### Changed

- Update README.md
- *(enum)* Add enumeration colorize and ensure diffstyle implement pflag.Value
- Wire the new enumeration type colorize and use pflag.Value interface to extract the allowed values
- *(app)* Collect findings/layouts then render

### Fixed

- *(changelog)* Pin git-cliff remote so compare links resolve
- Address analysis and discovery bugs (#42)
- Address UI and CLI bugs (#41)
- Legible CGA palette and 'total' in summary line (#44)

### Documentation

- Repair broken compare links in CHANGELOG
- Update readme and agents
- Document inspecting stdlib, dependency, and arbitrary library types
- Update AGENTS.md for v0.6.0 features

## [0.5.2] - 2026-05-25

### Changed

- Revert "fix: refine representativeType logic for generics"

### Fixed

- Refine representativeType logic for generics
- Refine representativeType logic for generics

### Documentation

- Add missing flags to AGENTS.md

## [0.5.1] - 2026-05-25

### Added

- *(ui)* Honor the NO_COLOR environment variable

## [0.5.0] - 2026-05-25

### Added

- *(layout)* Inspect generic structs with a representative type + disclaimer
- *(layout)* Keep generic fields source-faithful with per-field assume marker
- *(app)* Nudge -fix users toward fieldalignment (easter egg)

### Fixed

- *(app)* Report the module version for `go install` builds

### Documentation

- Use task (not make) in command examples

## [0.4.0] - 2026-05-25

### Added

- *(common)* Add Options and Finding size fields; Aligner/Inspector take Options
- *(structfilter)* Detect generated files and cache-line padding
- *(align)* Skip generated/cache-padded structs; capture sizes
- *(layout)* Skip generated/cache-padded structs
- *(loader)* Add -tests support (load _test.go files)
- *(app)* Wire -generated/-tests/-skip-cache-padded and -exclude package filter
- *(ui)* Show "type <Name> struct {" and the size-reduction percentage

### Changed

- Move main to the module root

### Documentation

- Regenerate diff.png for v0.4.0 output
- Normalize badges, add .github templates and AGENTS.md
- Refresh README/AGENTS for v0.4.0; render generic type params in the diff

## [0.3.0] - 2026-05-25

### Added

- *(common)* Add public contracts (domain types + interfaces)
- *(sizes)* Add common.Sizes adapter over go/types
- *(match)* Extract pattern parsing and glob matching
- *(loader)* Common.Loader over go/packages (Target mapping)
- *(textdiff)* Extract line diff over go-udiff
- *(align)* Common.Aligner producing findings as data
- *(layout)* Common.Inspector computing struct layouts as data
- *(ui)* Printer rendering + term helpers with golden tests

### Changed

- Update README.md
- *(common)* Generate DiffStyle as a uint8 enum via enumer
- Update README.md
- Update README.md
- Keep readSource verbatim; back testutil with a real on-disk file
- Wire app package; reduce main.go to an entrypoint

### Documentation

- Describe the v0.3.0 package layout and Task workflow

## [0.2.0] - 2026-05-24

### Added

- Support Go package patterns via go/packages (closes #5)

## [0.1.0] - 2026-05-24

### Added

- Initial commit: structalign CLI
- Add MIT license
- Add CI workflow: gofmt, vet, build, smoke test
- Add Tagged struct for -tags documentation
- Add Makefile and lower Go floor to 1.25
- Add GoReleaser-based releases and -version flag

### Changed

- Bind flags to a config struct; fix -width and -verbose help
- Modernize to Go 1.25 idioms

### Fixed

- Make screenshot target tolerate diff-mode exit code
- Skip generic types in -inspect mode

### Documentation

- Rework README for a consumer-facing tool
- Add colored diff screenshot to README
- Add latest-release badge
- Add CONTRIBUTING.md with dev workflow and release steps
- Add SECURITY.md and CODE_OF_CONDUCT.md
- *(_example)* Add Generic, FuncField, and MutexLast sample types
- Focus the README screenshot on the Record type
- Document file/dir args instead of unsupported ./... (#5)

[0.7.1]: https://github.com/peczenyj/structalign/compare/v0.7.0..v0.7.1
[0.7.0]: https://github.com/peczenyj/structalign/compare/v0.6.1..v0.7.0
[0.6.1]: https://github.com/peczenyj/structalign/compare/v0.6.0..v0.6.1
[0.6.0]: https://github.com/peczenyj/structalign/compare/v0.5.2..v0.6.0
[0.5.2]: https://github.com/peczenyj/structalign/compare/v0.5.1..v0.5.2
[0.5.1]: https://github.com/peczenyj/structalign/compare/v0.5.0..v0.5.1
[0.5.0]: https://github.com/peczenyj/structalign/compare/v0.4.0..v0.5.0
[0.4.0]: https://github.com/peczenyj/structalign/compare/v0.3.0..v0.4.0
[0.3.0]: https://github.com/peczenyj/structalign/compare/v0.2.0..v0.3.0
[0.2.0]: https://github.com/peczenyj/structalign/compare/v0.1.0..v0.2.0
[0.1.0]: https://github.com/peczenyj/structalign/tree/v0.1.0

<!-- generated by git-cliff -->
