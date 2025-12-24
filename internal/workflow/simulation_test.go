//nolint:paralleltest // Simulation tests
package workflow

import (
	"bytes"
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"log"
	"strings"
	"testing"
)

const builtInPrompt = "built-in prompt"

func TestRunSimulation(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	cfg := types.Config{
		SessionID:           "test-sim-session",
		Agent:               "claude",
		MaxReviewCycles:     3,
		MaxImplementRetries: 2,
		CommitEnabled:       true,
		DryRun:              true,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
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

func TestRunSimulation_NoCommit(t *testing.T) {
	// Test simulation without commit phase
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	cfg := types.Config{
		SessionID:           "test-sim-no-commit",
		Agent:               "claude",
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
		DryRun:              true,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	exitCode := RunSimulation(cfg)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := buf.String()

	// Verify simulation output doesn't contain commit phase
	if strings.Contains(output, "Phase: commit") {
		t.Errorf("Expected simulation without commit phase, but found commit in output:\n%s", output)
	}

	// But should contain implement and review
	if !strings.Contains(output, "Phase: implement") {
		t.Error("Expected simulation to contain implement phase")
	}
	if !strings.Contains(output, "Phase: review") {
		t.Error("Expected simulation to contain review phase")
	}

	// Should contain synthetic PASS reports
	if !strings.Contains(output, "Synthetic implement report: PASS") {
		t.Error("Expected simulation to contain synthetic implement PASS report")
	}
	if !strings.Contains(output, "Synthetic review report: PASS") {
		t.Error("Expected simulation to contain synthetic review PASS report")
	}

	// Should show completion message
	if !strings.Contains(output, "Simulated workflow completed successfully") {
		t.Error("Expected simulation to show completion message")
	}
}

func TestGetCommandFilePath_NoCommandFiles(t *testing.T) {
	cfg := types.Config{
		Agent:               "",
		AgentArgs:           nil,
		SessionID:           "",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
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
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/path/to/implement.md",
			ReviewPath:    "/path/to/review.md",
			CommitPath:    "/path/to/commit.md",
		},
		OutputFormat: output.FormatText,
		Sources:      map[string]string{},
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
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/path/to/implement.md",
			ReviewPath:    "",
			CommitPath:    "",
		},
		OutputFormat: output.FormatText,
		Sources:      map[string]string{},
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
