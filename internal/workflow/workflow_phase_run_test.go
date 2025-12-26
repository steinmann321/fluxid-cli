//nolint:exhaustruct,paralleltest // Tests use global mutex and incomplete structs
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

func TestRunImplementPhase_Success(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-impl-success-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
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

func TestRunImplementPhase_WithCommitViaRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-impl-commit-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
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

func TestRunReviewPhase_SuccessViaRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-review-success-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write review PASS report before calling runReviewPhase
	// With immediate checking, report must exist before the phase executes
	if err := ipc.WriteReport(sessionID, testPassReport); err != nil {
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

	sessionID := "test-commit-success-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
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

	sessionID := "test-implement-cmdfile-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cmdFiles := &config.ResolvedCommandFiles{
		ImplementPath: tmpDir + "/implement.md",
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        cmdFiles,
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
