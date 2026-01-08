package command

import (
	"fluxid-cli/internal/output"
	"testing"
)

func TestPrintInitializationStatus_TextFormat(t *testing.T) {
	t.Parallel()
	status := output.InitializationStatus{
		Version:             "test-version",
		SessionID:           "test-session",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		MaxCommitRetries:    100,
		TaskFile:            "",
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	exitCode := printInitializationStatus(status, output.FormatText)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for text format, got %d", exitCode)
	}
}

func TestPrintInitializationStatus_JSONFormat(t *testing.T) {
	t.Parallel()
	status := output.InitializationStatus{
		Version:             "test-version",
		SessionID:           "test-session",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		MaxCommitRetries:    100,
		TaskFile:            "",
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	exitCode := printInitializationStatus(status, output.FormatJSON)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for JSON format, got %d", exitCode)
	}
}

func TestPrintInitializationStatus_YAMLFormat(t *testing.T) {
	t.Parallel()
	status := output.InitializationStatus{
		Version:             "test-version",
		SessionID:           "test-session",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		MaxCommitRetries:    100,
		TaskFile:            "",
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	exitCode := printInitializationStatus(status, output.FormatYAML)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for YAML format, got %d", exitCode)
	}
}

func TestPrintInitializationStatus_DefaultFormat(t *testing.T) {
	t.Parallel()
	status := output.InitializationStatus{
		Version:             "test-version",
		SessionID:           "test-session",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		MaxCommitRetries:    100,
		TaskFile:            "",
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	// Test with an invalid/unknown format to trigger the default case
	exitCode := printInitializationStatus(status, output.Format("unknown"))
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for default format case, got %d", exitCode)
	}
}
