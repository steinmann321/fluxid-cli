//nolint:exhaustruct,paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"fmt"
	"sync"
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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	// Return FAIL reports to trigger retries
	reportCount := 0
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 5; i++ {
			<-time.After(50 * time.Millisecond)
			reportCount++
			if reportCount < 3 {
				_ = ipc.WriteReport(sessionID, testFailReport)
			} else {
				_ = ipc.WriteReport(sessionID, testPassReport)
			}
		}
	}()

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error after retries succeed, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	waitGroup.Wait()
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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	// Always return FAIL to exceed retries
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 10; i++ {
			<-time.After(50 * time.Millisecond)
			_ = ipc.WriteReport(sessionID, testFailReport)
		}
	}()

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error when max retries exceeded (should continue), got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (continue to next phase), got %d", exitCode)
	}

	waitGroup.Wait()
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
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	// Provide PASS report for implement phase
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		<-time.After(50 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, testPassReport)
	}()

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	waitGroup.Wait()
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
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
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
