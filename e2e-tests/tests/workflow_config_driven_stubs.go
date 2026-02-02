package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupWorkflowTest creates a temporary home directory with workflow config and command files.
// Returns tmpHome path and cleanup function.
func setupWorkflowTest(t *testing.T, root, fixtureName string) string {
	t.Helper()

	// Create temporary home with workflow config
	tmpHome := t.TempDir()
	configPath := filepath.Join(tmpHome, ".fluxid", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), permDir); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Load workflow fixture
	// #nosec G304 -- Test fixture path is trusted and controlled by test code
	fixtureContent, err := os.ReadFile(filepath.Join(root, "e2e-tests", "fixtures", "configs", fixtureName))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	if err := os.WriteFile(configPath, fixtureContent, permFile); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create command files directory
	commandsDir := filepath.Join(tmpHome, ".fluxid", "commands")
	if err := os.MkdirAll(commandsDir, permDir); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	return tmpHome
}

// createWorkflowCommandFiles creates dummy command files in the specified directory.
func createWorkflowCommandFiles(t *testing.T, tmpHome string, commands []string) {
	t.Helper()

	commandsDir := filepath.Join(tmpHome, ".fluxid", "commands")
	for _, cmd := range commands {
		filename := fmt.Sprintf("fluxid.%s.md", cmd)
		content := fmt.Sprintf("# %s\n", cmd)
		if err := os.WriteFile(filepath.Join(commandsDir, filename), []byte(content), permFile); err != nil {
			t.Fatalf("Failed to write command file %s: %v", filename, err)
		}
	}
}

// runFluxidWorkflow runs fluxid with the given task file in the specified home directory.
// Returns stdout+stderr combined output.
func runFluxidWorkflow(t *testing.T, root, tmpHome, taskPath string) string {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test code executes trusted test binary with controlled args
	cmd := exec.CommandContext(t.Context(), binPath, "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Output:\n%s", string(output))
		t.Fatalf("Command failed: %v", err)
	}

	return string(output)
}

// createTaskFile creates a simple task file for testing.
func createTaskFile(t *testing.T, tmpHome string) string {
	t.Helper()

	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("test task"), permFile); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}
	return taskPath
}

// createRetryTestStub creates a stub agent that fails first N attempts, then passes.
func createRetryTestStub(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := getRetryTestStubScript()

	if err := writeExecutableStub(stubPath, []byte(stubScript)); err != nil {
		t.Fatalf("Failed to create stub: %v", err)
	}
}

// getRetryTestStubScript returns the bash script for retry testing.
//
//nolint:funlen // Function length is acceptable - it's a single string literal.
func getRetryTestStubScript() string {
	return `#!/bin/bash
# Retry test stub - tracks implement step attempts and fails first N times

# Write report
FLUXID_BIN="$(dirname "$0")/fluxid"
# Use session ID for attempt tracking file to persist across stub invocations
IMPLEMENT_ATTEMPT_FILE="/tmp/fluxid_implement_attempts_${FLUXID_SESSION_ID}"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)

# Extract prompt from arguments (after -p flag)
PROMPT=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -p)
      PROMPT="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# Check if this is the implement step by looking for "Run implement command file" in the prompt
if echo "$PROMPT" | grep -q "Run implement command file"; then
  # This is the implement step - track attempts
  if [ -f "$IMPLEMENT_ATTEMPT_FILE" ]; then
    ATTEMPTS=$(cat "$IMPLEMENT_ATTEMPT_FILE")
  else
    ATTEMPTS=0
  fi

  # Increment attempt count
  ATTEMPTS=$((ATTEMPTS + 1))
  echo "$ATTEMPTS" > "$IMPLEMENT_ATTEMPT_FILE"

  echo "IMPLEMENT Attempt $ATTEMPTS" >&2

  # Fail first 2 attempts, pass on 3rd
  if [ "$ATTEMPTS" -le 2 ]; then
    echo "IMPLEMENT Attempt $ATTEMPTS: FAIL" >&2
    cat > "$REPORT_FILE" <<-REPORT_EOF
command: implement
artifact: stub-implement
timestamp: $TIMESTAMP
status: FAIL
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF
  else
    echo "IMPLEMENT Attempt $ATTEMPTS: PASS" >&2
    cat > "$REPORT_FILE" <<-REPORT_EOF
command: implement
artifact: stub-implement
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF
    # Clean up attempt tracker
    rm -f "$IMPLEMENT_ATTEMPT_FILE"
  fi
else
  # Other steps always pass
  cat > "$REPORT_FILE" <<-REPORT_EOF
command: other
artifact: stub-other
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF
fi

exit 0
`
}

