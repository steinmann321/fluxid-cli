package tests

import (
	"os"
	"strings"
	"testing"
)

// TestM02E01HomeConfigApplied validates that ~/.fluxid/config.yaml is read
// and values are applied correctly.
func TestM02E01HomeConfigApplied(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	configContent := `agent: claude
implement_retries: 5
iterations: 10
commit_enabled: false
`
	tmpHome := setupHomeWithConfig(t, configContent)
	output := runFluxidWithHome(t, root, tmpHome)

	// Verify home config values are applied
	if !strings.Contains(output, "Agent: claude (source: home)") {
		t.Errorf("Expected Agent: claude (source: home), got:\n%s", output)
	}

	if !strings.Contains(output, "Max Review Cycles: 10 (source: home)") {
		t.Errorf("Expected Max Review Cycles: 10 (source: home), got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 5 (source: home)") {
		t.Errorf("Expected Max Implement Retries: 5 (source: home), got:\n%s", output)
	}

	if !strings.Contains(output, "Commit Enabled: false (source: home)") {
		t.Errorf("Expected Commit Enabled: false (source: home), got:\n%s", output)
	}
}

// TestM02E01DefaultsWhenNoHomeConfig validates that defaults are used when
// ~/.fluxid/config.yaml doesn't exist.
func TestM02E01DefaultsWhenNoHomeConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home directory WITHOUT config file
	tmpHome := t.TempDir()

	// Run fluxid with custom HOME
	output := runFluxidWithHome(t, root, tmpHome)

	// Verify defaults are applied with correct sources
	if !strings.Contains(output, "Agent: claude (source: default)") {
		t.Errorf("Expected Agent: claude (source: default), got:\n%s", output)
	}

	if !strings.Contains(output, "Max Review Cycles: 20 (source: default)") {
		t.Errorf("Expected Max Review Cycles: 20 (source: default), got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 3 (source: default)") {
		t.Errorf("Expected Max Implement Retries: 3 (source: default), got:\n%s", output)
	}

	if !strings.Contains(output, "Commit Enabled: true (source: default)") {
		t.Errorf("Expected Commit Enabled: true (source: default), got:\n%s", output)
	}
}

// TestM02E01PartialHomeConfig validates that partial config files work correctly,
// with omitted keys using defaults.
func TestM02E01PartialHomeConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	configContent := `iterations: 15
`
	tmpHome := setupHomeWithConfig(t, configContent)
	output := runFluxidWithHome(t, root, tmpHome)

	// Verify partial override: iterations from home, others from defaults
	if !strings.Contains(output, "Max Review Cycles: 15 (source: home)") {
		t.Errorf("Expected Max Review Cycles: 15 (source: home), got:\n%s", output)
	}

	if !strings.Contains(output, "Agent: claude (source: default)") {
		t.Errorf("Expected Agent: claude (source: default), got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 3 (source: default)") {
		t.Errorf("Expected Max Implement Retries: 3 (source: default), got:\n%s", output)
	}

	if !strings.Contains(output, "Commit Enabled: true (source: default)") {
		t.Errorf("Expected Commit Enabled: true (source: default), got:\n%s", output)
	}
}

// TestM02E01InvalidTypeInConfig validates that invalid types are rejected with
// clear error messages.
func TestM02E01InvalidTypeInConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	configContent := `implement_retries: "three"
`
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
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	configContent := `implement_retries: 0
`
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
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	configContent := `iterations: 10
implement_retries: 5
`
	tmpHome := setupHomeWithConfig(t, configContent)
	output := runFluxidWithHomeAndArgs(t, root, tmpHome,
		"--fluxid-iterations", "25",
		"--fluxid-implement-retries", "7")

	// Verify CLI overrides home config
	if !strings.Contains(output, "Max Review Cycles: 25 (source: cli)") {
		t.Errorf("Expected Max Review Cycles: 25 (source: cli), got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 7 (source: cli)") {
		t.Errorf("Expected Max Implement Retries: 7 (source: cli), got:\n%s", output)
	}
}

// TestM02E01InitializationStatusFormat validates that the initialization status
// section appears with the expected format and all required fields.
func TestM02E01InitializationStatusFormat(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	output := runFluxidWithHome(t, root, tmpHome)

	// Verify initialization header
	if !strings.Contains(output, "=== fluxid Workflow Initialization ===") {
		t.Errorf("Missing initialization header")
	}

	// Verify all required fields are present with source indicators
	requiredFields := []string{
		"Agent:",
		"Session ID:",
		"Max Review Cycles:",
		"Max Implement Retries:",
		"Commit Enabled:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(output, field) {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// Verify source indicators are present
	if !strings.Contains(output, "(source: default)") && !strings.Contains(output, "(source: home)") {
		t.Errorf("Missing source indicators in output")
	}

	// Verify initialization appears BEFORE workflow execution
	initIdx := strings.Index(output, "=== fluxid Workflow Initialization ===")
	phaseIdx := strings.Index(output, "Review Cycle")

	if initIdx == -1 || phaseIdx == -1 || initIdx >= phaseIdx {
		t.Errorf("Initialization section should appear before workflow execution")
	}
}

// TestM02E01NoProjectStateModification validates that running fluxid with only
// home config doesn't create or modify files in the current directory.
func TestM02E01NoProjectStateModification(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpWorkDir := t.TempDir()
	configContent := `iterations: 5
`
	tmpHome := setupHomeWithConfig(t, configContent)

	// List files before running fluxid
	filesBefore, err := os.ReadDir(tmpWorkDir)
	if err != nil {
		t.Fatalf("Failed to read working dir before: %v", err)
	}

	// Run fluxid from the temporary working directory (using custom exec for Dir)
	runFluxidInDir(t, root, tmpHome, tmpWorkDir)

	// List files after running fluxid
	filesAfter, err := os.ReadDir(tmpWorkDir)
	if err != nil {
		t.Fatalf("Failed to read working dir after: %v", err)
	}

	// Verify no files were created or modified
	if len(filesAfter) != len(filesBefore) {
		t.Errorf("Expected no files to be created in working directory, but file count changed from %d to %d",
			len(filesBefore), len(filesAfter))
	}
}
