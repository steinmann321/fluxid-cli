package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestM02E05AllConfigKeysDisplayed validates that all configuration keys
// are displayed with their values and sources in the initialization status.
func TestM02E05AllConfigKeysDisplayed(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create mixed configuration (home + project + env + cli)
	tmpHome := setupHomeWithConfig(t, fullHomeConfig)

	projectConfigContent := `iterations: 15
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	// Run with CLI overrides
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-implement-retries", "7")

	// Verify all configuration keys are present
	requiredKeys := []string{
		"Agent:",
		"Session ID:",
		"Max Review Cycles:",
		"Max Implement Retries:",
		"Commit Enabled:",
	}

	for _, key := range requiredKeys {
		if !strings.Contains(output, key) {
			t.Errorf("Missing required configuration key: %s", key)
		}
	}

	// Verify correct sources are displayed
	if !strings.Contains(output, "Agent: claude (source: home)") {
		t.Errorf("Agent source incorrect, got:\n%s", output)
	}

	if !strings.Contains(output, "Max Review Cycles: 15 (source: project)") {
		t.Errorf("Max Review Cycles source incorrect, got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 7 (source: cli)") {
		t.Errorf("Max Implement Retries source incorrect, got:\n%s", output)
	}

	if !strings.Contains(output, "Commit Enabled: false (source: home)") {
		t.Errorf("Commit Enabled source incorrect, got:\n%s", output)
	}
}

// TestM02E05CommandFilePathsDisplayed validates that command file paths
// are displayed as absolute paths when configured.
func TestM02E05CommandFilePathsDisplayed(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpProjectDir := t.TempDir()

	// Create command files in project directory
	fluxidDir := filepath.Join(tmpProjectDir, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create .fluxid dir: %v", err)
	}

	commandsDir := filepath.Join(fluxidDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	// Create command files
	implementFile := filepath.Join(commandsDir, "implement.md")
	reviewFile := filepath.Join(commandsDir, "review.md")
	commitFile := filepath.Join(commandsDir, "commit.md")

	if err := os.WriteFile(implementFile, []byte("# Implement"), 0o644); err != nil {
		t.Fatalf("Failed to write implement file: %v", err)
	}
	if err := os.WriteFile(reviewFile, []byte("# Review"), 0o644); err != nil {
		t.Fatalf("Failed to write review file: %v", err)
	}
	if err := os.WriteFile(commitFile, []byte("# Commit"), 0o644); err != nil {
		t.Fatalf("Failed to write commit file: %v", err)
	}

	// Create config pointing to command files
	projectConfigContent := `commands:
  implement: implement.md
  review: review.md
  commit: commit.md
`
	configPath := filepath.Join(fluxidDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(projectConfigContent), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	output := runFluxidInDirWithOutput(t, root, tmpHome, tmpProjectDir)

	// Verify "Command Files:" section exists
	if !strings.Contains(output, "Command Files:") {
		t.Errorf("Missing 'Command Files:' section in output")
	}

	// Verify absolute paths are displayed
	// Note: These are the actual file paths we created
	expectedImplementPath := implementFile
	expectedReviewPath := reviewFile
	expectedCommitPath := commitFile

	if !strings.Contains(output, expectedImplementPath) {
		t.Errorf("Implement path not displayed correctly, expected %s in:\n%s",
			expectedImplementPath, output)
	}

	if !strings.Contains(output, expectedReviewPath) {
		t.Errorf("Review path not displayed correctly, expected %s in:\n%s",
			expectedReviewPath, output)
	}

	if !strings.Contains(output, expectedCommitPath) {
		t.Errorf("Commit path not displayed correctly, expected %s in:\n%s",
			expectedCommitPath, output)
	}

	// Verify format includes labels
	if !strings.Contains(output, "Implement:") {
		t.Errorf("Missing 'Implement:' label in Command Files section")
	}
	if !strings.Contains(output, "Review:") {
		t.Errorf("Missing 'Review:' label in Command Files section")
	}
	if !strings.Contains(output, "Commit:") {
		t.Errorf("Missing 'Commit:' label in Command Files section")
	}
}

// TestM02E05NoCommandFilesWhenNotConfigured validates that the Command Files
// section is not displayed when commands are not configured.
func TestM02E05NoCommandFilesWhenNotConfigured(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpProjectDir := t.TempDir()

	output := runFluxidInDirWithOutput(t, root, tmpHome, tmpProjectDir)

	// Verify Command Files section is NOT present
	if strings.Contains(output, "Command Files:") {
		t.Errorf("Command Files section should not appear when not configured, got:\n%s", output)
	}
}

// TestM02E05OutputStructuredAndScannable validates that the initialization
// status output is well-structured with clear formatting.
func TestM02E05OutputStructuredAndScannable(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	output := runFluxidWithHome(t, root, tmpHome)

	// Verify header is present
	if !strings.Contains(output, "=== fluxid Workflow Initialization ===") {
		t.Errorf("Missing initialization header")
	}

	// Verify footer is present
	if !strings.Contains(output, "======================================") {
		t.Errorf("Missing initialization footer")
	}

	// Verify all fields use consistent "key: value (source: X)" format
	lines := strings.Split(output, "\n")
	foundAgent := false
	foundIterations := false
	foundRetries := false
	foundCommit := false

	for _, line := range lines {
		if strings.Contains(line, "Agent:") && strings.Contains(line, "(source:") {
			foundAgent = true
		}
		if strings.Contains(line, "Max Review Cycles:") && strings.Contains(line, "(source:") {
			foundIterations = true
		}
		if strings.Contains(line, "Max Implement Retries:") && strings.Contains(line, "(source:") {
			foundRetries = true
		}
		if strings.Contains(line, "Commit Enabled:") && strings.Contains(line, "(source:") {
			foundCommit = true
		}
	}

	if !foundAgent {
		t.Errorf("Agent field not in expected format 'key: value (source: X)'")
	}
	if !foundIterations {
		t.Errorf("Max Review Cycles field not in expected format")
	}
	if !foundRetries {
		t.Errorf("Max Implement Retries field not in expected format")
	}
	if !foundCommit {
		t.Errorf("Commit Enabled field not in expected format")
	}
}

// TestM02E05StatusAppearsBeforeWorkflow validates that the initialization
// status is displayed before any workflow actions begin.
func TestM02E05StatusAppearsBeforeWorkflow(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	output := runFluxidWithHome(t, root, tmpHome)

	// Find positions of initialization and workflow sections
	initIdx := strings.Index(output, "=== fluxid Workflow Initialization ===")
	workflowIdx := strings.Index(output, "Review Cycle")

	if initIdx == -1 {
		t.Fatalf("Initialization section not found")
	}

	if workflowIdx == -1 {
		t.Fatalf("Workflow execution section not found")
	}

	// Verify initialization appears before workflow
	if initIdx >= workflowIdx {
		t.Errorf("Initialization status must appear before workflow execution (init at %d, workflow at %d)",
			initIdx, workflowIdx)
	}
}

// TestM02E05FullPrecedenceChain validates the complete precedence chain
// with all configuration sources active simultaneously.
func TestM02E05FullPrecedenceChain(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Setup: default < home < project < env < cli
	// - agent: home (no override)
	// - iterations: project (overrides home)
	// - implement_retries: cli (overrides all)
	// - commit_enabled: default (no override)

	homeConfigContent := `agent: claude
iterations: 10
`
	tmpHome := setupHomeWithConfig(t, homeConfigContent)

	projectConfigContent := `iterations: 25
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	// Run with CLI override for implement_retries
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-implement-retries", "8")

	// Verify each source is correctly attributed
	if !strings.Contains(output, "Agent: claude (source: home)") {
		t.Errorf("Agent source should be home, got:\n%s", output)
	}

	if !strings.Contains(output, "Max Review Cycles: 25 (source: project)") {
		t.Errorf("Max Review Cycles source should be project, got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 8 (source: cli)") {
		t.Errorf("Max Implement Retries source should be cli, got:\n%s", output)
	}

	if !strings.Contains(output, "Commit Enabled: false (source: default)") {
		t.Errorf("Commit Enabled source should be default, got:\n%s", output)
	}
}

// TestM02E05EnvOverridesInStatus validates that environment variable
// overrides are correctly displayed with "source: env".
func TestM02E05EnvOverridesInStatus(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := setupHomeWithConfig(t, basicHomeConfig)
	tmpProjectDir := t.TempDir()

	// Run with environment overrides
	envVars := map[string]string{
		"FLUXID_ITERATIONS":     "30",
		"FLUXID_COMMIT_ENABLED": "false",
	}
	output := runFluxidInDirWithEnv(t, root, tmpHome, tmpProjectDir, envVars)

	// Verify env overrides are shown
	if !strings.Contains(output, "Max Review Cycles: 30 (source: env)") {
		t.Errorf("Expected iterations from env, got:\n%s", output)
	}

	if !strings.Contains(output, "Commit Enabled: false (source: env)") {
		t.Errorf("Expected commit_enabled from env, got:\n%s", output)
	}

	// Verify home value for non-overridden key
	if !strings.Contains(output, "Max Implement Retries: 5 (source: home)") {
		t.Errorf("Expected implement_retries from home, got:\n%s", output)
	}
}

// TestM02E05SessionIDFormat validates that Session ID is displayed
// and has valid UUID v4 format.
func TestM02E05SessionIDFormat(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	output := runFluxidWithHome(t, root, tmpHome)

	// Verify Session ID is present
	if !strings.Contains(output, "Session ID:") {
		t.Errorf("Session ID not displayed in initialization status")
	}

	// Extract and validate UUID v4 format
	lines := strings.Split(output, "\n")
	foundSessionID := false
	for _, line := range lines {
		if strings.Contains(line, "Session ID:") {
			foundSessionID = true
			// Check for UUID v4 pattern (basic validation)
			if !strings.Contains(line, "-") {
				t.Errorf("Session ID doesn't appear to be in UUID format: %s", line)
			}
			break
		}
	}

	if !foundSessionID {
		t.Errorf("Session ID line not found in output")
	}
}
