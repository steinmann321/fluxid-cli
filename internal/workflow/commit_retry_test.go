//nolint:paralleltest,exhaustruct // Tests use global mutex and incomplete structs
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/ipc"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"fmt"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestRunCommitPhaseWithRetry_SuccessFirstAttempt(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-commit-success-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
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
	if err := ipc.WriteReport(sessionID, testCommitPassReport); err != nil {
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

	sessionID := "test-commit-all-fail-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
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
	if err := ipc.WriteReport(sessionID, testCommitFailReport); err != nil {
		t.Fatalf("Failed to write commit FAIL report: %v", err)
	}

	exitCode, err := runCommitPhaseWithRetry(cfg)
	if err == nil {
		t.Error("Expected error after all retries fail")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

func TestRunCommitPhaseWithRetry_AbortDuringRetry(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-commit-abort-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
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
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

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

	sessionID := "test-commit-status-pass-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Write PASS commit report
	if err := ipc.WriteReport(sessionID, testCommitPassReport); err != nil {
		t.Fatalf("Failed to write commit report: %v", err)
	}

	exitCode, err := checkCommitReportStatus(sessionID, 1)
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

	sessionID := "test-commit-status-fail-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Write FAIL commit report
	if err := ipc.WriteReport(sessionID, testCommitFailReport); err != nil {
		t.Fatalf("Failed to write commit FAIL report: %v", err)
	}

	exitCode, err := checkCommitReportStatus(sessionID, 1)
	if err != nil {
		t.Errorf("Expected no error for FAIL status, got: %v", err)
	}
	if exitCode != -1 {
		t.Errorf("Expected exit code -1 (signal to retry), got %d", exitCode)
	}
}

func TestCheckCommitReportStatus_AbortDuringCheck(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-commit-status-abort-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Set abort flag
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	exitCode, err := checkCommitReportStatus(sessionID, 1)
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}
