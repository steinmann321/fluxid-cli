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

// TestRunCommitPhase_AgentFailure tests commit phase with failing agent.
func TestRunCommitPhase_AgentFailure(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-commit-fail-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:    sessionID,
		Agent:        testAgentFalse,
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	exitCode, err := runCommitPhase(cfg)
	if err == nil {
		t.Error("Expected error when commit agent fails")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestRunCommitPhase_AgentFailsZeroExit tests commit with agent that returns error but no exit code.
func TestRunCommitPhase_AgentFailsZeroExit(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-commit-fail-zeroexit-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:    sessionID,
		Agent:        "/nonexistent/command/path", // Will fail to execute
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	exitCode, err := runCommitPhase(cfg)
	if err == nil {
		t.Error("Expected error when commit agent command not found")
	}
	// Could be exit code 1 (from error path) or other depending on exec failure
	_ = exitCode
}

// TestRunReviewPhase_AgentNonZeroExit tests review phase error handling with non-zero exit.
func TestRunReviewPhase_AgentNonZeroExit(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-review-nonzero-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:    sessionID,
		Agent:        testAgentFalse,
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Error("Expected error when review agent fails")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	if status != "" {
		t.Errorf("Expected empty status on error, got %s", status)
	}
}

// TestRunReviewPhase_AgentFailsZeroExit tests review with command execution error.
func TestRunReviewPhase_AgentFailsZeroExit(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-review-fail-zeroexit-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:    sessionID,
		Agent:        "/nonexistent/command/path",
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Error("Expected error when review agent command not found")
	}
	if status != "" {
		t.Errorf("Expected empty status on error, got %s", status)
	}
	_ = exitCode
}

// TestRunReviewPhase_ReportWaitAbort tests abort during review report wait.
func TestRunReviewPhase_ReportWaitAbort(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-review-wait-abort-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:    sessionID,
		Agent:        testAgentEcho,
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	// Set abort flag before calling runReviewPhase
	// With immediate report checking, abort must be set before the phase runs
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Error("Expected error due to abort")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130, got %d", exitCode)
	}
	if status != "" {
		t.Errorf("Expected empty status on abort, got %s", status)
	}
}

// TestWaitForValidReport_UnmarshalError tests that reports with invalid structure are treated as FAIL.
func TestWaitForValidReport_UnmarshalError(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-unmarshal-err-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Write a report with invalid structure (status is array instead of string)
	// With immediate checking, this should return FAIL immediately
	invalidStructure := `command: test
artifact: test
timestamp: "2024-01-01T00:00:00Z"
status: [PASS, FAIL]
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := ipc.WriteReport(sessionID, invalidStructure); err != nil {
		t.Fatalf("Failed to write invalid report: %v", err)
	}

	// waitForValidReport should immediately return FAIL for invalid structure
	status, err := waitForValidReport(sessionID, "test")
	if err != nil {
		t.Errorf("Expected no error (should return FAIL status), got: %v", err)
	}
	if status != statusFail {
		t.Errorf("Expected FAIL status for invalid structure, got %s", status)
	}
}
