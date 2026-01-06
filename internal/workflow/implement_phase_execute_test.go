//nolint:paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteImplementPhase_AgentExitCodeError(t *testing.T) {
	// Test that executeImplementPhase handles non-zero exit codes properly
	sessionID := "f8a9b0c1-2d3e-4f5a-6b7c-8d9e0f1a2b3c"
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false", // Always exits with code 1
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
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
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	// Test abort error handling in checkImplementReportStatus
	sessionID := "a9b0c1d2-3e4f-5a6b-7c8d-9e0f1a2b3c4d"
	tmpDir, cleanup := setupTestDataDir(t)
	defer cleanup()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Set abort flag before checking report
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

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
	sessionID := "b0c1d2e3-4f5a-6b7c-8d9e-0f1a2b3c4d5e"
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
	if err := storage.WriteReport(sessionID, failReport); err != nil {
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
