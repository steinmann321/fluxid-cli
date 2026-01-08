//nolint:exhaustruct,paralleltest // Tests use global mutex and incomplete structs
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/types"
	"testing"

	"go.uber.org/goleak"
)

// TestRunCommitPhase_AgentFailure tests commit phase with failing agent.
func TestRunCommitPhase_AgentFailure(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionReviewCommit

	cfg := types.Config{
		SessionID:    sessionID,
		SessionRoot:  "",
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

	sessionID := testSessionReviewCommit

	cfg := types.Config{
		SessionID:    sessionID,
		SessionRoot:  "",
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

	sessionID := testSessionReviewCommit

	cfg := types.Config{
		SessionID:    sessionID,
		SessionRoot:  "",
		Agent:        testAgentFalse,
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	status, exitCode, err := runReviewPhase(cfg)
	// With new behavior: missing reports are treated as FAIL, not errors
	if err == nil {
		// This is expected - missing report returns FAIL status, not error
		if status != statusFail {
			t.Errorf("Expected FAIL status for missing report, got %s", status)
		}
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for FAIL status, got %d", exitCode)
		}
	} else {
		t.Errorf("Expected no error (FAIL status instead), got error: %v", err)
	}
}

// TestRunReviewPhase_AgentFailsZeroExit tests review with command execution error.
func TestRunReviewPhase_AgentFailsZeroExit(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionReviewCommit

	cfg := types.Config{
		SessionID:    sessionID,
		SessionRoot:  "",
		Agent:        "/nonexistent/command/path",
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	status, exitCode, err := runReviewPhase(cfg)
	// With new behavior: missing reports are treated as FAIL, not errors
	if err == nil {
		// This is expected - missing report returns FAIL status, not error
		if status != statusFail {
			t.Errorf("Expected FAIL status for missing report, got %s", status)
		}
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for FAIL status, got %d", exitCode)
		}
	} else {
		t.Errorf("Expected no error (FAIL status instead), got error: %v", err)
	}
}

// TestRunReviewPhase_ReportWaitAbort tests abort during review report wait.
func TestRunReviewPhase_ReportWaitAbort(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionReviewCommit

	cfg := types.Config{
		SessionID:    sessionID,
		SessionRoot:  "",
		Agent:        testAgentEcho,
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	// Set abort flag before calling runReviewPhase
	// With immediate report checking, abort must be set before the phase runs
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

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

	sessionID := testSessionReviewCommit

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
	if err := storage.WriteReport(sessionID, invalidStructure); err != nil {
		t.Fatalf("Failed to write invalid report: %v", err)
	}

	// waitForValidReport should immediately return FAIL for invalid structure
	status, err := waitForValidReport(sessionID, "", "test")
	if err != nil {
		t.Errorf("Expected no error (should return FAIL status), got: %v", err)
	}
	if status != statusFail {
		t.Errorf("Expected FAIL status for invalid structure, got %s", status)
	}
}

// TestRunReviewPhase_Success2 tests successful review phase execution with pre-written report.
func TestRunReviewPhase_Success2(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "a0b1c2d3-e4f5-4a6b-7c8d-9e0f1a2b3c4d"

	cfg := types.Config{
		SessionID:    sessionID,
		SessionRoot:  "",
		Agent:        testAgentEcho,
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	// Pre-write PASS review report
	reviewPassReport := `command: test
artifact: test
timestamp: "2024-01-01T00:00:00Z"
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Continue with workflow
summary: Review completed successfully
`
	if err := storage.WriteReport(sessionID, reviewPassReport); err != nil {
		t.Fatalf("Failed to write review report: %v", err)
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}
}

// TestRunReviewPhase_FailStatus tests review phase with FAIL status.
func TestRunReviewPhase_FailStatus(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "b1c2d3e4-f5a6-4b7c-8d9e-0f1a2b3c4d5e"

	cfg := types.Config{
		SessionID:    sessionID,
		SessionRoot:  "",
		Agent:        testAgentEcho,
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
	}

	// Pre-write FAIL review report
	reviewFailReport := `command: test
artifact: test
timestamp: "2024-01-01T00:00:00Z"
status: FAIL
issues:
  blockers:
    - "Critical issue found"
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Fix critical issue
summary: Review found blocking issues
`
	if err := storage.WriteReport(sessionID, reviewFailReport); err != nil {
		t.Fatalf("Failed to write review report: %v", err)
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if status != statusFail {
		t.Errorf("Expected FAIL status, got %s", status)
	}
}

// TestCheckAbortBeforeCommit_NoAbort tests normal flow without abort.
func TestCheckAbortBeforeCommit_NoAbort(t *testing.T) {
	t.Parallel()

	exitCode, err := checkAbortBeforeCommit("test-session-id")
	if err != nil {
		t.Errorf("Expected no error when no abort, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 when no abort, got %d", exitCode)
	}
}

// TestExecuteCommitPhase_Success tests successful commit phase.
func TestExecuteCommitPhase_Success(t *testing.T) {
	// Cannot use t.Parallel() with setupTestDataDir() which calls t.Setenv()
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "c2d3e4f5-a6b7-4c8d-9e0f-1a2b3c4d5e6f"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := executeCommitPhase(cfg, 1)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// TestExecuteCommitPhase_Failure tests commit phase with failing agent.
func TestExecuteCommitPhase_Failure(t *testing.T) {
	// Cannot use t.Parallel() with setupTestDataDir() which calls t.Setenv()
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "d3e4f5a6-b7c8-4d9e-0f1a-2b3c4d5e6f7a"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               testAgentFalse,
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := executeCommitPhase(cfg, 1)
	if err == nil {
		t.Error("Expected error for failing agent")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}
