#!/bin/bash

# run-opencode-streaming.sh
# Runs OpenCode CLI non-interactively from the repo root,
# feeding the absolute project path to stdin and rendering
# a clean, human-friendly streaming output.
# Usage: ./run-opencode-streaming.sh [optional prompt]

set -euo pipefail

# Source common utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

# Resolve repo root
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
PROMPT="${1:-}"

# Ensure opencode CLI is available
if ! command -v opencode >/dev/null 2>&1; then
  echo "ERROR: 'opencode' CLI not found on PATH"
  echo "Please install OpenCode CLI and retry"
  exit 127
fi

start_ts=$(get_timestamp)

echo "OpenCode session: $REPO_ROOT"

# Stream output
(
  cd "$REPO_ROOT"
    if [[ "${PROMPT:-}" == /* ]]; then
      # Transform slash command + arg(s) to natural language sentence
      # Example: /fluxid.create-epics fluxid/milestones/m01-...md
      name_and_args="${PROMPT#/}"
      cmd_name="${name_and_args%% *}"
      cmd_args="${name_and_args#${cmd_name}}"
      cmd_args="${cmd_args# }"
      if [ -f "${REPO_ROOT}/commands/${cmd_name}.md" ]; then
        if [ -n "$cmd_args" ]; then
          opencode run "Read and execute the instructions from @commands/${cmd_name}.md using argument: ${cmd_args}" 2>&1
        else
          opencode run "Read and execute the instructions from @commands/${cmd_name}.md" 2>&1
        fi
      else
        opencode run "$PROMPT" 2>&1
      fi
    elif [ -n "${PROMPT:-}" ]; then
      opencode run "$PROMPT" 2>&1
    else
    echo "ERROR: No command or prompt provided to OpenCode wrapper" 1>&2
    exit 2
  fi
) | while IFS= read -r line; do
  if [ -n "$line" ]; then
    echo "$line"
  fi
done

exit_code=${PIPESTATUS[0]:-0}

end_ts=$(get_timestamp)
display_timing "$start_ts" "$end_ts" "OpenCode run" "simple"

exit "$exit_code"