// createReviewFailStub creates a stub agent that passes implement but fails review once (then passes).
// This enables testing that review FAIL triggers next iteration.
func createReviewFailStub(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := getReviewFailStubScript()

	if err := writeExecutableStub(stubPath, []byte(stubScript)); err != nil {
		t.Fatalf("Failed to create stub: %v", err)
	}
}

// getReviewFailStubScript returns the bash script for review FAIL testing.
//
//nolint:funlen // Function length is acceptable - it's a single string literal.
func getReviewFailStubScript() string {
	return `#!/bin/bash
# Review fail stub - passes implement, fails review first time, passes review second time

# Write report
FLUXID_BIN="$(dirname "$0")/fluxid"
# Use session ID for attempt tracking file to persist across stub invocations
REVIEW_ATTEMPT_FILE="/tmp/fluxid_review_attempts_${FLUXID_SESSION_ID}"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)

# Extract prompt from arguments (after -p flag)
PROMPT=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -p)
      PROMPT="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# Check if this is the review step by looking for "Run review command file" in the prompt
if echo "$PROMPT" | grep -q "Run review command file"; then
  # This is the review step - track attempts
  if [ -f "$REVIEW_ATTEMPT_FILE" ]; then
    ATTEMPTS=$(cat "$REVIEW_ATTEMPT_FILE")
  else
    ATTEMPTS=0
  fi

  # Increment attempt count
  ATTEMPTS=$((ATTEMPTS + 1))
  echo "$ATTEMPTS" > "$REVIEW_ATTEMPT_FILE"

  echo "REVIEW Attempt $ATTEMPTS" >&2

  # Fail first attempt, pass on second
  if [ "$ATTEMPTS" -eq 1 ]; then
    echo "REVIEW Attempt $ATTEMPTS: FAIL" >&2
    cat > "$REPORT_FILE" <<-REPORT_EOF
command: review
artifact: stub-review
timestamp: $TIMESTAMP
status: FAIL
issues:
  blockers:
    - message: "Test blocker for review FAIL iteration testing"
      location: "test"
      code: "TEST-001"
  defects: []
  concerns: []
  observations: []
  enhancements: []
summary: Review failed on first attempt (test scenario)
REPORT_EOF
  else
    echo "REVIEW Attempt $ATTEMPTS: PASS" >&2
    cat > "$REPORT_FILE" <<-REPORT_EOF
command: review
artifact: stub-review
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
summary: Review passed on second attempt (test scenario)
REPORT_EOF
    # Clean up attempt tracker
    rm -f "$REVIEW_ATTEMPT_FILE"
  fi
else
  # Implement and other steps always pass
  cat > "$REPORT_FILE" <<-REPORT_EOF
command: other
artifact: stub-other
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF
fi

exit 0
`
}

// createReviewAlwaysFailStub creates a stub agent that always fails review.
// This enables testing iteration exhaustion when review never passes.
func createReviewAlwaysFailStub(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := getReviewAlwaysFailStubScript()

	if err := writeExecutableStub(stubPath, []byte(stubScript)); err != nil {
		t.Fatalf("Failed to create stub: %v", err)
	}
}

// getReviewAlwaysFailStubScript returns the bash script for review always-fail testing.
func getReviewAlwaysFailStubScript() string {
	return `#!/bin/bash
# Review always-fail stub - passes implement, always fails review

# Write report
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)

# Extract prompt from arguments (after -p flag)
PROMPT=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -p)
      PROMPT="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# Check if this is the review step by looking for "Run review command file" in the prompt
if echo "$PROMPT" | grep -q "Run review command file"; then
  # Always fail review
  echo "REVIEW: FAIL (always fails for iteration exhaustion test)" >&2
  cat > "$REPORT_FILE" <<-REPORT_EOF
command: review
artifact: stub-review
timestamp: $TIMESTAMP
status: FAIL
issues:
  blockers:
    - message: "Test blocker - review always fails for iteration exhaustion testing"
      location: "test"
      code: "TEST-EXHAUST"
  defects: []
  concerns: []
  observations: []
  enhancements: []
summary: Review always fails (test scenario for iteration exhaustion)
REPORT_EOF
else
  # Implement and other steps always pass
  cat > "$REPORT_FILE" <<-REPORT_EOF
command: other
artifact: stub-other
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF
fi

exit 0
`
}
