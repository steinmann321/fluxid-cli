#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# Ensure GOPATH/bin is on PATH so installed tools are found
GOBIN_DIR="$(go env GOPATH 2>/dev/null)/bin"
if [[ -n "$GOBIN_DIR" ]]; then
  export PATH="$GOBIN_DIR:$PATH"
fi

# Collect staged changes (compatible with macOS bash)
CHANGED=$(git diff --cached --name-only)
if [[ -z "$CHANGED" ]]; then
  exit 0
fi

# Limit scope to project files: src/** and e2e-test/**
PROJECT_CHANGED=$(printf '%s\n' "$CHANGED" | grep -E '^(src/|e2e-test/)' || true)
if [[ -z "$PROJECT_CHANGED" ]]; then
  # No project files staged; skip checks
  exit 0
fi

# Format staged Go files only in project scope
GO_FILES=$(printf '%s\n' "$PROJECT_CHANGED" | grep -E '\.go$' || true)
if [[ -n "$GO_FILES" ]]; then
  command -v gofumpt >/dev/null || { echo "gofumpt not installed"; exit 1; }
  command -v goimports >/dev/null || { echo "goimports not installed"; exit 1; }
  echo "Formatting Go files with goimports/gofumpt..."
  for f in $GO_FILES; do
    [[ -f "$f" ]] || continue
    # Run goimports first, then gofumpt to canonicalize formatting
    goimports -w "$f"
    gofumpt -l -w "$f"
    git add "$f"
  done
fi


# Format staged Go files
GO_FILES=$(printf '%s\n' "$CHANGED" | grep -E '\.go$' || true)
if [[ -n "$GO_FILES" ]]; then
  command -v gofumpt >/dev/null || { echo "gofumpt not installed"; exit 1; }
  command -v goimports >/dev/null || { echo "goimports not installed"; exit 1; }
  echo "Formatting Go files with goimports/gofumpt..."
  for f in $GO_FILES; do
    [[ -f "$f" ]] || continue
    # Run goimports first, then gofumpt to canonicalize formatting
    goimports -w "$f"
    gofumpt -l -w "$f"
    git add "$f"
  done
fi

# Enforce go.mod/go.sum tidy
before=$(git status --porcelain)
go mod tidy
after=$(git status --porcelain)
if [[ "$before" != "$after" ]]; then
  echo "go.mod/go.sum not tidy. Run 'go mod tidy' and commit changes." >&2
  exit 1
fi

# Build (native)
if make -q build >/dev/null 2>&1; then
  echo "Building project..."
  make build
fi

# Max lines check
./hooks/check_max_lines.sh $PROJECT_CHANGED

# Secrets scanning
command -v gitleaks >/dev/null || { echo "gitleaks not installed"; exit 1; }
echo "Running gitleaks on staged changes..."
# Limit scan to project directories only
gitleaks protect --no-banner --staged --source .

# Linting (strict)
GL_BIN="$(command -v golangci-lint || true)"
if [[ -z "$GL_BIN" ]]; then
  echo "golangci-lint not installed"; exit 1
fi
# Require v2 binary; fallback to brew path if available
if ! "$GL_BIN" version 2>/dev/null | grep -q "version 2."; then
  if command -v /opt/homebrew/bin/golangci-lint >/dev/null 2>&1 && /opt/homebrew/bin/golangci-lint version 2>/dev/null | grep -q "version 2."; then
    GL_BIN="/opt/homebrew/bin/golangci-lint"
  else
    echo "golangci-lint v2 required; found v1. Please install v2 (see https://golangci-lint.run/usage/install/)." >&2
    exit 1
  fi
fi
# Auto-format via golangci-lint fmt (v2)
GO_FILES=$(printf '%s
' "$PROJECT_CHANGED" | grep -E '\.go$' || true)
if [[ -n "$GO_FILES" ]]; then
  echo "Formatting with golangci-lint fmt..."
  "$GL_BIN" fmt $GO_FILES || true
  for f in $GO_FILES; do
    [[ -f "$f" ]] && git add "$f"
  done
fi

echo "Running golangci-lint..."
# Limit lints to project dirs
"$GL_BIN" run ./src/... ./e2e-test/...

# Security static analysis
command -v gosec >/dev/null || { echo "gosec not installed"; exit 1; }
echo "Running gosec..."
gosec ./src/... ./e2e-test/...

# Vulnerability advisories
command -v govulncheck >/dev/null || { echo "govulncheck not installed"; exit 1; }
echo "Running govulncheck..."
govulncheck ./src/... ./e2e-test/...

# Coverage checks
./hooks/check_coverage.sh
