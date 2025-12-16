package main

import (
	"fluxid-loop/internal/ipc"
	"testing"
	"time"
)

//nolint:paralleltest // Simple workflow test
func TestExecuteWorkflow_DryRun(t *testing.T) {
	cfg := Config{
		SessionID:           "test-dry-run",
		Agent:               testAgentEcho,
		MaxReviewCycles:     2,
		MaxImplementRetries: 2,
		CommitEnabled:       true,
		DryRun:              true,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected zero exit code for dry-run, got %d", exitCode)
	}
}

func TestExecuteWorkflow_NoDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           "test-no-dry-run",
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode := executeWorkflow(cfg)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for failing workflow")
	}
}

func TestExecuteWorkflow_OutputFormats(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Test JSON
	cfg := Config{
		SessionID:           "test-json",
		Agent:               testAgentEcho,
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              true,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatJSON,
		Sources:             map[string]string{},
	}
	if exitCode := executeWorkflow(cfg); exitCode != 0 {
		t.Errorf("JSON output failed with exit code %d", exitCode)
	}

	// Test YAML
	cfg.OutputFormat = OutputFormatYAML
	cfg.SessionID = "test-yaml"
	if exitCode := executeWorkflow(cfg); exitCode != 0 {
		t.Errorf("YAML output failed with exit code %d", exitCode)
	}

	// Test default (text)
	cfg.OutputFormat = "unknown"
	cfg.SessionID = "test-default"
	if exitCode := executeWorkflow(cfg); exitCode != 0 {
		t.Errorf("Default output failed with exit code %d", exitCode)
	}
}

func TestExecuteWorkflow_SignalHandlerSetup(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           "test-signal",
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	// Use channel-based coordination instead of arbitrary sleep delays
	started := make(chan struct{})
	go func() {
		close(started) // Signal goroutine is ready
		implementReport := `command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: ` + statusPass + `
summary: Success
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
		_ = ipc.WriteReport(cfg.SessionID, implementReport)

		// Delay between reports to allow workflow to process each phase
		time.Sleep(100 * time.Millisecond)
		reviewReport := `command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: ` + statusPass + `
summary: Success
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
		_ = ipc.WriteReport(cfg.SessionID, reviewReport)
	}()
	<-started // Wait for goroutine to start

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected zero exit code, got %d", exitCode)
	}
}

func TestExecuteWorkflow_WorkflowError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           "test-error",
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode := executeWorkflow(cfg)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for workflow error")
	}
}

//nolint:paralleltest // Test doesn't use t
func TestPrintInitializationStatusYAML_Error(_ *testing.T) {
	cfg := Config{
		SessionID:           "test-yaml-status",
		Agent:               testAgentEcho,
		MaxReviewCycles:     2,
		MaxImplementRetries: 2,
		CommitEnabled:       true,
		OutputFormat:        OutputFormatYAML,
		Sources:             map[string]string{"agent": "default"},
		CommandFiles:        nil,
		AgentArgs:           []string{},
		DryRun:              false,
	}

	// This should work without error
	_ = PrintInitializationStatusYAML(cfg)
}

//nolint:paralleltest // Test doesn't use t
func TestPrintInitializationStatusJSON_AllFields(_ *testing.T) {
	cfg := Config{
		SessionID:           "test-json-full",
		Agent:               testAgentClaude,
		MaxReviewCycles:     3,
		MaxImplementRetries: 2,
		CommitEnabled:       true,
		OutputFormat:        OutputFormatJSON,
		Sources: map[string]string{
			"agent":             "config",
			"review_cycles":     "env",
			"implement_retries": "cli",
			"commit_enabled":    "default",
		},
		CommandFiles: nil,
		AgentArgs:    []string{},
		DryRun:       false,
	}

	// This should work without error
	_ = PrintInitializationStatusJSON(cfg)
}

//nolint:paralleltest // Simple workflow test
func TestExecuteWorkflow_SignalHandlerSkippedInDryRun(t *testing.T) {
	cfg := Config{
		SessionID:           "test-dry-run-no-signal",
		Agent:               "echo",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              true,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode := executeWorkflow(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run, got %d", exitCode)
	}
}

//nolint:paralleltest // Test doesn't use t
func TestPrintInitializationStatusText_AllSources(_ *testing.T) {
	cfg := Config{
		SessionID:           "test-text-sources",
		Agent:               testAgentClaude,
		MaxReviewCycles:     3,
		MaxImplementRetries: 2,
		CommitEnabled:       true,
		OutputFormat:        OutputFormatText,
		Sources: map[string]string{
			"agent":             "cli",
			"review_cycles":     "env",
			"implement_retries": "config",
			"commit_enabled":    "default",
		},
		CommandFiles: nil,
		AgentArgs:    []string{},
		DryRun:       false,
	}

	// This should work without error
	PrintInitializationStatusText(cfg)
}
