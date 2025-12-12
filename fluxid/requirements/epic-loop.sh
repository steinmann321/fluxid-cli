#!/usr/bin/env bash
# Epic Review Loop Script
# Runs configured implement and review commands for a single epic id
# Usage: ./epic-loop.sh [--codex|--claude|--opencode] <epic-id>

set -euo pipefail

EPIC_LOOP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$EPIC_LOOP_DIR/../../.." && pwd)"

# Support test mode with isolated directories
# When EPIC_LOOP_TEST_ROOT is set, use it instead of PROJECT_ROOT/fluxid
if [[ -n "${EPIC_LOOP_TEST_ROOT:-}" ]]; then
  FLUXID_ROOT="$EPIC_LOOP_TEST_ROOT"
  echo "[TEST MODE] Using isolated directory: $FLUXID_ROOT" >&2
else
  FLUXID_ROOT="$PROJECT_ROOT/fluxid"
fi

# Source common utilities and modules
source "$PROJECT_ROOT/.fluxid/scripts/lib/common.sh"
source "$PROJECT_ROOT/.fluxid/scripts/lib/validation.sh"
source "$PROJECT_ROOT/.fluxid/scripts/lib/loop_state.sh"
source "$EPIC_LOOP_DIR/lib/report-operations.sh"
source "$EPIC_LOOP_DIR/lib/progress-tracking.sh"
source "$EPIC_LOOP_DIR/lib/workflow-steps.sh"

# Configuration
CONFIG_FILE="$PROJECT_ROOT/.fluxid/scripts/loop/config.yaml"
if [[ ! -f "$CONFIG_FILE" ]]; then
  error "Loop config not found: $CONFIG_FILE"
  exit 1
fi
impl_rel=$(sed -n 's/^[[:space:]]*implement:[[:space:]]*\(.*\)$/\1/p' "$CONFIG_FILE" | tail -n1)
rev_rel=$(sed -n 's/^[[:space:]]*review:[[:space:]]*\(.*\)$/\1/p' "$CONFIG_FILE" | tail -n1)
if [[ -z "${impl_rel:-}" || -z "${rev_rel:-}" ]]; then
  error "Loop config missing implement/review entries"
  exit 1
fi
IMPLEMENT_COMMAND_FILE="$PROJECT_ROOT/.fluxid/commands/$impl_rel"
VALIDATE_COMMAND_FILE="$PROJECT_ROOT/.fluxid/commands/$rev_rel"
REPORT_FILE=$("$PROJECT_ROOT/.fluxid/scripts/command/files.sh" --report)
HISTORY_FILE=$("$PROJECT_ROOT/.fluxid/scripts/command/files.sh" --history)
IMPLEMENT_COMMAND_NAME="$(basename "${IMPLEMENT_COMMAND_FILE%.md}")"
VALIDATE_COMMAND_NAME="$(basename "${VALIDATE_COMMAND_FILE%.md}")"
COMMIT_CMD="$PROJECT_ROOT/.fluxid/scripts/commands/commit.sh"
PROGRESS_CMD="$PROJECT_ROOT/.fluxid/scripts/command/progress.sh"
MAX_ITERATIONS=20
AGENT="claude"
STREAMING_SCRIPT=""
DRY_RUN=false


# Handle Ctrl+C - exit immediately
trap 'echo ""; echo "Interrupted by user (Ctrl+C)"; exit 130' INT TERM

# Logging functions
log() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*"
}

error() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" >&2
}

step_start() {
  local step_name="$1"
  echo ""
  echo "────────────────────────────────────────────────────────"
  echo "▶ START: $step_name"
  echo "  Task: $EPIC_BASENAME"
  echo "  Time: $(date +'%H:%M:%S')"
  echo "────────────────────────────────────────────────────────"
}

step_end() {
  local step_name="$1"
  local status="$2"  # PASS, FAIL, or SKIP
  echo "────────────────────────────────────────────────────────"
  echo "◀ END: $step_name [$status]"
  echo "  Time: $(date +'%H:%M:%S')"
  echo "────────────────────────────────────────────────────────"
  echo ""
}

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --codex)
      AGENT="codex"
      shift
      ;;
    --claude)
      AGENT="claude"
      shift
      ;;
    --opencode)
      AGENT="opencode"
      shift
      ;;

    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -*)
      echo "Unknown option: $1" >&2
      echo "Usage: $0 [--codex|--claude|--opencode] [--dry-run] <epic-id>" >&2
      exit 1
      ;;
    *)
      EPIC_ID="$1"
      shift
      ;;
  esac
done

