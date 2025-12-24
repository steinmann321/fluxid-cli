package workflow

import (
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestRunImplementPhase_WithCommit(t *testing.T) {
	// Test implement phase with commit enabled
	sessionID := "test-implement-with-commit-" + time.Now().Format("20060102150405.000000")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
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
	defer goleak.VerifyNone(t)

	// Test successful implement phase with commit
	sessionID := "test-implement-success-commit-" + time.Now().Format("20060102150405.000000")
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	t.Setenv("XDG_DATA_HOME", tmpDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
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

func TestRunCommitPhase_Failure(t *testing.T) {
	// Test commit phase failure
	sessionID := "test-commit-failure-" + time.Now().Format("20060102150405.000000")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false", // Will fail
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
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

func TestExecuteCommitIfEnabled_Disabled(t *testing.T) {
	// Test that commit is skipped when disabled
	sessionID := "test-commit-disabled-" + time.Now().Format("20060102150405.000000")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false, // Disabled
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := executeCommitIfEnabled(cfg)
	if err != nil {
		t.Errorf("Expected no error when commit disabled, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
