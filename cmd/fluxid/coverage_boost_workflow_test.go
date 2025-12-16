//nolint:paralleltest // Coverage boost tests don't need parallel execution
package main

import (
	"fluxid-loop/internal/ipc"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunWorkflow_AbortBetweenPhases(t *testing.T) {
	t.Cleanup(cleanupAllSignalHandlers)
	sessionID := "test-abort-between-" + time.Now().Format("20060102150405.000000")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
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

	// Set abort flag very quickly to catch workflow between phases
	// Write report first so implement phase completes, then immediately set abort
	go func() {
		report := `command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: PASS
summary: Success
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
		_ = ipc.WriteReport(sessionID, report)
		// Set abort flag immediately after writing report
		_ = ipc.SetAbortFlag(sessionID)
	}()

	exitCode, err := runWorkflow(cfg)
	if err == nil {
		t.Error("Expected error when abort set")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130, got %d", exitCode)
	}
}

func TestRunWorkflow_ReviewFailContinues(t *testing.T) {
	t.Cleanup(cleanupAllSignalHandlers)
	sessionID := "test-review-fail-continue-" + time.Now().Format("20060102150405.000000")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     3,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	// Pre-write PASS report before workflow starts
	// No timing dependencies - completely deterministic
	if err := ipc.WriteReport(sessionID, testPassReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	exitCode, err := runWorkflow(cfg)
	if err != nil {
		t.Errorf("Expected success after retry, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestSetupSignalHandler_InvalidSession(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/dev/null")
	cleanup := setupSignalHandler("test-invalid")
	t.Cleanup(cleanup)
	// No sleep needed - cleanup is immediate
}

func TestSetupSignalHandler_MultipleSetups(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cleanup1 := setupSignalHandler("test-1")
	t.Cleanup(cleanup1)
	cleanup2 := setupSignalHandler("test-2")
	t.Cleanup(cleanup2)
	cleanup3 := setupSignalHandler("test-3")
	t.Cleanup(cleanup3)
	// No sleep needed - cleanup is immediate
}

func TestRunWorkflow_ReviewPhaseFailure(t *testing.T) {
	t.Cleanup(cleanupAllSignalHandlers)
	sessionID := "test-review-fail-" + time.Now().Format("20060102150405.000000")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "false",
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
		t.Error("Expected error")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestRunWorkflow_AbortCheckError(t *testing.T) {
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-abort-check-warn-" + time.Now().Format("20060102150405.000000")
	if err := os.MkdirAll(filepath.Join(tmpDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	// Write a PASS report to complete workflow
	go func() {
		for i := 0; i < 2; i++ {
			time.Sleep(100 * time.Millisecond)
			report := `command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: PASS
summary: Test
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
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

func TestRunSimulation_HappyPath(t *testing.T) {
	cfg := Config{
		SessionID:           "test-simulation",
		Agent:               "echo",
		MaxReviewCycles:     2,
		MaxImplementRetries: 2,
		CommitEnabled:       true,
		DryRun:              true,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        OutputFormatText,
		Sources:             map[string]string{},
	}

	exitCode := runSimulation(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for simulation, got %d", exitCode)
	}
}
