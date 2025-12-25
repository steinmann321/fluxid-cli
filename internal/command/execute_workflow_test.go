//nolint:exhaustruct // Test file with extensive struct setup
package command

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"testing"
)

func TestExecuteWorkflowWithJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           "test-json-session",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              true,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatJSON,
	}

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for JSON dry-run, got %d", exitCode)
	}
}

func TestExecuteWorkflowWithYAMLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           "test-yaml-session",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              true,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatYAML,
	}

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for YAML dry-run, got %d", exitCode)
	}
}

func TestExecuteWorkflowWithTextFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           "test-text-session",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              true,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for text dry-run, got %d", exitCode)
	}
}

func TestExecuteWorkflowDryRunMode(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           "test-dry-run",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     2,
		MaxImplementRetries: 2,
		DryRun:              true,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
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

	cfg := types.Config{
		SessionID:           "test-invalid-agent",
		Agent:               "nonexistent-agent-xyz",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// This will fail when trying to run the workflow
	exitCode := executeWorkflow(cfg)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid agent")
	}
}
