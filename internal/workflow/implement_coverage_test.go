//nolint:exhaustruct,paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"fmt"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// TestRunImplementPhase_RetryOnFailReport tests the retry loop when implement reports FAIL.
func TestRunImplementPhase_RetryOnFailReport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing-dependent retry test in short mode")
	}
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-impl-retry-fail-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxImplementRetries: 3,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write initial implement FAIL report to trigger first retry
	// The test simulates: 1st attempt FAIL, 2nd attempt FAIL, 3rd attempt PASS
	if err := ipc.WriteReport(sessionID, testImplementFailReport); err != nil {
		t.Fatalf("Failed to write initial implement FAIL report: %v", err)
	}

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error after retries succeed, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// TestRunImplementPhase_MaxRetriesExceeded tests that workflow continues when max implement retries are exceeded.
// The workflow should continue to commit (if enabled) and review phases even when all implement attempts fail.
func TestRunImplementPhase_MaxRetriesExceeded(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-impl-maxretry-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxImplementRetries: 2,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement FAIL report to simulate all retries failing
	// The workflow should exhaust all retries and continue to the next phase
	if err := ipc.WriteReport(sessionID, testImplementFailReport); err != nil {
		t.Fatalf("Failed to write implement FAIL report: %v", err)
	}

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error when max retries exceeded (should continue), got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (continue to next phase), got %d", exitCode)
	}
}

// TestRunImplementPhase_WithCommitEnabled tests implement phase with commit enabled.
func TestRunImplementPhase_WithCommitEnabled(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-impl-commit-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentTrue,
		AgentArgs:           []string{},
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report immediately before calling runImplementPhase
	// This ensures deterministic test behavior without timing dependencies
	if err := ipc.WriteReport(sessionID, testImplementPassReport); err != nil {
		t.Fatalf("Failed to write implement report: %v", err)
	}

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// TestRunImplementPhase_CommitPhaseFailure tests failure in commit phase.
func TestRunImplementPhase_CommitPhaseFailure(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-impl-commit-fail-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentFalse, // Will fail on commit phase
		AgentArgs:           []string{},
		MaxImplementRetries: 1,
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

	sessionID := "test-impl-agent-fail-noexit-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentFalse,
		AgentArgs:           []string{},
		MaxImplementRetries: 2,
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
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-impl-wait-abort-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Set abort flag before calling runImplementPhase
	// With immediate report checking, abort must be set before the phase runs
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error due to abort during report wait")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130, got %d", exitCode)
	}
}
