//nolint:paralleltest // E2E tests use shared infrastructure
package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// TestImplementRetriesExhaustedContinuesThroughCommit verifies that when all implement retries fail,
// the workflow continues through commit phase to review instead of terminating.
// Flow: implement (FAIL) → implement (FAIL) → commit → review (PASS) → workflow succeeds.
//
//nolint:funlen,cyclop // E2E test with setup and verification
func TestImplementRetriesExhaustedContinuesThroughCommit(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Setup
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	reportsDir := filepath.Join(os.TempDir(), "fluxid-reports")

	if err := os.MkdirAll(homeDir, 0o750); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o750); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}
	if err := os.MkdirAll(reportsDir, 0o750); err != nil {
		t.Fatalf("Failed to create reports dir: %v", err)
	}

	// Create config with commands section
	setupConfigWithCommands(t, homeDir, "claude")

	t.Cleanup(func() {
		_ = os.RemoveAll(reportsDir)
	})

	sessionID := "b9e3f4a0-e29b-41d4-a716-446655440000" // Valid UUID for test

	// Build fluxid binary
	root := getProjectRoot(t)
	buildFluxid(t, root)
	fluxidBin := filepath.Join(root, "bin", "fluxid")

	// Create mock agent that writes reports sequentially
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	stateFile := filepath.Join(tmpDir, "agent_state")
	if err := os.WriteFile(stateFile, []byte("0"), 0o644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	agentScript := filepath.Join(binDir, "codex")
	agentContent := `#!/bin/bash
set -e

# State file tracks how many times agent has been called
if [ ! -f "$STATE_FILE" ]; then
  echo "0" > "$STATE_FILE"
fi

CALL_COUNT=$(cat "$STATE_FILE")
NEXT_COUNT=$((CALL_COUNT + 1))
echo "$NEXT_COUNT" > "$STATE_FILE"

# Ensure reports directory exists
TMPDIR="${TMPDIR:-/tmp}"
REPORTS_DIR="${TMPDIR%/}/fluxid-reports"
mkdir -p "$REPORTS_DIR"

TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
REPORT_FILE="$REPORTS_DIR/${FLUXID_SESSION_ID}.yaml"

# Call 0-1: implement FAIL, Call 2: commit PASS, Call 3+: review PASS

if [ $CALL_COUNT -lt 2 ]; then
  # Implement FAIL
  cat > "$REPORT_FILE" <<'EOF'
command: fluxid.implement
artifact: test-artifact
timestamp: TIMESTAMP_PLACEHOLDER
status: FAIL
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
EOF
  sed -i.bak "s/TIMESTAMP_PLACEHOLDER/$TIMESTAMP/" "$REPORT_FILE"
  rm -f "${REPORT_FILE}.bak"
elif [ $CALL_COUNT -eq 2 ]; then
  # Commit PASS
  cat > "$REPORT_FILE" <<'EOF'
command: fluxid.commit
artifact: test-commit
timestamp: TIMESTAMP_PLACEHOLDER
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
EOF
  sed -i.bak "s/TIMESTAMP_PLACEHOLDER/$TIMESTAMP/" "$REPORT_FILE"
  rm -f "${REPORT_FILE}.bak"
else
  # Review PASS
  cat > "$REPORT_FILE" <<'EOF'
command: fluxid.review
artifact: test-review
timestamp: TIMESTAMP_PLACEHOLDER
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
EOF
  sed -i.bak "s/TIMESTAMP_PLACEHOLDER/$TIMESTAMP/" "$REPORT_FILE"
  rm -f "${REPORT_FILE}.bak"
fi

exit 0
`
	if err := os.WriteFile(agentScript, []byte(agentContent), 0o755); err != nil {
		t.Fatalf("Failed to write agent script: %v", err)
	}

	// Create task file
	taskPath := filepath.Join(homeDir, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}

	// Run fluxid workflow
	cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin,
		"--fluxid-iterations=1",
		"--fluxid-implement-retries=2",
		"--codex",
		"--file="+taskPath,
	)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+binDir+":"+os.Getenv("PATH"),
		"FLUXID_SESSION_ID="+sessionID,
		"STATE_FILE="+stateFile,
	)
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	// Verify workflow succeeded
	if err != nil {
		t.Errorf("Workflow should succeed even with implement failures, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Verify implement attempts were logged
	if !strings.Contains(outputStr, "Implement attempt 1/2") {
		t.Error("Expected to see first implement attempt in output")
	}
	if !strings.Contains(outputStr, "Implement attempt 2/2") {
		t.Error("Expected to see second implement attempt in output")
	}

	// Verify review phase was executed
	if !strings.Contains(outputStr, "Running review phase") {
		t.Error("Expected review phase to execute after implement retries exhausted")
	}

	// Verify workflow completed successfully
	if !strings.Contains(outputStr, "Status: SUCCESS") {
		t.Error("Expected workflow to complete successfully")
	}
}
