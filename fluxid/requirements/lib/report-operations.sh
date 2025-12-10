#!/usr/bin/env bash
# Report Operations Module
# Handles all report file parsing, validation, and status checking

# Check if report file exists
# Args: $1 - report file path
# Returns: 0 if exists, 1 if missing
report_exists() {
  local report_file="$1"
  [[ -f "$report_file" ]]
}

# Parse status field from report
# Args: $1 - report file path
# Returns: status string (PASS/FAIL) or empty if missing/invalid
# Prints: status to stdout
parse_report_status() {
  local report_file="$1"

  if [[ ! -f "$report_file" ]]; then
    return 1
  fi

  local status
  status="$(parse_yaml_field "$report_file" "status" 2>/dev/null || true)"

  if [[ -n "$status" ]]; then
    echo "$status"
    return 0
  fi

  return 1
}

# Parse command field from report
# Args: $1 - report file path
# Returns: command string or empty if missing/invalid
# Prints: command to stdout
parse_report_command() {
  local report_file="$1"

  if [[ ! -f "$report_file" ]]; then
    return 1
  fi

  local command
  command="$(parse_yaml_field "$report_file" "command" 2>/dev/null || true)"

  if [[ -n "$command" ]]; then
    echo "$command"
    return 0
  fi

  return 1
}

# Parse artifact field from report
# Args: $1 - report file path
# Returns: artifact string or empty if missing/invalid
# Prints: artifact to stdout
parse_report_artifact() {
  local report_file="$1"

  if [[ ! -f "$report_file" ]]; then
    return 1
  fi

  local artifact
  artifact="$(parse_yaml_field "$report_file" "artifact" 2>/dev/null || true)"

  if [[ -n "$artifact" ]]; then
    echo "$artifact"
    return 0
  fi

  return 1
}

# Validate report structure (YAML validity)
# Args: $1 - report file path
# Returns: 0 if valid, 1 if invalid
validate_report_structure() {
  local report_file="$1"
  local validation_script="$PROJECT_ROOT/.fluxid/scripts/command/validate-report.sh"

  if [[ ! -f "$report_file" ]]; then
    return 1
  fi

  if [[ ! -x "$validation_script" ]]; then
    # Fallback: basic check if validation script missing
    parse_yaml_field "$report_file" "command" >/dev/null 2>&1
    return $?
  fi

  "$validation_script" "$report_file" >/dev/null 2>&1
}

# Validate report metadata matches expectations
# Args: $1 - report file path
#       $2 - expected command name
#       $3 - expected artifact token
# Returns: 0 if match, 1 if mismatch
validate_report_metadata() {
  local report_file="$1"
  local expected_command="$2"
  local expected_artifact="$3"

  if [[ ! -f "$report_file" ]]; then
    return 1
  fi

  local actual_command
  local actual_artifact

  actual_command="$(parse_report_command "$report_file")"
  actual_artifact="$(parse_report_artifact "$report_file")"

  if [[ "$actual_command" != "$expected_command" ]]; then
    return 1
  fi

  if [[ "$actual_artifact" != "$expected_artifact" ]]; then
    return 1
  fi

  return 0
}

# Check if this is a loop restart scenario
# If report exists with PASS status for same artifact, delete it
# Args: $1 - report file path
#       $2 - epic id (with .md extension)
#       $3 - current command name
# Returns: 0 always (side effect: may delete report)
check_loop_restart() {
  local report_file="$1"
  local epic_id="$2"
  local current_command="$3"

  if [[ ! -f "$report_file" ]]; then
    return 0
  fi

  local report_command
  local report_status
  local report_artifact
  local artifact_token

  report_command="$(parse_report_command "$report_file" || true)"
  report_status="$(parse_report_status "$report_file" || true)"
  report_artifact="$(parse_report_artifact "$report_file" || true)"
  artifact_token="${epic_id%.md}"

  # If report is from current command, PASS status, and same artifact
  # This indicates a loop restart - delete report to start fresh
  if [[ "$report_command" == "$current_command" && \
        "$report_status" == "PASS" && \
        "$report_artifact" == "$artifact_token" ]]; then
    rm -f "$report_file" 2>/dev/null || true
  fi

  return 0
}
