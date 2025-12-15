#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# Ensure GOPATH/bin is on PATH so installed tools are found
GOBIN_DIR="$(go env GOPATH 2>/dev/null)/bin"
if [[ -n "$GOBIN_DIR" ]]; then
  export PATH="$GOBIN_DIR:$PATH"
fi

# Collect staged changes, use new paths for renames, ignore deletions
CHANGED_STATUS=$(git diff --cached --name-status)
CHANGED=$(printf '%s\n' "$CHANGED_STATUS" | awk '
  BEGIN { }
  {
    status=$1
    if (status ~ /^D/) { next }
    if (status ~ /^R/) { print $3; next }
    print $2
  }
')
if [[ -z "$CHANGED" ]]; then
  exit 0
fi

# Enforce repository layout: Go code must only exist under cmd/, internal/, pkg/, or e2e-tests/
INVALID_GO=$(printf '%s\n' "$CHANGED" | grep -E '\.go$' | grep -Ev '^(cmd/|internal/|pkg/|e2e-tests/)' || true)
if [[ -n "$INVALID_GO" ]]; then
  echo "\u2717 Invalid repository layout detected:" >&2
  echo "  Go files must reside only under cmd/ (binaries), internal/ (private packages)," >&2
  echo "  pkg/ (public packages), or e2e-tests/ (end-to-end tests)." >&2
  echo "  The following staged files violate this rule:" >&2
  printf '    %s\n' $INVALID_GO >&2
  echo "  Move code into the correct directory before committing." >&2
  exit 1
fi

# Global enforcement: no tracked Go files outside allowed directories
ALL_GO_TRACKED=$(git ls-files "*.go" || true)
# Exclude old paths that are being renamed in this commit
RENAME_OLD=$(printf '%s\n' "$CHANGED_STATUS" | awk '$1=="R" {print $2}')
INVALID_GO_TRACKED=$(printf '%s\n' "$ALL_GO_TRACKED" | while read -r f; do 
  # Skip if file is part of a staged rename (old path)
  if printf '%s\n' "$RENAME_OLD" | grep -Fxq "$f"; then
    continue
  fi
  # Enforce allowed directories
  if ! echo "$f" | grep -Eq '^(cmd/|internal/|pkg/|e2e-tests/)'; then
    echo "$f"
  fi
done)
if [[ -n "$INVALID_GO_TRACKED" ]]; then
  echo "\u2717 Invalid repository layout detected in tracked files:" >&2
  echo "  Go files must reside only under cmd/, internal/, pkg/, or e2e-tests/." >&2
  echo "  The following tracked files violate this rule:" >&2
  printf '    %s\n' $INVALID_GO_TRACKED >&2
  echo "  Move code into the correct directory before committing." >&2
  exit 1
fi



# Limit scope to project files: cmd/**, internal/**, pkg/**, and e2e-tests/**
PROJECT_CHANGED=$(printf '%s\n' "$CHANGED" | grep -E '^(cmd/|internal/|pkg/|e2e-tests/)' || true)
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
"$GL_BIN" run ./... 
# Run ruleguard separately to enforce architecture rules
command -v ruleguard >/dev/null || { echo "ruleguard not installed"; exit 1; }
echo "Running ruleguard..."
ruleguard -rules internal/linters/arch.rules ./...

# Security static analysis
command -v gosec >/dev/null || { echo "gosec not installed"; exit 1; }
echo "Running gosec..."
gosec ./... 
# Vulnerability advisories
command -v govulncheck >/dev/null || { echo "govulncheck not installed"; exit 1; }
echo "Running govulncheck..."
govulncheck ./... 
# Coverage checks
./hooks/check_coverage.sh