# Select streaming script based on agent
case "$AGENT" in
    claude)
        STREAMING_SCRIPT="$PROJECT_ROOT/.fluxid/scripts/run-claude-streaming.sh"
        ;;
    codex)
        STREAMING_SCRIPT="$PROJECT_ROOT/.fluxid/scripts/run-codex-streaming.sh"
        ;;
    opencode)
        STREAMING_SCRIPT="$PROJECT_ROOT/.fluxid/scripts/run-opencode-streaming.sh"
        ;;
    *)
        error "Unsupported agent: $AGENT (supported: claude, codex, opencode)"
        exit 1
        ;;
esac

# Validate inputs
if [[ -z "${EPIC_ID:-}" ]]; then
  echo "Usage: $0 [--codex|--claude|--opencode] [--dry-run] <epic-id>" >&2
  echo "" >&2
  echo "Options:" >&2
  echo "  --claude   Use claude CLI (default)" >&2
  echo "  --codex    Use codex CLI" >&2
  echo "  --opencode Use opencode CLI" >&2
  echo "  --dry-run  Run through loop once without executing actual commands (for testing)" >&2
  echo "" >&2
  echo "Example: $0 m01-e01-some-flow.md" >&2
  exit 1
fi

if [[ ! -f "$STREAMING_SCRIPT" ]]; then
  error "Streaming script not found: $STREAMING_SCRIPT"
  exit 1
fi

# Resolve paths
EPIC_BASENAME="$(basename "$EPIC_ID")"
EPIC_PATH="$FLUXID_ROOT/epics/$EPIC_BASENAME"
EPIC_ID_TOKEN="$(echo "${EPIC_BASENAME%.md}" | sed -E 's/^(m[0-9]+-e[0-9]+).*/\1/')"

if [[ ! -f "$EPIC_PATH" ]]; then
  error "Epic file not found for id '$EPIC_BASENAME' at $EPIC_PATH"
  exit 1
fi

# Resolve test file and directory (used for Playwright screenshot cleanup)
TEST_FILE=$("$PROJECT_ROOT/.fluxid/scripts/command/files.sh" --testfile "$EPIC_BASENAME")
TEST_DIR="$(dirname "$TEST_FILE")"

# Initial cleanup: remove stale screenshots
if [[ -n "$TEST_DIR" && -d "$TEST_DIR" ]]; then
  log "Cleaning up old screenshots in tests directory..."
  find "$TEST_DIR" -maxdepth 1 -name "*.png" -type f -delete 2>/dev/null || true
fi



# Helper function for dry-run mode to determine if implementation should succeed
should_impl_succeed_dryrun() {
  local current_outer=$1
  local current_inner=$2
  local control_file="$FLUXID_ROOT/tmp/loop-state/${EPIC_BASENAME%.md}.control"

  if [[ ! -f "$control_file" ]]; then
    return 0  # Default: succeed
  fi

  # Source the control file
  # shellcheck disable=SC1090
  source "$control_file"

  # Check for outer-specific success condition
  local outer_var="IMPL_SUCCESS_ON_INNER_OUTER_${current_outer}"
  if [[ -n "${!outer_var:-}" ]]; then
    [[ ${!outer_var} -eq $current_inner ]]
    return $?
  fi

  # Check for general success condition
  if [[ -n "${IMPL_SUCCESS_ON_INNER:-}" ]]; then
    [[ ${IMPL_SUCCESS_ON_INNER} -eq $current_inner ]]
    return $?
  fi

  return 0  # Default: succeed
}

# Helper function for dry-run mode to determine if review should succeed
should_review_succeed_dryrun() {
  local current_outer=$1
  local control_file="$FLUXID_ROOT/tmp/loop-state/${EPIC_BASENAME%.md}.control"

  if [[ ! -f "$control_file" ]]; then
    return 0  # Default: succeed
  fi

  # Source the control file
  # shellcheck disable=SC1090
  source "$control_file"

  if [[ -n "${REVIEW_SUCCESS_ON_OUTER:-}" ]]; then
    [[ ${REVIEW_SUCCESS_ON_OUTER} -eq $current_outer ]]
    return $?
  fi

  return 0  # Default: succeed
}

log "=== E2E Review Loop Started ==="
log "Agent: $AGENT"
log "Streaming Script: $STREAMING_SCRIPT"
log "Epic id: $EPIC_BASENAME"
log "Epic path: $EPIC_PATH"

log "Report file: $REPORT_FILE"
if [[ "$DRY_RUN" == true ]]; then
  log "DRY RUN MODE: Will simulate workflow without executing actual commands"
fi

