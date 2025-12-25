//nolint:paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecuteImplementPhase_AgentExitCodeError(t *testing.T) {
	// Test that executeImplementPhase handles non-zero exit codes properly
	sessionID := "test-exec-impl-exit-" + time.Now().Format("20060102150405.000000")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false", // Always exits with code 1
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        output.FormatText,
	}

	exitCode, err := executeImplementPhase(cfg, 1)
	if err == nil {
		t.Error("Expected error for non-zero exit code")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	if !strings.Contains(err.Error(), "failed with exit code") {
		t.Errorf("Expected error message to contain 'failed with exit code', got: %v", err)
	}
}

func TestCheckImplementReportStatus_Abort(t *testing.T) {
	// Test abort error handling in checkImplementReportStatus
	sessionID := "test-check-report-abort-" + time.Now().Format("20060102150405.000000")
	tmpDir, cleanup := setupTestDataDir(t)
	defer cleanup()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Set abort flag before checking report
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	exitCode, err := checkImplementReportStatus(sessionID, 1)
	if err == nil {
		t.Error("Expected abort error")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

func TestCheckImplementReportStatus_FailStatus(t *testing.T) {
	// Test that checkImplementReportStatus returns -1 for FAIL status
	sessionID := "test-check-fail-" + time.Now().Format("20060102150405.000000")
	tmpDir, cleanup := setupTestDataDir(t)
	defer cleanup()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Write FAIL report
	failReport := `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: FAIL
summary: Implementation failed
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Fix issues
`
	if err := ipc.WriteReport(sessionID, failReport); err != nil {
		t.Fatalf("Failed to write fail report: %v", err)
	}

	exitCode, err := checkImplementReportStatus(sessionID, 1)
	if err != nil {
		t.Errorf("Expected no error for FAIL status, got: %v", err)
	}
	if exitCode != -1 {
		t.Errorf("Expected exit code -1 for FAIL status, got %d", exitCode)
	}
}
