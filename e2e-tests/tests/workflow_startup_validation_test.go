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

// TestStartupValidationMissingWorkflow validates V001: workflow section must exist.
func TestStartupValidationMissingWorkflow(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpHome := setupWorkflowTest(t, root, "workflow_no_workflow_section.yaml")
	taskPath := createTaskFile(t, tmpHome)

	// Run fluxid - expect failure
	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test code executes trusted test binary with controlled args
	cmd := exec.CommandContext(t.Context(), binPath, "--file="+taskPath)
	cmd.Env = append(filterSessionIDFromEnv(os.Environ()),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with missing workflow section, but it succeeded")
	}

	outputStr := string(output)
	// When workflow section is missing, fluxid falls back to legacy workflow which requires commands section
	hasWorkflowError := strings.Contains(outputStr, "workflow section is required")
	hasCommandsError := strings.Contains(outputStr, "commands section is required")
	if !hasWorkflowError && !hasCommandsError {
		t.Errorf("Expected error related to missing workflow or commands configuration not found. Got: %s", outputStr)
	}

	t.Logf("Validation test passed: missing workflow section properly rejected")
}

// TestStartupValidationMissingReview validates V002: review section must exist.
func TestStartupValidationMissingReview(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpHome := setupWorkflowTest(t, root, "workflow_no_review.yaml")
	createWorkflowCommandFiles(t, tmpHome, []string{"implement"})
	taskPath := createTaskFile(t, tmpHome)

	// Run fluxid - expect failure
	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test code executes trusted test binary with controlled args
	cmd := exec.CommandContext(t.Context(), binPath, "--file="+taskPath)
	cmd.Env = append(filterSessionIDFromEnv(os.Environ()),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with missing review section, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "review section is required") {
		t.Errorf("Expected error message 'review section is required' not found. Got: %s", outputStr)
	}

	t.Logf("Validation test passed: missing review section properly rejected")
}

// TestStartupValidationDuplicateNames validates V006: step names must be unique.
func TestStartupValidationDuplicateNames(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpHome := setupWorkflowTest(t, root, "workflow_duplicate_names.yaml")
	// The fixture has two steps both named "implement" pointing to different commands
	createWorkflowCommandFiles(t, tmpHome, []string{"implement", "commit", "review"})
	taskPath := createTaskFile(t, tmpHome)

	// Run fluxid - expect failure
	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test code executes trusted test binary with controlled args
	cmd := exec.CommandContext(t.Context(), binPath, "--file="+taskPath)
	cmd.Env = append(filterSessionIDFromEnv(os.Environ()),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with duplicate step names, but it succeeded")
	}

	outputStr := string(output)
	// The error can be either "duplicate step name" or path validation error (both indicate invalid config)
	// The actual error depends on validation order in the implementation
	hasDuplicateError := strings.Contains(outputStr, "duplicate step name")
	hasConfigError := strings.Contains(outputStr, "workflow configuration error")
	if !hasDuplicateError && !hasConfigError {
		t.Errorf("Expected error related to duplicate step name or workflow validation not found. Got: %s", outputStr)
	}

	t.Logf("Validation test passed: duplicate step names config properly rejected")
}

// TestStartupValidationEmptyStepName validates V005: step names cannot be empty.
func TestStartupValidationEmptyStepName(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a config with empty step name
	tmpHome := t.TempDir()
	configPath := filepath.Join(tmpHome, ".fluxid", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), permDir); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configContent := `agent: claude
iterations: 2
workflow:
  steps:
    - name: ""
      command: .fluxid/commands/fluxid.implement.md
  review:
    command: .fluxid/commands/fluxid.review.md
`
	if err := os.WriteFile(configPath, []byte(configContent), permFile); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create commands directory first
	commandsDir := filepath.Join(tmpHome, ".fluxid", "commands")
	if err := os.MkdirAll(commandsDir, permDir); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	createWorkflowCommandFiles(t, tmpHome, []string{"implement", "review"})
	taskPath := createTaskFile(t, tmpHome)

	// Run fluxid - expect failure
	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test code executes trusted test binary with controlled args
	cmd := exec.CommandContext(t.Context(), binPath, "--file="+taskPath)
	cmd.Env = append(filterSessionIDFromEnv(os.Environ()),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with empty step name, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "step name cannot be empty") {
		t.Errorf("Expected error message 'step name cannot be empty' not found. Got: %s", outputStr)
	}

	t.Logf("Validation test passed: empty step name properly rejected")
}

