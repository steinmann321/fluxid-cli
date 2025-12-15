//nolint:paralleltest // Output tests with stdout capture
package main

import (
	"bytes"
	"fluxid-loop/internal/config"
	"os"
	"strings"
	"testing"
)

func TestPrintInitializationStatusText_AllFields(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	cfg := Config{
		SessionID:           "text-session-123",
		Agent:               "claude",
		AgentArgs:           []string{"--arg1", "value1"},
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		CommitEnabled:       true,
		DryRun:              false,
		OutputFormat:        OutputFormatText,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/path/to/implement.md",
			ReviewPath:    "/path/to/review.md",
			CommitPath:    "/path/to/commit.md",
		},
		Sources: map[string]string{
			"agent":             "cli",
			"iterations":        "config",
			"implement_retries": "env",
			"commit_enabled":    "default",
		},
	}

	// Print status
	PrintInitializationStatusText(cfg)

	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Verify all expected fields are present
	expectedStrings := []string{
		"=== fluxid Workflow Initialization ===",
		"Agent: claude (source: cli)",
		"Session ID: text-session-123",
		"Max Review Cycles: 10 (source: config)",
		"Max Implement Retries: 3 (source: env)",
		"Commit Enabled: true (source: default)",
		"Command Files:",
		"Implement: /path/to/implement.md",
		"Review: /path/to/review.md",
		"Commit: /path/to/commit.md",
		"Agent Args: [--arg1 value1]",
		"======================================",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s", expected, output)
		}
	}
}

func TestPrintInitializationStatusText_MinimalFields(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	cfg := Config{
		SessionID:           "text-session-minimal",
		Agent:               "claude",
		AgentArgs:           nil, // No agent args
		MaxReviewCycles:     5,
		MaxImplementRetries: 2,
		CommitEnabled:       false,
		DryRun:              false,
		OutputFormat:        OutputFormatText,
		CommandFiles:        nil, // No command files
		Sources: map[string]string{
			"agent":             "default",
			"iterations":        "default",
			"implement_retries": "default",
			"commit_enabled":    "default",
		},
	}

	// Print status
	PrintInitializationStatusText(cfg)

	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Verify required fields are present
	expectedStrings := []string{
		"=== fluxid Workflow Initialization ===",
		"Agent: claude (source: default)",
		"Session ID: text-session-minimal",
		"Max Review Cycles: 5 (source: default)",
		"Max Implement Retries: 2 (source: default)",
		"Commit Enabled: false (source: default)",
		"======================================",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s", expected, output)
		}
	}

	// Verify optional fields are NOT present
	unexpectedStrings := []string{
		"Command Files:",
		"Agent Args:",
	}

	for _, unexpected := range unexpectedStrings {
		if strings.Contains(output, unexpected) {
			t.Errorf("Expected output to NOT contain %q, but it did.\nFull output:\n%s", unexpected, output)
		}
	}
}

func TestPrintInitializationStatusText_EmptyAgentArgs(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	cfg := Config{
		SessionID:           "text-session-empty-args",
		Agent:               "claude",
		AgentArgs:           []string{}, // Empty slice (not nil)
		MaxReviewCycles:     5,
		MaxImplementRetries: 2,
		CommitEnabled:       true,
		DryRun:              false,
		OutputFormat:        OutputFormatText,
		CommandFiles:        nil,
		Sources: map[string]string{
			"agent":             "default",
			"iterations":        "default",
			"implement_retries": "default",
			"commit_enabled":    "default",
		},
	}

	// Print status
	PrintInitializationStatusText(cfg)

	// Close writer and restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Empty agent args should not be displayed
	if strings.Contains(output, "Agent Args:") {
		t.Errorf("Expected output to NOT contain 'Agent Args:' for empty slice, but it did.\nFull output:\n%s", output)
	}
}
