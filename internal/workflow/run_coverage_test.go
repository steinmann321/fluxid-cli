//nolint:exhaustruct // Run function coverage tests
package workflow

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"testing"
	"time"
)

// TestRun_AbortAfterImplementPhase tests abort flag check after implement phase completes.
func TestRun_AbortAfterImplementPhase(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-run-abort-after-impl-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	// Set abort flag after implement phase runs
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, testPassReport)
		time.Sleep(50 * time.Millisecond)
		_ = ipc.SetAbortFlag(sessionID)
	}()

	exitCode, err := Run(cfg)
	if err == nil {
		t.Error("Expected error when abort flag is set after implement")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

// TestRun_ReviewCycleFAILContinuation tests workflow continuing after FAIL review.
func TestRun_ReviewCycleFAILContinuation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-review-fail-continue-" + time.Now().Format("20060102150405")

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxReviewCycles:     3,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	// Return FAIL for first review cycle, PASS for second
	reportCount := 0
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(50 * time.Millisecond)
			reportCount++
			// First cycle: implement PASS, review FAIL
			// Second cycle: implement PASS, review PASS
			switch {
			case reportCount <= 2:
				_ = ipc.WriteReport(sessionID, testPassReport)
			case reportCount == 3:
				_ = ipc.WriteReport(sessionID, testFailReport)
			default:
				_ = ipc.WriteReport(sessionID, testPassReport)
			}
		}
	}()

	exitCode, err := Run(cfg)
	if err != nil {
		t.Errorf("Expected no error after retry succeeds, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
