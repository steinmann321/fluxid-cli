package workflow

import (
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunImplementPhase_AbortDuringImplement(t *testing.T) {
	sessionID := "test-abort-implement-" + time.Now().Format("20060102150405.000000")
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatal(err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentEcho,
		MaxReviewCycles:     1,
		MaxImplementRetries: 2,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
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
	sessionID := "test-retries-" + time.Now().Format("20060102150405.000000")
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentTrue,
		MaxReviewCycles:     1,
		MaxImplementRetries: 3,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	// Pre-write PASS report before workflow starts
	// No timing dependencies - completely deterministic
	if err := ipc.WriteReport(sessionID, testPassReport); err != nil {
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

func TestRunImplementPhase_AllRetriesFail(t *testing.T) {
	sessionID := "test-all-fail-" + time.Now().Format("20060102150405.000000")
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               testAgentTrue,
		MaxReviewCycles:     1,
		MaxImplementRetries: 2,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	// Pre-write FAIL report before workflow starts
	// No timing dependencies - completely deterministic
	if err := ipc.WriteReport(sessionID, testFailReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error when retries exhausted (should continue), got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (continue to next phase), got %d", exitCode)
	}
}

func TestRunImplementPhase_AgentFailure(t *testing.T) {
	sessionID := "test-agent-fail-" + time.Now().Format("20060102150405.000000")
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error for failing agent")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestRunImplementPhase_NonexistentAgent(t *testing.T) {
	sessionID := "test-nonexistent-" + time.Now().Format("20060102150405.000000")
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "nonexistent-agent-12345",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestRunCommitPhase_Disabled(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)

	cfg := types.Config{
		SessionID:           "test-commit-disabled",
		Agent:               testAgentTrue,
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
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
		Agent:               testAgentTrue,
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
	}

	exitCode, err := runCommitPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error for commit phase, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for commit phase, got %d", exitCode)
	}
}
