//nolint:exhaustruct // Review and commit phase coverage tests
package workflow

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"testing"
	"time"
)

// TestRunCommitPhase_AgentFailure tests commit phase with failing agent.
func TestRunCommitPhase_AgentFailure(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-commit-fail-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:     sessionID,
		Agent:         testAgentFalse,
		AgentArgs:     []string{},
		CommitEnabled: true,
		DryRun:        false,
		CommandFiles:  &config.ResolvedCommandFiles{},
		OutputFormat:  output.FormatText,
		Sources:       map[string]string{},
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-commit-fail-zeroexit-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:     sessionID,
		Agent:         "/nonexistent/command/path", // Will fail to execute
		AgentArgs:     []string{},
		CommitEnabled: true,
		DryRun:        false,
		CommandFiles:  &config.ResolvedCommandFiles{},
		OutputFormat:  output.FormatText,
		Sources:       map[string]string{},
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-review-nonzero-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:    sessionID,
		Agent:        testAgentFalse,
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
		Sources:      map[string]string{},
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-review-fail-zeroexit-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:    sessionID,
		Agent:        "/nonexistent/command/path",
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
		Sources:      map[string]string{},
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-review-wait-abort-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:    sessionID,
		Agent:        testAgentEcho,
		AgentArgs:    []string{},
		DryRun:       false,
		CommandFiles: &config.ResolvedCommandFiles{},
		OutputFormat: output.FormatText,
		Sources:      map[string]string{},
	}

	// Set abort flag while waiting for report
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = ipc.SetAbortFlag(sessionID)
	}()

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

// TestWaitForValidReport_UnmarshalError tests handling of report unmarshal errors.
func TestWaitForValidReport_UnmarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-unmarshal-err-" + time.Now().Format("20060102150405")

	// Write a report that passes validation but has unmarshal issues, then a good one
	go func() {
		time.Sleep(50 * time.Millisecond)
		// This will pass basic YAML validation but might cause unmarshal issues
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
		_ = ipc.WriteReport(sessionID, invalidStructure)
		time.Sleep(100 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, testPassReport)
	}()

	status, err := waitForValidReport(sessionID, "test")
	if err != nil {
		t.Errorf("Expected no error after valid report, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}
}
