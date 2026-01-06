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

func TestRunImplementPhase_Success(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionPhaseRun

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report immediately before calling runImplementPhase
	// This ensures deterministic test behavior without timing dependencies
	if err := storage.WriteReport(sessionID, testImplementPassReport); err != nil {
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

func TestRunImplementPhase_WithCommitViaRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionPhaseRun

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report immediately before calling runImplementPhase
	// This ensures deterministic test behavior without timing dependencies
	if err := storage.WriteReport(sessionID, testImplementPassReport); err != nil {
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

func TestRunReviewPhase_SuccessViaRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionPhaseRun

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write review PASS report before calling runReviewPhase
	// With immediate checking, report must exist before the phase executes
	if err := storage.WriteReport(sessionID, testPassReport); err != nil {
		t.Fatalf("Failed to write review report: %v", err)
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if status != statusPass {
		t.Errorf("Expected status PASS, got %s", status)
	}
}

func TestRunCommitPhase_SuccessViaRun(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "550e8400-e29b-41d4-a716-446655440000"

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := runCommitPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunImplementPhase_WithCommandFile(t *testing.T) {
	defer goleak.VerifyNone(t)

	tmpDir, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := testSessionPhaseRun

	cmdFiles := &config.ResolvedCommandFiles{
		ImplementPath: tmpDir + "/implement.md",
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        cmdFiles,
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report immediately before calling runImplementPhase
	// This ensures deterministic test behavior without timing dependencies
	if err := storage.WriteReport(sessionID, testImplementPassReport); err != nil {
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
