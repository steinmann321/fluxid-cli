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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	// Start goroutine to write reports after a brief delay
	go func() {
		<-time.After(100 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, testPassReport)
	}()

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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	failReport := `command: "test"
artifact: "test-artifact"
timestamp: "2024-01-01T00:00:00Z"
status: FAIL
issues:
  blockers:
    - message: "Test failed"
  defects: []
  concerns: []
  observations: []
  enhancements: []
summary: "Test failed"
`

	callCount := 0
	go func() {
		for callCount < 4 {
			<-time.After(100 * time.Millisecond)
			callCount++
			// First implement: PASS, first review: FAIL
			// Second implement: PASS, second review: PASS
			switch callCount {
			case 1, 3:
				_ = ipc.WriteReport(sessionID, testPassReport)
			case 2:
				_ = ipc.WriteReport(sessionID, failReport)
			}
		}
	}()

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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	go func() {
		<-time.After(100 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, testPassReport)
	}()

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
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        &config.ResolvedCommandFiles{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	_, err := Run(cfg)
	if err == nil {
		t.Error("Expected error due to invalid session ID, got nil")
	}
}
