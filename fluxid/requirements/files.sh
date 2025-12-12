#!/usr/bin/env bash
# File path manager for workflow files
# Usage: ./files.sh --report|--history|--testfile <epic-id>
# Returns absolute path for the requested file type

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# Support test mode with isolated directories
# When EPIC_LOOP_TEST_ROOT is set, use it instead of PROJECT_ROOT/fluxid
if [[ -n "${EPIC_LOOP_TEST_ROOT:-}" ]]; then
  FLUXID_ROOT="$EPIC_LOOP_TEST_ROOT"
else
  FLUXID_ROOT="$PROJECT_ROOT/fluxid"
fi

# Optional project env (for test root etc.)
ENV_FILE="$FLUXID_ROOT/config/env.sh"
if [[ -f "$ENV_FILE" ]]; then
  # shellcheck source=/dev/null
  source "$ENV_FILE"
fi

FLAG="${1:-}"

case "$FLAG" in
  --report)
    echo "$FLUXID_ROOT/reports/workflow-report.yaml"
    ;;
  --history)
    echo "$FLUXID_ROOT/tmp/workflow-loop-history.md"
    ;;
  --testfile)
    if [[ $# -lt 2 ]]; then
      echo "Error: --testfile requires <epic-id> argument" >&2
      echo "Usage: $0 --testfile mXX-eYY-<slug>.md" >&2
      exit 1
    fi
    EPIC_ID="$2"

    EPIC_BASENAME="$(basename "$EPIC_ID")"
    if [[ "$EPIC_BASENAME" != m??-e??-*.md ]]; then
      echo "Error: epic id must match pattern mXX-eYY-<slug>.md (got: $EPIC_BASENAME)" >&2
      exit 1
    fi

    EPIC_PATH="$FLUXID_ROOT/epics/$EPIC_BASENAME"
    if [[ ! -f "$EPIC_PATH" ]]; then
      echo "Error: epic file not found for id '$EPIC_BASENAME' at $EPIC_PATH" >&2
      exit 1
    fi

    TEST_DIR="${E2E_TEST_ROOT:-"$PROJECT_ROOT/e2e-tests/tests"}"
    mkdir -p "$TEST_DIR"

    TEST_BASENAME="${EPIC_BASENAME%.md}.spec.ts"
    TEST_PATH="$TEST_DIR/$TEST_BASENAME"

    if [[ ! -f "$TEST_PATH" ]]; then
      : > "$TEST_PATH"
    fi

    echo "$TEST_PATH"
    ;;
  *)
    echo "Error: Flag required" >&2
    echo "Usage: $0 --report|--history|--testfile <epic-id>" >&2
    echo "" >&2
    echo "Flags:" >&2
    echo "  --report             Get workflow report file path" >&2
    echo "  --history            Get workflow history file path" >&2
    echo "  --testfile <epic-id> Get or create deterministic test file for epic id (mXX-eYY-<slug>.md)" >&2
    exit 1
    ;;
esac
