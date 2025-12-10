#!/usr/bin/env bash
# Validates fluxid/progress.yaml structure using yq (no Python).
# Usage: ./validate-progress.sh <progress-file-path>
# Exit: 0 on success, 1 on validation failure
#
# Conceptual schema: .fluxid/templates/progress-schema.yaml

set -euo pipefail

PROGRESS_FILE="${1:-}"

if [[ -z "$PROGRESS_FILE" ]]; then
  echo "Error: No progress file specified" >&2
  echo "Usage: $0 <progress-file-path>" >&2
  exit 1
fi

if [[ ! -f "$PROGRESS_FILE" ]]; then
  echo "Error: Progress file not found: $PROGRESS_FILE" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
SCHEMA_PATH="$PROJECT_ROOT/.fluxid/templates/progress-schema.yaml"

if [[ ! -f "$SCHEMA_PATH" ]]; then
  echo "Error: Progress schema not found: $SCHEMA_PATH" >&2
  exit 1
fi

if ! command -v yq >/dev/null 2>&1; then
  echo "Error: yq is not installed. Please install it (e.g., brew install yq)." >&2
  exit 1
fi

errors=()

add_error() {
  errors+=("$1")
}

# Basic required fields
project_val="$(yq -r '.project // ""' "$PROGRESS_FILE")"
if [[ -z "$project_val" || "$project_val" == "null" ]]; then
  add_error "Missing or empty required field: project"
fi

last_updated_val="$(yq -r '.last_updated // ""' "$PROGRESS_FILE")"
if [[ -z "$last_updated_val" || "$last_updated_val" == "null" ]]; then
  add_error "Missing or empty required field: last_updated"
elif ! [[ "$last_updated_val" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
  add_error "last_updated must be YYYY-MM-DD, got: $last_updated_val"
fi

has_milestones="$(yq -r 'has("milestones")' "$PROGRESS_FILE" 2>/dev/null || echo "false")"
if [[ "$has_milestones" != "true" ]]; then
  add_error "Missing required field: milestones"
fi

milestones_type="$(yq -r '.milestones | type' "$PROGRESS_FILE" 2>/dev/null || echo "null")"
if [[ "$milestones_type" != "!!seq" ]]; then
  add_error "Field 'milestones' must be an array"
fi

# Validate milestones and epics
milestone_count="$(yq -r '.milestones | length' "$PROGRESS_FILE" 2>/dev/null || echo "0")"

for ((i=0; i<milestone_count; i++)); do
  mid="$(yq -r ".milestones[$i].id // \"\"" "$PROGRESS_FILE")"
  martifact="$(yq -r ".milestones[$i].artifact // \"\"" "$PROGRESS_FILE")"
  mstatus="$(yq -r ".milestones[$i].status // \"\"" "$PROGRESS_FILE")"

  label="milestones[$i]"
  if [[ -n "$mid" && "$mid" != "null" ]]; then
    label="milestone id '$mid'"
  fi

  if [[ -z "$mid" || "$mid" == "null" ]]; then
    add_error "$label missing required field: id"
  fi
  if [[ -z "$martifact" || "$martifact" == "null" ]]; then
    add_error "$label missing required field: artifact"
  fi
  if [[ -z "$mstatus" || "$mstatus" == "null" ]]; then
    add_error "$label missing required field: status"
  else
    case "$mstatus" in
      pending|in_progress|implement|review|done|error) ;;
      *) add_error "$label has invalid status '$mstatus' (allowed: pending, in_progress, implement, review, done, error)" ;;
    esac
  fi

  epics_type="$(yq -r ".milestones[$i].epics | type" "$PROGRESS_FILE" 2>/dev/null || echo "null")"
  if [[ "$epics_type" != "!!seq" ]]; then
    add_error "$label.epics must be an array"
    continue
  fi

  epic_count="$(yq -r ".milestones[$i].epics | length" "$PROGRESS_FILE" 2>/dev/null || echo "0")"
  for ((j=0; j<epic_count; j++)); do
    eid="$(yq -r ".milestones[$i].epics[$j].id // \"\"" "$PROGRESS_FILE")"
    eartifact="$(yq -r ".milestones[$i].epics[$j].artifact // \"\"" "$PROGRESS_FILE")"
    estatus="$(yq -r ".milestones[$i].epics[$j].status // \"\"" "$PROGRESS_FILE")"

    elabel="$label.epics[$j]"
    if [[ -n "$eid" && "$eid" != "null" ]]; then
      elabel="$label epic id '$eid'"
    fi

    if [[ -z "$eid" || "$eid" == "null" ]]; then
      add_error "$elabel missing required field: id"
    fi
    if [[ -z "$eartifact" || "$eartifact" == "null" ]]; then
      add_error "$elabel missing required field: artifact"
    fi
    if [[ -z "$estatus" || "$estatus" == "null" ]]; then
      add_error "$elabel missing required field: status"
    else
      case "$estatus" in
        pending|implement|review|done|error) ;;
        *) add_error "$elabel has invalid status '$estatus' (allowed: pending, implement, review, done, error)" ;;
      esac
    fi
  done
done

if (( ${#errors[@]} > 0 )); then
  echo "Progress validation failed (schema: $SCHEMA_PATH):" >&2
  for e in "${errors[@]}"; do
    echo "  - $e" >&2
  done
  echo >&2
  echo "Progress file: $PROGRESS_FILE" >&2
  exit 1
fi

exit 0
