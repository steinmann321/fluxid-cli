//nolint:exhaustruct,paralleltest // Tests use global mutex and incomplete structs
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

// TestAbortError_Error tests the AbortError Error() method.
func TestAbortError_Error(t *testing.T) {
	t.Parallel()
	err := &AbortError{
		ExitCode: 130,
		Message:  "Workflow aborted by signal",
	}

	expected := "Workflow aborted by signal"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

// TestRunImplementPhase_AbortBeforeRetry tests abort checking across retries.
func TestRunImplementPhase_AbortBeforeRetry(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-abort-retry-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false", // Fails immediately
		AgentArgs:           []string{},
		MaxImplementRetries: 3,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Set abort flag before running
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatal(err)
	}

	// Should detect abort and return error
	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error when abort flag is set")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

// TestWaitForValidReport_NoReportReturnsFAIL tests handling when no report is written.
func TestWaitForValidReport_NoReportReturnsFAIL(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-noreport-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Don't write any report - should return FAIL immediately
	status, err := waitForValidReport(sessionID, "test-phase")
	if err != nil {
		t.Errorf("Expected no error when report missing (should return FAIL), got: %v", err)
	}
	if status != statusFail {
		t.Errorf("Expected status FAIL when no report exists, got: %s", status)
	}
}

func TestRun_ImplementPhaseAbort(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-run-impl-abort-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Set abort flag to trigger abort
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	exitCode, err := Run(cfg)
	if err == nil {
		t.Error("Expected error for aborted workflow")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

func TestRunReviewPhase_Abort(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-review-abort-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Set abort flag before calling runReviewPhase
	// With immediate report checking, abort must be set before the phase runs
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Error("Expected error for aborted review phase")
	}
	if exitCode != 130 && exitCode != 1 {
		t.Errorf("Expected exit code 130 or 1 for abort, got %d", exitCode)
	}
	// Status can be empty string when error occurs
	_ = status
}

func TestRunImplementPhase_Abort(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-impl-abort-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Set abort flag before implement phase
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error for aborted implement phase")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

func TestRunReviewPhase_AgentCommandFail(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-review-agentfail-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "/bin/false", // Command that always fails
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	_, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Error("Expected error when agent command fails")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when agent fails")
	}
}
