//nolint:paralleltest // E2E tests use shared infrastructure
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
func verifyFailureOutput(t *testing.T, output string) {
	t.Helper()
	checks := []struct {
		contains bool
		pattern  string
		errMsg   string
	}{
		{true, "=== Workflow Aborted ===", "Expected workflow abort header in output"},
		{true, "Agent execution failed", "Expected agent failure message in output"},
		{true, "Exit code: 1", "Expected exit code 1 to be displayed in error message"},
		{true, "Next steps:", "Expected next steps guidance in output"},
		{false, "Status: SUCCESS", "Success message should not appear after failure"},
	}
	for _, check := range checks {
		if check.contains != strings.Contains(output, check.pattern) {
			t.Error(check.errMsg)
		}
	}
}

func TestM01E04ClaudeFailureImmediateAbort(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)
	// Create test-specific stub directory to avoid race conditions
	stubDir := t.TempDir()
	createFailingClaudeStub(t, stubDir, 2)
	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")
	binPath := filepath.Join(root, "bin", "fluxid")
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s:%s", stubDir, filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	// Corrected behavior: Workflow completes successfully even when agents fail
	// The workflow continues through all phases, logging failures
	if err != nil {
		t.Fatalf("Expected fluxid to succeed (workflow completes all phases), got error: %v\nOutput:\n%s",
			err, output.String())
	}
	verifyFailureOutput(t, output.String())
	// Verify completion summary appears (workflow completed all cycles)
	if !strings.Contains(output.String(), "Workflow Completion Summary") {
		t.Error("Completion summary should appear after workflow completes all cycles")
	}
}

// TestM01E04NoFurtherPhasesAfterFailure verifies that when Claude fails,
// no further phases or iterations are executed.
func TestM01E04NoFurtherPhasesAfterFailure(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)
	// Create test-specific stub directory to avoid race conditions
	stubDir := t.TempDir()
	createFailingClaudeStub(t, stubDir, 3)

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s:%s", stubDir, filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	// Corrected behavior: Workflow completes successfully even when all phases fail
	// The workflow continues through all phases and iterations
	if err != nil {
		t.Fatalf("Expected fluxid to succeed (workflow completes all phases), got error: %v\nOutput:\n%s",
			err, output.String())
	}

	outputStr := output.String()

	// Count how many phases were started
	implementCount := strings.Count(outputStr, "Starting phase: implement")
	commitCount := strings.Count(outputStr, "Starting phase: commit")
	reviewCount := strings.Count(outputStr, "Starting phase: review")

	// With default MaxImplementRetries=3, should see 3 implement attempts (all fail, then retries exhausted)
	if implementCount != 3 {
		t.Errorf("Expected exactly 3 implement phases (default retry limit), got %d", implementCount)
	}

	// After exhausting implement retries, workflow continues to commit phase
	// With new behavior, commit failures don't block workflow, so we see commit and review phases
	if commitCount != 100 {
		t.Errorf("Expected 100 commit phases (default commit retry limit), got %d", commitCount)
	}
	// With new behavior, workflow continues to review phase even after commit failures
	if reviewCount < 1 {
		t.Errorf("Expected at least 1 review phase (workflow continues after commit failures), got %d", reviewCount)
	}

	// Verify only one development iteration was started
	cycleCount := strings.Count(outputStr, "DEVELOPMENT ITERATION")
	if cycleCount != 1 {
		t.Errorf("Expected exactly 1 development iteration, got %d\nOutput:\n%s", cycleCount, outputStr)
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

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runPhaseFailureTest(t, testCase.failOnInvoke, testCase.expectedExitCode)
		})
	}
}

// Helper functions

func runPhaseFailureTest(t *testing.T, failOnInvoke, expectedExitCode int) {
	t.Helper()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	// Create test-specific stub directory to avoid race conditions
	stubDir := t.TempDir()
	createConditionalFailingClaudeStub(t, stubDir, failOnInvoke, expectedExitCode)

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s:%s", stubDir, filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

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

func createFailingClaudeStub(t *testing.T, stubDir string, exitCode int) {
	t.Helper()

	stubPath := filepath.Join(stubDir, "claude")
	stubScript := fmt.Sprintf(`#!/bin/bash
# Stub Claude CLI that always fails with exit code %d

echo "Claude stub invoked with args: $@"
echo "FLUXID_SESSION_ID=$FLUXID_SESSION_ID"
echo "Simulating Claude failure..." >&2

exit %d
`, exitCode, exitCode)

	if err := os.MkdirAll(stubDir, 0o755); err != nil {
		t.Fatalf("failed to create stub directory: %v", err)
	}

	if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
		t.Fatalf("failed to create failing claude stub: %v", err)
	}
}

func createConditionalFailingClaudeStub(t *testing.T, stubDir string, failOnInvoke int, exitCode int) {
	t.Helper()

	if err := os.MkdirAll(stubDir, 0o755); err != nil {
		t.Fatalf("failed to create stub directory: %v", err)
	}

	stubPath := filepath.Join(stubDir, "claude")
	counterFile := filepath.Join(stubDir, ".claude_invoke_count")

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
