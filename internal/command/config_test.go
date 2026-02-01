//nolint:paralleltest,exhaustruct,godot // Test file with structs and literals
package command

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildFinalConfigWithInvalidOutputFormat(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommandFiles:     &config.ResolvedCommandFiles{},
	}

	invalidFormat := "invalid-format"
	args := &CLIArgs{
		AgentArgs:           []string{},
		CLIOutputFormat:     &invalidFormat,
		CLIDryRun:           boolPtr(false),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
	}

	_, err := buildFinalConfig(resolved, args, nil, nil, "")
	if err == nil {
		t.Error("Expected error for invalid output format, got nil")
	}
}

func TestBuildFinalConfigWithDryRun(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommandFiles:     &config.ResolvedCommandFiles{},
	}

	args := &CLIArgs{
		AgentArgs:           []string{"--model", "test"},
		CLIOutputFormat:     nil,
		CLIDryRun:           boolPtr(true),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
	}

	cfg, err := buildFinalConfig(resolved, args, nil, nil, "")
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if !cfg.DryRun {
		t.Error("Expected DryRun to be true")
	}
}

func TestBuildFinalConfigWithSessionID(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommandFiles:     &config.ResolvedCommandFiles{},
	}

	args := &CLIArgs{
		AgentArgs:           []string{},
		CLIOutputFormat:     nil,
		CLIDryRun:           boolPtr(true),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLITaskFilePath:     strPtr("/abs/task.txt"),
	}

	// Test with FLUXID_SESSION_ID set
	t.Setenv("FLUXID_SESSION_ID", "test-session-123")

	cfg, err := buildFinalConfig(resolved, args, nil, nil, "")
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if cfg.SessionID != "test-session-123" {
		t.Errorf("Expected SessionID to be test-session-123, got %s", cfg.SessionID)
	}
}

func TestBuildFinalConfigWithJSONFormat(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommandFiles:     &config.ResolvedCommandFiles{},
	}

	jsonFormat := testFormatJSON
	args := &CLIArgs{
		AgentArgs:           []string{},
		CLIOutputFormat:     &jsonFormat,
		CLIDryRun:           boolPtr(true),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLITaskFilePath:     strPtr("/abs/task.txt"),
	}

	cfg, err := buildFinalConfig(resolved, args, nil, nil, "")
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if cfg.OutputFormat != output.FormatJSON {
		t.Errorf("Expected OutputFormat to be JSON, got %v", cfg.OutputFormat)
	}
}

func TestBuildFinalConfigWithYAMLFormat(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommandFiles:     &config.ResolvedCommandFiles{},
	}

	yamlFormat := "yaml"
	args := &CLIArgs{
		AgentArgs:           []string{},
		CLIOutputFormat:     &yamlFormat,
		CLIDryRun:           boolPtr(true),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLITaskFilePath:     strPtr("/abs/task.txt"),
	}

	cfg, err := buildFinalConfig(resolved, args, nil, nil, "")
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if cfg.OutputFormat != output.FormatYAML {
		t.Errorf("Expected OutputFormat to be YAML, got %v", cfg.OutputFormat)
	}
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string { return &s }

func TestBuildWorkflowIfConfigured_NoConfig(t *testing.T) {
	workflow, err := buildWorkflowIfConfigured(nil, nil, "", 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if workflow != nil {
		t.Error("Expected nil workflow when no config is provided")
	}
}

func TestBuildWorkflowIfConfigured_ProjectConfig(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "task.txt")
	if err := os.WriteFile(taskFile, []byte("test task"), 0o600); err != nil {
		t.Fatalf("Failed to create task file: %v", err)
	}

	projectConfig := &config.ProjectConfig{
		Workflow: &config.WorkflowConfig{
			Steps: []config.WorkflowStepConfig{
				{
					Name:    "test-step",
					Command: taskFile,
					Retries: 1,
				},
			},
			Review: config.ReviewStepConfig{
				Command: taskFile,
				Retries: 1,
			},
		},
	}

	workflow, err := buildWorkflowIfConfigured(projectConfig, nil, tmpDir, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if workflow == nil {
		t.Error("Expected workflow to be created")
	}
}

func TestBuildWorkflowIfConfigured_HomeConfig(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "task.txt")
	if err := os.WriteFile(taskFile, []byte("test task"), 0o600); err != nil {
		t.Fatalf("Failed to create task file: %v", err)
	}

	homeConfig := &config.HomeConfig{
		Workflow: &config.WorkflowConfig{
			Steps: []config.WorkflowStepConfig{
				{
					Name:    "test-step",
					Command: taskFile,
					Retries: 1,
				},
			},
			Review: config.ReviewStepConfig{
				Command: taskFile,
				Retries: 1,
			},
		},
	}

	workflow, err := buildWorkflowIfConfigured(nil, homeConfig, tmpDir, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if workflow == nil {
		t.Error("Expected workflow to be created")
	}
}

func TestBuildWorkflowIfConfigured_InvalidConfig(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		Workflow: &config.WorkflowConfig{
			Steps: []config.WorkflowStepConfig{
				{
					Name:    "invalid-step",
					Command: "/nonexistent/file.txt",
					Retries: 1,
				},
			},
			Review: config.ReviewStepConfig{
				Command: "/nonexistent/file.txt",
				Retries: 1,
			},
		},
	}

	_, err := buildWorkflowIfConfigured(projectConfig, nil, "", 1)
	if err == nil {
		t.Error("Expected error for invalid workflow config")
	}
}

func TestGetSessionID_FromEnv(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "test-session-id")
	sessionID := getSessionID()
	if sessionID != "test-session-id" {
		t.Errorf("Expected test-session-id, got %s", sessionID)
	}
}

func TestGetSessionID_GenerateNew(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")
	sessionID := getSessionID()
	if sessionID == "" {
		t.Error("Expected non-empty session ID")
	}
}

func TestCustomConfigToProjectConfig_WithNil(t *testing.T) {
	result := customConfigToProjectConfig(nil)
	if result != nil {
		t.Error("Expected nil result for nil input")
	}
}

func TestCustomConfigToProjectConfig_WithValidConfig(t *testing.T) {
	agent := "claude"
	implementRetries := 5
	commitRetries := 3
	iterations := 10

	customCfg := &config.CustomConfig{
		Agent:            &agent,
		AgentArgs:        []string{"--model", "test"},
		ImplementRetries: &implementRetries,
		CommitRetries:    &commitRetries,
		Iterations:       &iterations,
		Commands:         &config.Commands{},
		Workflow:         &config.WorkflowConfig{},
	}

	result := customConfigToProjectConfig(customCfg)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Agent == nil || *result.Agent != "claude" {
		t.Errorf("Expected agent claude, got %v", result.Agent)
	}
	if len(result.AgentArgs) != 2 {
		t.Errorf("Expected 2 agent args, got %d", len(result.AgentArgs))
	}
	if result.ImplementRetries == nil || *result.ImplementRetries != 5 {
		t.Error("Expected ImplementRetries to be 5")
	}
}

func TestLoadAllConfigsWithDir_InvalidDir(t *testing.T) {
	nonexistentPath := "/nonexistent/directory/that/does/not/exist"
	homeConfig, projectConfig, configDir, exitCode := loadAllConfigsWithDir(&nonexistentPath)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for nonexistent directory")
	}
	if homeConfig != nil || projectConfig != nil || configDir != "" {
		t.Error("Expected nil configs and empty configDir for invalid path")
	}
}

func TestLoadConfigsWithCustom_InvalidPath(t *testing.T) {
	_, _, exitCode := loadConfigsWithCustom("/nonexistent/file.yaml")
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid path")
	}
}
