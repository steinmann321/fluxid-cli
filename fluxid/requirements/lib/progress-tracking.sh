#!/usr/bin/env bash
# Progress Tracking Module
# Handles all progress.yaml read/write operations

# Set task status in progress.yaml
# Args: $1 - epic id token (e.g., m01-e03)
#       $2 - status (pending|implement|review|done|error)
# Returns: 0 if success, 1 if failed or progress command not found
set_task_status() {
  local epic_id_token="$1"
  local status="$2"
  local progress_cmd="${PROGRESS_CMD:-}"

  if [[ -z "$epic_id_token" ]]; then
    return 1
  fi

  if [[ -z "$progress_cmd" || ! -x "$progress_cmd" ]]; then
    return 0  # Soft fail if progress command not available
  fi

  "$progress_cmd" --epic-status-set "$epic_id_token" "$status" 2>/dev/null || return 1
}

# Get current task status from progress.yaml
# Args: $1 - epic id token (e.g., m01-e03)
# Returns: 0 if success, 1 if failed
# Prints: status string to stdout
get_task_status() {
  local epic_id_token="$1"
  local progress_cmd="${PROGRESS_CMD:-}"

  if [[ -z "$epic_id_token" ]]; then
    return 1
  fi

  if [[ -z "$progress_cmd" || ! -x "$progress_cmd" ]]; then
    return 1
  fi

  local status
  status="$("$progress_cmd" --epic-status-get "$epic_id_token" 2>/dev/null || true)"

  if [[ -n "$status" ]]; then
    echo "$status"
    return 0
  fi

  return 1
}

# Mark task as complete in progress.yaml
# Args: $1 - epic id token (e.g., m01-e03)
# Returns: 0 if success, 1 if failed
mark_task_complete() {
  local epic_id_token="$1"
  local progress_cmd="${PROGRESS_CMD:-}"

  if [[ -z "$epic_id_token" ]]; then
    return 1
  fi

  if [[ -z "$progress_cmd" || ! -x "$progress_cmd" ]]; then
    return 0  # Soft fail if progress command not available
  fi

  "$progress_cmd" --epic-complete "$epic_id_token" 2>/dev/null || return 1
}

# Check if task should be skipped (already done)
# Args: $1 - epic id token (e.g., m01-e03)
# Returns: 0 if should skip, 1 if should process
should_skip_task() {
  local epic_id_token="$1"
  local current_status

  current_status="$(get_task_status "$epic_id_token" 2>/dev/null || true)"

  if [[ "$current_status" == "done" ]]; then
    return 0  # Skip
  fi

  return 1  # Process
}
