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
	sessionID := "test-abort-between"
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

	go func() {
		time.Sleep(100 * time.Millisecond)
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
		time.Sleep(50 * time.Millisecond)
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
	sessionID := "test-review-fail-continue"
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

	reportCount := 0
	go func() {
		for i := 0; i < 4; i++ {
			time.Sleep(100 * time.Millisecond)
			reportCount++
			status := "PASS"
			if reportCount == 2 {
				status = "FAIL" // First review fails
			}
			report := `command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: ` + status + `
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
		t.Errorf("Expected success after retry, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestSetupSignalHandler_InvalidSession(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/dev/null")
	setupSignalHandler("test-invalid")
	time.Sleep(10 * time.Millisecond)
}

func TestSetupSignalHandler_MultipleSetups(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	setupSignalHandler("test-1")
	setupSignalHandler("test-2")
	setupSignalHandler("test-3")
	time.Sleep(10 * time.Millisecond)
}

func TestRunWorkflow_ReviewPhaseFailure(t *testing.T) {
	sessionID := "test-review-fail"
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-abort-check-warn"
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
