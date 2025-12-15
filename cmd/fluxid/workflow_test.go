package main

import (
	"fluxid-loop/internal/ipc"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunWorkflow_ImmediateAbort(t *testing.T) {
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

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "echo",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runWorkflow(cfg)
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
	// Test workflow with multiple review cycles (all fail)
	sessionID := "test-multi-cycle"
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "false",
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runWorkflow(cfg)
	// Should fail at first implement phase
	if err == nil {
		t.Error("Expected error for failing workflow")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestRunWorkflow_ChecksAbortBeforeReview(t *testing.T) {
	// This test can't easily verify the abort check without setting it mid-execution
	// But it exercises the code path
	sessionID := "test-workflow-abort-before-review"
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "false",
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runWorkflow(cfg)
	// Will fail on first implement attempt
	if err == nil {
		t.Error("Expected error")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestRunWorkflow_SuccessFirstCycle(t *testing.T) {
	// Test successful workflow completion in first cycle
	sessionID := "test-workflow-success"
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	t.Setenv("XDG_DATA_HOME", tmpDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	// Write reports asynchronously to simulate agent responses
	go func() {
		time.Sleep(100 * time.Millisecond)
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

		time.Sleep(100 * time.Millisecond)
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

	exitCode, err := runWorkflow(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunWorkflow_FailThenPass(t *testing.T) {
	// Test workflow that fails first cycle then passes
	sessionID := "test-workflow-retry-pass"
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	t.Setenv("XDG_DATA_HOME", tmpDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     3,
		MaxImplementRetries: 2,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	cycleCount := 0
	go func() {
		for i := 0; i < 4; i++ {
			time.Sleep(100 * time.Millisecond)
			cycleCount++

			var report string
			if cycleCount <= 2 {
				// First cycle: FAIL
				report = `command: test-phase
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: FAIL
summary: Need improvement
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Retry
`
			} else {
				// Second cycle: PASS
				report = `command: test-phase
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
			}
			_ = ipc.WriteReport(sessionID, report)
		}
	}()

	exitCode, err := runWorkflow(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
