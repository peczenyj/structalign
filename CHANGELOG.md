# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Stop merge commits leaking into the changelog and pin the GitHub remote

## [0.5.2] - 2026-05-25

### Changed

- Merge tag 'v0.5.1' into devel
- Revert "fix: refine representativeType logic for generics"
- Merge pull request #22 from peczenyj/pr/refine-layout-logic
- Merge pull request #23 from peczenyj/feature/ci-go-consistent

### Fixed

- Refine representativeType logic for generics
- Refine representativeType logic for generics

### Documentation

- Add missing flags to AGENTS.md

## [0.5.1] - 2026-05-25

### Added

- *(ui)* Honor the NO_COLOR environment variable

### Changed

- Merge tag 'v0.5.0' into devel
- Merge pull request #21 from peczenyj/feature/no-color
- Merge branch 'release/v0.5.1'

## [0.5.0] - 2026-05-25

### Added

- *(layout)* Inspect generic structs with a representative type + disclaimer
- *(layout)* Keep generic fields source-faithful with per-field assume marker
- *(app)* Nudge -fix users toward fieldalignment (easter egg)

### Changed

- Merge tag 'v0.4.0' into devel
- Merge pull request #16 from peczenyj/feature/version-buildinfo
- Merge pull request #17 from peczenyj/feature/inspect-generics
- Merge pull request #18 from peczenyj/feature/fix-easter-egg
- Merge pull request #19 from peczenyj/docs/task-commands
- Merge branch 'release/v0.5.0'

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

- Merge tag 'v0.3.0' into devel
- Move main to the module root
- Merge pull request #14 from peczenyj/feature/v0.4.0-filters-and-display
- Merge pull request #15 from peczenyj/feature/v0.4.0-docs
- Merge branch 'release/v0.4.0'

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

- Merge tag 'v0.2.0' into devel
- Update README.md
- Merge pull request #7 from peczenyj/feature/tooling-task-golangci
- *(common)* Generate DiffStyle as a uint8 enum via enumer
- Merge pull request #8 from peczenyj/feature/pkg-common
- Merge pull request #9 from peczenyj/feature/internal-sizes-match
- Update README.md
- Update README.md
- Merge pull request #11 from peczenyj/feature/codecov
- Merge pull request #10 from peczenyj/feature/internal-loader
- Keep readSource verbatim; back testutil with a real on-disk file
- Wire app package; reduce main.go to an entrypoint
- Merge pull request #12 from peczenyj/feature/v0.3.0-decoupling
- Merge pull request #13 from peczenyj/feature/v0.3.0-followup
- Merge branch 'release/v0.3.0'

### Documentation

- Describe the v0.3.0 package layout and Task workflow

## [0.2.0] - 2026-05-24

### Added

- Support Go package patterns via go/packages (closes #5)

### Changed

- Merge tag 'v0.1.0' into devel
- Merge pull request #6 from peczenyj/feature/5-recursive-package-pattern
- Merge branch 'release/v0.2.0'

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
- Merge branch 'release/v0.1.0'

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

[0.5.2]: https://github.com///compare/v0.5.1..v0.5.2
[0.5.1]: https://github.com///compare/v0.5.0..v0.5.1
[0.5.0]: https://github.com///compare/v0.4.0..v0.5.0
[0.4.0]: https://github.com///compare/v0.3.0..v0.4.0
[0.3.0]: https://github.com///compare/v0.2.0..v0.3.0
[0.2.0]: https://github.com///compare/v0.1.0..v0.2.0
[0.1.0]: https://github.com///tree/v0.1.0

<!-- generated by git-cliff -->
