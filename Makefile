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

VIS_VERSION ?= 9.1.9
VIS_DIR     := internal/output/html/assets
VIS_FILE    := vis-network.min.js
VIS_URL     := https://unpkg.com/vis-network@$(VIS_VERSION)/standalone/umd/$(VIS_FILE)

.PHONY: all build run install test vet fmt tidy clean snapshot verify-assets vendor-vis vendor-vis-update

all: tidy vet verify-assets build

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

verify-assets: ## Verify the vendored vis-network bundle matches its recorded SHA-256
	cd $(VIS_DIR) && sha256sum -c $(VIS_FILE).sha256

vendor-vis: ## Re-fetch vis-network from upstream and verify it still matches the pin
	curl -fsSL $(VIS_URL) -o $(VIS_DIR)/$(VIS_FILE)
	$(MAKE) verify-assets

vendor-vis-update: ## Re-fetch vis-network and rewrite the pin (use when bumping VIS_VERSION)
	curl -fsSL $(VIS_URL) -o $(VIS_DIR)/$(VIS_FILE)
	cd $(VIS_DIR) && sha256sum $(VIS_FILE) > $(VIS_FILE).sha256
	@echo "updated $(VIS_FILE) + checksum; now update version/size/sha in $(VIS_DIR)/README.md"
