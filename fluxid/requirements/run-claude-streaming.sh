#!/bin/bash

# run-claude-streaming.sh
# Runs Claude with streaming JSON output and displays clean real-time status messages
# Usage: ./run-claude-streaming.sh "your prompt here"

# Source common utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

PROMPT="$1"

if [ -z "$PROMPT" ]; then
    echo "ERROR: No prompt provided"
    echo "Usage: $0 \"your prompt here\""
    exit 1
fi

# Track state
show_thinking=false
current_text=""
current_tool_name=""
current_tool_desc=""
in_text_block=false
in_tool_block=false
text_block_index=-1
text_block_has_content=false

# Run Claude with partial messages enabled and parse JSON stream
claude --dangerously-skip-permissions \
    --output-format stream-json \
    --verbose \
    --include-partial-messages \
    -p "$PROMPT" 2>&1 | while IFS= read -r line; do

    # Parse JSON fields
    event_type=$(echo "$line" | jq -r '.type // empty' 2>/dev/null)

    case "$event_type" in
        "system")
            subtype=$(echo "$line" | jq -r '.subtype // empty' 2>/dev/null)
            if [ "$subtype" = "init" ]; then
                model=$(echo "$line" | jq -r '.model // empty' 2>/dev/null)
                echo "Model: $model"
            fi
            ;;

        "stream_event")
            stream_type=$(echo "$line" | jq -r '.event.type // empty' 2>/dev/null)

            case "$stream_type" in
                "message_start")
                    current_text=""
                    text_block_has_content=false
                    ;;

                "content_block_start")
                    content_type=$(echo "$line" | jq -r '.event.content_block.type // empty' 2>/dev/null)
                    block_index=$(echo "$line" | jq -r '.event.index // empty' 2>/dev/null)

                    if [ "$content_type" = "text" ]; then
                        in_text_block=true
                        text_block_index="$block_index"
                        text_block_has_content=false
                    elif [ "$content_type" = "tool_use" ]; then
                        in_tool_block=true
                        current_tool_name=$(echo "$line" | jq -r '.event.content_block.name // empty' 2>/dev/null)
                    fi
                    ;;

                "content_block_delta")
                    block_index=$(echo "$line" | jq -r '.event.index // empty' 2>/dev/null)
                    delta_type=$(echo "$line" | jq -r '.event.delta.type // empty' 2>/dev/null)

                    if [ "$delta_type" = "text_delta" ]; then
                        text_delta=$(echo "$line" | jq -r '.event.delta.text // empty' 2>/dev/null)
                        if [ -n "$text_delta" ] && [ "$text_delta" != "null" ]; then
                            echo -n "$text_delta"
                            text_block_has_content=true
                            current_text="${current_text}${text_delta}"
                        fi
                    elif [ "$delta_type" = "input_json_delta" ]; then
                        :
                    fi
                    ;;

                "content_block_stop")
                    if [ "$in_text_block" = true ]; then
                        if [ "$text_block_has_content" = true ]; then
                            echo ""
                        fi
                        in_text_block=false
                        text_block_has_content=false
                        current_text=""
                    elif [ "$in_tool_block" = true ]; then
                        in_tool_block=false
                    fi
                    ;;

                "message_delta")
                    :
                    ;;

                "message_stop")
                    :
                    ;;
            esac
            ;;

        "assistant")
            tool_name=$(echo "$line" | jq -r '.tool_call.name // empty' 2>/dev/null)
            tool_desc=$(echo "$line" | jq -r '.tool_call.description // empty' 2>/dev/null)

            if [ -n "$tool_name" ] && [ "$tool_name" != "null" ]; then
                echo "[TOOL] $tool_name"
                if [ -n "$tool_desc" ] && [ "$tool_desc" != "null" ]; then
                    echo "  $tool_desc"
                fi
            fi
            ;;

        "tool_result")
            tool_name=$(echo "$line" | jq -r '.tool_call.name // empty' 2>/dev/null)
            if [ -n "$tool_name" ] && [ "$tool_name" != "null" ]; then
                echo "[DONE] $tool_name"
            fi
            ;;

        "result")
            subtype=$(echo "$line" | jq -r '.subtype // empty' 2>/dev/null)
            duration_ms=$(echo "$line" | jq -r '.duration_ms // empty' 2>/dev/null)
            num_turns=$(echo "$line" | jq -r '.num_turns // empty' 2>/dev/null)
            cost=$(echo "$line" | jq -r '.total_cost_usd // empty' 2>/dev/null)

            echo ""
            if [ "$subtype" = "success" ]; then
                echo "SUCCESS: Duration ${duration_ms}ms, Turns $num_turns, Cost \$$cost"
            else
                error_msg=$(echo "$line" | jq -r '.error // "Unknown error"' 2>/dev/null)
                echo "ERROR: $error_msg"
            fi
            ;;
    esac
done

# Capture exit code
exit_code=${PIPESTATUS[0]}

# Return exit code from Claude
exit $exit_code
