package tests

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	permExecutable = 0o755 // Executable file permissions for test stubs
)

// writeExecutableStub writes a bash script with executable permissions.
// This is a test helper that requires 0755 permissions for shell scripts.
//
// #nosec G306 -- Test stubs require executable permissions
//
//nolint:wrapcheck // Test helper, error context clear from usage
func writeExecutableStub(path string, content []byte) error {
	return os.WriteFile(path, content, permExecutable)
}

// createStreamingStubClaude creates a stub that generates streaming output over time.
func createStreamingStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Streaming output stub - generates output over time

echo "Claude stub: Starting streaming output test"

# Generate burst output with small delays
for i in {1..10}; do
  echo "BURST_LINE $i: $(date +%s%N)"
  sleep 0.05
done

echo "FLUXID_SESSION_ID=$FLUXID_SESSION_ID"

# Write report so workflow can proceed
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
"$FLUXID_BIN" ipc write-report --session "$FLUXID_SESSION_ID" <<REPORT_EOF
command: test
artifact: stub-test
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

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

// createInteractiveStubClaude creates a stub that prompts for input and echoes it back.
func createInteractiveStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Interactive stub - prompts for input and echoes it back

echo "Claude stub: Interactive test"

# Only prompt during implement phase
if echo "$@" | grep -q "Implement the required"; then
  echo "PROMPT: Enter your name:"
  read -r response
  echo "RECEIVED: $response"
fi

# Write report so workflow can proceed
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
"$FLUXID_BIN" ipc write-report --session "$FLUXID_SESSION_ID" <<REPORT_EOF
command: test
artifact: stub-test
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

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

// createLargeOutputStubClaude creates a stub that generates many lines to test buffer handling.
func createLargeOutputStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Large output stub - generates many lines to test buffer handling

echo "Claude stub: Large output test"

# Generate 1500 lines of output
for i in {1..1500}; do
  echo "LARGE_OUTPUT_LINE $i: Lorem ipsum dolor sit amet, consectetur adipiscing elit."
done

# Write report so workflow can proceed
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
"$FLUXID_BIN" ipc write-report --session "$FLUXID_SESSION_ID" <<REPORT_EOF
command: test
artifact: stub-test
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

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

// createMixedStreamStubClaude creates a stub that outputs to both stdout and stderr.
func createMixedStreamStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Mixed stream stub - outputs to both stdout and stderr

echo "Claude stub: Mixed stream test"

for i in {1..5}; do
  echo "STDOUT: MSG_$i message on stdout"
  echo "STDERR: MSG_$i message on stderr" >&2
  sleep 0.02
done

# Write report so workflow can proceed
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
"$FLUXID_BIN" ipc write-report --session "$FLUXID_SESSION_ID" <<REPORT_EOF
command: test
artifact: stub-test
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

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

// createWorkflowContinuationStubClaude creates a stub that is interactive during
// implement phase but silent for other phases.
func createWorkflowContinuationStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Workflow continuation stub - interactive during implement, silent for others

# Check which phase we're in
if echo "$@" | grep -q "Implement the required"; then
  echo "IMPLEMENT_PROMPT: Ready to implement? (type anything to continue)"
  read -r response
  echo "IMPLEMENT_RESPONSE: Got '$response', continuing..."
elif echo "$@" | grep -q "Create a git commit"; then
  echo "Commit phase executing..."
elif echo "$@" | grep -q "Review the implementation"; then
  echo "Review phase executing..."
fi

# Write report so workflow can proceed
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
"$FLUXID_BIN" ipc write-report --session "$FLUXID_SESSION_ID" <<REPORT_EOF
command: test
artifact: stub-test
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

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

// createLongRunningStub creates a Claude stub that adds delays to allow abort signals.
// The stub sleeps briefly before writing report to give time for abort.
func createLongRunningStub(t *testing.T, root string, _ int) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Stub that sleeps briefly to allow time for abort signals

echo "Claude stub: Starting phase..."

# Sleep to give time for abort signal
sleep 0.5

# Write report so workflow can proceed to next phase
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
"$FLUXID_BIN" ipc write-report --session "$FLUXID_SESSION_ID" <<REPORT_EOF
command: test
artifact: stub-test
timestamp: $TIMESTAMP
status: FAIL
issues:
  blockers: []
  defects: []
  concerns: []
  observations:
    - message: "Testing abort between phases"
  enhancements: []
REPORT_EOF

echo "Phase completed"
exit 0
`

	if err := writeExecutableStub(stubPath, []byte(stubScript)); err != nil {
		t.Fatalf("Failed to create long-running stub: %v", err)
	}
}
