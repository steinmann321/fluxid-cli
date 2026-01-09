//nolint:paralleltest // Simulation tests
package workflow

import (
	"bytes"
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"log"
	"strings"
	"testing"
)

func TestRunSimulation(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	cfg := types.Config{
		SessionID:           "00000000-0000-4000-8000-000000000005",
		SessionRoot:         "",
		Agent:               "claude",
		MaxReviewCycles:     3,
		MaxImplementRetries: 2,
		MaxCommitRetries:    100,
		DryRun:              true,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	exitCode := RunSimulation(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := buf.String()

	// Verify simulation output contains expected sections
	expectedStrings := []string{
		"Simulation Plan",
		"Review Cycle 1/3",
		"Implement attempt 1/2",
		"Phase: implement",
		"Phase: commit",
		"Phase: review",
		"Synthetic implement report: PASS",
		"Synthetic review report: PASS",
		"Simulated workflow completed successfully",
		"End Simulation",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected simulation output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
		}
	}
}

// TestRunSimulation_NoCommit removed - commit phase is always enabled in v2.0

func TestGetCommandFilePath_NoCommandFiles(t *testing.T) {
	cfg := types.Config{
		Agent:               "",
		AgentArgs:           nil,
		SessionID:           "",
		SessionRoot:         "",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	result := getCommandFilePath(cfg, "implement")
	if result != builtInPrompt {
		t.Errorf("Expected %q, got %q", builtInPrompt, result)
	}
}

func TestGetCommandFilePath_WithCommandFiles(t *testing.T) {
	cfg := types.Config{
		Agent:               "",
		AgentArgs:           nil,
		SessionID:           "",
		SessionRoot:         "",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/path/to/implement.md",
			ReviewPath:    "/path/to/review.md",
			CommitPath:    "/path/to/commit.md",
		},
		OutputFormat: output.FormatText,
		TaskFilePath: "",
	}

	tests := []struct {
		phase    string
		expected string
	}{
		{"implement", "/path/to/implement.md"},
		{"review", "/path/to/review.md"},
		{"commit", "/path/to/commit.md"},
		{"unknown", builtInPrompt},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			result := getCommandFilePath(cfg, tt.phase)
			if result != tt.expected {
				t.Errorf("For phase %q, expected %q, got %q", tt.phase, tt.expected, result)
			}
		})
	}
}

func TestGetCommandFilePath_PartialCommandFiles(t *testing.T) {
	cfg := types.Config{
		Agent:               "",
		AgentArgs:           nil,
		SessionID:           "",
		SessionRoot:         "",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/path/to/implement.md",
			ReviewPath:    "",
			CommitPath:    "",
		},
		OutputFormat: output.FormatText,
		TaskFilePath: "",
	}

	// Test that we fall back to built-in prompt for missing command files
	result := getCommandFilePath(cfg, "review")
	if result != builtInPrompt {
		t.Errorf("Expected %q for missing review file, got %q", builtInPrompt, result)
	}

	result = getCommandFilePath(cfg, "commit")
	if result != builtInPrompt {
		t.Errorf("Expected %q for missing commit file, got %q", builtInPrompt, result)
	}

	// But implement should return the configured path
	result = getCommandFilePath(cfg, "implement")
	if result != "/path/to/implement.md" {
		t.Errorf("Expected '/path/to/implement.md', got %q", result)
	}
}
