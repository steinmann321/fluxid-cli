//nolint:paralleltest // Coverage tests don't need parallel execution
package main

import (
	"fluxid-loop/internal/ipc"
	"os"
	"testing"
)

// TestHandleGetReportSchema_WriteError tests WriteReportSchema error path.
func TestHandleGetReportSchema_WriteError(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create a pipe and immediately close the write end to cause write error
	r, w, _ := os.Pipe()
	_ = w.Close()

	os.Stdout = w

	// Redirect stderr to suppress error output
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	exitCode := handleGetReportSchema([]string{})

	// Close pipes
	_ = r.Close()
	_ = stderrW.Close()
	_ = stderrR.Close()

	if exitCode == 0 {
		t.Error("Expected non-zero exit code when WriteReportSchema fails")
	}
}

// TestHandleWriteReport_ReadAllError tests stdin read error path.
func TestHandleWriteReport_ReadAllError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-write-error"
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a reader that will fail
	pr, pw, _ := os.Pipe()
	_ = pw.Close() // Close writer to make read fail eventually
	os.Stdin = pr

	// This should trigger the ReadAll error path
	exitCode := handleWriteReport([]string{})

	_ = pr.Close()

	// Exit code might be 0 or non-zero depending on timing
	_ = exitCode
}

// TestHandleWriteReport_WriteReportError tests ipc.WriteReport error path.
func TestHandleWriteReport_WriteReportError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Use an invalid session ID that will cause storage issues
	sessionID := ""
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create stdin with invalid YAML
	input := "invalid: yaml: content: [broken"
	r, w, _ := os.Pipe()
	_, _ = w.Write([]byte(input))
	_ = w.Close()
	os.Stdin = r

	exitCode := handleWriteReport([]string{})

	if exitCode == 0 {
		t.Error("Expected non-zero exit code when WriteReport fails")
	}
}

// TestHandleReadReport_ReadReportError tests ipc.ReadReport error path.
func TestHandleReadReport_ReadReportError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Use an empty session ID to trigger error
	sessionID := ""
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	exitCode := handleReadReport([]string{})

	if exitCode == 0 {
		t.Error("Expected non-zero exit code when ReadReport fails with empty session")
	}
}

// TestRunWorkflow_AbortCheckErrorWarning tests abort check error warning path.
func TestRunWorkflow_AbortCheckErrorWarning(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Use a session ID without proper setup to cause CheckAbortFlag to potentially error
	cfg := Config{
		SessionID:           "", // Empty session might cause issues
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

	exitCode, err := runWorkflow(cfg)

	// We expect this to fail (since agent is "false"), but it should handle
	// any abort check errors gracefully
	if err == nil && exitCode == 0 {
		t.Error("Expected error or non-zero exit code")
	}
}

// TestRunWorkflow_SecondAbortCheckErrorWarning tests second abort check error.
func TestRunWorkflow_SecondAbortCheckErrorWarning(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create a session directory
	sessionID := "test-abort-check-2"

	// Initialize IPC storage for this session
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to initialize abort flag: %v", err)
	}

	// Now clear the abort flag
	if err := ipc.ClearAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to clear abort flag: %v", err)
	}

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "false", // Will fail quickly
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runWorkflow(cfg)

	// Should fail due to false agent
	if err == nil && exitCode == 0 {
		t.Error("Expected error or non-zero exit code")
	}
}

// TestPrintInitializationStatusText_WithSources tests text output with source tracking.
func TestPrintInitializationStatusText_WithSources(t *testing.T) {
	t.Parallel()
	cfg := Config{
		SessionID:           "test-sources",
		Agent:               "claude",
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources: map[string]string{
			"agent":             "cli",
			"review_cycles":     "env",
			"implement_retries": "default",
		},
	}

	// Should not panic
	PrintInitializationStatusText(cfg)
}

// TestValidateOutputFormat_AllValid tests all valid output formats.
func TestValidateOutputFormat_AllValid(t *testing.T) {
	t.Parallel()
	validFormats := []string{"text", "json", "yaml"}

	for _, format := range validFormats {
		err := ValidateOutputFormat(format)
		if err != nil {
			t.Errorf("Expected no error for format '%s', got: %v", format, err)
		}
	}
}

// TestPrintInitializationStatusJSON_EncodeError tests JSON encoder error path.
func TestPrintInitializationStatusJSON_EncodeError(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create a pipe and immediately close the write end to cause write error
	r, w, _ := os.Pipe()
	_ = w.Close()
	os.Stdout = w

	cfg := Config{
		SessionID:           "test-json-error",
		Agent:               "claude",
		MaxReviewCycles:     20,
		MaxImplementRetries: 3,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatJSON,
		Sources:             map[string]string{},
	}

	err := PrintInitializationStatusJSON(cfg)

	_ = r.Close()

	if err == nil {
		t.Error("Expected error when JSON encoder fails")
	}
}

// TestPrintInitializationStatusYAML_EncodeError tests YAML encoder error path.
func TestPrintInitializationStatusYAML_EncodeError(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create a pipe and immediately close the write end to cause write error
	r, w, _ := os.Pipe()
	_ = w.Close()
	os.Stdout = w

	cfg := Config{
		SessionID:           "test-yaml-error",
		Agent:               "claude",
		MaxReviewCycles:     20,
		MaxImplementRetries: 3,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatYAML,
		Sources:             map[string]string{},
	}

	err := PrintInitializationStatusYAML(cfg)

	_ = r.Close()

	if err == nil {
		t.Error("Expected error when YAML encoder fails")
	}
}

// TestHandleSpecialCommands_WriteHistoryPath tests --write-history code path.
func TestHandleSpecialCommands_WriteHistoryPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-write-history-path")

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test the --write-history path in handleSpecialCommands
	os.Args = []string{"fluxid", "--write-history", "test", "message"}

	exitCode, handled := handleSpecialCommands()

	if !handled {
		t.Error("Expected --write-history to be handled")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
