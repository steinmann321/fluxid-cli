//nolint:paralleltest // Workflow tests with subprocess execution
package workflow

import (
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"strings"
	"testing"
)

func TestRunCommitPhase(t *testing.T) {
	// This test verifies runCommitPhase calls runPhase correctly
	// We can't easily test the full flow without mocking, so we test error handling
	cfg := types.Config{
		SessionID:           "test-session",
		Agent:               "nonexistent-agent",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runCommitPhase(cfg)
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	if !strings.Contains(err.Error(), "commit phase failed") {
		t.Errorf("Expected error message about commit phase, got: %v", err)
	}
}

func TestRunReviewPhase(t *testing.T) {
	// This test verifies runReviewPhase calls runPhase correctly
	// We can't easily test the full flow without mocking, so we test error handling
	cfg := types.Config{
		SessionID:           "test-session",
		Agent:               "nonexistent-agent",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	if status != "" {
		t.Errorf("Expected empty status on error, got: %s", status)
	}
	if !strings.Contains(err.Error(), "review phase failed") {
		t.Errorf("Expected error message about review phase, got: %v", err)
	}
}

func TestRunPhase_ExitCodeExtraction(t *testing.T) {
	// Test that runPhase correctly extracts exit codes from failed commands
	// We use false command which always exits with code 1
	cfg := types.Config{
		SessionID:           "test-exit-code-session",
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runPhase(cfg, "test-phase", "test prompt")
	if err == nil {
		t.Error("Expected error for failing command")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("Expected failure message, got: %v", err)
	}
}

func TestRunPhase_Success(t *testing.T) {
	// Test that runPhase returns 0 for successful commands
	// We use true command which always exits with code 0
	cfg := types.Config{
		SessionID:           "test-success-session",
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runPhase(cfg, "test-phase", "test prompt")
	if err != nil {
		t.Errorf("Expected no error for successful command, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunCommitPhase_CommitDisabled(t *testing.T) {
	// Test that commit phase works when enabled
	cfg := types.Config{
		SessionID:           "test-commit-disabled",
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runCommitPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error for successful commit phase, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunReviewPhase_NonZeroExitCode(t *testing.T) {
	// Test that review phase fails on non-zero exit code
	cfg := types.Config{
		SessionID:           "test-review-nonzero",
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	status, exitCode, err := runReviewPhase(cfg)
	if err == nil {
		t.Error("Expected error for non-zero exit code")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	if status != "" {
		t.Errorf("Expected empty status on error, got: %s", status)
	}
}

func TestRunCommitPhase_Success(t *testing.T) {
	// Test successful commit phase
	cfg := types.Config{
		SessionID:           "test-commit-success",
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runCommitPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
