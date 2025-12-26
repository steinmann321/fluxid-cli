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

func TestRun_SingleCycleSuccess(t *testing.T) {
	defer goleak.VerifyNone(t)

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-run-single-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

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

	// Write implement PASS report immediately before calling Run()
	// This ensures deterministic test behavior without timing dependencies
	if err := ipc.WriteReport(sessionID, testImplementPassReport); err != nil {
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
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "test-run-abort-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Set abort flag before starting
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

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

	sessionID := "test-run-multi-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     3,
		MaxImplementRetries: 1,
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

	sessionID := "test-agent-args-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		AgentArgs:           []string{"-n", "test"},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	// Write implement PASS report immediately before calling Run()
	// This ensures deterministic test behavior without timing dependencies
	if err := ipc.WriteReport(sessionID, testImplementPassReport); err != nil {
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
		Agent:               "echo",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
	}

	_, err := Run(cfg)
	if err == nil {
		t.Error("Expected error due to invalid session ID, got nil")
	}
}
