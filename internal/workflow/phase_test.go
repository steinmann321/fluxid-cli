//nolint:paralleltest // Workflow tests with subprocess execution
package workflow

import (
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"strings"
	"testing"
)

func TestRunCommitPhase(t *testing.T) {
	// This test verifies runCommitPhase calls runPhase correctly
	// We can't easily test the full flow without mocking, so we test error handling
	cfg := types.Config{
		SessionID:           "00000000-0000-4000-8000-000000000002",
		SessionRoot:         "",
		Agent:               "nonexistent-agent",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
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
	// This test verifies runReviewPhase handles missing reports correctly
	// With the new quiet behavior, missing reports are treated as FAIL (not errors)
	cfg := types.Config{
		SessionID:           "00000000-0000-4000-8000-000000000002",
		SessionRoot:         "",
		Agent:               "nonexistent-agent",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	status, exitCode, err := runReviewPhase(cfg)

	// With new behavior: missing reports are treated as FAIL, not errors
	// waitForValidReport returns (statusFail, nil) when report can't be read
	if err == nil {
		// This is expected - missing report returns FAIL status, not error
		if status != statusFail {
			t.Errorf("Expected FAIL status for missing report, got: %s", status)
		}
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for FAIL status (not an error), got: %d", exitCode)
		}
	} else {
		t.Errorf("Expected no error (FAIL status instead), got error: %v", err)
	}
}

func TestRunPhase_ExitCodeExtraction(t *testing.T) {
	// Test that runPhase correctly extracts exit codes from failed commands
	// We use false command which always exits with code 1
	cfg := types.Config{
		SessionID:           "00000000-0000-4000-8000-000000000003",
		SessionRoot:         "",
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
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
		SessionID:           "00000000-0000-4000-8000-000000000004",
		SessionRoot:         "",
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
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
		SessionRoot:         "",
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
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
	// Test that review phase handles agent failure (missing report treated as FAIL)
	cfg := types.Config{
		SessionID:           "test-review-nonzero",
		SessionRoot:         "",
		Agent:               "false",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	status, exitCode, err := runReviewPhase(cfg)
	// With new behavior: missing reports are treated as FAIL, not errors
	if err == nil {
		// This is expected - missing report returns FAIL status, not error
		if status != statusFail {
			t.Errorf("Expected FAIL status for missing report, got: %s", status)
		}
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for FAIL status, got: %d", exitCode)
		}
	} else {
		t.Errorf("Expected no error (FAIL status instead), got error: %v", err)
	}
}

func TestRunCommitPhase_Success(t *testing.T) {
	// Test successful commit phase
	cfg := types.Config{
		SessionID:           "test-commit-success",
		SessionRoot:         "",
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	exitCode, err := runCommitPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
