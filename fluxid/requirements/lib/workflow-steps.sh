#!/usr/bin/env bash
# Workflow Steps Module
# Handles execution of implement, commit, and review steps via agent

# Invoke implement step via agent
# Args: $1 - epic id (filename with .md)
#       $2 - report file path
# Returns: 0 if agent execution succeeded, 1 if failed
invoke_implement_step() {
  local epic_id="$1"
  local report_file="$2"
  local implement_cmd_file="${IMPLEMENT_COMMAND_FILE:-}"
  local streaming_script="${STREAMING_SCRIPT:-}"

  if [[ -z "$streaming_script" || ! -f "$streaming_script" ]]; then
    return 1
  fi

  local instruction
  instruction="Read and execute the instructions in \`$implement_cmd_file\` for epic id \`$epic_id\`."
  instruction+=" Use the shared workflow report protocol: write a PURE YAML report to the path from \`.fluxid/scripts/command/files.sh --report\`"
  instruction+=" following \`.fluxid/templates/report-schema.yaml\` and validate it with \`.fluxid/scripts/command/validate-report.sh\`."

  # Execute agent with instruction
  if ! "$streaming_script" "$instruction"; then
    return 1
  fi

  return 0
}

# Invoke commit step via agent
# Returns: 0 if commit succeeded, 1 if failed
invoke_commit_step() {
  local commit_script="${COMMIT_CMD:-}"
  local streaming_script="${STREAMING_SCRIPT:-}"

  if [[ -z "$commit_script" || ! -x "$commit_script" ]]; then
    return 1
  fi

  # Execute commit script (which internally calls the agent)
  if ! STREAMING_SCRIPT="$streaming_script" "$commit_script"; then
    return 1
  fi

  return 0
}

# Invoke review step via agent
# Args: $1 - epic id (filename with .md)
#       $2 - report file path
# Returns: 0 if agent execution succeeded, 1 if failed
invoke_review_step() {
  local epic_id="$1"
  local report_file="$2"
  local review_cmd_file="${VALIDATE_COMMAND_FILE:-}"
  local streaming_script="${STREAMING_SCRIPT:-}"

  if [[ -z "$streaming_script" || ! -f "$streaming_script" ]]; then
    return 1
  fi

  local instruction
  instruction="Read and execute the instructions in \`$review_cmd_file\` for epic id \`$epic_id\`."
  instruction+=" Use the shared workflow report protocol: write a PURE YAML report to the path from \`.fluxid/scripts/command/files.sh --report\`"
  instruction+=" following \`.fluxid/templates/report-schema.yaml\` and validate it with \`.fluxid/scripts/command/validate-report.sh\`."

  # Execute agent with instruction
  if ! "$streaming_script" "$instruction"; then
    return 1
  fi

  return 0
}

# Retry a workflow step if report is missing
# Args: $1 - step function name (e.g., invoke_implement_step)
#       $2 - epic id
#       $3 - report file path
#       $4 - max retries (default 3)
# Returns: 0 if succeeded (report exists), 1 if all retries failed
retry_with_missing_report() {
  local step_function="$1"
  local epic_id="$2"
  local report_file="$3"
  local max_retries="${4:-3}"
  local retry_count=0

  while [[ $retry_count -le $max_retries ]]; do
    if [[ $retry_count -gt 0 ]]; then
      log "Retry attempt $retry_count/$max_retries for missing report"
    fi

    # Execute step function
    if ! "$step_function" "$epic_id" "$report_file"; then
      return 1  # Agent execution failed
    fi

    # Check if report exists
    if report_exists "$report_file"; then
      return 0  # Success
    fi

    # Report missing, retry
    if [[ $retry_count -lt $max_retries ]]; then
      log "Report missing after attempt $((retry_count + 1))/$((max_retries + 1)), retrying..."
      retry_count=$((retry_count + 1))
    else
      return 1  # All retries exhausted
    fi
  done

  return 1
}
