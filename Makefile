BINARY := structalign
PKG    := ./cmd/structalign
SAMPLE := ./_example

.DEFAULT_GOAL := build

.PHONY: build
build: ## Build the binary into ./$(BINARY)
	go build -o $(BINARY) $(PKG)

.PHONY: install
install: ## Install the binary into $GOBIN / $GOPATH/bin
	go install $(PKG)

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: fmt
fmt: ## Format the code in place
	gofmt -w cmd _example

.PHONY: fmt-check
fmt-check: ## Fail if any file is not gofmt'd
	@unformatted=$$(gofmt -l cmd _example); \
	if [ -n "$$unformatted" ]; then \
		echo "Not gofmt'd:"; echo "$$unformatted"; exit 1; \
	fi

.PHONY: smoke
smoke: ## Exercise both modes against the bundled sample
	go run $(PKG) -inspect $(SAMPLE)
	@if go run $(PKG) $(SAMPLE); then \
		echo "expected non-zero exit when reorderings are found"; exit 1; \
	fi

.PHONY: check
check: fmt-check vet build smoke ## Run everything CI runs

.PHONY: changelog
changelog: ## Regenerate CHANGELOG.md from commit history (needs git-cliff)
	git-cliff -o CHANGELOG.md

.PHONY: changelog-unreleased
changelog-unreleased: ## Print the pending (unreleased) changelog entries
	git-cliff --unreleased

.PHONY: release
release: ## Write CHANGELOG.md for a release: make release TAG=v0.1.0
	@test -n "$(TAG)" || { echo "usage: make release TAG=vX.Y.Z"; exit 1; }
	git-cliff --tag $(TAG) -o CHANGELOG.md

.PHONY: clean
clean: ## Remove the built binary
	rm -f $(BINARY)

.PHONY: help
help: ## List targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort | awk 'BEGIN {FS = ":.*?## "} {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
