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
		MaxCommitRetries:    100,
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
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write initial implement PASS report before calling Run()
	// The workflow will: implement (PASS) -> commit -> review (FAIL) -> implement (PASS) -> commit -> review (PASS)
	// We pre-write the first implement report to start the workflow deterministically
	if err := ipc.WriteReport(sessionID, testImplementPassReport); err != nil {
		t.Fatalf("Failed to write initial implement report: %v", err)
	}

	exitCode, err := Run(cfg)
	if err != nil {
		t.Errorf("Expected no error after retry succeeds, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
