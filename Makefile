.PHONY: build install uninstall clean test help

# Binary name
BINARY_NAME=fluxid

# Installation directories
INSTALL_DIR=/usr/local/bin
BUILD_DIR=bin

# Go build flags
LDFLAGS=-ldflags "-s -w"

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the fluxid binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/fluxid
	@echo "✓ Built $(BUILD_DIR)/$(BINARY_NAME)"

install: build ## Install fluxid to /usr/local/bin (requires sudo)
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "✓ Run 'fluxid --help' to verify installation"

install-user: build ## Install fluxid to ~/bin (no sudo required)
	@echo "Installing $(BINARY_NAME) to ~/bin..."
	@mkdir -p ~/bin
	@install -m 755 $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)
	@echo "✓ Installed to ~/bin/$(BINARY_NAME)"
	@if echo $$PATH | grep -q "$$HOME/bin"; then \
		echo "✓ ~/bin is in your PATH"; \
		echo "✓ Run 'fluxid --help' to verify installation"; \
	else \
		echo "⚠ ~/bin is NOT in your PATH"; \
		echo "  Add this to your ~/.zshrc or ~/.bashrc:"; \
		echo "    export PATH=\"\$$HOME/bin:\$$PATH\""; \
		echo "  Then run: source ~/.zshrc"; \
	fi

uninstall: ## Uninstall fluxid from /usr/local/bin
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_DIR)..."
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Uninstalled from $(INSTALL_DIR)"

uninstall-user: ## Uninstall fluxid from ~/bin
	@echo "Uninstalling $(BINARY_NAME) from ~/bin..."
	@rm -f ~/bin/$(BINARY_NAME)
	@echo "✓ Uninstalled from ~/bin"

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) .tmp
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

.DEFAULT_GOAL := help
