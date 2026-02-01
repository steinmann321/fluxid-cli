//nolint:paralleltest,exhaustruct,funlen,goconst,dupl // Test file with test fixtures
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteStepWithRetrySuccess(t *testing.T) {
	tempDir := t.TempDir()
	sessionRoot := tempDir
	sessionID := "00000000-0000-4000-8000-000000000300"

	// Create PASS report
	reportContent := `command: "test command"
artifact: "test artifact"
timestamp: "2026-01-01T00:00:00Z"
status: PASS
summary: "Test passed"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	reportPath := filepath.Join(sessionRoot, sessionID, "report.yaml")
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatalf("Failed to create report dir: %v", err)
	}
	if err := os.WriteFile(reportPath, []byte(reportContent), 0o600); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         sessionRoot,
		Agent:               "claude",
		MaxImplementRetries: 2,
		OutputFormat:        output.FormatText,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "test-implement.txt",
		},
	}

	step := types.WorkflowStep{
		Name:            "implement",
		CommandFilePath: "test-implement.txt",
		Retries:         2,
		IsReview:        false,
		Order:           0,
	}

	logger := &Logger{OutputFormat: output.FormatText}

	// Mock agent
	mockBinDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0o755); err != nil {
		t.Fatalf("Failed to create mock bin dir: %v", err)
	}
	mockScript := filepath.Join(mockBinDir, "claude")
	scriptContent := `#!/bin/bash
exit 0
`
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("Failed to write mock script: %v", err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	err := ExecuteStepWithRetry(cfg, step, 1, logger)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestExecuteStepWithRetryFailExhausted(t *testing.T) {
	tempDir := t.TempDir()
	sessionRoot := tempDir
	sessionID := "00000000-0000-4000-8000-000000000301"

	// Create FAIL report
	reportContent := `command: "test command"
artifact: "test artifact"
timestamp: "2026-01-01T00:00:00Z"
status: FAIL
summary: "Test failed"
issues:
  blockers:
    - message: "Test error"
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	reportPath := filepath.Join(sessionRoot, sessionID, "report.yaml")
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatalf("Failed to create report dir: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         sessionRoot,
		Agent:               "claude",
		MaxImplementRetries: 1,
		OutputFormat:        output.FormatText,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "test-implement.txt",
		},
	}

	step := types.WorkflowStep{
		Name:            "implement",
		CommandFilePath: "test-implement.txt",
		Retries:         2,
		IsReview:        false,
		Order:           0,
	}

	logger := &Logger{OutputFormat: output.FormatText}

	// Mock agent
	mockBinDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0o755); err != nil {
		t.Fatalf("Failed to create mock bin dir: %v", err)
	}
	mockScript := filepath.Join(mockBinDir, "claude")
	scriptContent := `#!/bin/bash
exit 0
`
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("Failed to write mock script: %v", err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	// Write FAIL report before execution
	if err := os.WriteFile(reportPath, []byte(reportContent), 0o600); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	err := ExecuteStepWithRetry(cfg, step, 1, logger)
	// Should not error, just continue to next step
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestExecuteStepWithRetryAgentFailure(t *testing.T) {
	tempDir := t.TempDir()
	sessionRoot := tempDir
	sessionID := "00000000-0000-4000-8000-000000000302"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         sessionRoot,
		Agent:               "nonexistent-agent",
		MaxImplementRetries: 1,
		OutputFormat:        output.FormatText,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "test-implement.txt",
		},
	}

	step := types.WorkflowStep{
		Name:            "implement",
		CommandFilePath: "test-implement.txt",
		Retries:         1,
		IsReview:        false,
		Order:           0,
	}

	logger := &Logger{OutputFormat: output.FormatText}

	err := ExecuteStepWithRetry(cfg, step, 1, logger)
	// Should not error, continues to next step after exhausting retries
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestExecuteStepWithRetryInfiniteRetries(t *testing.T) {
	tempDir := t.TempDir()
	sessionRoot := tempDir
	sessionID := "00000000-0000-4000-8000-000000000303"

	// Create PASS report
	reportContent := `command: "test command"
artifact: "test artifact"
timestamp: "2026-01-01T00:00:00Z"
status: PASS
summary: "Test passed"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	reportPath := filepath.Join(sessionRoot, sessionID, "report.yaml")
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatalf("Failed to create report dir: %v", err)
	}
	if err := os.WriteFile(reportPath, []byte(reportContent), 0o600); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         sessionRoot,
		Agent:               "claude",
		MaxImplementRetries: 2,
		OutputFormat:        output.FormatText,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "test-implement.txt",
		},
	}

	step := types.WorkflowStep{
		Name:            "implement",
		CommandFilePath: "test-implement.txt",
		Retries:         0, // Infinite retries (test maxRetries == 0 branch)
		IsReview:        false,
		Order:           0,
	}

	logger := &Logger{OutputFormat: output.FormatText}

	// Mock agent
	mockBinDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0o755); err != nil {
		t.Fatalf("Failed to create mock bin dir: %v", err)
	}
	mockScript := filepath.Join(mockBinDir, "claude")
	scriptContent := `#!/bin/bash
exit 0
`
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("Failed to write mock script: %v", err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	err := ExecuteStepWithRetry(cfg, step, 1, logger)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestExecuteStepWithRetryJSONOutput(t *testing.T) {
	tempDir := t.TempDir()
	sessionRoot := tempDir
	sessionID := "00000000-0000-4000-8000-000000000304"

	// Create PASS report
	reportContent := `command: "test command"
artifact: "test artifact"
timestamp: "2026-01-01T00:00:00Z"
status: PASS
summary: "Test passed"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	reportPath := filepath.Join(sessionRoot, sessionID, "report.yaml")
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatalf("Failed to create report dir: %v", err)
	}
	if err := os.WriteFile(reportPath, []byte(reportContent), 0o600); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         sessionRoot,
		Agent:               "claude",
		MaxImplementRetries: 2,
		OutputFormat:        output.FormatJSON, // Test JSON format
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "test-implement.txt",
		},
	}

	step := types.WorkflowStep{
		Name:            "implement",
		CommandFilePath: "test-implement.txt",
		Retries:         1,
		IsReview:        false,
		Order:           0,
	}

	logger := &Logger{OutputFormat: output.FormatJSON}

	// Mock agent
	mockBinDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0o755); err != nil {
		t.Fatalf("Failed to create mock bin dir: %v", err)
	}
	mockScript := filepath.Join(mockBinDir, "claude")
	scriptContent := `#!/bin/bash
exit 0
`
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("Failed to write mock script: %v", err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	err := ExecuteStepWithRetry(cfg, step, 1, logger)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
