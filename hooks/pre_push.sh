#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"


# Full E2E regression shield
# Run unit tests with race detector (fast: ~20s)
# Run e2e tests without race detector (integration tests, race detector adds minimal value)
echo "Running unit tests with race detector..."
go test -race -short -count=1 -timeout=3m ./cmd/fluxid/... ./internal/...

echo "Running e2e integration tests..."
go test -count=1 -timeout=3m ./e2e-tests/...
