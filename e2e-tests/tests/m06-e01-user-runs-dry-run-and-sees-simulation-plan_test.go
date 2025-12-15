package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestM06E01DryRunSimulationBasic validates basic dry-run functionality.
func TestM06E01DryRunSimulationBasic(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-dry-run failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify dry-run header
	if !strings.Contains(output, "=== DRY RUN MODE - Simulation Only ===") {
		t.Errorf("Missing dry-run header in output:\n%s", output)
	}

	// Verify initialization section appears
	if !strings.Contains(output, "=== fluxid Workflow Initialization ===") {
		t.Errorf("Missing initialization section in output:\n%s", output)
	}

	// Verify simulation plan section
	if !strings.Contains(output, "=== Simulation Plan ===") {
		t.Errorf("Missing simulation plan section in output:\n%s", output)
	}

	// Verify simulation end marker
	if !strings.Contains(output, "=== End Simulation ===") {
		t.Errorf("Missing simulation end marker in output:\n%s", output)
	}

	// Verify exit code is 0
	// (already verified by cmd.Run() succeeding)
}

// TestM06E01DryRunShowsPhasesAndIterations validates the simulation plan output format.
func TestM06E01DryRunShowsPhasesAndIterations(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-dry-run failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify "Would execute:" lines with iteration/retry/phase
	expectedPatterns := []string{
		"Would execute: Iteration 1, Retry 1, Phase: implement",
		"Would execute: Iteration 1, Retry 1, Phase: review",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Expected pattern %q not found in output:\n%s", pattern, output)
		}
	}

	// Verify command file paths are shown
	if !strings.Contains(output, "Command file:") {
		t.Errorf("Missing command file path in output:\n%s", output)
	}

	// Should show either "built-in prompt" (no config) or actual paths (if home config exists)
	if !strings.Contains(output, "built-in prompt") && !strings.Contains(output, "Command file:") {
		t.Errorf("Expected command file information in output:\n%s", output)
	}
}

// TestM06E01DryRunNoAgentProcessSpawned validates that no agent process is created.
func TestM06E01DryRunNoAgentProcessSpawned(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-dry-run failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify stub agent is NOT invoked (stub prints "Claude stub invoked")
	if strings.Contains(output, "Claude stub invoked") {
		t.Errorf("Agent process was spawned in dry-run mode - output contains stub marker:\n%s", output)
	}

	// Verify no phase execution markers from actual workflow
	if strings.Contains(output, "Starting phase:") {
		t.Errorf("Actual phase execution detected in dry-run mode:\n%s", output)
	}

	// Verify no report waiting messages
	if strings.Contains(output, "Waiting for") && strings.Contains(output, "report") {
		t.Errorf("Report waiting detected in dry-run mode:\n%s", output)
	}
}

// TestM06E01DryRunWithCommitPhase validates commit phase appears when enabled.
func TestM06E01DryRunWithCommitPhase(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Test with commit enabled via config
	t.Run("commit enabled", func(t *testing.T) {
		t.Parallel()
		// Create temp directory with project config that enables commit
		tmpDir := t.TempDir()
		configDir := setupConfigDir(t, tmpDir)
		configPath := filepath.Join(configDir, "config.yaml")
		writeRawConfigFile(t, configPath, "commit_enabled: true\n")

		binPath := filepath.Join(root, "bin", "fluxid")
		cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stdout

		if err := cmd.Run(); err != nil {
			t.Fatalf("fluxid --fluxid-dry-run failed: %v\nOutput:\n%s", err, stdout.String())
		}

		output := stdout.String()

		// Verify commit phase appears
		if !strings.Contains(output, "Phase: commit") {
			t.Errorf("Expected commit phase in simulation with commit enabled:\n%s", output)
		}
	})

	// Test with commit disabled
	t.Run("commit disabled", func(t *testing.T) {
		t.Parallel()
		binPath := filepath.Join(root, "bin", "fluxid")
		cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--fluxid-no-commit", "--claude")

		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stdout

		if err := cmd.Run(); err != nil {
			t.Fatalf("fluxid --fluxid-dry-run --fluxid-no-commit failed: %v\nOutput:\n%s", err, stdout.String())
		}

		output := stdout.String()

		// Verify commit phase does NOT appear
		if strings.Contains(output, "Phase: commit") {
			t.Errorf("Commit phase should not appear when disabled:\n%s", output)
		}
	})
}

// TestM06E01DryRunWithConfigValues validates dry-run uses resolved configuration.
func TestM06E01DryRunWithConfigValues(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath,
		"--fluxid-dry-run",
		"--fluxid-iterations", "5",
		"--fluxid-implement-retries", "2",
		"--claude")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-dry-run with config failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify config values are shown in initialization
	if !strings.Contains(output, "Max Review Cycles: 5") {
		t.Errorf("Expected iterations=5 in output:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 2") {
		t.Errorf("Expected implement retries=2 in output:\n%s", output)
	}
}

// TestM06E01DryRunWithInvalidConfigFails validates configuration errors are caught.
func TestM06E01DryRunWithInvalidConfigFails(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath,
		"--fluxid-dry-run",
		"--fluxid-iterations", "0", // Invalid: must be >= 1
		"--claude")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected error with invalid config, but succeeded")
	}

	// Verify error occurs before simulation
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "must be a positive integer") {
		t.Errorf("Expected validation error in stderr, got:\n%s", stderrOutput)
	}
}
