package tests

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestM06E05LoopProgressionWithSyntheticPASS validates dry-run simulates loop progression with PASS reports.
func TestM06E05LoopProgressionWithSyntheticPASS(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath,
		"--fluxid-dry-run",
		"--fluxid-iterations", "3",
		"--fluxid-implement-retries", "2",
		"--claude")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-dry-run failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify simulation plan shows loop progression
	expectedPatterns := []string{
		"=== Simulation Plan ===",
		"Review Cycle 1/3",
		"Implement attempt 1/2",
		"Would execute: Iteration 1, Retry 1, Phase: implement",
		"Synthetic implement report: PASS",
		"Would execute: Iteration 1, Retry 1, Phase: review",
		"Synthetic review report: PASS",
		"Simulated workflow completed successfully after 1 review cycle(s)",
		"=== End Simulation ===",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Expected pattern %q not found in output:\n%s", pattern, output)
		}
	}

	// Verify no actual agent was invoked
	if strings.Contains(output, "Claude stub invoked") {
		t.Errorf("Agent should not be invoked in dry-run mode:\n%s", output)
	}
}

// TestM06E05PhaseTogglesRespectedInSimulation validates --fluxid-no-commit removes commit from plan.
func TestM06E05PhaseTogglesRespectedInSimulation(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath,
		"--fluxid-dry-run",
		"--fluxid-no-commit",
		"--claude")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-dry-run --fluxid-no-commit failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify commit phase does NOT appear when disabled
	if strings.Contains(output, "Phase: commit") {
		t.Errorf("Commit phase should not appear when disabled:\n%s", output)
	}

	// Verify implement and review phases are present
	if !strings.Contains(output, "Phase: implement") {
		t.Error("Expected implement phase in simulation")
	}
	if !strings.Contains(output, "Phase: review") {
		t.Error("Expected review phase in simulation")
	}

	// Verify synthetic PASS reports are present
	if !strings.Contains(output, "Synthetic implement report: PASS") {
		t.Error("Expected synthetic implement PASS report")
	}
	if !strings.Contains(output, "Synthetic review report: PASS") {
		t.Error("Expected synthetic review PASS report")
	}
}

// TestM06E05LoopCountsHonoredInSimulation validates loop counts affect simulation plan.
func TestM06E05LoopCountsHonoredInSimulation(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath,
		"--fluxid-dry-run",
		"--fluxid-iterations", "5",
		"--fluxid-implement-retries", "3",
		"--claude")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-dry-run with custom loop counts failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify configured loop counts appear in output
	expectedPatterns := []string{
		"Review Cycle 1/5",
		"Implement attempt 1/3",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Expected pattern %q not found in output:\n%s", pattern, output)
		}
	}

	// Simulation should complete after first cycle with PASS
	if !strings.Contains(output, "Simulated workflow completed successfully after 1 review cycle(s)") {
		t.Error("Expected simulation to complete after first cycle with synthetic PASS")
	}
}

// TestM06E05ExitCodeZeroOnSuccessfulSimulation validates exit code 0 on successful dry-run.
func TestM06E05ExitCodeZeroOnSuccessfulSimulation(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath,
		"--fluxid-dry-run",
		"--claude")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	// Run and verify exit code 0
	err := cmd.Run()
	if err != nil {
		t.Errorf("Expected exit code 0, but command failed: %v\nOutput:\n%s", err, stdout.String())
	}

	// Verify completion message is present
	output := stdout.String()
	if !strings.Contains(output, "Simulated workflow completed successfully") {
		t.Error("Expected completion message in output")
	}
}

// TestM06E05FirstRetrySucceedsWithSyntheticPASS validates first retry gets synthetic PASS.
func TestM06E05FirstRetrySucceedsWithSyntheticPASS(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath,
		"--fluxid-dry-run",
		"--fluxid-implement-retries", "5",
		"--claude")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-dry-run failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify only retry 1 is attempted (synthetic PASS ends retry loop)
	if strings.Contains(output, "Implement attempt 2/5") {
		t.Error("Expected first retry to succeed with synthetic PASS, but found retry 2")
	}

	// Verify retry 1 shows synthetic PASS
	if !strings.Contains(output, "Implement attempt 1/5") {
		t.Error("Expected implement attempt 1/5 in output")
	}
	if !strings.Contains(output, "Synthetic implement report: PASS") {
		t.Error("Expected synthetic implement PASS report")
	}
}

// TestM06E05CommitEnabledShowsCommitPhase validates commit appears when enabled.
func TestM06E05CommitEnabledShowsCommitPhase(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temp directory with project config that enables commit
	tmpDir := t.TempDir()
	configDir := setupConfigDir(t, tmpDir)
	configPath := filepath.Join(configDir, "config.yaml")
	writeRawConfigFile(t, configPath, "commit_enabled: true\n")

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")
	cmd.Dir = tmpDir

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-dry-run with commit enabled failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify commit phase appears
	if !strings.Contains(output, "Phase: commit") {
		t.Errorf("Expected commit phase in simulation with commit enabled:\n%s", output)
	}

	// Verify all phases appear in correct order
	implementIdx := strings.Index(output, "Phase: implement")
	commitIdx := strings.Index(output, "Phase: commit")
	reviewIdx := strings.Index(output, "Phase: review")

	if implementIdx == -1 || commitIdx == -1 || reviewIdx == -1 {
		t.Error("Expected all three phases (implement, commit, review) in output")
	}

	if implementIdx >= commitIdx || commitIdx >= reviewIdx {
		t.Error("Expected phases in order: implement -> commit -> review")
	}
}
