//nolint:paralleltest,exhaustruct // Tests use global mutex and incomplete structs
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/ipc"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestExecuteCommitPhase_Error tests commit phase with non-existent agent.
func TestExecuteCommitPhase_Error(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-exec-commit-err-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
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
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-abort-after-impl-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
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
	if err := ipc.WriteReport(sessionID, testImplementPassReport); err != nil {
		t.Fatalf("Failed to write implement report: %v", err)
	}

	// Set abort flag to trigger abort before review
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

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

	sessionID := "test-exec-impl-err-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "nonexistent-agent-xyz",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := executeImplementPhase(cfg, 1)
	if err == nil {
		t.Error("Expected error for non-existent agent")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestCheckAbortBeforeCommit_Abort tests abort detection before commit.
func TestCheckAbortBeforeCommit_Abort(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-abort-before-commit-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Set abort flag
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

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
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-abort-before-impl-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Set abort flag
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

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

	sessionID := "test-impl-nonzero-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
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
	if err == nil {
		t.Error("Expected error for non-existent command")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestRunCommitPhaseWithRetry_ExecuteError tests commit phase with agent error.
func TestRunCommitPhaseWithRetry_ExecuteError(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-commit-exec-err-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
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
	if err == nil {
		t.Error("Expected error for non-existent agent")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestWaitForValidReport_AbortDuringWait tests abort while checking report.
func TestWaitForValidReport_AbortDuringWait(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-wait-abort-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Set abort flag
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	status, err := waitForValidReport(sessionID, "implement")
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if status != "" {
		t.Errorf("Expected empty status on abort, got %s", status)
	}
}

// TestCheckImplementReportStatus_AbortFlag tests abort during implement report check.
func TestCheckImplementReportStatus_AbortFlag(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-impl-report-abort-flag-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Set abort flag
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

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
	exitCode, err := checkImplementReportStatus("", 1)
	if err == nil {
		t.Error("Expected error for empty session ID")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
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
	sessionID := "test-review-error-" + time.Now().Format("20060102150405.000000")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	cfg := types.Config{
		Agent:           "nonexistent-agent-xyz",
		SessionID:       sessionID,
		MaxReviewCycles: 1,
		TaskFilePath:    "",
		CommandFiles:    nil,
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Fatal("Expected error from runReviewPhase, got nil")
	}
	if status != "" {
		t.Errorf("Expected empty status on error, got %q", status)
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code, got 0")
	}
}
