package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	sourceTypeCLI = "cli"
	agentClaude   = "claude"
)

// InitializationStatusJSON represents the JSON output structure.
// v2.0: source tracking removed (Phase 9), commit_enabled removed (Phase 10).
type InitializationStatusJSON struct {
	SessionID           string            `json:"session_id"`
	Agent               string            `json:"agent"`
	MaxReviewCycles     int               `json:"max_review_cycles"`
	MaxImplementRetries int               `json:"max_implement_retries"`
	CommandFiles        *CommandFilesJSON `json:"command_files,omitempty"`
	AgentArgs           []string          `json:"agent_args,omitempty"`
}

// CommandFilesJSON represents command file paths in JSON output.
type CommandFilesJSON struct {
	Implement string `json:"implement"`
	Review    string `json:"review"`
	Commit    string `json:"commit"`
}

// TestM06E02JSONOutputBasic validates basic JSON output functionality.
func TestM06E02JSONOutputBasic(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-output=json", "--fluxid-dry-run", "--claude")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --fluxid-output json failed: %v\nStdout:\n%s\nStderr:\n%s",
			err, stdout.String(), stderr.String())
	}

	output := stdout.String()

	// Verify JSON is valid
	var status InitializationStatusJSON
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput:\n%s", err, output)
	}

	// v2.0: source tracking removed (Phase 9), commit_enabled removed (Phase 10)
	// Verify required fields are present
	if status.SessionID == "" {
		t.Errorf("Expected session_id to be non-empty, got: %q", status.SessionID)
	}
	if status.Agent != agentClaude {
		t.Errorf("Expected agent to be 'claude', got: %q", status.Agent)
	}
	if status.MaxReviewCycles <= 0 {
		t.Errorf("Expected max_review_cycles to be positive, got: %d", status.MaxReviewCycles)
	}
	if status.MaxImplementRetries <= 0 {
		t.Errorf("Expected max_implement_retries to be positive, got: %d", status.MaxImplementRetries)
	}
}

// TestM06E02JSONOutputWithConfig validates JSON output with configuration values.
func TestM06E02JSONOutputWithConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithOutputFormat(t, root, "json",
		"--fluxid-iterations=5",
		"--fluxid-implement-retries=2")

	var status InitializationStatusJSON
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput:\n%s", err, output)
	}

	// v2.0: source tracking removed (Phase 9)
	// Verify configured values using helper
	verifyConfigValues(t,
		status.MaxReviewCycles, status.MaxImplementRetries,
		5, 2)
}

// TestM06E02DefaultFormatIsText validates that default output format is text.
func TestM06E02DefaultFormatIsText(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidDefaultFormat(t, root)

	verifyDefaultTextFormat(t, output, func(output string) error {
		var status InitializationStatusJSON
		return json.Unmarshal([]byte(output), &status)
	}, "JSON")
}

// TestM06E02UnknownFormatRejected validates that unknown format values are rejected.
//
//nolint:dupl // Test functions intentionally similar for different error cases
func TestM06E02UnknownFormatRejected(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-output=xml", "--claude")
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

	// Verify error message mentions 'xml'
	if !strings.Contains(stderrOutput, "xml") {
		t.Errorf("Expected error message to mention 'xml', got:\n%s", stderrOutput)
	}
}

// TestM06E02JSONWithDryRun validates JSON output works with dry-run mode.
func TestM06E02JSONWithDryRun(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithOutputFormat(t, root, "json")

	// Verify JSON is valid
	var status InitializationStatusJSON
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput:\n%s", err, output)
	}

	// Verify dry-run header does NOT appear in JSON mode
	if strings.Contains(output, "=== DRY RUN MODE") {
		t.Errorf("Expected no dry-run text header in JSON output, got:\n%s", output)
	}
}

// TestM06E02JSONOutputStructure validates the complete JSON structure.
func TestM06E02JSONOutputStructure(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	output := runFluxidWithOutputFormat(t, root, "json", "arg1", "arg2")

	var status InitializationStatusJSON
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput:\n%s", err, output)
	}

	// v2.0: source tracking removed (Phase 9)
	// Verify agent args using helper
	verifyAgentArgsAndSource(t, status.AgentArgs, status.Agent, agentClaude)
}
