//nolint:exhaustruct,paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"testing"

	"go.uber.org/goleak"
)

// TestRunImplementPhase_CommitPhaseFailure tests failure in commit phase.
func TestRunImplementPhase_CommitPhaseFailure(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "b8c9d0e1-f2a3-4b5c-6d7e-8f9a0b1c2d3e"

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentFalse, // Will fail on commit phase
		AgentArgs:           []string{},
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// No report needed since commit will fail before waiting

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error when commit phase fails")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestRunImplementPhase_AgentFailsNoExit tests implement with agent that fails without exit code.
func TestRunImplementPhase_AgentFailsNoExit(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "c9d0e1f2-a3b4-5c6d-7e8f-9a0b1c2d3e4f"

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentFalse,
		AgentArgs:           []string{},
		MaxImplementRetries: 2,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Since agent fails, we won't get to report checking
	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error when agent fails")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestRunImplementPhase_ReportWaitAbort tests abort during implement report wait.
func TestRunImplementPhase_ReportWaitAbort(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "d0e1f2a3-b4c5-6d7e-8f9a-0b1c2d3e4f5a"

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Set abort flag before calling runImplementPhase
	// With immediate report checking, abort must be set before the phase runs
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error due to abort during report wait")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130, got %d", exitCode)
	}
}
