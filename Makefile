SHELL := /bin/bash
GO ?= go
PB_BIN ?= $(HOME)/pocketbase
PB_DATA ?= pb_data

.PHONY: dev test cli pb-serve

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
