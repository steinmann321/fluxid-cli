#!/usr/bin/env bash
set -euo pipefail

GLOBAL_THRESHOLD="90"

echo "Running Go tests with coverage..."
COV_FILE="${TMPDIR:-/tmp}/fluxid-coverage.out"
# Exclude e2e-tests from coverage calculation
PKGS=$(go list ./... | grep -v '/e2e-tests/')
go test -failfast -timeout=10m -covermode=atomic -coverprofile="${COV_FILE}" $PKGS 

[[ -f "${COV_FILE}" ]] || { echo "coverage.out not found" >&2; exit 1; }

total_line=$(go tool cover -func="${COV_FILE}" | awk '/^total:/ {print $0}')
coverage=$(awk '/^total:/ {gsub(/%/,"",$3); print $3}' <<< "${total_line}")

awk -v cov="${coverage}" -v thresh="${GLOBAL_THRESHOLD}" 'BEGIN { if (cov < thresh) exit 1 }' || {
  echo "✗ Coverage gate failed: ${coverage}% (minimum ${GLOBAL_THRESHOLD}%)." >&2
  echo "  Use the 1/3 happy : 2/3 unhappy rule when writing tests:" >&2
  echo "  - One third of tests should assert successful, happy-path behavior" >&2
  echo "  - Two thirds should target edge cases, failures, and resilience" >&2
  echo "  Strengthen assertions and add failure-mode coverage until the gate passes." >&2
  exit 1
}

echo "✓ Coverage OK: ${coverage}%"

