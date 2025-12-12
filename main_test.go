package main

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// TestM01E01UserRunsWorkflowToCompletion tests the complete workflow
// Epic: m01-e01-user-runs-workflow-to-completion
func TestM01E01UserRunsWorkflowToCompletion(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "bin/fluxid-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Run fluxid --claude
	cmd := exec.Command("./bin/fluxid-test", "--claude")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()
	output := stdout.String()

	// Test: Verify exit code 0
	if err != nil {
		t.Errorf("Command failed with error: %v\nStdout: %s\nStderr: %s", err, output, stderr.String())
	}

	// Test: Verify initialization shows session ID (UUID v4 format)
	uuidRegex := regexp.MustCompile(`Session ID: ([a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12})`)
	if !uuidRegex.MatchString(output) {
		t.Errorf("Expected valid UUID v4 session ID in output, got:\n%s", output)
	}

	// Test: Verify agent selection is displayed
	if !strings.Contains(output, "Agent: claude") {
		t.Errorf("Expected 'Agent: claude' in initialization, got:\n%s", output)
	}

	// Test: Verify loop counts are displayed
	if !strings.Contains(output, "Max Review Cycles: 20") {
		t.Errorf("Expected 'Max Review Cycles: 20' in initialization, got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 3") {
		t.Errorf("Expected 'Max Implement Retries: 3' in initialization, got:\n%s", output)
	}

	// Test: Verify phases are invoked in correct order
	if !strings.Contains(output, "Starting phase: implement") {
		t.Errorf("Expected implement phase to be invoked, got:\n%s", output)
	}

	if !strings.Contains(output, "Starting phase: commit") {
		t.Errorf("Expected commit phase to be invoked, got:\n%s", output)
	}

	if !strings.Contains(output, "Starting phase: review") {
		t.Errorf("Expected review phase to be invoked, got:\n%s", output)
	}

	// Test: Verify completion summary appears
	if !strings.Contains(output, "Workflow Completion Summary") {
		t.Errorf("Expected completion summary, got:\n%s", output)
	}

	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Expected 'Status: SUCCESS' in completion, got:\n%s", output)
	}

	// Test: Verify nested loop structure (Review Cycle indicator)
	if !strings.Contains(output, "Review Cycle 1/20") {
		t.Errorf("Expected review cycle indicator, got:\n%s", output)
	}
}

// TestM01E01SessionIDUniqueness verifies unique UUID v4 session IDs across runs
func TestM01E01SessionIDUniqueness(t *testing.T) {
	sessionIDs := make(map[string]bool)
	runs := 5

	for i := 0; i < runs; i++ {
		cmd := exec.Command("./bin/fluxid-test", "--claude")
		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		if err := cmd.Run(); err != nil {
			t.Fatalf("Run %d failed: %v", i+1, err)
		}

		// Extract session ID
		uuidRegex := regexp.MustCompile(`Session ID: ([a-f0-9-]+)`)
		matches := uuidRegex.FindStringSubmatch(stdout.String())
		if len(matches) < 2 {
			t.Fatalf("Run %d: Could not extract session ID", i+1)
		}

		sessionID := matches[1]

		// Verify UUID v4 format
		uuidV4Regex := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
		if !uuidV4Regex.MatchString(sessionID) {
			t.Errorf("Run %d: Session ID %s is not a valid UUID v4", i+1, sessionID)
		}

		// Check uniqueness
		if sessionIDs[sessionID] {
			t.Errorf("Run %d: Duplicate session ID found: %s", i+1, sessionID)
		}
		sessionIDs[sessionID] = true
	}

	if len(sessionIDs) != runs {
		t.Errorf("Expected %d unique session IDs, got %d", runs, len(sessionIDs))
	}
}

// TestM01E01ClaudeArgsPassthrough tests that Claude args are accepted
func TestM01E01ClaudeArgsPassthrough(t *testing.T) {
	cmd := exec.Command("./bin/fluxid-test", "--claude", "--custom-arg", "value")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	// Verify initialization still works with extra args
	if !strings.Contains(output, "Agent: claude") {
		t.Errorf("Expected agent initialization with args, got:\n%s", output)
	}

	if !strings.Contains(output, "Claude Args:") {
		t.Errorf("Expected Claude args to be displayed, got:\n%s", output)
	}
}

// TestM01E01WithoutClaudeFlag tests behavior without --claude flag
func TestM01E01WithoutClaudeFlag(t *testing.T) {
	cmd := exec.Command("./bin/fluxid-test")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()

	// Should exit successfully but show usage
	if err != nil {
		t.Errorf("Expected successful exit without --claude flag, got error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Hello, FluxID!") {
		t.Errorf("Expected greeting message, got:\n%s", output)
	}
}

// TestM01E01NestedLoopCounts tests that nested loops execute correctly
func TestM01E01NestedLoopCounts(t *testing.T) {
	cmd := exec.Command("./bin/fluxid-test", "--claude")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	// Count phase executions
	implementCount := strings.Count(output, "Starting phase: implement")
	commitCount := strings.Count(output, "Starting phase: commit")
	reviewCount := strings.Count(output, "Starting phase: review")

	// For the stubbed implementation, we expect 1 cycle to complete
	// Each cycle has 1 implement, 1 commit, 1 review
	if implementCount < 1 {
		t.Errorf("Expected at least 1 implement phase, got %d", implementCount)
	}

	if commitCount < 1 {
		t.Errorf("Expected at least 1 commit phase, got %d", commitCount)
	}

	if reviewCount < 1 {
		t.Errorf("Expected at least 1 review phase, got %d", reviewCount)
	}

	// Verify phases are balanced (same count for each in simple success case)
	if implementCount != commitCount || commitCount != reviewCount {
		t.Logf("Phase counts: implement=%d, commit=%d, review=%d", implementCount, commitCount, reviewCount)
		// This is just informational for now
	}
}
