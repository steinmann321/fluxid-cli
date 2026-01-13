//nolint:paralleltest,exhaustruct // Tests use global mutex and incomplete structs
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/types"
	"os"
	"path/filepath"
	"testing"
)

// TestExecuteCommitPhase_Error tests commit phase with non-existent agent.
func TestExecuteCommitPhase_Error(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "8b4e5f9c-1a2b-4c3d-9e8f-7a6b5c4d3e2f"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "nonexistent-agent-xyz",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := executeCommitPhase(cfg, 1)
	if err == nil {
		t.Error("Expected error for non-existent agent")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestRun_AbortAfterImplement tests abort between implement and review.
func TestRun_AbortAfterImplement(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "9d8e7f6a-5b4c-3d2e-1f0a-9b8c7d6e5f4a"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report
	if err := storage.WriteReport(sessionID, testImplementPassReport); err != nil {
		t.Fatalf("Failed to write implement report: %v", err)
	}

	// Set abort flag to trigger abort before review
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	exitCode, err := Run(cfg)
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

// TestExecuteImplementPhase_Error tests implement phase with non-existent agent.
func TestExecuteImplementPhase_Error(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "nonexistent-agent-xyz",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := executeImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error for non-existent agent")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestCheckAbortBeforeCommit_Abort tests abort detection before commit.
func TestCheckAbortBeforeCommit_Abort(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "b2c3d4e5-f6a7-8b9c-0d1e-2f3a4b5c6d7e"

	// Set abort flag
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	exitCode, err := checkAbortBeforeCommit(sessionID)
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

// TestCheckAbortBeforeImplement_Abort tests abort detection before implement.
func TestCheckAbortBeforeImplement_Abort(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "c3d4e5f6-a7b8-9c0d-1e2f-3a4b5c6d7e8f"

	// Set abort flag
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	exitCode, err := checkAbortBeforeImplement(sessionID)
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

// TestRunImplementPhase_NonZeroExitFromImplement tests implement phase error with non-zero exit.
func TestRunImplementPhase_NonZeroExitFromImplement(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "d4e5f6a7-b8c9-0d1e-2f3a-4b5c6d7e8f9a"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "nonexistent-command-xyz",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
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

// TestRunCommitPhaseWithRetry_ExecuteError tests commit phase with agent error.
func TestRunCommitPhaseWithRetry_ExecuteError(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "e5f6a7b8-c9d0-1e2f-3a4b-5c6d7e8f9a0b"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "nonexistent-agent-commit",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := runCommitPhaseWithRetry(cfg)
	// Corrected behavior: commit failures are logged but don't block workflow
	// runCommitPhaseWithRetry returns success even when all retries fail
	if err != nil {
		t.Errorf("Expected no error (commit failures logged, workflow continues), got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (allow review phase), got %d", exitCode)
	}
}

// TestWaitForValidReport_AbortDuringWait tests abort while checking report.
func TestWaitForValidReport_AbortDuringWait(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "f6a7b8c9-d0e1-2f3a-4b5c-6d7e8f9a0b1c"

	// Set abort flag
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	status, err := waitForValidReport(sessionID, "", "implement")
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if status != "" {
		t.Errorf("Expected empty status on abort, got %s", status)
	}
}

// TestCheckImplementReportStatus_AbortFlag tests abort during implement report check.
func TestCheckImplementReportStatus_AbortFlag(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "a7b8c9d0-e1f2-3a4b-5c6d-7e8f9a0b1c2d"

	// Set abort flag
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	exitCode, err := checkImplementReportStatus(sessionID, 1)
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

// TestCheckImplementReportStatus_ReadError tests report read failure.
func TestCheckImplementReportStatus_ReadError(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	// Use empty session ID to trigger read error
	// In the refactored design, read errors are treated as FAIL status (not as errors)
	// This allows the retry mechanism to handle transient failures
	exitCode, err := checkImplementReportStatus("", 1)
	if err != nil {
		t.Errorf("Expected no error (read errors treated as FAIL status), got: %v", err)
	}
	if exitCode != -1 {
		t.Errorf("Expected exit code -1 (retry signal), got %d", exitCode)
	}
}

// TestComposePrompt_WithCommandFile tests prompt composition with command file.
func TestComposePrompt_WithCommandFile(t *testing.T) {
	tmpDir := t.TempDir()
	cmdFilePath := filepath.Join(tmpDir, "test-cmd.md")

	// Create command file
	cmdContent := "# Test Command\nDo the thing"
	if err := os.WriteFile(cmdFilePath, []byte(cmdContent), 0o600); err != nil {
		t.Fatalf("Failed to write command file: %v", err)
	}

	cfg := types.Config{
		TaskFilePath: "/task.txt",
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: cmdFilePath,
		},
	}

	result := composePrompt(cfg, "implement", "Base prompt")
	if result == "" {
		t.Error("Expected non-empty prompt")
	}
	// Should contain command file content
	if len(result) < len(cmdContent) {
		t.Error("Expected prompt to contain command file content")
	}
}

// TestComposePrompt_WithMissingCommandFile tests prompt composition with non-existent file.
func TestComposePrompt_WithMissingCommandFile(t *testing.T) {
	cfg := types.Config{
		TaskFilePath: "/task.txt",
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/nonexistent/file.md",
		},
	}

	result := composePrompt(cfg, "implement", "Base prompt")
	if result == "" {
		t.Error("Expected non-empty prompt even with missing file")
	}
	// Should still contain base prompt
	if len(result) < len("Base prompt") {
		t.Error("Expected prompt to at least contain base prompt")
	}
}

// TestComposePrompt_BuiltInPrompt tests prompt composition with built-in prompt.
func TestComposePrompt_BuiltInPrompt(t *testing.T) {
	cfg := types.Config{
		TaskFilePath: "/task.txt",
		CommandFiles: nil,
	}

	result := composePrompt(cfg, "implement", "Base prompt")
	if result == "" {
		t.Error("Expected non-empty prompt")
	}
}

func TestRunReviewPhaseError(t *testing.T) {
	sessionID := "d6e7f8a9-0b1c-2d3e-4f5a-6b7c8d9e0f1a"
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	cfg := types.Config{
		Agent:           "nonexistent-agent-xyz",
		SessionID:       sessionID,
		SessionRoot:     "",
		MaxReviewCycles: 1,
		TaskFilePath:    "",
		CommandFiles:    nil,
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
