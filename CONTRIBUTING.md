# Contributing

Thanks for your interest in structalign! This document covers the development
workflow, commit conventions, and the release process.

## Prerequisites

- **Go 1.25+** (the floor set by `golang.org/x/tools`).
- [`git-cliff`](https://git-cliff.org) — to regenerate `CHANGELOG.md`.
- [`goreleaser`](https://goreleaser.com) — to validate/build releases locally.
- Python 3 + Pillow — only needed to regenerate the README screenshot.

## Development workflow

The whole program is a single file, `cmd/structalign/main.go`; `_example/` holds
sample structs used for manual testing and in the docs.

```sh
make build          # -> ./structalign
make check          # gofmt, vet, build, and a smoke test against ./_example  (what CI runs)
make help           # list all targets
```

Run `make check` before pushing — it mirrors the CI job exactly. CI also runs the
matrix across Go 1.25.0 and stable.

When you change diff or inspect output, refresh the README screenshot:

```sh
make screenshot     # regenerates docs/diff.png (needs python3 + pillow)
```

## Commit messages

Commits follow [Conventional Commits](https://www.conventionalcommits.org). This
is what drives the changelog, so the prefix matters:

| Prefix | Changelog section |
|---|---|
| `feat:` | Added |
| `fix:` | Fixed |
| `perf:` / `refactor:` / `style:` / `revert:` | Changed |
| `deprecate:` | Deprecated |
| `remove:` | Removed |
| `security:` | Security |
| `docs:` | Documentation |
| `chore:` / `ci:` / `build:` / `test:` | (excluded — not user-facing) |

Mark breaking changes with a `!` (e.g. `feat!:`) or a `BREAKING CHANGE:` footer.

## Branching (git-flow)

- `main` — production; every commit is a tagged release.
- `devel` — integration branch and the **default** branch; day-to-day work targets it.
- `feature/*`, `release/*`, `hotfix/*` — the usual git-flow temporaries.

Do not commit directly to `main`.

## Releasing

Releases use [git-flow](https://github.com/nvie/gitflow) to cut the version and
[GoReleaser](https://goreleaser.com) (triggered by the tag push) to build the
cross-platform binaries and publish the GitHub Release. Example for `v0.1.0`:

```sh
# 1. Start the release branch off devel.
git flow release start v0.1.0

# 2. Stamp the changelog for this version and commit it.
make release TAG=v0.1.0
git add CHANGELOG.md
git commit -m "chore(release): v0.1.0"

# 3. (Recommended) dry-run the release build locally — no publish.
make release-check          # validate .goreleaser.yaml
make snapshot               # build all artifacts into dist/

# 4. Finish: merges release -> main, tags v0.1.0, back-merges into devel.
git flow release finish v0.1.0

# 5. Push everything, including the tag (the tag is what triggers the release).
git push origin main devel --tags

# 6. Verify.
gh run list --workflow=release.yml
gh release view v0.1.0
```

Notes:

- The **tag** (`vX.Y.Z`) is what fires `.github/workflows/release.yml`; if you
  forget `--tags`, nothing happens. The version is derived from the tag and
  embedded in the binary (`structalign -version`).
- Release notes are GitHub's native generated notes
  (`changelog.use: github-native` in `.goreleaser.yaml`). The in-repo
  `CHANGELOG.md` is git-cliff-generated and bundled into the release archives.
- `GITHUB_TOKEN` is provided automatically by Actions — no secret to configure
  for a binaries-only release.
