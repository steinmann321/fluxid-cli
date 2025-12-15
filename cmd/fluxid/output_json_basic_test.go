//nolint:paralleltest // Output tests with log capture
package main

import (
	"encoding/json"
	"fluxid-loop/internal/config"
	"os"
	"testing"
)

const testAgentName = "claude"

func TestPrintInitializationStatusJSONBasic(t *testing.T) {
	// Create a temporary file to capture stdout
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-output")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Redirect stdout
	os.Stdout = tmpFile

	cfg := Config{
		SessionID:           "test-json-session-123",
		Agent:               testAgentName,
		AgentArgs:           []string{},
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
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

	// Parse the JSON to verify it's valid
	var status InitializationStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		t.Errorf("Output is not valid JSON: %v\nOutput:\n%s", err, output)
	}

	// Verify the contents
	if status.SessionID != "test-json-session-123" {
		t.Errorf("Expected SessionID 'test-json-session-123', got '%s'", status.SessionID)
	}

	if status.Agent != testAgentName {
		t.Errorf("Expected Agent '%s', got '%s'", testAgentName, status.Agent)
	}

	if status.MaxReviewCycles != 10 {
		t.Errorf("Expected MaxReviewCycles 10, got %d", status.MaxReviewCycles)
	}

	if status.MaxImplementRetries != 3 {
		t.Errorf("Expected MaxImplementRetries 3, got %d", status.MaxImplementRetries)
	}

	if status.CommitEnabled != true {
		t.Errorf("Expected CommitEnabled true, got %v", status.CommitEnabled)
	}
}

func TestPrintInitializationStatusJSONWithCommandFiles(t *testing.T) {
	// Create a temporary file to capture stdout
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-output-cmd")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Redirect stdout
	os.Stdout = tmpFile

	cfg := Config{
		SessionID:           "test-with-files-123",
		Agent:               testAgentName,
		AgentArgs:           []string{},
		MaxReviewCycles:     5,
		MaxImplementRetries: 2,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/path/to/implement.sh",
			ReviewPath:    "/path/to/review.sh",
			CommitPath:    "/path/to/commit.sh",
		},
		OutputFormat: OutputFormatJSON,
		Sources: map[string]string{
			"agent":             "cli",
			"iterations":        "project",
			"implement_retries": "home",
			"commit_enabled":    "env",
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

	// Verify command files are included
	if status.CommandFiles == nil {
		t.Error("Expected CommandFiles to be set, got nil")
	} else {
		if status.CommandFiles.Implement != "/path/to/implement.sh" {
			t.Errorf("Expected Implement path '/path/to/implement.sh', got '%s'", status.CommandFiles.Implement)
		}

		if status.CommandFiles.Review != "/path/to/review.sh" {
			t.Errorf("Expected Review path '/path/to/review.sh', got '%s'", status.CommandFiles.Review)
		}

		if status.CommandFiles.Commit != "/path/to/commit.sh" {
			t.Errorf("Expected Commit path '/path/to/commit.sh', got '%s'", status.CommandFiles.Commit)
		}
	}

	// Verify sources are included
	if status.AgentSource != "cli" {
		t.Errorf("Expected AgentSource 'cli', got '%s'", status.AgentSource)
	}

	if status.ReviewCyclesSource != "project" {
		t.Errorf("Expected ReviewCyclesSource 'project', got '%s'", status.ReviewCyclesSource)
	}

	if status.ImplementRetriesSource != "home" {
		t.Errorf("Expected ImplementRetriesSource 'home', got '%s'", status.ImplementRetriesSource)
	}

	if status.CommitEnabledSource != "env" {
		t.Errorf("Expected CommitEnabledSource 'env', got '%s'", status.CommitEnabledSource)
	}
}

func TestPrintInitializationStatusJSONWithAgentArgs(t *testing.T) {
	// Create a temporary file to capture stdout
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-output-args")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Redirect stdout
	os.Stdout = tmpFile

	agentArgs := []string{"--model", "gpt-4", "--temperature", "0.7"}

	cfg := Config{
		SessionID:           "test-agent-args",
		Agent:               testAgentName,
		AgentArgs:           agentArgs,
		MaxReviewCycles:     3,
		MaxImplementRetries: 1,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        OutputFormatJSON,
		Sources: map[string]string{
			"agent":             "cli",
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

	// Parse the JSON to verify it's valid
	var status InitializationStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Verify agent args are included
	if len(status.AgentArgs) != len(agentArgs) {
		t.Errorf("Expected %d agent args, got %d", len(agentArgs), len(status.AgentArgs))
	}

	for i, arg := range agentArgs {
		if status.AgentArgs[i] != arg {
			t.Errorf("Expected agent arg[%d] '%s', got '%s'", i, arg, status.AgentArgs[i])
		}
	}
}

func TestPrintInitializationStatusJSONNoAgentArgs(t *testing.T) {
	// Create a temporary file to capture stdout
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-output-no-args")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Redirect stdout
	os.Stdout = tmpFile

	cfg := Config{
		SessionID:           "test-no-args",
		Agent:               testAgentName,
		AgentArgs:           []string{}, // Empty args
		MaxReviewCycles:     2,
		MaxImplementRetries: 1,
		CommitEnabled:       false,
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

	// Parse the JSON to verify it's valid
	var status InitializationStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Verify agent args is nil (not present in JSON when empty)
	if status.AgentArgs != nil {
		t.Errorf("Expected AgentArgs to be nil when empty, got %v", status.AgentArgs)
	}
}
