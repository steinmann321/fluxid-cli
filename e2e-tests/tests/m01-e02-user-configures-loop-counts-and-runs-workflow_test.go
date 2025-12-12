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

// TestM01E02ConfigureLoopCounts validates that --fluxid-iterations and
// --fluxid-implement-retries flags override defaults correctly.
func TestM01E02ConfigureLoopCounts(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithClaude(t, root,
		"--fluxid-iterations", "5",
		"--fluxid-implement-retries", "2",
		"--fluxid-no-commit")

	// Verify initialization displays overrides
	if !strings.Contains(output, "Max Review Cycles: 5") {
		t.Errorf("Expected Max Review Cycles: 5 in initialization, got output:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 2") {
		t.Errorf("Expected Max Implement Retries: 2 in initialization, got output:\n%s", output)
	}

	// Verify session ID is present
	uuidV4Pattern := `[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}`
	sessionIDPattern := regexp.MustCompile(`Session ID: ` + uuidV4Pattern)
	if !sessionIDPattern.MatchString(output) {
		t.Errorf("Missing or invalid session ID in output")
	}

	// Verify workflow executes with correct counts
	if !strings.Contains(output, "Review Cycle 1/5") {
		t.Errorf("Expected Review Cycle 1/5, got output:\n%s", output)
	}

	// Verify exit code 0 (implicit - test passed if runFluxidWithClaude didn't fail)
}

// TestM01E02InvalidIterationsZero validates that --fluxid-iterations 0 is rejected.
func TestM01E02InvalidIterationsZero(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations", "0")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected fluxid to fail with --fluxid-iterations 0, but it succeeded")
	}

	// Verify exit code is 1
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Errorf("Expected exit code 1, got: %v", err)
	}

	// Verify error message mentions positive integer requirement
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "positive integer") && !strings.Contains(errOutput, "≥1") {
		t.Errorf("Expected error message about positive integer requirement, got: %s", errOutput)
	}

	if !strings.Contains(errOutput, "0") {
		t.Errorf("Expected error message to show invalid value 0, got: %s", errOutput)
	}
}

// TestM01E02InvalidIterationsNegative validates that negative iterations are rejected.
func TestM01E02InvalidIterationsNegative(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations", "-1")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected fluxid to fail with --fluxid-iterations -1, but it succeeded")
	}

	// Verify exit code is 1
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Errorf("Expected exit code 1, got: %v", err)
	}

	// Verify error message
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "positive integer") && !strings.Contains(errOutput, "≥1") {
		t.Errorf("Expected error message about positive integer requirement, got: %s", errOutput)
	}
}

// TestM01E02InvalidIterationsNonInteger validates that non-integer values are rejected.
func TestM01E02InvalidIterationsNonInteger(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations", "abc")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected fluxid to fail with --fluxid-iterations abc, but it succeeded")
	}

	// Verify exit code is 1
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Errorf("Expected exit code 1, got: %v", err)
	}

	// Verify error message mentions invalid value
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "abc") {
		t.Errorf("Expected error message to show invalid value 'abc', got: %s", errOutput)
	}
}

// TestM01E02InvalidRetriesZero validates that --fluxid-implement-retries 0 is rejected.
func TestM01E02InvalidRetriesZero(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-implement-retries", "0")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected fluxid to fail with --fluxid-implement-retries 0, but it succeeded")
	}

	// Verify exit code is 1
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Errorf("Expected exit code 1, got: %v", err)
	}

	// Verify error message
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "positive integer") && !strings.Contains(errOutput, "≥1") {
		t.Errorf("Expected error message about positive integer requirement, got: %s", errOutput)
	}
}

// TestM01E02DefaultsAppliedWhenFlagsOmitted validates that defaults are used
// when override flags are not provided.
func TestM01E02DefaultsAppliedWhenFlagsOmitted(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithClaude(t, root, "--fluxid-no-commit")

	// Verify defaults are shown
	if !strings.Contains(output, "Max Review Cycles: 20") {
		t.Errorf("Expected default Max Review Cycles: 20, got output:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 3") {
		t.Errorf("Expected default Max Implement Retries: 3, got output:\n%s", output)
	}
}

// TestM01E02PartialOverride validates that only one flag can be overridden
// while the other uses default.
func TestM01E02PartialOverride(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithClaude(t, root, "--fluxid-iterations", "7", "--fluxid-no-commit")

	// Verify custom iteration count
	if !strings.Contains(output, "Max Review Cycles: 7") {
		t.Errorf("Expected Max Review Cycles: 7, got output:\n%s", output)
	}

	// Verify default retries
	if !strings.Contains(output, "Max Implement Retries: 3") {
		t.Errorf("Expected default Max Implement Retries: 3, got output:\n%s", output)
	}
}

// TestM01E02SuccessfulCompletion validates that workflow completes successfully
// with custom loop counts and exits with code 0.
func TestM01E02SuccessfulCompletion(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithClaude(t, root,
		"--fluxid-iterations", "5",
		"--fluxid-implement-retries", "2",
		"--fluxid-no-commit")

	// Verify completion summary
	if !strings.Contains(output, "=== Workflow Completion Summary ===") {
		t.Errorf("Missing completion summary header")
	}

	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Missing success status")
	}

	// Exit code 0 is implicit - runFluxidWithClaude would fail otherwise
}

// TestM01E02ClaudeArgsPassthroughWithOverrides validates that Claude args
// work correctly even when fluxid-specific flags are used.
func TestM01E02ClaudeArgsPassthroughWithOverrides(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath,
		"--claude",
		"--fluxid-iterations", "5",
		"--fluxid-implement-retries", "2",
		"--custom-arg", "value")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid with overrides and custom args failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify overrides are applied
	if !strings.Contains(output, "Max Review Cycles: 5") {
		t.Errorf("Expected Max Review Cycles: 5")
	}

	if !strings.Contains(output, "Max Implement Retries: 2") {
		t.Errorf("Expected Max Implement Retries: 2")
	}

	// Verify custom args are passed through
	if !strings.Contains(output, "--custom-arg") || !strings.Contains(output, "value") {
		t.Errorf("Custom arguments not passed through to Claude")
	}
}
