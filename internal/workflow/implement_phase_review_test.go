//nolint:paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-cli/internal/ipc"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestRunReviewPhase_Failure(t *testing.T) {
	// Test review phase failure
	sessionID := "test-review-failure-" + time.Now().Format("20060102150405.000000")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false", // Will fail
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
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
	defer goleak.VerifyNone(t)

	// Test successful review phase
	sessionID := "test-review-success-" + time.Now().Format("20060102150405.000000")
	tmpDir, cleanup := setupTestDataDir(t)
	defer cleanup()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
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
