#!/usr/bin/env bash
# Progress manager for workflow progress.yaml
# Usage:
#   ./progress.sh --file
#   ./progress.sh --epic-complete <epic-id>
#   ./progress.sh --milestone-complete <milestone-id>
#
# Notes:
#   - Follows the same PROJECT_ROOT resolution pattern as files.sh
#   - Uses lib/progress-yaml.sh for actual YAML updates (requires yq)

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

resolve_progress_file() {
  local candidates=(
    "$FLUXID_ROOT/progress.yaml"
    "$PROJECT_ROOT/.fluxid/progress.yaml"
    "$SCRIPT_DIR/../progress.yaml"
    "$SCRIPT_DIR/../../progress.yaml"
  )

  for candidate in "${candidates[@]}"; do
    if [[ -f "$candidate" ]]; then
      echo "$candidate"
      return 0
    fi
  done

  echo "Error: progress.yaml not found. Checked:" >&2
  for candidate in "${candidates[@]}"; do
    echo " - $candidate" >&2
  done
  return 1
}

print_usage() {
  cat <<EOF
Usage: $(basename "$0") <flag> [args]

Flags:
  --file                         Print resolved progress.yaml path
  --epic-complete <epic-id>      Mark epic as complete in progress.yaml (id like m01-e01)
  --milestone-complete <id>      Mark milestone as complete in progress.yaml (id like m01)
  --epic-status-get <epic-id>    Get status for epic (id like m01-e01)
  --epic-status-set <epic-id> <status>
                                 Set status for epic (e.g., pending|implement|review|done|error)
  --milestone-status-get <id>    Get status for milestone (id like m01)
  --milestone-status-set <id> <status>
                                 Set status for milestone (e.g., pending|in_progress|done|error)
  -h, --help                     Show this help message
EOF
}

FLAG="${1:-}"

case "$FLAG" in
  --file)
    PROGRESS_FILE="$(resolve_progress_file)" || exit 1
    echo "$PROGRESS_FILE"
    ;;

  --epic-complete)
    if [[ $# -lt 2 ]]; then
      echo "Error: --epic-complete requires <epic-id> argument (e.g., m01-e01)" >&2
      print_usage
      exit 1
    fi

    EPIC_ID="$2"
    PROGRESS_FILE="$(resolve_progress_file)" || exit 1

    # Only source progress-yaml when we actually need it (yq required)
    # shellcheck source=/dev/null
    source "$PROJECT_ROOT/.fluxid/scripts/lib/progress-yaml.sh"

    mark_epic_complete "$PROGRESS_FILE" "$EPIC_ID"
    ;;

  --milestone-complete)
    if [[ $# -lt 2 ]]; then
      echo "Error: --milestone-complete requires <milestone-id> argument (e.g., m01)" >&2
      print_usage
      exit 1
    fi

    MILESTONE_ID="$2"
    PROGRESS_FILE="$(resolve_progress_file)" || exit 1

    # shellcheck source=/dev/null
    source "$PROJECT_ROOT/.fluxid/scripts/lib/progress-yaml.sh"

    mark_milestone_complete "$PROGRESS_FILE" "$MILESTONE_ID"
    ;;
  --epic-status-get)
    if [[ $# -lt 2 ]]; then
      echo "Error: --epic-status-get requires <epic-id> argument (e.g., m01-e01)" >&2
      print_usage
      exit 1
    fi

    EPIC_ID="$2"
    PROGRESS_FILE="$(resolve_progress_file)" || exit 1
    yq -r ".milestones[].epics[] | select(.id == \"$EPIC_ID\") | .status" "$PROGRESS_FILE" 2>/dev/null || true
    ;;

  --epic-status-set)
    if [[ $# -lt 3 ]]; then
      echo "Error: --epic-status-set requires <epic-id> and <status> arguments" >&2
      print_usage
      exit 1
    fi

    EPIC_ID="$2"
    STATUS="$3"
    PROGRESS_FILE="$(resolve_progress_file)" || exit 1
    yq -i "(.milestones[].epics[] | select(.id == \"$EPIC_ID\") | .status) = \"$STATUS\"" "$PROGRESS_FILE"
    ;;

  --milestone-status-get)
    if [[ $# -lt 2 ]]; then
      echo "Error: --milestone-status-get requires <milestone-id> argument (e.g., m01)" >&2
      print_usage
      exit 1
    fi

    MILESTONE_ID="$2"
    PROGRESS_FILE="$(resolve_progress_file)" || exit 1
    yq -r ".milestones[] | select(.id == \"$MILESTONE_ID\") | .status" "$PROGRESS_FILE" 2>/dev/null || true
    ;;

  --milestone-status-set)
    if [[ $# -lt 3 ]]; then
      echo "Error: --milestone-status-set requires <milestone-id> and <status> arguments" >&2
      print_usage
      exit 1
    fi

    MILESTONE_ID="$2"
    STATUS="$3"
    PROGRESS_FILE="$(resolve_progress_file)" || exit 1
    yq -i "(.milestones[] | select(.id == \"$MILESTONE_ID\") | .status) = \"$STATUS\"" "$PROGRESS_FILE"
    ;;

  -h|--help|help)
    print_usage
    ;;

  *)
    echo "Error: Unknown or missing flag: ${FLAG:-<none>}" >&2
    print_usage
    exit 1
    ;;
esac
