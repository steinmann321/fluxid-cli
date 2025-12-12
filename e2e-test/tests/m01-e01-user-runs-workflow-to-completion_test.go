package tests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestM01E01UserRunsWorkflowToCompletion validates the full workflow execution
// with default settings and verifies exit code 0.
func TestM01E01UserRunsWorkflowToCompletion(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithClaude(t, root)

	verifyInitialization(t, output)
	verifyPhaseExecution(t, output)
	verifyCompletionSummary(t, output)
}

// TestM01E01SessionIDUniqueness verifies that each run generates a unique UUID v4.
func TestM01E01SessionIDUniqueness(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	sessionIDs := make(map[string]bool)
	uuidV4Pattern := `([0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})`
	sessionIDPattern := regexp.MustCompile(`Session ID: ` + uuidV4Pattern)

	// Run fluxid 3 times and collect session IDs
	for i := 0; i < 3; i++ {
		binPath := filepath.Join(root, "bin", "fluxid")
		cmd := exec.CommandContext(t.Context(), binPath, "--claude")
		cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stdout

		if err := cmd.Run(); err != nil {
			t.Fatalf("Run %d failed: %v", i+1, err)
		}

		matches := sessionIDPattern.FindStringSubmatch(stdout.String())
		if len(matches) < 2 {
			t.Fatalf("Run %d: no session ID found", i+1)
		}

		sessionID := matches[1]

		// Verify UUID v4 format (version 4 has specific bits set)
		if !isValidUUIDv4(sessionID) {
			t.Errorf("Run %d: session ID %s is not a valid UUID v4", i+1, sessionID)
		}

		if sessionIDs[sessionID] {
			t.Errorf("Run %d: duplicate session ID %s", i+1, sessionID)
		}
		sessionIDs[sessionID] = true
	}

	if len(sessionIDs) != 3 {
		t.Errorf("Expected 3 unique session IDs, got %d", len(sessionIDs))
	}
}

// TestM01E01ClaudeArgsPassthrough verifies that arbitrary Claude arguments
// are accepted and passed through correctly.
func TestM01E01ClaudeArgsPassthrough(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--custom-arg", "value", "--another-flag")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid with custom args failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify custom args are shown in initialization
	if !strings.Contains(output, "Claude Args:") {
		t.Errorf("Claude Args not displayed in output")
	}

	// Verify custom args appear in output (from stub echo)
	hasCustomArg := strings.Contains(output, "--custom-arg")
	hasValue := strings.Contains(output, "value")
	hasAnotherFlag := strings.Contains(output, "--another-flag")
	if !hasCustomArg || !hasValue || !hasAnotherFlag {
		t.Errorf("Custom arguments not passed through to Claude")
	}
}

// TestM01E01WithoutClaudeFlag verifies that the CLI exits with error when
// --claude flag is not provided.
func TestM01E01WithoutClaudeFlag(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected fluxid without --claude to fail, but it succeeded")
	}

	// Verify exit code is 1
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Errorf("Expected exit code 1, got: %v", err)
	}

	// Verify usage message
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("Expected usage message in stderr, got: %s", stderr.String())
	}
}

// TestM01E01SessionIDPropagation verifies that FLUXID_SESSION_ID environment
// variable is propagated to child Claude processes.
func TestM01E01SessionIDPropagation(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --claude failed: %v", err)
	}

	output := stdout.String()

	// Extract session ID from initialization output
	uuidV4Pattern := `([0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})`
	sessionIDPattern := regexp.MustCompile(`Session ID: ` + uuidV4Pattern)
	matches := sessionIDPattern.FindStringSubmatch(output)
	if len(matches) < 2 {
		t.Fatal("Could not find session ID in output")
	}
	sessionID := matches[1]

	// Verify stub Claude received the session ID (stub echoes it)
	expectedEnv := fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID)
	if !strings.Contains(output, expectedEnv) {
		t.Errorf("FLUXID_SESSION_ID not propagated to Claude process (expected %s in output)", expectedEnv)
	}
}

// Helper functions

func getProjectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	root, err := findProjectRoot(wd)
	if err != nil {
		t.Fatalf("find project root failed: %v", err)
	}

	return root
}

func buildFluxid(t *testing.T, root string) {
	t.Helper()

	// Build fluxid binary
	build := exec.CommandContext(t.Context(), "go", "build", "-o", "bin/fluxid", "./cmd/fluxid")
	build.Dir = root

	var stderr bytes.Buffer
	build.Stderr = &stderr

	if err := build.Run(); err != nil {
		t.Fatalf("build failed: %v\nStderr: %s", err, stderr.String())
	}
}

func createStubClaude(t *testing.T, root string) {
	t.Helper()

	// Create stub Claude binary for testing
	stubPath := filepath.Join(root, "bin", "claude")

	stubScript := `#!/bin/bash
# Stub Claude CLI for testing

# Echo all arguments to demonstrate passthrough
echo "Claude stub invoked with args: $@"

# Echo environment variables for validation
echo "FLUXID_SESSION_ID=$FLUXID_SESSION_ID"

# Simulate successful execution
exit 0
`

	if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
		t.Fatalf("failed to create stub claude: %v", err)
	}
}

func isValidUUIDv4(uuid string) bool {
	// UUID v4 has format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	// where y is one of [8, 9, a, b]
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return pattern.MatchString(uuid)
}

func runFluxidWithClaude(t *testing.T, root string, args ...string) string {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmdArgs := append([]string{"--claude"}, args...)
	cmd := exec.CommandContext(t.Context(), binPath, cmdArgs...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --claude failed: %v\nOutput:\n%s", err, stdout.String())
	}

	return stdout.String()
}

func verifyInitialization(t *testing.T, output string) {
	t.Helper()

	if !strings.Contains(output, "=== fluxid Workflow Initialization ===") {
		t.Errorf("Missing initialization header in output")
	}

	if !strings.Contains(output, "Agent: claude") {
		t.Errorf("Missing agent selection in output")
	}

	uuidV4Pattern := `[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}`
	sessionIDPattern := regexp.MustCompile(`Session ID: ` + uuidV4Pattern)
	if !sessionIDPattern.MatchString(output) {
		t.Errorf("Missing or invalid session ID (expected UUID v4 format)")
	}

	if !strings.Contains(output, "Max Review Cycles: 20") {
		t.Errorf("Missing max review cycles in output")
	}

	if !strings.Contains(output, "Max Implement Retries: 3") {
		t.Errorf("Missing max implement retries in output")
	}
}

func verifyPhaseExecution(t *testing.T, output string) {
	t.Helper()

	if !strings.Contains(output, "Review Cycle 1/20") {
		t.Errorf("Missing review cycle indicator")
	}

	if !strings.Contains(output, "Starting phase: implement") {
		t.Errorf("Missing implement phase")
	}

	if !strings.Contains(output, "Starting phase: commit") {
		t.Errorf("Missing commit phase")
	}

	if !strings.Contains(output, "Starting phase: review") {
		t.Errorf("Missing review phase")
	}
}

func verifyCompletionSummary(t *testing.T, output string) {
	t.Helper()

	if !strings.Contains(output, "=== Workflow Completion Summary ===") {
		t.Errorf("Missing completion summary header")
	}

	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Missing success status in completion summary")
	}

	if !strings.Contains(output, "All workflow loops completed.") {
		t.Errorf("Missing completion message")
	}
}
