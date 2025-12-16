//nolint:paralleltest // Workflow tests with subprocess execution
package main

import (
	"fluxid-loop/internal/ipc"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunImplementPhase_WithAbort(t *testing.T) {
	// Test that implement phase checks abort flag
	sessionID := "test-implement-abort-session"
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	t.Setenv("XDG_DATA_HOME", tmpDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Set abort flag
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "echo",
		MaxReviewCycles:     1,
		MaxImplementRetries: 3,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error for aborted implement phase")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

func TestRunImplementPhase_MaxRetries(t *testing.T) {
	// Test that implement phase fails after max retries
	cfg := Config{
		SessionID:           "test-retries-session",
		Agent:               "nonexistent-agent-xyz",
		MaxReviewCycles:     1,
		MaxImplementRetries: 2,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error after max retries")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "failed after") && !strings.Contains(errMsg, "retries") &&
		!strings.Contains(errMsg, "failed with exit code") {
		t.Errorf("Expected retry-related error message, got: %v", err)
	}
}

func TestRunImplementPhase_NonZeroExitCode(t *testing.T) {
	// Test that implement phase aborts on non-zero exit code
	cfg := Config{
		SessionID:           "test-nonzero-exit",
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 3,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error for non-zero exit code")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	if !strings.Contains(err.Error(), "failed with exit code") {
		t.Errorf("Expected exit code error message, got: %v", err)
	}
}

func TestRunImplementPhase_WithCommit(t *testing.T) {
	// Test implement phase with commit enabled
	sessionID := "test-implement-with-commit"
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runImplementPhase(cfg)
	// Should fail at implement phase
	if err == nil {
		t.Error("Expected error")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestRunImplementPhase_SuccessWithCommit(t *testing.T) {
	// Test successful implement phase with commit
	sessionID := "test-implement-success-commit"
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	t.Setenv("XDG_DATA_HOME", tmpDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	// Write report in background to simulate agent behavior
	// Use channels for reliable coordination instead of arbitrary sleeps
	done := make(chan struct{})
	started := make(chan struct{})
	go func() {
		defer close(done)
		close(started) // Signal goroutine is ready
		report := `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Implementation successful
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Continue
`
		_ = ipc.WriteReport(sessionID, report)
	}()
	<-started // Wait for goroutine to start

	exitCode, err := runImplementPhase(cfg)
	<-done // Wait for goroutine to complete
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunImplementPhase_FailRetryThenPass(t *testing.T) {
	// Test implement phase with FAIL then PASS
	sessionID := "test-implement-retry-pass"
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	t.Setenv("XDG_DATA_HOME", tmpDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 3,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	go writeReportWithRetry(sessionID, 3, 2)

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunCommitPhase_Failure(t *testing.T) {
	// Test commit phase failure
	sessionID := "test-commit-failure"
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "false", // Will fail
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runCommitPhase(cfg)
	if err == nil {
		t.Error("Expected error for failed commit phase")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestRunReviewPhase_Failure(t *testing.T) {
	// Test review phase failure
	sessionID := "test-review-failure"
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "false", // Will fail
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	_, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Error("Expected error for failed review phase")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestRunReviewPhase_Success(t *testing.T) {
	// Test successful review phase
	sessionID := "test-review-success"
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	t.Setenv("XDG_DATA_HOME", tmpDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := Config{
		SessionID:           sessionID,
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

	// Use channel-based coordination instead of arbitrary sleep
	started := make(chan struct{})
	go func() {
		close(started) // Signal goroutine is ready
		report := `command: test-review
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Review passed
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Complete
`
		_ = ipc.WriteReport(sessionID, report)
	}()
	<-started // Wait for goroutine to start

	status, exitCode, err := runReviewPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if status != "PASS" {
		t.Errorf("Expected status PASS, got: %s", status)
	}
}
