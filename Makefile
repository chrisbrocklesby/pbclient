SHELL := /bin/bash
GO ?= go
PB_BIN ?= $(HOME)/pocketbase
PB_DATA ?= pb_data
TAG = $(if $(filter v%,$(VERSION)),$(VERSION),v$(VERSION))

.PHONY: dev test cli pb-serve release

## Run the demo HTTP server (uses default PocketBase credentials configured in cmd/demo/main.go)
http-dev:
	$(GO) run ./cmd/dev

## Run all tests
test:
	$(GO) test ./...

## Build CLI binary for the demo (outputs ./bin/dev)
cli:
	mkdir -p bin
	$(GO) build -o bin/dev ./cmd/dev

## Start local PocketBase (binary at $(PB_BIN), data dir $(PB_DATA))
pb-dev:
	"$(PB_BIN)" serve --http=127.0.0.1:8090 --dir="$(PB_DATA)"

## Create and push a git release tag (usage: make release VERSION=v1.0.4)
release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required (example: make release VERSION=v1.0.4)"; exit 1; fi
	@echo "$(TAG)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$' || (echo "VERSION must be semver like v1.2.3 (or 1.2.3)"; exit 1)
	@echo "Releasing tag $(TAG)"
	@git diff --quiet || (echo "Working tree has unstaged changes"; exit 1)
	@git diff --cached --quiet || (echo "Working tree has staged but uncommitted changes"; exit 1)
	@git rev-parse "refs/tags/$(TAG)" >/dev/null 2>&1 && (echo "Tag $(TAG) already exists"; exit 1) || true
	git tag "$(TAG)"
	git push origin "$(TAG)"
