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

func TestRunReviewPhase_Failure(t *testing.T) {
	// Test review phase failure (missing report treated as FAIL)
	sessionID := "b4c5d6e7-8f9a-0b1c-2d3e-4f5a6b7c8d9e"
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
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

	status, exitCode, err := runReviewPhase(cfg)
	// With new behavior: missing reports are treated as FAIL, not errors
	if err == nil {
		// This is expected - missing report returns FAIL status, not error
		if status != statusFail {
			t.Errorf("Expected FAIL status for missing report, got %q", status)
		}
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for FAIL status, got %d", exitCode)
		}
	} else {
		t.Errorf("Expected no error (FAIL status instead), got error: %v", err)
	}
}

func TestRunReviewPhase_Success(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Test successful review phase
	sessionID := "c5d6e7f8-9a0b-1c2d-3e4f-5a6b7c8d9e0f"
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
		_ = storage.WriteReport(sessionID, report)
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
