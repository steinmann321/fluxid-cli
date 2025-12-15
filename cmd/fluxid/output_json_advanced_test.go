//nolint:paralleltest // Output tests with log capture
package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestPrintInitializationStatusJSONOutputIsIndented(t *testing.T) {
	// Create a temporary file to capture stdout
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-output-indent")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Redirect stdout
	os.Stdout = tmpFile

	cfg := Config{
		SessionID:           "test-indent",
		Agent:               "claude",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        OutputFormatJSON,
		Sources: map[string]string{
			"agent":             "default",
			"iterations":        "default",
			"implement_retries": "default",
			"commit_enabled":    "default",
		},
	}

	err = PrintInitializationStatusJSON(cfg)

	// Restore stdout
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("PrintInitializationStatusJSON returned error: %v", err)
	}

	// Read the output
	_ = tmpFile.Close()
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	output := string(data)

	// Verify output is indented (contains newlines and spaces)
	if !strings.Contains(output, "\n") {
		t.Error("Expected JSON output to be indented with newlines")
	}

	// Verify it starts with { and ends with }\n
	trimmed := strings.TrimSpace(output)
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		t.Errorf("Expected JSON output to be wrapped in braces, got: %s", trimmed)
	}
}

func TestPrintInitializationStatusJSONWithComplexSources(t *testing.T) {
	// Create a temporary file to capture stdout
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-output-sources")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Redirect stdout
	os.Stdout = tmpFile

	cfg := Config{
		SessionID:           "test-sources",
		Agent:               "claude",
		AgentArgs:           []string{},
		MaxReviewCycles:     5,
		MaxImplementRetries: 2,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        OutputFormatJSON,
		Sources: map[string]string{
			"agent":             "home (/home/user/.fluxid/config.yaml)",
			"iterations":        "project (./.fluxid/config.yaml)",
			"implement_retries": "env (FLUXID_IMPLEMENT_RETRIES)",
			"commit_enabled":    "cli",
		},
	}

	err = PrintInitializationStatusJSON(cfg)

	// Restore stdout
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("PrintInitializationStatusJSON returned error: %v", err)
	}

	// Read the output
	_ = tmpFile.Close()
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	// Parse the JSON to verify it's valid
	var status InitializationStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Verify sources with paths are correctly included
	if !strings.Contains(status.AgentSource, "/home/user/.fluxid/config.yaml") {
		t.Errorf("Expected AgentSource to contain config path, got '%s'", status.AgentSource)
	}

	if !strings.Contains(status.ReviewCyclesSource, "./.fluxid/config.yaml") {
		t.Errorf("Expected ReviewCyclesSource to contain config path, got '%s'", status.ReviewCyclesSource)
	}
}

func TestPrintInitializationStatusJSONEncoderError(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create a pipe that we can close to force an encoder error
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Redirect stdout to the write end
	os.Stdout = w

	// Close the write end immediately to force an encoding error
	_ = w.Close()

	cfg := Config{
		SessionID:           "test-encoder-error",
		Agent:               "claude",
		AgentArgs:           []string{},
		MaxReviewCycles:     1,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        OutputFormatJSON,
		Sources: map[string]string{
			"agent":             "default",
			"iterations":        "default",
			"implement_retries": "default",
			"commit_enabled":    "default",
		},
	}

	err = PrintInitializationStatusJSON(cfg)

	// Restore stdout
	os.Stdout = oldStdout
	_ = r.Close()

	// Should return an error when encoding fails
	if err == nil {
		t.Error("Expected PrintInitializationStatusJSON to return error when stdout is closed")
	}

	// Verify error message mentions encoding
	if !strings.Contains(err.Error(), "failed to encode") {
		t.Errorf("Expected error message to mention encoding failure, got: %v", err)
	}
}
