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

func TestRun_SingleCycleSuccess(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionRun

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report immediately before calling Run()
	// This ensures deterministic test behavior without timing dependencies
	if err := storage.WriteReport(sessionID, testImplementPassReport); err != nil {
		t.Fatalf("Failed to write implement report: %v", err)
	}

	exitCode, err := Run(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRun_AbortBeforeImplement(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionRun

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Set abort flag before starting
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	exitCode, err := Run(cfg)
	if err == nil {
		t.Error("Expected error for aborted workflow")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}

func TestRun_MultipleReviewCycles(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionRun

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     3,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write initial implement PASS report before calling Run()
	// The workflow will: implement (PASS) -> commit -> review (FAIL) -> implement (PASS) -> commit -> review (PASS)
	// We pre-write the first implement report to start the workflow deterministically
	if err := storage.WriteReport(sessionID, testImplementPassReport); err != nil {
		t.Fatalf("Failed to write initial implement report: %v", err)
	}

	exitCode, err := Run(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRun_WithAgentArgs(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionRun

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		AgentArgs:           []string{"-n", "test"},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report immediately before calling Run()
	// This ensures deterministic test behavior without timing dependencies
	if err := storage.WriteReport(sessionID, testImplementPassReport); err != nil {
		t.Fatalf("Failed to write implement report: %v", err)
	}

	exitCode, err := Run(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRun_ReadReportFailure(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	// Use invalid session ID to trigger read error
	sessionID := ""

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := Run(cfg)
	// Corrected behavior: Even with invalid session ID, workflow completes all cycles
	// Missing reports are treated as FAIL status, not errors
	if err != nil {
		t.Errorf("Expected no error (FAIL status used for failures), got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (workflow completes), got %d", exitCode)
	}
}

func TestRun_AllReviewCyclesExhausted(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "a1b2c3d4-e5f6-4071-8c9d-0e1f2a3b4c5d"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               testAgentEcho,
		AgentArgs:           []string{},
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report before calling Run()
	if err := storage.WriteReport(sessionID, testImplementPassReport); err != nil {
		t.Fatalf("Failed to write implement report: %v", err)
	}

	// Run() will exhaust both review cycles with FAIL status
	// The workflow continues through all cycles and returns success (exit code 0)
	exitCode, err := Run(cfg)
	if err != nil {
		t.Errorf("Expected no error when cycles exhausted, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 when cycles exhausted, got %d", exitCode)
	}
}

func TestRun_ImplementPhaseError(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "b2c3d4e5-f6a7-4182-9d0e-1f2a3b4c5d6e"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               testAgentFalse, // Agent that fails
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Corrected behavior: Run() continues to review even after commit failures
	// Missing reports are treated as FAIL status, workflow completes all cycles
	exitCode, err := Run(cfg)
	if err != nil {
		t.Errorf("Expected no error (FAIL status used for failures), got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (workflow completes), got %d", exitCode)
	}
}

func TestRun_ReviewPhaseError(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "c3d4e5f6-a7b8-4293-0e1f-2a3b4c5d6e7f"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               testAgentEcho, // Use echo to pass implement/commit
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report
	if err := storage.WriteReport(sessionID, testImplementPassReport); err != nil {
		t.Fatalf("Failed to write implement report: %v", err)
	}

	// Don't write a review report - this will cause waitForValidReport to treat it as FAIL
	// which returns empty status and causes Run() to continue (not error)

	exitCode, err := Run(cfg)
	// When review returns FAIL status (not error), Run() completes all cycles successfully
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
