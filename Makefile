.PHONY: build test lint install clean snapshot help

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -s -w -X github.com/tradekit-dev/tradekit-cli/internal/cmd.version=$(VERSION) \
           -X github.com/tradekit-dev/tradekit-cli/internal/cmd.commit=$(COMMIT) \
           -X github.com/tradekit-dev/tradekit-cli/internal/cmd.date=$(DATE)

## build: Compile the CLI binary
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/tradekit ./cmd/tradekit

## install: Install the CLI to $GOPATH/bin
install:
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./cmd/tradekit

## test: Run all tests
test:
	go test -race ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/ dist/

## snapshot: Build release snapshot (goreleaser)
snapshot:
	goreleaser release --snapshot --clean

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
