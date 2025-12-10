#!/bin/bash

# run-codex-streaming.sh
# Runs Codex CLI non-interactively, captures JSONL event stream,
# and renders a clean, human-friendly status with final output.
# Usage: ./run-codex-streaming.sh "your prompt here" [--profile PROFILE] [--sandbox SANDBOX]

set -e

# Source common utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

PROMPT="$1"
shift  # Remove prompt from args

if [ -z "$PROMPT" ]; then
    echo "ERROR: No prompt provided"
    echo "Usage: $0 \"your prompt here\" [--profile PROFILE] [--sandbox SANDBOX]"
    exit 1
fi

# Parse optional arguments
PROFILE_ARG=""
SANDBOX_ARG="danger-full-access"  # Default sandbox mode

while [[ $# -gt 0 ]]; do
    case "$1" in
        --profile)
            PROFILE_ARG="$2"
            shift 2
            ;;
        --sandbox)
            SANDBOX_ARG="$2"
            shift 2
            ;;
        *)
            echo "WARNING: Unknown argument: $1"
            shift
            ;;
    esac
done

# Resolve repo root and Codex home
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)

# Set CODEX_HOME: prefer repo-local .codex, fallback to user home, or use codex default
if [ -d "$REPO_ROOT/.codex" ]; then
    export CODEX_HOME="$REPO_ROOT/.codex"
elif [ -d "$HOME/.codex" ]; then
    export CODEX_HOME="$HOME/.codex"
fi
# else: Let codex use its own default location

# Temp file for capturing the last assistant message
TMP_OUT_FILE=$(mktemp /tmp/codex_last_msg.XXXXXX)
cleanup() { rm -f "$TMP_OUT_FILE" 2>/dev/null || true; }
trap cleanup EXIT

# Build codex command arguments
CODEX_ARGS=()
if [ -n "$PROFILE_ARG" ]; then
    CODEX_ARGS+=(--profile "$PROFILE_ARG")
fi
CODEX_ARGS+=(exec "$PROMPT")
CODEX_ARGS+=(--sandbox "$SANDBOX_ARG")
CODEX_ARGS+=(-C "$REPO_ROOT")
CODEX_ARGS+=(-o "$TMP_OUT_FILE")

# If jq is missing, fall back to raw Codex output so user sees something
if ! command -v jq >/dev/null 2>&1; then
    echo "WARNING: jq not found; showing raw Codex output"
    codex "${CODEX_ARGS[@]}"
    # Show last message if present
    if [ -s "$TMP_OUT_FILE" ]; then
        echo "RESPONSE:"
        cat "$TMP_OUT_FILE"
    fi
    exit $?
fi

# Start Codex and parse JSONL events (Codex CLI schema)
codex "${CODEX_ARGS[@]}" --json 2>&1 | while IFS= read -r line; do
    # Try to parse the event type; ignore non-JSON
    event_type=$(echo "$line" | jq -r '.type // empty' 2>/dev/null)

    case "$event_type" in
        "thread.started")
            echo "Codex session started"
            ;;

        "turn.started")
            :
            ;;

        "item.started")
            item_type=$(echo "$line" | jq -r '.item.type // empty' 2>/dev/null)
            if [ "$item_type" = "command_execution" ]; then
                cmd=$(echo "$line" | jq -r '.item.command // empty' 2>/dev/null)
                if [ -n "$cmd" ] && [ "$cmd" != "null" ]; then
                    echo "[RUN] $cmd"
                fi
            fi
            ;;

        "item.completed")
            item_type=$(echo "$line" | jq -r '.item.type // empty' 2>/dev/null)
            case "$item_type" in
                "reasoning")
                    txt=$(echo "$line" | jq -r '.item.text // empty' 2>/dev/null)
                    if [ -n "$txt" ] && [ "$txt" != "null" ]; then
                        echo "[REASONING]"
                        echo "$txt"
                    fi
                    ;;
                "agent_message")
                    txt=$(echo "$line" | jq -r '.item.text // empty' 2>/dev/null)
                    if [ -n "$txt" ] && [ "$txt" != "null" ]; then
                        echo "$txt"
                    fi
                    ;;
                "command_execution")
                    out=$(echo "$line" | jq -r '.item.aggregated_output // empty' 2>/dev/null)
                    exit_code=$(echo "$line" | jq -r '.item.exit_code // empty' 2>/dev/null)
                    if [ -n "$out" ] && [ "$out" != "null" ]; then
                        echo "[OUTPUT]"
                        echo "$out"
                    fi
                    if [ -n "$exit_code" ] && [ "$exit_code" != "null" ]; then
                        if [ "$exit_code" -eq 0 ]; then
                            echo "[DONE] Command exited 0"
                        else
                            echo "[ERROR] Command exited $exit_code"
                        fi
                    fi
                    ;;
            esac
            ;;

        "error")
            err_msg=$(echo "$line" | jq -r '.message // .error.message // "Unknown error"' 2>/dev/null)
            echo "ERROR: $err_msg"
            ;;

        "turn.failed")
            err_msg=$(echo "$line" | jq -r '.error.message // "Unknown error"' 2>/dev/null)
            echo "ERROR: $err_msg"
            ;;

        "turn.completed")
            :
            ;;

        *)
            # Unknown/non-JSON line; print raw for visibility
            if [ -n "$line" ]; then
                # If it's JSON but with unhandled fields, show minimal raw
                if echo "$line" | jq -e . >/dev/null 2>&1; then
                    raw_type=$(echo "$line" | jq -r '.type // empty' 2>/dev/null)
                    case "$raw_type" in
                        "item.updated")
                            # Try best-effort partial output keys
                            part=$(echo "$line" | jq -r '.item.partial_output // .delta.output // .delta.text // .data.output // .data.text // empty' 2>/dev/null)
                            if [ -n "$part" ] && [ "$part" != "null" ]; then
                                echo -n "$part"
                            fi
                            ;;
                        *)
                            # Suppress overly chatty metadata; only show if clearly useful
                            :
                            ;;
                    esac
                else
                    echo "$line"
                fi
            fi
            ;;
    esac
done

# If no assistant message was printed, fall back to last message file
if [ -s "$TMP_OUT_FILE" ]; then
    echo "RESPONSE:"
    cat "$TMP_OUT_FILE"
fi

# Exit with Codex's exit code
exit ${PIPESTATUS[0]}
