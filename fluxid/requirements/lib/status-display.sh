#!/usr/bin/env bash
# Status Display Module
# Provides persistent status header at top of terminal

# Terminal control sequences
SAVE_CURSOR="\033[s"
RESTORE_CURSOR="\033[u"
CLEAR_LINE="\033[2K"
CURSOR_TO_TOP="\033[H"
CURSOR_HOME="\033[1;1H"
HIDE_CURSOR="\033[?25l"
SHOW_CURSOR="\033[?25h"

# Set terminal window title
# Args: $1 - title text
set_terminal_title() {
  echo -ne "\033]0;$1\007"
}

# Status header configuration
STATUS_HEADER_LINES=9
STATUS_SCROLL_REGION_START=$((STATUS_HEADER_LINES + 1))
FIXED_TABLE_WIDTH=80

# Global status variables
CURRENT_EPIC=""
CURRENT_ITERATION=""
CURRENT_MAX_ITERATIONS=""
START_TIME=""
STATUS_IMPLEMENT="PENDING"
STATUS_COMMIT="PENDING"
STATUS_REVIEW="PENDING"
STATUS_UPDATER_PID=""

# Background updater function
status_background_updater() {
  while true; do
    sleep 5
    status_update
  done
}

# Initialize status display
# Sets up terminal with fixed header and scrolling region
status_init() {
  local epic_name="$1"
  local max_iterations="${2:-5}"

  CURRENT_EPIC="$epic_name"
  CURRENT_ITERATION="0"
  CURRENT_MAX_ITERATIONS="$max_iterations"
  START_TIME=$(date +%s)
  STATUS_IMPLEMENT="PENDING"
  STATUS_COMMIT="PENDING"
  STATUS_REVIEW="PENDING"

  # Clear screen and hide cursor
  clear
  echo -ne "$HIDE_CURSOR"

  # Set scroll region (from line 5 to bottom)
  local term_height=$(tput lines)
  tput csr $STATUS_SCROLL_REGION_START $term_height

  # Draw initial header
  status_update

  # Move cursor to scroll region
  tput cup $STATUS_SCROLL_REGION_START 0

  # Start background updater for duration
  status_background_updater &
  STATUS_UPDATER_PID=$!
}

# Calculate duration since start
get_duration() {
  local now=$(date +%s)
  local elapsed=$((now - START_TIME))
  local hours=$((elapsed / 3600))
  local minutes=$(((elapsed % 3600) / 60))
  printf "%02d:%02d" $hours $minutes
}

# Format start time
get_start_time() {
  date -r "$START_TIME" +%H:%M:%S 2>/dev/null || echo "N/A"
}

# Generate fixed-width separator line
generate_separator() {
  printf '%80s\n' '' | tr ' ' '-'
}

# Truncate text to fit within fixed width
truncate_text() {
  local text="$1"
  local max_len="$2"
  if [[ ${#text} -gt $max_len ]]; then
    echo "${text:0:$((max_len-3))}..."
  else
    echo "$text"
  fi
}

# Update status header
status_update() {
  # Update terminal window title
  local duration=$(get_duration)
  set_terminal_title "[$CURRENT_ITERATION/$CURRENT_MAX_ITERATIONS] $CURRENT_EPIC - $duration"

  # Save current cursor position
  echo -ne "$SAVE_CURSOR"

  # Move to top and draw header
  tput cup 0 0

  # Truncate task name if needed
  local task_display=$(truncate_text "$CURRENT_EPIC" 55)

  # Line 1: Separator
  echo -ne "$CLEAR_LINE"
  generate_separator

  # Line 2: Task
  echo -ne "$CLEAR_LINE"
  printf "| %-20.20s | %-55.55s |\n" "Task" "$task_display"

  # Line 3: Iteration
  echo -ne "$CLEAR_LINE"
  printf "| %-20.20s | %-55.55s |\n" "Iteration" "$CURRENT_ITERATION/$CURRENT_MAX_ITERATIONS"

  # Line 4: Started
  echo -ne "$CLEAR_LINE"
  local start_time=$(get_start_time)
  printf "| %-20.20s | %-55.55s |\n" "Started" "$start_time"

  # Line 5: Duration
  echo -ne "$CLEAR_LINE"
  printf "| %-20.20s | %-55.55s |\n" "Duration" "$duration"

  # Line 6: Separator
  echo -ne "$CLEAR_LINE"
  generate_separator

  # Line 7: Implement step
  echo -ne "$CLEAR_LINE"
  printf "| %-20.20s | %-55.55s |\n" "Implement" "$STATUS_IMPLEMENT"

  # Line 8: Commit step
  echo -ne "$CLEAR_LINE"
  printf "| %-20.20s | %-55.55s |\n" "Commit/Refactor" "$STATUS_COMMIT"

  # Line 9: Review step
  echo -ne "$CLEAR_LINE"
  printf "| %-20.20s | %-55.55s |\n" "Review" "$STATUS_REVIEW"

  # Restore cursor position and flush output
  echo -ne "$RESTORE_CURSOR"
  tput cup $STATUS_SCROLL_REGION_START 0 2>/dev/null || true
}

# Update just the step status
status_set_step() {
  local step="$1"
  local status="$2"

  case "$step" in
    IMPLEMENT*|Implement*)
      STATUS_IMPLEMENT="$status"
      ;;
    COMMIT*|Commit*)
      STATUS_COMMIT="$status"
      ;;
    REVIEW*|Review*)
      STATUS_REVIEW="$status"
      ;;
  esac

  status_update
}

# Update just the iteration
status_set_iteration() {
  local iteration="$1"
  CURRENT_ITERATION="$iteration"

  # Reset all statuses to pending at start of new iteration
  STATUS_IMPLEMENT="PENDING"
  STATUS_COMMIT="PENDING"
  STATUS_REVIEW="PENDING"

  status_update
}

# Cleanup status display
# Restores terminal to normal state
status_cleanup() {
  # Kill background updater
  if [[ -n "$STATUS_UPDATER_PID" ]]; then
    kill "$STATUS_UPDATER_PID" 2>/dev/null || true
    wait "$STATUS_UPDATER_PID" 2>/dev/null || true
  fi

  # Kill all child processes of this script
  local script_pid=$$
  pkill -P "$script_pid" 2>/dev/null || true

  # Reset scroll region
  local term_height=$(tput lines)
  tput csr 0 $term_height

  # Show cursor
  echo -ne "$SHOW_CURSOR"

  # Move cursor to bottom
  tput cup $term_height 0
}

# Trap handler for cleanup on exit
status_trap_cleanup() {
  status_cleanup
  exit "${1:-0}"
}
