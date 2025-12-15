//nolint:paralleltest // Output tests with log capture
package main

import (
	"fluxid-loop/internal/config"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPrintInitializationStatusYAMLBasic(t *testing.T) {
	// Create a temporary file to capture stdout
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-output")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			t.Errorf("Failed to close temp file: %v", err)
		}
	}()

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Redirect stdout
	os.Stdout = tmpFile

	cfg := Config{
		SessionID:           "test-yaml-session-123",
		Agent:               "claude",
		AgentArgs:           []string{},
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		CommitEnabled:       true,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        OutputFormatYAML,
		Sources: map[string]string{
			"agent":             "default",
			"iterations":        "default",
			"implement_retries": "default",
			"commit_enabled":    "default",
		},
	}

	err = PrintInitializationStatusYAML(cfg)

	// Restore stdout
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("PrintInitializationStatusYAML failed: %v", err)
	}

	// Read output
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	output := string(content)

	// Verify YAML contains expected fields
	if !strings.Contains(output, "session_id: test-yaml-session-123") {
		t.Errorf("Expected YAML to contain session_id, got:\n%s", output)
	}

	if !strings.Contains(output, "agent: claude") {
		t.Errorf("Expected YAML to contain agent, got:\n%s", output)
	}

	if !strings.Contains(output, "max_review_cycles: 10") {
		t.Errorf("Expected YAML to contain max_review_cycles, got:\n%s", output)
	}

	// Verify it's valid YAML by parsing
	var result map[string]interface{}
	if err := yaml.Unmarshal(content, &result); err != nil {
		t.Errorf("Output is not valid YAML: %v\nOutput:\n%s", err, output)
	}
}

func TestPrintInitializationStatusYAMLWithCommandFiles(t *testing.T) {
	// Create a temporary file to capture stdout
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-output")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			t.Errorf("Failed to close temp file: %v", err)
		}
	}()

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Redirect stdout
	os.Stdout = tmpFile

	cfg := Config{
		SessionID:           "test-yaml-with-files-456",
		Agent:               "claude",
		AgentArgs:           []string{"--arg1", "value1"},
		MaxReviewCycles:     5,
		MaxImplementRetries: 2,
		CommitEnabled:       false,
		DryRun:              true,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/path/to/implement.md",
			ReviewPath:    "/path/to/review.md",
			CommitPath:    "/path/to/commit.md",
		},
		OutputFormat: OutputFormatYAML,
		Sources: map[string]string{
			"agent":             "cli",
			"iterations":        "project",
			"implement_retries": "env",
			"commit_enabled":    "home",
		},
	}

	err = PrintInitializationStatusYAML(cfg)

	// Restore stdout
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("PrintInitializationStatusYAML failed: %v", err)
	}

	// Read output
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	output := string(content)

	// Verify command files are present
	if !strings.Contains(output, "command_files:") {
		t.Errorf("Expected YAML to contain command_files, got:\n%s", output)
	}

	if !strings.Contains(output, "/path/to/implement.md") {
		t.Errorf("Expected YAML to contain implement path, got:\n%s", output)
	}

	// Verify agent args are present
	if !strings.Contains(output, "agent_args:") {
		t.Errorf("Expected YAML to contain agent_args, got:\n%s", output)
	}

	// Verify source attributions
	if !strings.Contains(output, "agent_source: cli") {
		t.Errorf("Expected YAML to contain agent_source: cli, got:\n%s", output)
	}

	// Verify it's valid YAML
	var result map[string]interface{}
	if err := yaml.Unmarshal(content, &result); err != nil {
		t.Errorf("Output is not valid YAML: %v\nOutput:\n%s", err, output)
	}

	// Verify structure
	if result["session_id"] != "test-yaml-with-files-456" {
		t.Errorf("Expected session_id to be test-yaml-with-files-456, got: %v", result["session_id"])
	}
}
