//nolint:paralleltest // E2E tests use shared infrastructure
package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestM08E01AllFailReportsWithIterationsExhausted verifies that the workflow
// completes successfully even when all phases (implement, commit, review) report
// FAIL status across multiple iterations until MaxReviewCycles is exhausted.
//
// This test ensures that:
// 1. Implement phase FAIL reports don't abort the workflow
// 2. Commit phase FAIL reports don't abort the workflow
// 3. Review phase FAIL reports don't abort the workflow
// 4. Workflow only exits when MaxReviewCycles is exhausted
//
// Configuration: MaxImplementRetries=2, MaxCommitRetries=2, MaxReviewCycles=2
//
// NOTE: Not using t.Parallel() because FAIL stubs conflict with PASS stubs
// when they overwrite the same agent binaries in root/bin.
//
//nolint:cyclop,funlen // E2E test with multiple validation checks
func TestM08E01AllFailReportsWithIterationsExhausted(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create Claude stub that always reports FAIL for all phases
	createClaudeFormatStubFail(t, root)

	tmpHome := t.TempDir()
	configDir := setupConfigDir(t, tmpHome)
	createCommandFiles(t, configDir)

	// Create config with custom retry values
	configContent := fmt.Sprintf(`agent: claude
iterations: 2
implement_retries: 2
commit_retries: 2
commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, configDir, configDir, configDir)
	configPath := filepath.Join(configDir, "config.yaml")
	writeRawConfigFile(t, configPath, configContent)

	binPath := filepath.Join(root, "bin", "fluxid")
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("test task"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run with config-specified iterations and retries
	cmd := exec.CommandContext(
		t.Context(),
		binPath,
		"--claude",
		"--file="+taskPath,
	)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	// Workflow should complete successfully even with all FAIL reports
	if err != nil {
		t.Fatalf(
			"Expected workflow to succeed (completes all phases despite FAIL reports), got error: %v\nOutput:\n%s",
			err,
			string(output),
		)
	}

	outputStr := string(output)

	// Verify workflow ran through all development iterations
	developmentIterationCount := strings.Count(outputStr, "DEVELOPMENT ITERATION")
	if developmentIterationCount != 2 {
		t.Errorf(
			"Expected 2 development iterations (MaxReviewCycles=2), got %d\nOutput:\n%s",
			developmentIterationCount,
			outputStr,
		)
	}

	// Verify implement phase ran with retries in each iteration
	// Each iteration should have MaxImplementRetries attempts
	implementPhaseCount := strings.Count(outputStr, "Starting phase: implement")
	expectedImplementCount := 2 * 2 // 2 iterations * 2 retries per iteration
	if implementPhaseCount != expectedImplementCount {
		t.Errorf(
			"Expected %d implement phases (2 iterations * 2 retries), got %d\nOutput:\n%s",
			expectedImplementCount,
			implementPhaseCount,
			outputStr,
		)
	}

	// Verify commit phase ran with retries in each iteration
	// Each iteration should have MaxCommitRetries attempts
	commitPhaseCount := strings.Count(outputStr, "Starting phase: commit")
	expectedCommitCount := 2 * 2 // 2 iterations * 2 retries per iteration
	if commitPhaseCount != expectedCommitCount {
		t.Errorf(
			"Expected %d commit phases (2 iterations * 2 retries), got %d\nOutput:\n%s",
			expectedCommitCount,
			commitPhaseCount,
			outputStr,
		)
	}

	// Verify review phase ran in each iteration
	reviewPhaseCount := strings.Count(outputStr, "Starting phase: review")
	if reviewPhaseCount != 2 {
		t.Errorf("Expected 2 review phases (1 per iteration), got %d\nOutput:\n%s", reviewPhaseCount, outputStr)
	}

	// Verify all phases reported FAIL status
	if !strings.Contains(outputStr, "Status:  FAIL") && !strings.Contains(outputStr, "Status: FAIL") {
		t.Error("Expected FAIL status reports in output")
	}

	// Verify exhaustion messages appeared
	if !strings.Contains(outputStr, "All") && !strings.Contains(outputStr, "attempts") {
		t.Error("Expected exhaustion messages for retries")
	}

	// Verify workflow completion summary
	if !strings.Contains(outputStr, "Workflow Completion Summary") ||
		!strings.Contains(outputStr, "Status: SUCCESS") {
		t.Error("Expected successful workflow completion summary despite FAIL reports")
	}
}

// TestM08E01EdgeCaseConfigOneOneOne tests the workflow with the minimum
// configuration: 1 implement retry, 1 commit retry, 1 development iteration.
//
// This edge case ensures the workflow behaves correctly even with minimal
// retry/iteration counts, and that all phases execute as configured without
// premature abortion.
//
// NOTE: Not using t.Parallel() because FAIL stubs conflict with PASS stubs
// when they overwrite the same agent binaries in root/bin.
//
//nolint:cyclop,funlen // E2E test with multiple validation checks
func TestM08E01EdgeCaseConfigOneOneOne(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create Claude stub that always reports FAIL for all phases
	createClaudeFormatStubFail(t, root)

	tmpHome := t.TempDir()
	configDir := setupConfigDir(t, tmpHome)
	createCommandFiles(t, configDir)

	// Create config with edge case retry values (1,1,1)
	configContent := fmt.Sprintf(`agent: claude
iterations: 1
implement_retries: 1
commit_retries: 1
commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, configDir, configDir, configDir)
	configPath := filepath.Join(configDir, "config.yaml")
	writeRawConfigFile(t, configPath, configContent)

	binPath := filepath.Join(root, "bin", "fluxid")
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("test task"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run with edge case configuration: 1,1,1
	cmd := exec.CommandContext(
		t.Context(),
		binPath,
		"--claude",
		"--file="+taskPath,
	)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	// Workflow should complete successfully even with edge case config and FAIL reports
	if err != nil {
		t.Fatalf(
			"Expected workflow to succeed with edge case config (1,1,1), got error: %v\nOutput:\n%s",
			err,
			string(output),
		)
	}

	outputStr := string(output)

	// Verify exactly 1 development iteration
	developmentIterationCount := strings.Count(outputStr, "DEVELOPMENT ITERATION")
	if developmentIterationCount != 1 {
		t.Errorf("Expected 1 development iteration, got %d\nOutput:\n%s", developmentIterationCount, outputStr)
	}

	// Verify exactly 1 implement phase attempt
	implementPhaseCount := strings.Count(outputStr, "Starting phase: implement")
	if implementPhaseCount != 1 {
		t.Errorf("Expected 1 implement phase, got %d\nOutput:\n%s", implementPhaseCount, outputStr)
	}

	// Verify exactly 1 commit phase attempt
	commitPhaseCount := strings.Count(outputStr, "Starting phase: commit")
	if commitPhaseCount != 1 {
		t.Errorf("Expected 1 commit phase, got %d\nOutput:\n%s", commitPhaseCount, outputStr)
	}

	// Verify exactly 1 review phase
	reviewPhaseCount := strings.Count(outputStr, "Starting phase: review")
	if reviewPhaseCount != 1 {
		t.Errorf("Expected 1 review phase, got %d\nOutput:\n%s", reviewPhaseCount, outputStr)
	}

	// Verify all phases reported FAIL status
	if !strings.Contains(outputStr, "Status:  FAIL") && !strings.Contains(outputStr, "Status: FAIL") {
		t.Error("Expected FAIL status reports in output")
	}

	// Verify workflow completion summary
	if !strings.Contains(outputStr, "Workflow Completion Summary") ||
		!strings.Contains(outputStr, "Status: SUCCESS") {
		t.Errorf(
			"Expected successful workflow completion summary with edge case config, got:\n%s",
			outputStr,
		)
	}

	// Verify no unexpected errors or aborts
	if strings.Contains(outputStr, "Workflow Aborted") {
		t.Error("Workflow should not abort with FAIL reports, it should complete all configured iterations")
	}
}
