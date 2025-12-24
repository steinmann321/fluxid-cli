package workflow

import (
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunReviewPhase_Failure(t *testing.T) {
	// Test review phase failure
	sessionID := "test-review-failure-" + time.Now().Format("20060102150405.000000")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false", // Will fail
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
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
	sessionID := "test-review-success-" + time.Now().Format("20060102150405.000000")
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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
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
