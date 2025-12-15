//nolint:exhaustruct // Test file with extensive struct setup
package main

import (
	"fluxid-loop/internal/config"
	"testing"
)

func TestExecuteWorkflowWithJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           "test-json-session",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              true,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        OutputFormatJSON,
		Sources:             map[string]string{},
	}

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for JSON dry-run, got %d", exitCode)
	}
}

func TestExecuteWorkflowWithYAMLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           "test-yaml-session",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              true,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        OutputFormatYAML,
		Sources:             map[string]string{},
	}

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for YAML dry-run, got %d", exitCode)
	}
}

func TestExecuteWorkflowWithTextFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           "test-text-session",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              true,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for text dry-run, got %d", exitCode)
	}
}

func TestExecuteWorkflowDryRunMode(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           "test-dry-run",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     2,
		MaxImplementRetries: 2,
		CommitEnabled:       false,
		DryRun:              true,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run mode, got %d", exitCode)
	}
}

func TestExecuteWorkflowWithInvalidAgent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("PATH", "") // Empty PATH to ensure agent not found

	cfg := Config{
		SessionID:           "test-invalid-agent",
		Agent:               "nonexistent-agent-xyz",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	// This will fail when trying to run the workflow
	exitCode := executeWorkflow(cfg)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid agent")
	}
}
