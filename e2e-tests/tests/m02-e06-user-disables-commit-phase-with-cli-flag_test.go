package tests

import (
	"strings"
	"testing"
)

const (
	homeConfigCommitEnabled = `commit_enabled: true
`
)

// TestM02E06NoCommitFlagDisablesCommitPhase validates that --fluxid-no-commit
// prevents commit phase execution while keeping review phase mandatory.
func TestM02E06NoCommitFlagDisablesCommitPhase(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config with commit enabled
	tmpHome := setupHomeWithConfig(t, homeConfigCommitEnabled)
	tmpProjectDir := createProjectWithConfig(t, "")

	// Run with --fluxid-no-commit flag
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-no-commit",
	)

	// Verify initialization shows commit disabled
	if !strings.Contains(output, "Commit Enabled: false (source: cli)") {
		t.Errorf("Expected Commit Enabled: false (source: cli), got:\n%s", output)
	}

	// Verify commit phase never executed
	if strings.Contains(output, "Running commit phase...") {
		t.Errorf("Commit phase should not execute when --fluxid-no-commit is set, got:\n%s", output)
	}

	// Verify review phase still runs
	if !strings.Contains(output, "Running review phase...") {
		t.Errorf("Review phase should still execute when --fluxid-no-commit is set, got:\n%s", output)
	}

	// Verify workflow completes successfully
	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Workflow should complete successfully with --fluxid-no-commit, got:\n%s", output)
	}
}

// TestM02E06CommitPhaseRunsWithoutFlag validates that commit phase executes
// normally when --fluxid-no-commit is not set.
func TestM02E06CommitPhaseRunsWithoutFlag(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config with commit enabled
	tmpHome := setupHomeWithConfig(t, homeConfigCommitEnabled)
	tmpProjectDir := createProjectWithConfig(t, "")

	// Run WITHOUT --fluxid-no-commit flag
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir)

	// Verify initialization shows commit enabled
	if !strings.Contains(output, "Commit Enabled: true (source: home)") {
		t.Errorf("Expected Commit Enabled: true (source: home), got:\n%s", output)
	}

	// Verify commit phase executes
	if !strings.Contains(output, "Running commit phase...") {
		t.Errorf("Commit phase should execute when commit_enabled=true, got:\n%s", output)
	}

	// Verify review phase runs
	if !strings.Contains(output, "Running review phase...") {
		t.Errorf("Review phase should execute, got:\n%s", output)
	}
}

// TestM02E06NoCommitFlagOverridesConfig validates that --fluxid-no-commit
// overrides commit_enabled from config files.
func TestM02E06NoCommitFlagOverridesConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create project config with commit enabled
	tmpHome := t.TempDir()
	tmpProjectDir := createProjectWithConfig(t, homeConfigCommitEnabled)

	// Run with --fluxid-no-commit (should override project config)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-no-commit",
	)

	// Verify CLI flag overrides config
	if !strings.Contains(output, "Commit Enabled: false (source: cli)") {
		t.Errorf("Expected Commit Enabled: false (source: cli), got:\n%s", output)
	}

	// Verify commit phase skipped
	if strings.Contains(output, "Running commit phase...") {
		t.Errorf("Commit phase should be skipped, got:\n%s", output)
	}
}

// TestM02E06ReviewPhaseMandatory validates that review phase always runs
// regardless of --fluxid-no-commit flag.
func TestM02E06ReviewPhaseMandatory(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		args          []string
		expectCommit  bool
		expectReview  bool
		commitEnabled string
	}{
		{
			name:          "with_no_commit_flag",
			args:          []string{"--fluxid-no-commit"},
			expectCommit:  false,
			expectReview:  true,
			commitEnabled: "false (source: cli)",
		},
		{
			name:          "without_no_commit_flag_default",
			args:          []string{},
			expectCommit:  false,
			expectReview:  true,
			commitEnabled: "false (source: default)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := getProjectRoot(t)
			buildFluxid(t, root)
			createStubClaude(t, root)

			tmpHome := t.TempDir()
			tmpProjectDir := t.TempDir()

			output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir, tc.args...)

			// Verify commit enabled status
			expectedStatus := "Commit Enabled: " + tc.commitEnabled
			if !strings.Contains(output, expectedStatus) {
				t.Errorf("Expected '%s', got:\n%s", expectedStatus, output)
			}

			// Verify commit phase execution
			hasCommitPhase := strings.Contains(output, "Running commit phase...")
			if hasCommitPhase != tc.expectCommit {
				if tc.expectCommit {
					t.Errorf("Expected commit phase to run, but it didn't. Output:\n%s", output)
				} else {
					t.Errorf("Expected commit phase to be skipped, but it ran. Output:\n%s", output)
				}
			}

			// Verify review phase ALWAYS runs
			if !strings.Contains(output, "Running review phase...") {
				t.Errorf("Review phase must always run, got:\n%s", output)
			}
		})
	}
}

// TestM02E06PhaseExecutionOrder validates that when commit is enabled,
// phases execute in correct order: implement -> commit -> review.
func TestM02E06PhaseExecutionOrder(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create config with commit enabled
	tmpHome := setupHomeWithConfig(t, homeConfigCommitEnabled)
	tmpProjectDir := t.TempDir()

	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir)

	// Find phase execution order by string positions
	implementIdx := strings.Index(output, "Implement attempt 1/3...")
	commitIdx := strings.Index(output, "Running commit phase...")
	reviewIdx := strings.Index(output, "Running review phase...")

	// Verify all phases present
	if implementIdx == -1 {
		t.Error("Implement phase not found in output")
	}
	if commitIdx == -1 {
		t.Error("Commit phase not found in output")
	}
	if reviewIdx == -1 {
		t.Error("Review phase not found in output")
	}

	// Verify order: implement < commit < review
	if implementIdx > commitIdx {
		t.Error("Implement phase should occur before commit phase")
	}
	if commitIdx > reviewIdx {
		t.Error("Commit phase should occur before review phase")
	}
}

// TestM02E06PhaseExecutionOrderNoCommit validates that when commit is disabled,
// phases execute in order: implement -> review (commit skipped).
func TestM02E06PhaseExecutionOrderNoCommit(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpProjectDir := t.TempDir()

	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-no-commit",
	)

	// Find phase execution order by string positions
	implementIdx := strings.Index(output, "Implement attempt 1/3...")
	reviewIdx := strings.Index(output, "Running review phase...")

	// Verify implement and review phases present
	if implementIdx == -1 {
		t.Error("Implement phase not found in output")
	}
	if reviewIdx == -1 {
		t.Error("Review phase not found in output")
	}

	// Verify commit phase NOT present
	if strings.Contains(output, "Running commit phase...") {
		t.Error("Commit phase should not execute when --fluxid-no-commit is set")
	}

	// Verify order: implement < review
	if implementIdx > reviewIdx {
		t.Error("Implement phase should occur before review phase")
	}
}
