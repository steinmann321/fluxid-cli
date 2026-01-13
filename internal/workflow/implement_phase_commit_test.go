//nolint:paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/types"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/goleak"
)

func TestRunImplementPhase_WithCommit(t *testing.T) {
	// Test implement phase with commit enabled
	sessionID := "e1f2a3b4-5c6d-7e8f-9a0b-1c2d3e4f5a6b"
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	exitCode, err := runImplementPhase(cfg)
	// Corrected behavior: commit failures don't block workflow
	if err != nil {
		t.Errorf("Expected no error (commit failures logged, workflow continues), got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (allow review phase), got: %d", exitCode)
	}
}

func TestRunImplementPhase_SuccessWithCommit(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Test successful implement phase with commit
	sessionID := "f2a3b4c5-6d7e-8f9a-0b1c-2d3e4f5a6b7c"
	tmpDir, cleanup := setupTestDataDir(t)
	defer cleanup()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
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
		_ = storage.WriteReport(sessionID, report)
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
	sessionID := "a3b4c5d6-7e8f-9a0b-1c2d-3e4f5a6b7c8d"
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "false", // Will fail
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	exitCode, err := runCommitPhase(cfg)
	if err == nil {
		t.Error("Expected error for failed commit phase")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestExecuteCommitIfEnabled_Disabled removed - commits are always enabled in v2.0
