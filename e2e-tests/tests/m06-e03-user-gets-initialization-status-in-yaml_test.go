package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	yamlSourceTypeCLI = "cli"
	yamlAgentClaude   = "claude"
)

// InitializationStatusYAML represents the YAML output structure.
// v2.0: source tracking removed (Phase 9), commit_enabled removed (Phase 10).
type InitializationStatusYAML struct {
	SessionID           string            `yaml:"session_id"`
	Agent               string            `yaml:"agent"`
	MaxReviewCycles     int               `yaml:"max_review_cycles"`
	MaxImplementRetries int               `yaml:"max_implement_retries"`
	CommandFiles        *CommandFilesYAML `yaml:"command_files,omitempty"`
	AgentArgs           []string          `yaml:"agent_args,omitempty"`
}

// CommandFilesYAML represents command file paths in YAML output.
type CommandFilesYAML struct {
	Implement string `yaml:"implement"`
	Review    string `yaml:"review"`
	Commit    string `yaml:"commit"`
}

// TestM06E03YAMLOutputBasic validates basic YAML output functionality.
func TestM06E03YAMLOutputBasic(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-output=yaml", "--fluxid-dry-run", "--claude")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-output yaml failed: %v\nStdout:\n%s\nStderr:\n%s",
			err, stdout.String(), stderr.String())
	}

	output := stdout.String()

	// Verify YAML is valid
	var status InitializationStatusYAML
	if err := yaml.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("Failed to parse YAML output: %v\nOutput:\n%s", err, output)
	}

	// Verify required fields are present
	if status.SessionID == "" {
		t.Errorf("Expected session_id to be non-empty, got: %q", status.SessionID)
	}
	if status.Agent != "claude" {
		t.Errorf("Expected agent to be 'claude', got: %q", status.Agent)
	}
	if status.MaxReviewCycles <= 0 {
		t.Errorf("Expected max_review_cycles to be positive, got: %d", status.MaxReviewCycles)
	}
	if status.MaxImplementRetries <= 0 {
		t.Errorf("Expected max_implement_retries to be positive, got: %d", status.MaxImplementRetries)
	}
}

// TestM06E03YAMLOutputWithConfig validates YAML output with configuration values.
func TestM06E03YAMLOutputWithConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithOutputFormat(t, root, "yaml",
		"--fluxid-iterations=5",
		"--fluxid-implement-retries=2")

	var status InitializationStatusYAML
	if err := yaml.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("Failed to parse YAML output: %v\nOutput:\n%s", err, output)
	}

	// v2.0: source tracking removed (Phase 9)
	// Verify configured values using helper
	verifyConfigValues(t,
		status.MaxReviewCycles, status.MaxImplementRetries,
		5, 2)
}

// TestM06E03DefaultFormatIsText validates that default output format is text when flag omitted.
func TestM06E03DefaultFormatIsText(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidDefaultFormat(t, root)

	verifyDefaultTextFormat(t, output, func(output string) error {
		var status InitializationStatusYAML
		return yaml.Unmarshal([]byte(output), &status)
	}, "YAML")
}

// TestM06E03UnknownFormatRejected validates that unknown format values are rejected.
//
//nolint:dupl // Test functions intentionally similar for different error cases
func TestM06E03UnknownFormatRejected(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-output=toml", "--claude")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Verify command exits with error
	if err == nil {
		t.Fatalf("Expected command to fail with unknown format, but it succeeded.\nStdout:\n%s\nStderr:\n%s",
			stdout.String(), stderr.String())
	}

	stderrOutput := stderr.String()

	// Verify error message mentions unsupported format
	if !strings.Contains(stderrOutput, "unsupported output format") {
		t.Errorf("Expected error message about unsupported format, got:\n%s", stderrOutput)
	}

	// Verify error message mentions 'toml'
	if !strings.Contains(stderrOutput, "toml") {
		t.Errorf("Expected error message to mention 'toml', got:\n%s", stderrOutput)
	}
}

// TestM06E03YAMLWithDryRun validates YAML output works with dry-run mode.
func TestM06E03YAMLWithDryRun(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithOutputFormat(t, root, "yaml")

	// Verify YAML is valid
	var status InitializationStatusYAML
	if err := yaml.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("Failed to parse YAML output: %v\nOutput:\n%s", err, output)
	}

	// Verify dry-run header does NOT appear in YAML mode
	if strings.Contains(output, "=== DRY RUN MODE") {
		t.Errorf("Expected no dry-run text header in YAML output, got:\n%s", output)
	}
}

// TestM06E03YAMLOutputStructure validates the complete YAML structure.
func TestM06E03YAMLOutputStructure(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithOutputFormat(t, root, "yaml", "arg1", "arg2")

	var status InitializationStatusYAML
	if err := yaml.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("Failed to parse YAML output: %v\nOutput:\n%s", err, output)
	}

	// v2.0: source tracking removed (Phase 9)
	// Verify agent args using helper
	verifyAgentArgsAndSource(t, status.AgentArgs, status.Agent, yamlAgentClaude)
}
