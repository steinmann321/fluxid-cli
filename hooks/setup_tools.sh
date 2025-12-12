#!/usr/bin/env bash
set -euo pipefail

# Installs required tooling for strict pre-commit gate
# Tools: gofumpt, goimports, golangci-lint, gosec, govulncheck, gitleaks, ruleguard
need() { command -v "$1" >/dev/null 2>&1 || return 0; return 1; }

GO_BIN_DIR="$(go env GOPATH 2>/dev/null)/bin"
mkdir -p "${GO_BIN_DIR}"

install_go_tool() {
  local pkg="$1"
  echo "Installing ${pkg} ..."
  GO111MODULE=on go install "${pkg}"
}

# gofumpt
if ! command -v gofumpt >/dev/null 2>&1; then
  install_go_tool mvdan.cc/gofumpt@latest
else
  echo "gofumpt already installed"
fi

# goimports
if ! command -v goimports >/dev/null 2>&1; then
  install_go_tool golang.org/x/tools/cmd/goimports@latest
else
  echo "goimports already installed"
fi

# gosec
if ! command -v gosec >/dev/null 2>&1; then
  install_go_tool github.com/securego/gosec/v2/cmd/gosec@latest
else
  echo "gosec already installed"
fi

# govulncheck
if ! command -v govulncheck >/dev/null 2>&1; then
  install_go_tool golang.org/x/vuln/cmd/govulncheck@latest
else
  echo "govulncheck already installed"
fi

# gitleaks
if ! command -v gitleaks >/dev/null 2>&1; then
  if command -v brew >/dev/null 2>&1; then
    echo "Installing gitleaks via Homebrew..."
    brew install gitleaks
  else
    echo "Installing gitleaks via 'go install' (v8)..."
    install_go_tool github.com/gitleaks/gitleaks/v8@latest
  fi
else
  echo "gitleaks already installed"
fi

# golangci-lint (install latest official binary with ruleguard bundled)
GCL_BIN="$(go env GOPATH 2>/dev/null)/bin/golangci-lint"
install_golangci_official() {
  local dest_bin
  dest_bin="$(go env GOPATH 2>/dev/null)/bin"
  echo "Installing golangci-lint (latest) via official installer..."
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$dest_bin"
}

if ! command -v golangci-lint >/dev/null 2>&1; then
  install_golangci_official
else
  # Verify ruleguard is available in the current binary; reinstall if missing
  if ! golangci-lint help linters 2>/dev/null | grep -q "ruleguard"; then
    install_golangci_official
  else
    echo "golangci-lint already installed with ruleguard"
  fi
fi

# ruleguard
if ! command -v ruleguard >/dev/null 2>&1; then
  echo "Installing ruleguard via 'go install'..."
  GO111MODULE=on go install github.com/quasilyte/go-ruleguard/cmd/ruleguard@latest
else
  echo "ruleguard already installed"
fi

# PATH hint
if ! echo ":$PATH:" | grep -q ":${GO_BIN_DIR}:"; then
  echo
  echo "NOTE: Add ${GO_BIN_DIR} to your PATH so installed tools are found."
  echo "Example: export PATH=\"${GO_BIN_DIR}:$PATH\""
fi

echo "All required tools are installed or already present."