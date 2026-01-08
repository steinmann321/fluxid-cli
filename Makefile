.PHONY: build install uninstall clean test help

# Binary name
BINARY_NAME=fluxid

# Installation directory - CANONICAL across all platforms
# Use GOPATH/bin for cross-platform consistency (Windows, macOS, Linux)
# Package managers (Homebrew, apt, rpm, Chocolatey) handle their own paths
INSTALL_DIR=$(shell go env GOPATH)/bin

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build fluxid using GoReleaser (local platform only)
	@echo "Building $(BINARY_NAME) with GoReleaser..."
	@goreleaser build --snapshot --clean --single-target
	@echo "✓ Built dist/$(BINARY_NAME)_*_*/$(BINARY_NAME)"

install: build ## Install fluxid to $$GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp dist/$(BINARY_NAME)_*_*/$(BINARY_NAME)* $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "✓ Run 'fluxid version' to verify installation"

uninstall: ## Uninstall fluxid from $$GOPATH/bin
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME).exe
	@echo "✓ Uninstalled from $(INSTALL_DIR)"

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf dist .tmp
	@rm -f coverage*.out .tmp*.out
	@echo "✓ Cleaned"

test: ## Run all tests
	@echo "Running unit tests with race detector..."
	@go test -race -short -count=1 -timeout=3m ./cmd/fluxid/... ./internal/...
	@echo "Running e2e integration tests..."
	@go test -count=1 -timeout=3m ./e2e-tests/...
	@echo "✓ All tests passed"

test-e2e: ## Run only e2e tests
	@echo "Running e2e integration tests..."
	@go test -count=1 -timeout=3m ./e2e-tests/...

test-unit: ## Run only unit tests
	@echo "Running unit tests with race detector..."
	@go test -race -short -count=1 -timeout=3m ./cmd/fluxid/... ./internal/...

coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p .tmp/coverage
	@go test -short -covermode=atomic -coverprofile=.tmp/coverage/coverage.out ./cmd/fluxid/... ./internal/...
	@go tool cover -func=.tmp/coverage/coverage.out | grep total:

release-test: ## Test full release build (all platforms)
	@echo "Building release snapshot for all platforms..."
	@goreleaser build --snapshot --clean
	@echo "✓ Built binaries for all platforms in dist/"

.DEFAULT_GOAL := help
