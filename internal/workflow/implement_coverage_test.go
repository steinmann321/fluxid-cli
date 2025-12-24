//nolint:exhaustruct // Implement phase coverage tests
package workflow

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"testing"
	"time"
)

// TestRunImplementPhase_RetryOnFailReport tests the retry loop when implement reports FAIL.
func TestRunImplementPhase_RetryOnFailReport(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-impl-retry-fail-" + time.Now().Format("20060102150405")

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
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(50 * time.Millisecond)
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
}

// TestRunImplementPhase_MaxRetriesExceeded tests exceeding max implement retries.
func TestRunImplementPhase_MaxRetriesExceeded(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-impl-maxretry-" + time.Now().Format("20060102150405")

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
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(50 * time.Millisecond)
			_ = ipc.WriteReport(sessionID, testFailReport)
		}
	}()

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error when max retries exceeded")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

// TestRunImplementPhase_WithCommitEnabled tests implement phase with commit enabled.
func TestRunImplementPhase_WithCommitEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-impl-commit-" + time.Now().Format("20060102150405")

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
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, testPassReport)
	}()

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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-impl-commit-fail-" + time.Now().Format("20060102150405")

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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-impl-agent-fail-noexit-" + time.Now().Format("20060102150405")

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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-impl-wait-abort-" + time.Now().Format("20060102150405")

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

	// Set abort flag while waiting for implement report
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = ipc.SetAbortFlag(sessionID)
	}()

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error due to abort during report wait")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130, got %d", exitCode)
	}
}