# Check if task already done
if should_skip_task "$EPIC_ID_TOKEN"; then
  log "Epic '$EPIC_ID_TOKEN' already marked as done in progress.yaml – skipping."
  exit 0
fi

# Determine where to resume based on current status and report state
CURRENT_STATUS=$(get_task_status "$EPIC_ID_TOKEN" 2>/dev/null || echo "pending")
RESUME_FROM="implement"

log "Current task status: $CURRENT_STATUS"

if [[ "$CURRENT_STATUS" == "review" ]]; then
  # Check if we can resume from review
  artifact_token="${EPIC_BASENAME%.md}"
  if [[ -f "$REPORT_FILE" ]] && \
     validate_report_structure "$REPORT_FILE" 2>/dev/null && \
     validate_report_metadata "$REPORT_FILE" "$IMPLEMENT_COMMAND_NAME" "$artifact_token" 2>/dev/null; then
    impl_status="$(parse_report_status "$REPORT_FILE" 2>/dev/null || echo "")"
    if [[ "$impl_status" == "PASS" ]]; then
      RESUME_FROM="review"
      log "Resuming from REVIEW phase (valid PASS report exists)"
    else
      log "Report exists but status is not PASS, starting from IMPLEMENT"
      rm -f "$REPORT_FILE" || true
    fi
  else
    log "No valid report found, starting from IMPLEMENT"
    rm -f "$REPORT_FILE" || true
    set_task_status "$EPIC_ID_TOKEN" "pending" || true
  fi
elif [[ "$CURRENT_STATUS" == "implement" ]]; then
  log "Resuming from IMPLEMENT phase"
  # Clean up any stale report
  if [[ -f "$REPORT_FILE" ]]; then
    rm -f "$REPORT_FILE" || true
  fi
else
  log "Starting from IMPLEMENT phase"
  set_task_status "$EPIC_ID_TOKEN" "pending" || true
  # Clean up any stale report
  if [[ -f "$REPORT_FILE" ]]; then
    rm -f "$REPORT_FILE" || true
  fi
fi



TOTAL_TESTS=1
SUCCESSFUL=0

log ""
log "--- Processing epic: $EPIC_BASENAME ---"

# Clear terminal for clean display
clear
# ═══════════════════════════════════════════════════════════
# MAIN WORKFLOW LOOP - Two nested loops
# Outer loop: 0-19 (max 20 iterations)
# Inner loop: 0-2 (max 3 implementation attempts per outer)
# Display: 1-based (Outer 1/20, Inner 1/3)
# ═══════════════════════════════════════════════════════════

outer=0

while [[ $outer -lt $MAX_ITERATIONS ]]; do
  # Inner loop for implementation attempts (max 3 per outer iteration)
  inner=0
  impl_passed=false

  # Skip implement/commit/validate if resuming from review
  if [[ "$RESUME_FROM" != "review" ]]; then

    # Inner loop: Try implementing up to 3 times
    while [[ $inner -lt 3 ]]; do
      # Display counters (1-based for humans: outer+1, inner+1)
      outer_display=$((outer + 1))
      inner_display=$((inner + 1))

# ───────────────────────────────────────────────────────
# IMPLEMENT PHASE
# ───────────────────────────────────────────────────────
step_start "Outer $outer_display/20, Inner $inner_display/3: IMPLEMENT"
set_task_status "$EPIC_ID_TOKEN" "implement" || log "Warning: failed to set status to 'implement'"

if [[ "$DRY_RUN" == true ]]; then
  # Dry-run mode: check control file
  artifact_token="${EPIC_BASENAME%.md}"

  if should_impl_succeed_dryrun "$outer" "$inner"; then
    impl_status="PASS"
    log "[DRY-RUN] Outer $outer_display, Inner $inner_display: IMPLEMENT PASS"
  else
    impl_status="FAIL"
    log "[DRY-RUN] Outer $outer_display, Inner $inner_display: IMPLEMENT FAIL"
  fi

  mkdir -p "$(dirname "$REPORT_FILE")"
  cat > "$REPORT_FILE" << EOF
command: $IMPLEMENT_COMMAND_NAME
artifact: $artifact_token
timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
status: $impl_status
summary: DRY RUN - Outer $outer_display, Inner $inner_display - $impl_status
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
EOF
else
  # Real mode
  if ! invoke_implement_step "$EPIC_BASENAME" "$REPORT_FILE"; then
    step_end "IMPLEMENT" "FAIL"
    loop_state_set "$EPIC_BASENAME" "error"
    error "Implement step failed for epic $EPIC_BASENAME (Outer $outer_display, Inner $inner_display)"
    error "CRITICAL: Halting workflow loop due to implement step failure."
    exit 1
  fi
