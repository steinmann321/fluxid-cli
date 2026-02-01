package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateWorkflowConfig_NilConfig(t *testing.T) {
	t.Parallel()

	err := ValidateWorkflowConfig(nil, "")
	if err == nil {
		t.Fatal("Expected error for nil config, got nil")
	}
	if err.Error() != "workflow section is required in config.yaml" {
		t.Errorf("Expected 'workflow section is required' error, got: %v", err)
	}
}

func TestValidateWorkflowConfig_MissingReviewCommand(t *testing.T) {
	t.Parallel()

	cfg := &WorkflowConfig{
		Steps: []WorkflowStepConfig{
			{Name: "step1", Command: "cmd.md", Retries: 1},
		},
		Review: ReviewStepConfig{Command: "", Retries: 1},
	}

	err := ValidateWorkflowConfig(cfg, "")
	if err == nil {
		t.Fatal("Expected error for missing review command, got nil")
	}
}

func TestValidateWorkflowConfig_NoCustomSteps(t *testing.T) {
	t.Parallel()

	cfg := &WorkflowConfig{
		Steps:  []WorkflowStepConfig{},
		Review: ReviewStepConfig{Command: "review.md", Retries: 1},
	}

	err := ValidateWorkflowConfig(cfg, "")
	if err == nil {
		t.Fatal("Expected error for no custom steps, got nil")
	}
	if err.Error() != "at least one custom workflow step is required before review" {
		t.Errorf("Expected 'at least one custom step' error, got: %v", err)
	}
}

func TestValidateWorkflowConfig_EmptyStepName(t *testing.T) {
	t.Parallel()

	cfg := &WorkflowConfig{
		Steps: []WorkflowStepConfig{
			{Name: "", Command: "cmd.md", Retries: 1},
		},
		Review: ReviewStepConfig{Command: "review.md", Retries: 1},
	}

	err := ValidateWorkflowConfig(cfg, "")
	if err == nil {
		t.Fatal("Expected error for empty step name, got nil")
	}
	if err.Error() != "step name cannot be empty or whitespace-only" {
		t.Errorf("Expected 'step name cannot be empty' error, got: %v", err)
	}
}

func TestValidateWorkflowConfig_WhitespaceOnlyStepName(t *testing.T) {
	t.Parallel()

	cfg := &WorkflowConfig{
		Steps: []WorkflowStepConfig{
			{Name: "   ", Command: "cmd.md", Retries: 1},
		},
		Review: ReviewStepConfig{Command: "review.md", Retries: 1},
	}

	err := ValidateWorkflowConfig(cfg, "")
	if err == nil {
		t.Fatal("Expected error for whitespace-only step name, got nil")
	}
}

func TestValidateWorkflowConfig_DuplicateStepNames(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdFile1 := filepath.Join(tmpDir, "cmd1.md")
	cmdFile2 := filepath.Join(tmpDir, "cmd2.md")
	reviewFile := filepath.Join(tmpDir, "review.md")
	if err := os.WriteFile(cmdFile1, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cmdFile2, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reviewFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &WorkflowConfig{
		Steps: []WorkflowStepConfig{
			{Name: "implement", Command: cmdFile1, Retries: 1},
			{Name: "implement", Command: cmdFile2, Retries: 1},
		},
		Review: ReviewStepConfig{Command: reviewFile, Retries: 1},
	}

	err := ValidateWorkflowConfig(cfg, tmpDir)
	if err == nil {
		t.Fatal("Expected error for duplicate step names, got nil")
	}
	if err.Error() != "duplicate step name: implement" {
		t.Errorf("Expected 'duplicate step name' error, got: %v", err)
	}
}

func TestValidateWorkflowConfig_NegativeStepRetries(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdFile := filepath.Join(tmpDir, "cmd.md")
	if err := os.WriteFile(cmdFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &WorkflowConfig{
		Steps: []WorkflowStepConfig{
			{Name: "step1", Command: cmdFile, Retries: -1},
		},
		Review: ReviewStepConfig{Command: cmdFile, Retries: 1},
	}

	err := ValidateWorkflowConfig(cfg, tmpDir)
	if err == nil {
		t.Fatal("Expected error for negative retries, got nil")
	}
	if err.Error() != "step step1: retries cannot be negative" {
		t.Errorf("Expected 'retries cannot be negative' error, got: %v", err)
	}
}

func TestValidateWorkflowConfig_NegativeReviewRetries(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdFile := filepath.Join(tmpDir, "cmd.md")
	reviewFile := filepath.Join(tmpDir, "review.md")
	if err := os.WriteFile(cmdFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reviewFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &WorkflowConfig{
		Steps: []WorkflowStepConfig{
			{Name: "step1", Command: cmdFile, Retries: 1},
		},
		Review: ReviewStepConfig{Command: reviewFile, Retries: -1},
	}

	err := ValidateWorkflowConfig(cfg, tmpDir)
	if err == nil {
		t.Fatal("Expected error for negative review retries, got nil")
	}
}

func TestValidateWorkflowConfig_ValidConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdFile := filepath.Join(tmpDir, "cmd.md")
	reviewFile := filepath.Join(tmpDir, "review.md")
	if err := os.WriteFile(cmdFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reviewFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &WorkflowConfig{
		Steps: []WorkflowStepConfig{
			{Name: "step1", Command: cmdFile, Retries: 3},
			{Name: "step2", Command: cmdFile, Retries: 5},
		},
		Review: ReviewStepConfig{Command: reviewFile, Retries: 1},
	}

	err := ValidateWorkflowConfig(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

func TestValidateCommandPath_FileNotFound(t *testing.T) {
	t.Parallel()

	err := ValidateCommandPath("/nonexistent/file.md", "/tmp")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestValidateCommandPath_Directory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	err := ValidateCommandPath(tmpDir, "")
	if err == nil {
		t.Fatal("Expected error for directory path, got nil")
	}
}

func TestValidateCommandPath_ValidFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdFile := filepath.Join(tmpDir, "cmd.md")
	if err := os.WriteFile(cmdFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := ValidateCommandPath(cmdFile, tmpDir)
	if err != nil {
		t.Errorf("Expected no error for valid file, got: %v", err)
	}
}

func TestValidateCommandPath_RelativePath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cmdFile := filepath.Join(tmpDir, "cmd.md")
	if err := os.WriteFile(cmdFile, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := ValidateCommandPath("cmd.md", tmpDir)
	if err != nil {
		t.Errorf("Expected no error for relative path, got: %v", err)
	}
}
