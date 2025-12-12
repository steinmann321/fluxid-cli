#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"


# Full E2E regression shield: run race-enabled tests for all packages
echo "Running full test suite with race detector..."
go test -race -count=1 -timeout=15m ./...
