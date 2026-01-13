//nolint:paralleltest,exhaustruct // Tests use global mutex and incomplete structs
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/types"
	"testing"

	"go.uber.org/goleak"
)

func TestRunCommitPhaseWithRetry_SuccessFirstAttempt(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "550e8400-e29b-41d4-a716-446655440000"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    3,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write PASS commit report before calling function
	if err := storage.WriteReport(sessionID, testCommitPassReport); err != nil {
		t.Fatalf("Failed to write commit report: %v", err)
	}

	exitCode, err := runCommitPhaseWithRetry(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunCommitPhaseWithRetry_AllRetriesFail(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "6ba7b810-9dad-11d1-80b4-00c04fd430c1"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write FAIL report that will persist through all retries
	if err := storage.WriteReport(sessionID, testCommitFailReport); err != nil {
		t.Fatalf("Failed to write commit FAIL report: %v", err)
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

func TestRunCommitPhaseWithRetry_AbortDuringRetry(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    3,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Set abort flag
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	exitCode, err := runCommitPhaseWithRetry(cfg)
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

func TestCheckCommitReportStatus_Pass(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "7c9e6679-7425-40de-944b-e07fc1f90ae7"

	// Write PASS commit report
	if err := storage.WriteReport(sessionID, testCommitPassReport); err != nil {
		t.Fatalf("Failed to write commit report: %v", err)
	}

	exitCode, err := checkCommitReportStatus(sessionID, "", 1)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestCheckCommitReportStatus_Fail(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "c73bcdcc-2669-4bf6-81d3-e4ae73fb11fd"

	// Write FAIL commit report
	if err := storage.WriteReport(sessionID, testCommitFailReport); err != nil {
		t.Fatalf("Failed to write commit FAIL report: %v", err)
	}

	exitCode, err := checkCommitReportStatus(sessionID, "", 1)
	if err != nil {
		t.Errorf("Expected no error for FAIL status, got: %v", err)
	}
	if exitCode != -1 {
		t.Errorf("Expected exit code -1 (signal to retry), got %d", exitCode)
	}
}

func TestCheckCommitReportStatus_AbortDuringCheck(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "3d6f7e0a-8c0f-4e38-ad73-7f6d91e8b2c1"

	// Set abort flag
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	exitCode, err := checkCommitReportStatus(sessionID, "", 1)
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}
