package tests

import (
	"os"
	"path/filepath"
	"testing"
)

const filePerms = 0o755

// createClaudeFormatStub creates a Claude stub that emits stream-json format.
//
//nolint:lll // JSON test data requires long lines
func createClaudeFormatStub(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Claude format stub - emits stream-json format

# Emit Claude stream-json format
cat <<'EOF'
{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_test","type":"message","role":"assistant","content":[],"model":"claude-3-5-sonnet-20241022"}}}
{"type":"stream_event","event":{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}}
{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Claude streaming message"}}}
{"type":"stream_event","event":{"type":"content_block_stop","index":0}}
{"type":"stream_event","event":{"type":"message_delta","delta":{"stop_reason":"end_turn"}}}
{"type":"stream_event","event":{"type":"message_stop"}}
EOF

# Write PASS report so workflow can proceed
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)
cat > "$REPORT_FILE" <<-REPORT_EOF
command: test
artifact: claude-test
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF

exit 0
`

	if err := os.WriteFile(stubPath, []byte(stubScript), filePerms); err != nil {
		t.Fatalf("failed to create claude stub: %v", err)
	}
}

// createCodexFormatStub creates a Codex stub that emits JSONL format.
func createCodexFormatStub(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "codex")
	stubScript := `#!/bin/bash
# Codex format stub - emits JSONL format

# Emit Codex JSONL format
cat <<'EOF'
{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"Codex agent message","role":"assistant"}}
{"type":"item.completed","item":{"id":"item_2","type":"reasoning","text":"Codex reasoning step"}}
{"type":"turn.completed","turn":{"id":"turn_1","status":"completed","usage":{"input_tokens":100,"output_tokens":50}}}
EOF

# Write PASS report so workflow can proceed
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)
cat > "$REPORT_FILE" <<-REPORT_EOF
command: test
artifact: codex-test
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF

exit 0
`

	if err := os.WriteFile(stubPath, []byte(stubScript), filePerms); err != nil {
		t.Fatalf("failed to create codex stub: %v", err)
	}
}

// createOpencodeFormatStub creates an Opencode stub that emits JSON format.
func createOpencodeFormatStub(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "opencode")
	stubScript := `#!/bin/bash
# Opencode format stub - emits JSON format

# Emit Opencode JSON format
cat <<'EOF'
{"type":"text","part":{"text":"Opencode text output"},"index":0}
{"type":"step_finish","step":{"status":"completed","duration":1.5}}
EOF

# Write PASS report so workflow can proceed
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)
cat > "$REPORT_FILE" <<-REPORT_EOF
command: test
artifact: opencode-test
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF

exit 0
`

	if err := os.WriteFile(stubPath, []byte(stubScript), filePerms); err != nil {
		t.Fatalf("failed to create opencode stub: %v", err)
	}
}

// createGeminiFormatStub creates a Gemini stub that emits stream-json format.
func createGeminiFormatStub(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "gemini")
	stubScript := `#!/bin/bash
# Gemini format stub - emits stream-json format

# Emit Gemini JSON format
cat <<'EOF'
{"type":"message","role":"assistant","content":"Gemini assistant response","model":"gemini-pro"}
{"type":"result","status":"success","finish_reason":"stop"}
EOF

# Write PASS report so workflow can proceed
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)
cat > "$REPORT_FILE" <<-REPORT_EOF
command: test
artifact: gemini-test
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF

exit 0
`

	if err := os.WriteFile(stubPath, []byte(stubScript), filePerms); err != nil {
		t.Fatalf("failed to create gemini stub: %v", err)
	}
}

// createClaudeFormatStubFail creates a Claude stub that generates FAIL reports.
//
//nolint:lll // JSON test data requires long lines
func createClaudeFormatStubFail(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Claude format stub - emits stream-json and FAIL report

# Emit Claude stream-json format
cat <<'EOF'
{"type":"stream_event","event":{"type":"message_start","message":{"id":"msg_test","type":"message","role":"assistant","content":[],"model":"claude-3-5-sonnet-20241022"}}}
{"type":"stream_event","event":{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}}
{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Claude encountered an issue"}}}
{"type":"stream_event","event":{"type":"content_block_stop","index":0}}
{"type":"stream_event","event":{"type":"message_stop"}}
EOF

# Write FAIL report
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)
cat > "$REPORT_FILE" <<-REPORT_EOF
command: test
artifact: claude-test
timestamp: $TIMESTAMP
status: FAIL
issues:
  blockers:
    - message: "Test blocker issue"
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF

exit 0
`

	if err := os.WriteFile(stubPath, []byte(stubScript), filePerms); err != nil {
		t.Fatalf("failed to create claude fail stub: %v", err)
	}
}

// createCodexFormatStubFail creates a Codex stub that generates FAIL reports.
//
//nolint:lll // JSON test data requires long lines
func createCodexFormatStubFail(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "codex")
	stubScript := `#!/bin/bash
# Codex format stub - emits JSONL and FAIL report

# Emit Codex JSONL format
cat <<'EOF'
{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"Codex encountered an issue","role":"assistant"}}
{"type":"turn.completed","turn":{"id":"turn_1","status":"failed"}}
EOF

# Write FAIL report
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)
cat > "$REPORT_FILE" <<-REPORT_EOF
command: test
artifact: codex-test
timestamp: $TIMESTAMP
status: FAIL
issues:
  blockers:
    - message: "Test blocker issue"
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF

exit 0
`

	if err := os.WriteFile(stubPath, []byte(stubScript), filePerms); err != nil {
		t.Fatalf("failed to create codex fail stub: %v", err)
	}
}

// createOpencodeFormatStubFail creates an Opencode stub that generates FAIL reports.
func createOpencodeFormatStubFail(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "opencode")
	stubScript := `#!/bin/bash
# Opencode format stub - emits JSON and FAIL report

# Emit Opencode JSON format
cat <<'EOF'
{"type":"text","part":{"text":"Opencode encountered an issue"},"index":0}
{"type":"step_finish","step":{"status":"failed","error":"Test error"}}
EOF

# Write FAIL report
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)
cat > "$REPORT_FILE" <<-REPORT_EOF
command: test
artifact: opencode-test
timestamp: $TIMESTAMP
status: FAIL
issues:
  blockers:
    - message: "Test blocker issue"
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF

exit 0
`

	if err := os.WriteFile(stubPath, []byte(stubScript), filePerms); err != nil {
		t.Fatalf("failed to create opencode fail stub: %v", err)
	}
}

// createGeminiFormatStubFail creates a Gemini stub that generates FAIL reports.
func createGeminiFormatStubFail(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "gemini")
	stubScript := `#!/bin/bash
# Gemini format stub - emits JSON and FAIL report

# Emit Gemini JSON format
cat <<'EOF'
{"role":"assistant","content":"Gemini encountered an issue","model":"gemini-pro"}
{"type":"result","status":"error","error":"Test error"}
EOF

# Write FAIL report
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)
cat > "$REPORT_FILE" <<-REPORT_EOF
command: test
artifact: gemini-test
timestamp: $TIMESTAMP
status: FAIL
issues:
  blockers:
    - message: "Test blocker issue"
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF

exit 0
`

	if err := os.WriteFile(stubPath, []byte(stubScript), filePerms); err != nil {
		t.Fatalf("failed to create gemini fail stub: %v", err)
	}
}
