//nolint:exhaustruct,funlen // Test file with test fixtures
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/types"
	"os"
	"path/filepath"
	"testing"
)

func TestRunConfigDrivenWorkflowSuccess(t *testing.T) {
	tempDir := t.TempDir()
	sessionRoot := tempDir
	sessionID := "00000000-0000-4000-8000-000000000200"

	// Create test report file
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
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		OutputFormat:        output.FormatText,
		Agent:               "claude",
		Workflow: &types.Workflow{
			MaxIterations:    2,
			CurrentIteration: 0,
			Steps: []types.WorkflowStep{
				{
					Name:            "implement",
					CommandFilePath: "test-implement.txt",
					Retries:         1,
					IsReview:        false,
					Order:           0,
				},
				{
					Name:            "review",
					CommandFilePath: "test-review.txt",
					Retries:         1,
					IsReview:        true,
					Order:           1,
				},
			},
		},
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "test-implement.txt",
			ReviewPath:    "test-review.txt",
			CommitPath:    "test-commit.txt",
		},
	}

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

	exitCode, err := runConfigDrivenWorkflow(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", exitCode)
	}
}

func TestRunConfigDrivenWorkflowIterationExhaustion(t *testing.T) {
	tempDir := t.TempDir()
	sessionRoot := tempDir
	sessionID := "00000000-0000-4000-8000-000000000201"

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
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		OutputFormat:        output.FormatText,
		Workflow: &types.Workflow{
			MaxIterations:    1,
			CurrentIteration: 0,
			Steps: []types.WorkflowStep{
				{
					Name:            "review",
					CommandFilePath: "test-review.txt",
					Retries:         1,
					IsReview:        true,
					Order:           0,
				},
			},
		},
		CommandFiles: &config.ResolvedCommandFiles{
			ReviewPath: "test-review.txt",
		},
	}

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

	exitCode, err := runConfigDrivenWorkflow(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (iterations exhausted), got: %d", exitCode)
	}

	// Verify report exists
	if _, err := storage.ReadReport(sessionID, sessionRoot); err != nil {
		t.Errorf("Expected report to exist: %v", err)
	}
}

//nolint:paralleltest // Test uses file system operations that should not run in parallel
func TestRunFallbackToLegacy(t *testing.T) {
	tempDir := t.TempDir()
	sessionRoot := tempDir
	sessionID := "00000000-0000-4000-8000-000000000300"

	// Create test report file for legacy workflow
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
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    1,
		OutputFormat:        output.FormatText,
		Agent:               "claude",
		Workflow:            nil, // No workflow config, should trigger legacy fallback
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "echo 'implement'",
			ReviewPath:    "echo 'review'",
			CommitPath:    "echo 'commit'",
		},
		DryRun: false,
	}

	exitCode, err := Run(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", exitCode)
	}
}
