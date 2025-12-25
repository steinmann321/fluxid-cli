//nolint:exhaustruct,paralleltest // Tests use global mutex and incomplete structs
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

// TestRun_AbortAfterImplementPhase tests abort flag check after implement phase completes.
func TestRun_AbortAfterImplementPhase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing-dependent abort test in short mode")
	}
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-run-abort-after-impl-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report before calling Run()
	// With immediate checking, report must exist before runImplementPhase() executes
	if err := ipc.WriteReport(sessionID, testPassReport); err != nil {
		t.Fatalf("Failed to write implement report: %v", err)
	}

	// Set abort flag before calling Run()
	// This will be caught after implement phase completes and before review phase starts
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

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
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-review-fail-continue-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxReviewCycles:     3,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Return FAIL for first review cycle, PASS for second
	reportCount := 0
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 10; i++ {
			<-time.After(50 * time.Millisecond)
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

	waitGroup.Wait()
}
