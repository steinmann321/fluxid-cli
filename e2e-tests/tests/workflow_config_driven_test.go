//nolint:paralleltest // E2E tests use shared infrastructure
package tests

import (
	"strings"
	"testing"
)

// TestConfigDrivenWorkflowMinimal validates workflow with 1 custom step + review.
// Verifies: implement → review → exit.
func TestConfigDrivenWorkflowMinimal(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := setupWorkflowTest(t, root, "workflow_minimal.yaml")
	createWorkflowCommandFiles(t, tmpHome, []string{"implement", "review"})
	taskPath := createTaskFile(t, tmpHome, "test task")

	output := runFluxidWorkflow(t, root, tmpHome, taskPath)

	// Verify steps executed in order
	if !strings.Contains(output, "implement") {
		t.Error("Implement step not found in output")
	}
	if !strings.Contains(output, "review") {
		t.Error("Review step not found in output")
	}

	// Verify review is last
	implementIdx := strings.Index(output, "implement")
	reviewIdx := strings.LastIndex(output, "review")
	if implementIdx >= reviewIdx {
		t.Error("Review step did not execute after implement step")
	}

	t.Logf("Minimal workflow test passed")
}

// TestConfigDrivenWorkflowStandard validates workflow with 3 steps (implement → commit → review).
// This matches the old hardcoded 3-step behavior for backward compatibility verification.
func TestConfigDrivenWorkflowStandard(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := setupWorkflowTest(t, root, "workflow_standard.yaml")
	createWorkflowCommandFiles(t, tmpHome, []string{"implement", "commit", "review"})
	taskPath := createTaskFile(t, tmpHome, "test task")

	output := runFluxidWorkflow(t, root, tmpHome, taskPath)

	// Verify all 3 steps executed
	expectedSteps := []string{"implement", "commit", "review"}
	for _, step := range expectedSteps {
		if !strings.Contains(output, step) {
			t.Errorf("Step %q not found in output", step)
		}
	}

	// Verify execution order: implement → commit → review
	implementIdx := strings.Index(output, "implement")
	commitIdx := strings.Index(output, "commit")
	reviewIdx := strings.LastIndex(output, "review")

	if implementIdx >= commitIdx || commitIdx >= reviewIdx {
		t.Error("Steps did not execute in correct order (implement → commit → review)")
	}

	t.Logf("Standard workflow test passed")
}

// TestConfigDrivenWorkflowExtended validates workflow with 5 steps.
// Verifies: design → implement → test → review → exit.
func TestConfigDrivenWorkflowExtended(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := setupWorkflowTest(t, root, "workflow_extended.yaml")
	createWorkflowCommandFiles(t, tmpHome, []string{"implement", "implement-e2e", "review"})
	taskPath := createTaskFile(t, tmpHome, "test task")

	output := runFluxidWorkflow(t, root, tmpHome, taskPath)

	// Verify all steps executed
	expectedSteps := []string{"design", "implement", "test", "review"}
	for _, step := range expectedSteps {
		if !strings.Contains(output, step) {
			t.Errorf("Step %q not found in output", step)
		}
	}

	// Verify execution order
	designIdx := strings.Index(output, "design")
	implementIdx := strings.Index(output, "implement")
	testIdx := strings.Index(output, "test")
	reviewIdx := strings.LastIndex(output, "review")

	if designIdx >= implementIdx || implementIdx >= testIdx || testIdx >= reviewIdx {
		t.Error("Steps did not execute in correct order (design → implement → test → review)")
	}

	t.Logf("Extended workflow test passed")
}

// TestConfigDrivenWorkflowStepRetries validates step-specific retry configuration.
// Verifies that different steps can have different retry limits configured.
func TestConfigDrivenWorkflowStepRetries(t *testing.T) {
	t.Skip("Retry behavior test requires mock agent - skipping for now")

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createRetryTestStub(t, root)

	tmpHome := setupWorkflowTest(t, root, "workflow_standard.yaml")
	createWorkflowCommandFiles(t, tmpHome, []string{"implement", "commit", "review"})
	taskPath := createTaskFile(t, tmpHome, "test task")

	output := runFluxidWorkflow(t, root, tmpHome, taskPath)

	// Verify retry attempts occurred for implement step
	verifyRetryAttempts(t, output)
	verifyRetryFailuresAndSuccess(t, output)

	t.Logf("Step retry test passed")
}

// verifyRetryAttempts checks that all 3 retry attempts occurred.
func verifyRetryAttempts(t *testing.T, output string) {
	t.Helper()

	for i := 1; i <= 3; i++ {
		attemptMsg := "IMPLEMENT Attempt " + string(rune('0'+i))
		if !strings.Contains(output, attemptMsg) {
			t.Errorf("IMPLEMENT Attempt %d not found - retries not working\nOutput:\n%s", i, output)
		}
	}
}

// verifyRetryFailuresAndSuccess checks FAIL→FAIL→PASS pattern.
func verifyRetryFailuresAndSuccess(t *testing.T, output string) {
	t.Helper()

	expectedPatterns := []string{
		"IMPLEMENT Attempt 1: FAIL",
		"IMPLEMENT Attempt 2: FAIL",
		"IMPLEMENT Attempt 3: PASS",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Pattern %q not found\nOutput:\n%s", pattern, output)
		}
	}
}
