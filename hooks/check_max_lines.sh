#!/usr/bin/env bash
set -euo pipefail

MAX_LINES=400
status=0

for file in "$@"; do
  # Only enforce on cmd/**, internal/**, pkg/**, and e2e-tests/**
  case "$file" in
    cmd/*|internal/*|pkg/*|e2e-tests/*) ;;
    *) continue ;;
  esac

  case "$file" in
    *.go|*.sh|*.md|*.yml|*.yaml|*.txt|Makefile)
      ;;
    *)
      continue
      ;;
  esac

  [[ -f "$file" ]] || continue

  lines=$(wc -l < "$file" | tr -d '[:space:]')
  if [[ "$lines" -gt "$MAX_LINES" ]]; then
    echo "✗ $file has ${lines} lines (maximum ${MAX_LINES})." >&2
    echo "  Enforce modularization: split this file into cohesive components" >&2
    echo "  with clear boundaries (functions, small types, and packages)." >&2
    echo "  Aim for perfectly modular design: no monoliths, single responsibility," >&2
    echo "  and reusable units. Refactor before committing." >&2
    status=1
  fi
done

if [[ "$status" -ne 0 ]]; then
  echo "Pre-commit 'max-lines' failed: some files exceed ${MAX_LINES} lines." >&2
  exit "$status"
fi
