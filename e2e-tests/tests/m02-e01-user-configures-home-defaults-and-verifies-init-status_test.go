//nolint:paralleltest // E2E tests use shared infrastructure
package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestM02E01HomeConfigApplied validates that ~/.fluxid/config.yaml is read
// and values are applied correctly.
func TestM02E01HomeConfigApplied(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	configContent := `agent: claude
implement_retries: 5
iterations: 10`
	tmpHome := setupHomeWithConfig(t, configContent)
	output := runFluxidWithHome(t, root, tmpHome)

	// v2.0: source tracking removed (Phase 9), just verify values
	verifyConfigLine(t, output, "Agent: claude", "")
	verifyConfigLine(t, output, "Max Review Cycles: 10", "")
	verifyConfigLine(t, output, "Max Implement Retries: 5", "")
	// v2.0: commit_enabled removed (Phase 10 - commits always enabled)
}

// TestM02E01DefaultsWhenNoHomeConfig validates that defaults are used when
// ~/.fluxid/config.yaml doesn't exist.
func TestM02E01DefaultsWhenNoHomeConfig(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home with minimal v2.0 config (need commands section)
	tmpHome := t.TempDir()
	fluxidDir := createHomeConfigDir(t, tmpHome)
	// Create minimal config with only commands section (other values will use defaults)
	minimalConfig := fmt.Sprintf(`commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, fluxidDir, fluxidDir, fluxidDir)
	writeHomeConfig(t, fluxidDir, minimalConfig)

	// Run fluxid with custom HOME
	output := runFluxidWithHome(t, root, tmpHome)

	// v2.0: source tracking removed (Phase 9), just verify default values
	verifyConfigLine(t, output, "Agent: claude", "")
	verifyConfigLine(t, output, "Max Review Cycles: 20", "")
	verifyConfigLine(t, output, "Max Implement Retries: 3", "")
	// v2.0: commit_enabled removed (Phase 10 - commits always enabled)
}

// TestM02E01PartialHomeConfig validates that partial config files work correctly,
// with omitted keys using defaults.
func TestM02E01PartialHomeConfig(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	configContent := `iterations: 15`
	tmpHome := setupHomeWithConfig(t, configContent)
	output := runFluxidWithHome(t, root, tmpHome)

	// v2.0: source tracking removed (Phase 9)
	// Verify partial override: iterations from home, others from defaults
	verifyConfigLine(t, output, "Max Review Cycles: 15", "")
	verifyConfigLine(t, output, "Agent: claude", "")
	verifyConfigLine(t, output, "Max Implement Retries: 3", "")
	// v2.0: commit_enabled removed (Phase 10 - commits always enabled)
}

// TestM02E01InvalidTypeInConfig validates that invalid types are rejected with
// clear error messages.
func TestM02E01InvalidTypeInConfig(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	configContent := `implement_retries: "three"`
	tmpHome := setupHomeWithConfig(t, configContent)
	errOutput, exitCode := runFluxidExpectError(t, root, tmpHome)

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", exitCode)
	}

	// Verify error message mentions the config file and parsing issue
	if !strings.Contains(errOutput, "Error loading home configuration") {
		t.Errorf("Expected error about home configuration, got: %s", errOutput)
	}
}

// TestM02E01InvalidValueInConfig validates that invalid values (e.g., zero or negative)
// are rejected with clear error messages.
func TestM02E01InvalidValueInConfig(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	configContent := `implement_retries: 0`
	tmpHome := setupHomeWithConfig(t, configContent)
	errOutput, exitCode := runFluxidExpectError(t, root, tmpHome)

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", exitCode)
	}

	// Verify error message mentions positive integer requirement
	if !strings.Contains(errOutput, "positive integer") && !strings.Contains(errOutput, "≥1") {
		t.Errorf("Expected error about positive integer requirement, got: %s", errOutput)
	}
}

// TestM02E01CLIOverridesHomeConfig validates that CLI flags override home config values
// and source is correctly reported as "cli".
func TestM02E01CLIOverridesHomeConfig(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := setupHomeWithConfig(t, basicHomeConfig)
	output := runFluxidWithHomeAndArgs(t, root, tmpHome,
		"--fluxid-iterations=25",
		"--fluxid-implement-retries=7")

	// v2.0: source tracking removed (Phase 9)
	// Verify CLI overrides home config
	verifyConfigLine(t, output, "Max Review Cycles: 25", "")
	verifyConfigLine(t, output, "Max Implement Retries: 7", "")
}

// TestM02E01InitializationStatusFormat validates that the initialization status
// section appears with the expected format and all required fields.
func TestM02E01InitializationStatusFormat(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home with minimal v2.0 config (need commands section)
	tmpHome := t.TempDir()
	fluxidDir := createHomeConfigDir(t, tmpHome)
	minimalConfig := fmt.Sprintf(`commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, fluxidDir, fluxidDir, fluxidDir)
	writeHomeConfig(t, fluxidDir, minimalConfig)
	output := runFluxidWithHome(t, root, tmpHome)

	// Verify initialization header
	if !strings.Contains(output, "Workflow Initialization") {
		t.Errorf("Missing initialization header")
	}

	// v2.0: source tracking removed (Phase 9), commit_enabled removed (Phase 10)
	// Verify all required fields are present
	requiredFields := []string{
		"Agent:",
		"Session ID:",
		"Max Review Cycles:",
		"Max Implement Retries:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(output, field) {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// Verify initialization appears BEFORE workflow execution
	initIdx := strings.Index(output, "Workflow Initialization")
	phaseIdx := strings.Index(output, "REVIEW CYCLE")

	if initIdx == -1 || phaseIdx == -1 || initIdx >= phaseIdx {
		t.Errorf("Initialization section should appear before workflow execution")
	}
}

// TestM02E01NoProjectStateModification validates that running fluxid with only
// home config doesn't create or modify files in the current directory.
func TestM02E01NoProjectStateModification(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpWorkDir := t.TempDir()
	tmpSessionRoot := t.TempDir() // Separate directory for session files
	configContent := `iterations: 5
`
	tmpHome := setupHomeWithConfig(t, configContent)

	// List files before running fluxid
	filesBefore, err := os.ReadDir(tmpWorkDir)
	if err != nil {
		t.Fatalf("Failed to read working dir before: %v", err)
	}

	// Run fluxid from the temporary working directory with session root outside workdir
	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)
	cmd.Dir = tmpWorkDir
	cmd.Env = append(os.Environ(),
		"HOME="+tmpHome,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
		"FLUXID_SESSION_ROOT="+tmpSessionRoot,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, stdout.String())
	}

	// List files after running fluxid
	filesAfter, err := os.ReadDir(tmpWorkDir)
	if err != nil {
		t.Fatalf("Failed to read working dir after: %v", err)
	}

	// Verify no files were created or modified in working directory
	if len(filesAfter) != len(filesBefore) {
		var newFiles []string
		for _, f := range filesAfter {
			newFiles = append(newFiles, f.Name())
		}
		t.Errorf(
			"Expected no files in working directory, but file count changed from %d to %d. Created: %v",
			len(filesBefore), len(filesAfter), newFiles)
	}
}
