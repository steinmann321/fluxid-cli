//nolint:paralleltest // E2E tests use shared infrastructure
package tests

import (
	"bytes"
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
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithClaude(t, root, "--fluxid-iterations=1")

	verifyInitialization(t, output)
	verifyPhaseExecution(t, output)
	verifyCompletionSummary(t, output)
}

// TestM01E01SessionIDUniqueness verifies that each run generates a unique UUID v4.
func TestM01E01SessionIDUniqueness(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	sessionIDs := make(map[string]bool)
	uuidV4Pattern := `([0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})`
	sessionIDPattern := regexp.MustCompile(`Session ID: ` + uuidV4Pattern)

	// Run fluxid 3 times and collect session IDs
	for runIndex := 0; runIndex < 3; runIndex++ {
		binPath := filepath.Join(root, "bin", "fluxid")
		// Create a dummy task file in home
		taskPath := filepath.Join(tmpHome, "task.txt")
		if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
			t.Fatalf("Failed to write task file: %v", err)
		}
		cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations=1", "--file="+taskPath)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
			"HOME="+tmpHome,
		)

		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stdout

		if err := cmd.Run(); err != nil {
			t.Fatalf("Run %d failed: %v", runIndex+1, err)
		}

		matches := sessionIDPattern.FindStringSubmatch(stdout.String())
		if len(matches) < 2 {
			t.Fatalf("Run %d: no session ID found", runIndex+1)
		}

		sessionID := matches[1]

		// Verify UUID v4 format (version 4 has specific bits set)
		if !isValidUUIDv4(sessionID) {
			t.Errorf("Run %d: session ID %s is not a valid UUID v4", runIndex+1, sessionID)
		}

		if sessionIDs[sessionID] {
			t.Errorf("Run %d: duplicate session ID %s", runIndex+1, sessionID)
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
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	testAgentArgsPassthrough(t, root, "--claude", "Agent Args:", []string{"--custom-arg", "value", "--another-flag"})
}

// TestM01E01WithoutClaudeFlag verifies that when no agent flag is provided,
// the system uses the default agent (claude) and completes successfully.
func TestM01E01WithoutClaudeFlag(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root) // Create stub claude since it's the default agent

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	// Create a dummy task file in home
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	if err != nil {
		t.Fatalf(
			"Expected fluxid without flags to succeed with default agent, but it failed: %v\nOutput:\n%s",
			err, stdout.String(),
		)
	}

	output := stdout.String()

	// Verify default agent is used
	if !strings.Contains(output, "Agent: claude") {
		t.Errorf("Expected default agent 'claude' to be used, got: %s", output)
	}

	// Verify workflow completes successfully
	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Expected successful completion with default agent, got: %s", output)
	}
}

// TestM01E01SessionIDPropagation verifies that FLUXID_SESSION_ID environment
// variable is propagated to child Claude processes.
func TestM01E01SessionIDPropagation(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	// Create a dummy task file in home
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

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
	expectedEnv := "FLUXID_SESSION_ID=" + sessionID
	if !strings.Contains(output, expectedEnv) {
		t.Errorf("FLUXID_SESSION_ID not propagated to Claude process (expected %s in output)", expectedEnv)
	}
}

// Helper functions

func isValidUUIDv4(uuid string) bool {
	// UUID v4 has format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	// where y is one of [8, 9, a, b]
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return pattern.MatchString(uuid)
}

func runFluxidWithClaude(t *testing.T, root string, args ...string) string {
	t.Helper()

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	cmdArgs := append([]string{"--claude"}, args...)
	// Create dummy task file in home for real runs
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}
	cmdArgs = append(cmdArgs, "--file="+taskPath)
	cmd := exec.CommandContext(t.Context(), binPath, cmdArgs...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

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

	if !strings.Contains(output, "Workflow Initialization") {
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

	if !strings.Contains(output, "Max Review Cycles: 1") {
		t.Errorf("Missing max review cycles in output")
	}

	if !strings.Contains(output, "Max Implement Retries: 3") {
		t.Errorf("Missing max implement retries in output")
	}
}

func verifyPhaseExecution(t *testing.T, output string) {
	t.Helper()

	if !strings.Contains(output, "Review Cycle 1/1") {
		t.Errorf("Missing review cycle indicator")
	}

	if !strings.Contains(output, "Starting phase: implement") {
		t.Errorf("Missing implement phase")
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
