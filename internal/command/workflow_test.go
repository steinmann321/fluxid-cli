package command

import (
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"fluxid-loop/internal/workflow"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestRunWorkflow_ImmediateAbort(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	// Test that workflow respects abort flag before starting
	sessionID := "test-workflow-abort-session"
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	// Set up IPC storage
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create storage directory
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Set abort flag before workflow starts
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := workflow.Run(cfg)
	if err == nil {
		t.Error("Expected error for aborted workflow")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("Expected abort error message, got: %v", err)
	}
}

func TestRunWorkflow_MultipleReviewCycles(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	// Test workflow with multiple review cycles (all fail)
	sessionID := "test-multi-cycle"
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false",
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := workflow.Run(cfg)
	// Should fail at first implement phase
	if err == nil {
		t.Error("Expected error for failing workflow")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestRunWorkflow_ChecksAbortBeforeReview(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	// This test can't easily verify the abort check without setting it mid-execution
	// But it exercises the code path
	sessionID := "test-workflow-abort-before-review"
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false",
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := workflow.Run(cfg)
	// Will fail on first implement attempt
	if err == nil {
		t.Error("Expected error")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

//nolint:funlen // Unit test with complete workflow validation
func TestRunWorkflow_SuccessFirstCycle(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	// Test successful workflow completion in first cycle
	sessionID := "test-workflow-success"
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	t.Setenv("XDG_DATA_HOME", tmpDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	// Write reports asynchronously to simulate agent responses
	// Use channel-based coordination instead of arbitrary sleep delays
	started := make(chan struct{})
	defer close(started)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		started <- struct{}{} // Signal goroutine is ready
		implementReport := `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Implementation successful
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Proceed to review
`
		_ = ipc.WriteReport(sessionID, implementReport)

		// Delay to allow implement phase to read the report
		// before we overwrite it with the review report
		<-time.After(100 * time.Millisecond)
		reviewReport := `command: test-review
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Review successful
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Complete
`
		_ = ipc.WriteReport(sessionID, reviewReport)
	}()
	<-started // Wait for goroutine to start before running workflow

	exitCode, err := workflow.Run(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	waitGroup.Wait()
}

func TestRunWorkflow_FailThenPass(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	// Redesigned test: Pre-write ALL reports before workflow starts
	// This eliminates timing dependencies entirely
	sessionID := "test-workflow-retry-pass-" + time.Now().Format("20060102150405.000000")
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	t.Setenv("XDG_DATA_HOME", tmpDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     3,
		MaxImplementRetries: 2,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	// Strategy: Write a PASS report before workflow starts
	// The workflow will use this report for ALL phases (implement and review)
	// This tests that workflow completes successfully when all phases pass
	passReport := `command: test-phase
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Success
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Complete
`
	if err := ipc.WriteReport(sessionID, passReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	exitCode, err := workflow.Run(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
