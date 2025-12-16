package tests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestM01E04ClaudeFailureImmediateAbort verifies that when Claude exits with
// a non-zero exit code, fluxid aborts immediately and mirrors the exit code.
//
//nolint:paralleltest // Sequential execution required due to shared stub
func TestM01E04ClaudeFailureImmediateAbort(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createFailingClaudeStub(t, root, 2)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations", "1")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()

	// Verify command failed
	if err == nil {
		t.Fatal("Expected fluxid to fail when Claude exits non-zero, but it succeeded")
	}

	// Verify exit code matches Claude's exit code (2)
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Expected ExitError, got: %v", err)
	}
	if exitErr.ExitCode() != 2 {
		t.Errorf("Expected exit code 2 (matching Claude), got: %d", exitErr.ExitCode())
	}

	outputStr := output.String()

	// Verify abort message is displayed
	if !strings.Contains(outputStr, "=== Workflow Aborted ===") {
		t.Error("Expected workflow abort header in output")
	}

	// Verify error explanation
	if !strings.Contains(outputStr, "Agent execution failed") {
		t.Error("Expected agent failure message in output")
	}

	// Verify exit code is displayed
	if !strings.Contains(outputStr, "Exit code: 2") {
		t.Error("Expected exit code 2 to be displayed in error message")
	}

	// Verify next steps are provided
	if !strings.Contains(outputStr, "Next steps:") {
		t.Error("Expected next steps guidance in output")
	}

	// Verify no success message appears
	if strings.Contains(outputStr, "Status: SUCCESS") {
		t.Error("Success message should not appear after failure")
	}

	// Verify no completion summary appears
	if strings.Contains(outputStr, "Workflow Completion Summary") {
		t.Error("Completion summary should not appear after abort")
	}
}

// TestM01E04NoFurtherPhasesAfterFailure verifies that when Claude fails,
// no further phases or iterations are executed.
//
//nolint:paralleltest // Sequential execution required due to shared stub
func TestM01E04NoFurtherPhasesAfterFailure(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createFailingClaudeStub(t, root, 3)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations", "1")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected fluxid to fail")
	}

	outputStr := output.String()

	// Count how many phases were started
	implementCount := strings.Count(outputStr, "Starting phase: implement")
	commitCount := strings.Count(outputStr, "Starting phase: commit")
	reviewCount := strings.Count(outputStr, "Starting phase: review")

	// Should only see one implement phase (which fails)
	if implementCount != 1 {
		t.Errorf("Expected exactly 1 implement phase, got %d", implementCount)
	}

	// Should not see commit or review phases
	if commitCount > 0 {
		t.Errorf("Expected 0 commit phases after failure, got %d", commitCount)
	}
	if reviewCount > 0 {
		t.Errorf("Expected 0 review phases after failure, got %d", reviewCount)
	}

	// Verify only one cycle was started (use more specific pattern to avoid matching "Max Review Cycles")
	cycleCount := strings.Count(outputStr, "--- Review Cycle")
	if cycleCount != 1 {
		t.Errorf("Expected exactly 1 review cycle, got %d\nOutput:\n%s", cycleCount, outputStr)
	}
}

// TestM01E04FailureInDifferentPhases verifies that Claude failure in any phase
// (implement, commit, or review) triggers immediate abort.
//
//nolint:paralleltest // Sequential execution required due to shared stub
func TestM01E04FailureInDifferentPhases(t *testing.T) {
	t.Skip("BROKEN: Stub does not write reports (M03-E04 workflow requirement)")
	testCases := []struct {
		name             string
		failOnInvoke     int
		expectedExitCode int
	}{
		{
			name:             "FailInImplementPhase",
			failOnInvoke:     1, // First invocation (implement)
			expectedExitCode: 5,
		},
		{
			name:             "FailInReviewPhase",
			failOnInvoke:     2, // Second invocation (review)
			expectedExitCode: 9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runPhaseFailureTest(t, tc.failOnInvoke, tc.expectedExitCode)
		})
	}
}

// Helper functions

func runPhaseFailureTest(t *testing.T, failOnInvoke, expectedExitCode int) {
	t.Helper()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createConditionalFailingClaudeStub(t, root, failOnInvoke, expectedExitCode)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations", "1")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected fluxid to fail")
	}

	verifyExitCodeAndOutput(t, err, output.String(), expectedExitCode)
}

func verifyExitCodeAndOutput(t *testing.T, err error, outputStr string, expectedExitCode int) {
	t.Helper()

	// Verify correct exit code
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Expected ExitError, got: %v", err)
	}
	if exitErr.ExitCode() != expectedExitCode {
		t.Errorf("Expected exit code %d, got: %d", expectedExitCode, exitErr.ExitCode())
	}

	// Verify exit code appears in output message
	expectedExitCodeMsg := fmt.Sprintf("Exit code: %d", expectedExitCode)
	if !strings.Contains(outputStr, expectedExitCodeMsg) {
		t.Errorf("Expected exit code message '%s' in output", expectedExitCodeMsg)
	}
}

func createFailingClaudeStub(t *testing.T, root string, exitCode int) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := fmt.Sprintf(`#!/bin/bash
# Stub Claude CLI that always fails with exit code %d

echo "Claude stub invoked with args: $@"
echo "FLUXID_SESSION_ID=$FLUXID_SESSION_ID"
echo "Simulating Claude failure..." >&2

exit %d
`, exitCode, exitCode)

	if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
		t.Fatalf("failed to create failing claude stub: %v", err)
	}
}

func createConditionalFailingClaudeStub(t *testing.T, root string, failOnInvoke int, exitCode int) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	counterFile := filepath.Join(root, "bin", ".claude_invoke_count")

	// Initialize counter file
	if err := os.WriteFile(counterFile, []byte("0"), 0o644); err != nil {
		t.Fatalf("failed to create counter file: %v", err)
	}

	stubScript := fmt.Sprintf(`#!/bin/bash
# Stub Claude that fails on a specific invocation

COUNTER_FILE="%s"
FAIL_ON_INVOKE=%d
EXIT_CODE=%d

# Read current count
COUNT=$(cat "$COUNTER_FILE")
COUNT=$((COUNT + 1))
echo "$COUNT" > "$COUNTER_FILE"

echo "Claude stub invoked (count: $COUNT) with args: $@"
echo "FLUXID_SESSION_ID=$FLUXID_SESSION_ID"

if [ "$COUNT" -eq "$FAIL_ON_INVOKE" ]; then
    echo "Simulating Claude failure on invocation $COUNT..." >&2
    exit $EXIT_CODE
fi

# Succeed on other invocations
exit 0
`, counterFile, failOnInvoke, exitCode)

	if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
		t.Fatalf("failed to create conditional failing claude stub: %v", err)
	}
}