fi
      step_end "IMPLEMENT" "DONE"

      # ───────────────────────────────────────────────────────
      # COMMIT PHASE
      # ───────────────────────────────────────────────────────
      step_start "COMMIT"
      if [[ "$DRY_RUN" == true ]]; then
        log "[DRY-RUN] Skipping commit step"
      else
        if ! invoke_commit_step; then
          step_end "COMMIT" "FAIL"
          loop_state_set "$EPIC_BASENAME" "error"
          error "Commit step failed after implement step for epic $EPIC_BASENAME"
          exit 1
        fi
      fi
      step_end "COMMIT" "DONE"

      # ───────────────────────────────────────────────────────
      # CHECK IMPLEMENT REPORT
      # ───────────────────────────────────────────────────────
      step_start "VALIDATE REPORT"

      # Check report exists
      if ! report_exists "$REPORT_FILE"; then
        error "Report file missing after implement step: $REPORT_FILE"
        step_end "VALIDATE REPORT" "FAIL"
        inner=$((inner + 1))
        continue  # Try next inner iteration
      fi

      # Validate structure
      if ! validate_report_structure "$REPORT_FILE"; then
        error "Report structure validation failed"
        step_end "VALIDATE REPORT" "FAIL"
        inner=$((inner + 1))
        continue
      fi

      # Validate metadata
      artifact_token="${EPIC_BASENAME%.md}"
      if ! validate_report_metadata "$REPORT_FILE" "$IMPLEMENT_COMMAND_NAME" "$artifact_token"; then
        error "Report metadata validation failed (expected command: $IMPLEMENT_COMMAND_NAME, artifact: $artifact_token)"
        step_end "VALIDATE REPORT" "FAIL"
        inner=$((inner + 1))
        continue
      fi

      # Read status
      impl_status="$(parse_report_status "$REPORT_FILE")"

      if [[ -z "$impl_status" ]]; then
        error "Report status field missing in $REPORT_FILE"
        step_end "VALIDATE REPORT" "FAIL"
        inner=$((inner + 1))
        continue
      fi

      log "Implementation report status: $impl_status"

      # Check if PASS
      if [[ "$impl_status" == "PASS" ]]; then
        log "✅ Implementation PASS on Outer $outer_display, Inner $inner_display"
        step_end "VALIDATE REPORT" "PASS"
        impl_passed=true
        break  # Exit inner loop, proceed to review
      else
        log "Implementation status is $impl_status"
        step_end "VALIDATE REPORT" "FAIL"
        inner=$((inner + 1))
        # Continue to next inner iteration
      fi
    done  # End inner loop

  else
    log "Skipping IMPLEMENT/COMMIT/VALIDATE phases (resuming from REVIEW)"
    impl_passed=true  # Assume implementation passed if resuming from review
  fi

  # Clear resume flag after first iteration
  RESUME_FROM="implement"

  # ───────────────────────────────────────────────────────
  # REVIEW PHASE (always executed after inner loop)
  # ───────────────────────────────────────────────────────
  # Cleanup screenshots before review
  if [[ -n "$TEST_DIR" && -d "$TEST_DIR" ]]; then
    log "Cleaning up screenshots before review step..."
    find "$TEST_DIR" -maxdepth 1 -name "*.png" -type f -delete 2>/dev/null || true
  fi

  outer_display=$((outer + 1))
step_start "Outer $outer_display/20: REVIEW"
set_task_status "$EPIC_ID_TOKEN" "review" || log "Warning: failed to set status to 'review'"

# Track if review should fail
review_failed=false
review_status=""

if [[ "$DRY_RUN" == true ]]; then
  # Dry-run mode: check control file
  artifact_token="${EPIC_BASENAME%.md}"

  if should_review_succeed_dryrun "$outer"; then
    report_status="PASS"
    log "[DRY-RUN] Outer $outer_display: REVIEW PASS"
  else
    report_status="FAIL"
    log "[DRY-RUN] Outer $outer_display: REVIEW FAIL"
  fi

  mkdir -p "$(dirname "$REPORT_FILE")"
  cat > "$REPORT_FILE" << EOF
command: $VALIDATE_COMMAND_NAME
artifact: $artifact_token
timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
status: $report_status
summary: DRY RUN - Outer $outer_display REVIEW - $report_status
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
EOF
else
  # Real mode
  if ! invoke_review_step "$EPIC_BASENAME" "$REPORT_FILE"; then
    loop_state_set "$EPIC_BASENAME" "error"
    error "Review step failed for epic $EPIC_BASENAME (Outer $outer_display)"
    error "CRITICAL: Halting workflow loop due to review step failure."
    step_end "REVIEW" "FAIL"
    exit 1
  fi
