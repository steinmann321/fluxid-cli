//nolint:exhaustruct // Workflow error path tests
package workflow

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"testing"
	"time"
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-abort-retry-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false", // Fails immediately
		AgentArgs:           []string{},
		MaxImplementRetries: 3,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
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

// TestWaitForValidReport_NoReport tests handling when no report is written.
func TestWaitForValidReport_NoReport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode (takes 5 minutes)")
	}
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-noreport-" + time.Now().Format("20060102150405")

	// Don't write any report - should timeout
	_, err := waitForValidReport(sessionID, "test-phase")
	if err == nil {
		t.Error("Expected timeout error when no report is written, got nil")
	}
}

func TestRun_MaxCyclesExceeded(t *testing.T) {
	t.Skip("TODO: Fix test timing issues - reports not being picked up reliably")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-run-maxcycles-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	// Always return FAIL reports to exceed max cycles
	failReport := `command: "test"
artifact: "test-artifact"
timestamp: "2024-01-01T00:00:00Z"
status: FAIL
issues:
  blockers:
    - message: "Test failed"
  defects: []
  concerns: []
  observations: []
  enhancements: []
summary: "Test failed"
`

	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(50 * time.Millisecond)
			_ = ipc.WriteReport(sessionID, failReport)
		}
	}()

	exitCode, err := Run(cfg)
	if err == nil {
		t.Error("Expected error for exceeded max cycles")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for exceeded max cycles")
	}
}

func TestRun_ImplementPhaseAbort(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-run-impl-abort-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-review-abort-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	// Set abort flag before review phase
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = ipc.SetAbortFlag(sessionID)
	}()

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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-impl-abort-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-review-agentfail-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "/bin/false", // Command that always fails
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	_, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Error("Expected error when agent command fails")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when agent fails")
	}
}