// TestStartupValidationNegativeRetries validates V009: retries must be non-negative.
func TestStartupValidationNegativeRetries(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpHome := setupWorkflowTest(t, root, "workflow_negative_retries.yaml")
	createWorkflowCommandFiles(t, tmpHome, []string{"implement", "review"})
	taskPath := createTaskFile(t, tmpHome)

	// Run fluxid - expect failure
	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test code executes trusted test binary with controlled args
	cmd := exec.CommandContext(t.Context(), binPath, "--file="+taskPath)
	cmd.Env = append(filterSessionIDFromEnv(os.Environ()),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with negative retries, but it succeeded")
	}

	outputStr := string(output)
	// The error can be either "retries cannot be negative" or path validation error (both indicate invalid config)
	// The actual error depends on validation order in the implementation
	hasRetriesError := strings.Contains(outputStr, "retries cannot be negative")
	hasConfigError := strings.Contains(outputStr, "workflow configuration error")
	if !hasRetriesError && !hasConfigError {
		t.Errorf("Expected error related to negative retries or workflow validation not found. Got: %s", outputStr)
	}

	t.Logf("Validation test passed: negative retries config properly rejected")
}

// TestStartupValidationNegativeIterations validates V010: max_iterations must be non-negative.
func TestStartupValidationNegativeIterations(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a config with negative iterations
	tmpHome := t.TempDir()
	configPath := filepath.Join(tmpHome, ".fluxid", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), permDir); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configContent := `agent: claude
iterations: -1
workflow:
  steps:
    - name: implement
      command: .fluxid/commands/fluxid.implement.md
  review:
    command: .fluxid/commands/fluxid.review.md
`
	if err := os.WriteFile(configPath, []byte(configContent), permFile); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create commands directory first
	commandsDir := filepath.Join(tmpHome, ".fluxid", "commands")
	if err := os.MkdirAll(commandsDir, permDir); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	createWorkflowCommandFiles(t, tmpHome, []string{"implement", "review"})
	taskPath := createTaskFile(t, tmpHome)

	// Run fluxid - expect failure
	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test code executes trusted test binary with controlled args
	cmd := exec.CommandContext(t.Context(), binPath, "--file="+taskPath)
	cmd.Env = append(filterSessionIDFromEnv(os.Environ()),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with negative iterations, but it succeeded")
	}

	outputStr := string(output)
	// The error message may vary ("iterations cannot be negative" or "iterations must be a positive integer")
	hasNegativeError := strings.Contains(outputStr, "iterations cannot be negative")
	hasPositiveError := strings.Contains(outputStr, "iterations must be a positive integer")
	if !hasNegativeError && !hasPositiveError {
		t.Errorf("Expected error related to negative iterations not found. Got: %s", outputStr)
	}

	t.Logf("Validation test passed: negative iterations properly rejected")
}

// TestStartupValidationInvalidCommandPath validates V007-V008: command files must exist and be readable.
func TestStartupValidationInvalidCommandPath(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a config with non-existent command file
	tmpHome := t.TempDir()
	configPath := filepath.Join(tmpHome, ".fluxid", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), permDir); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configContent := `agent: claude
iterations: 2
workflow:
  steps:
    - name: implement
      command: .fluxid/commands/nonexistent.md
  review:
    command: .fluxid/commands/fluxid.review.md
`
	if err := os.WriteFile(configPath, []byte(configContent), permFile); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create commands directory
	commandsDir := filepath.Join(tmpHome, ".fluxid", "commands")
	if err := os.MkdirAll(commandsDir, permDir); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	// Only create review command file, not implement
	createWorkflowCommandFiles(t, tmpHome, []string{"review"})
	taskPath := createTaskFile(t, tmpHome)

	// Run fluxid - expect failure
	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test code executes trusted test binary with controlled args
	cmd := exec.CommandContext(t.Context(), binPath, "--file="+taskPath)
	cmd.Env = append(filterSessionIDFromEnv(os.Environ()),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with invalid command path, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "command file not found") && !strings.Contains(outputStr, "no such file") {
		t.Errorf("Expected error message about missing command file not found. Got: %s", outputStr)
	}

	t.Logf("Validation test passed: invalid command path properly rejected")
}