fi

  # Check review report
  if ! report_exists "$REPORT_FILE"; then
    error "Report file missing after review step: $REPORT_FILE"
    review_failed=true
  elif ! validate_report_structure "$REPORT_FILE"; then
    error "Review report structure validation failed"
    review_failed=true
  elif ! validate_report_metadata "$REPORT_FILE" "$VALIDATE_COMMAND_NAME" "$artifact_token"; then
    error "Review report metadata validation failed (expected command: $VALIDATE_COMMAND_NAME, artifact: $artifact_token)"
    review_failed=true
  else
    review_status="$(parse_report_status "$REPORT_FILE")"
    if [[ -z "$review_status" ]]; then
      error "Review report status field missing in $REPORT_FILE"
      review_failed=true
    fi
  fi

  # Handle review result
  if [[ "$review_failed" == true ]]; then
    step_end "REVIEW" "FAIL"
    log "Review validation failed, continuing to next outer iteration..."
    outer=$((outer + 1))
    continue
  fi

  log "Review report status: $review_status"

  if [[ "$review_status" == "PASS" ]]; then
    log "✅ Review PASS for epic $EPIC_BASENAME on outer $outer_display"
    step_end "REVIEW" "PASS"

    set_task_status "$EPIC_ID_TOKEN" "done" || log "Warning: failed to set status to 'done'"
    mark_task_complete "$EPIC_ID_TOKEN" || log "Warning: failed to mark epic complete"

    SUCCESSFUL=$((SUCCESSFUL + 1))

    # Final cleanup: remove screenshots after success
    if [[ -n "$TEST_DIR" && -d "$TEST_DIR" ]]; then
      log "Cleaning up screenshots after successful completion..."
      find "$TEST_DIR" -maxdepth 1 -name "*.png" -type f -delete 2>/dev/null || true
    fi

    if [[ -n "${HISTORY_FILE:-}" && -f "$HISTORY_FILE" ]]; then
      : > "$HISTORY_FILE" || true
      log "Truncated workflow history file: $HISTORY_FILE"
    fi

    log ""
    log "=== E2E Review Loop Completed Successfully ==="
    log "Total tests: $TOTAL_TESTS"
    log "Successful: $SUCCESSFUL"

    exit 0
  fi

  # Review status is not PASS
  log "❌ Review status is $review_status, looping back to implement..."
  step_end "REVIEW" "FAIL"

  # Cleanup artifacts after every 5th outer iteration
  if [[ $(((outer + 1) % 5)) -eq 0 ]]; then
    log ""
    log "═══════════════════════════════════════════════════════════"
    log "🧹 CLEANUP: Outer $outer_display (every 5th loop) - removing all artifacts"
    log "═══════════════════════════════════════════════════════════"

    # Periodic cleanup: remove screenshots every 5 loops
    if [[ -n "$TEST_DIR" && -d "$TEST_DIR" ]]; then
      log "Removing screenshots from tests directory..."
      find "$TEST_DIR" -maxdepth 1 -name "*.png" -type f -delete 2>/dev/null || true
    fi

    # Clean up report file
    if [[ -f "$REPORT_FILE" ]]; then
      log "Removing report file: $REPORT_FILE"
      rm -f "$REPORT_FILE" || true
    fi

    # Truncate history file
    if [[ -n "${HISTORY_FILE:-}" && -f "$HISTORY_FILE" ]]; then
      : > "$HISTORY_FILE" || true
      log "Truncated workflow history file: $HISTORY_FILE"
    fi

    # Reset task status to pending (as if loop restarted)
    set_task_status "$EPIC_ID_TOKEN" "pending" || log "Warning: failed to reset status to 'pending'"

    log "✅ Cleanup complete - loop resetting as if fresh start"
    log "═══════════════════════════════════════════════════════════"
    log ""
  fi

  outer=$((outer + 1))
done  # End outer loop

# Max iterations exhausted
set_task_status "$EPIC_ID_TOKEN" "error" || log "Warning: failed to set status to 'error'"
error "CRITICAL: Max iterations ($MAX_ITERATIONS) reached without success for epic $EPIC_BASENAME"
error "Halting workflow loop."

log ""
log "=== E2E Review Loop Failed ==="
log "Total tests: $TOTAL_TESTS"
log "Successful: $SUCCESSFUL"

exit 1
