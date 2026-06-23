BINARY      := infrasight
PKG         := github.com/Aakhri-Pastaa/infrasight
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.1.0-dev)
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE        ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w \
	-X $(PKG)/pkg/version.Version=$(VERSION) \
	-X $(PKG)/pkg/version.Commit=$(COMMIT) \
	-X $(PKG)/pkg/version.Date=$(DATE)

GO          ?= go

.PHONY: all build run install test vet fmt tidy clean snapshot

all: tidy vet build

build: ## Build the binary into ./bin (Linux target)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BINARY) ./cmd/infrasight

run: build ## Build then run a scan
	./bin/$(BINARY) scan

install: ## Install into $GOBIN / $GOPATH/bin
	CGO_ENABLED=0 $(GO) install -trimpath -ldflags '$(LDFLAGS)' ./cmd/infrasight

test: ## Run unit tests
	$(GO) test ./...

vet: ## Static analysis
	$(GO) vet ./...

fmt: ## Format all sources
	$(GO) fmt ./...

tidy: ## Sync go.mod / go.sum
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -rf bin dist

snapshot: ## Build a local multi-platform snapshot via GoReleaser
	goreleaser release --snapshot --clean
