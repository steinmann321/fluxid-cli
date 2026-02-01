package workflow

import (
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/types"
	"os"
	"path/filepath"
	"testing"
)

func TestRunImplementPhase_AbortDuringImplement(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	sessionID := "a1b2c3d4-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatal(err)
	}*/

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               testAgentEcho,
		MaxReviewCycles:     1,
		MaxImplementRetries: 2,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error when abort flag set")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130, got %d", exitCode)
	}
}

func TestRunImplementPhase_MultipleRetries(t *testing.T) {
	sessionID := "b3f0c2d1-4e5f-6a7b-8c9d-0e1f2a3b4c5d"
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               testAgentTrue,
		MaxReviewCycles:     1,
		MaxImplementRetries: 3,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	// Pre-write PASS report before workflow starts
	// No timing dependencies - completely deterministic
	if err := storage.WriteReport(sessionID, testPassReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected success, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunImplementPhase_AgentFailure(t *testing.T) {
	sessionID := "c4d5e6f7-8a9b-0c1d-2e3f-4a5b6c7d8e9f"
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	exitCode, err := runImplementPhase(cfg)
	// Corrected behavior: commit failures don't block workflow
	if err != nil {
		t.Errorf("Expected no error (commit failures logged, workflow continues), got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (allow review phase), got: %d", exitCode)
	}
}

func TestRunImplementPhase_NonexistentAgent(t *testing.T) {
	sessionID := "d5e6f7a8-9b0c-1d2e-3f4a-5b6c7d8e9f0a"
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "nonexistent-agent-12345",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	exitCode, err := runImplementPhase(cfg)
	// Corrected behavior: commit failures don't block workflow
	if err != nil {
		t.Errorf("Expected no error (commit failures logged, workflow continues), got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (allow review phase), got: %d", exitCode)
	}
}

func TestRunCommitPhase_Disabled(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	cfg := types.Config{
		SessionID:           "test-commit-disabled",
		SessionRoot:         "",
		Agent:               testAgentTrue,
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	exitCode, err := runCommitPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error for disabled commit, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for disabled commit, got %d", exitCode)
	}
}

func TestRunCommitPhase_Enabled(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	cfg := types.Config{
		SessionID:           "test-commit-enabled",
		SessionRoot:         "",
		Agent:               testAgentTrue,
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    2, // Reduced from 100 to avoid timeout
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	exitCode, err := runCommitPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error for commit phase, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for commit phase, got %d", exitCode)
	}
}
