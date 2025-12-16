package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestM02E02ProjectOverridesHome validates that project config overrides home config
// and the initialization status correctly shows "source: project" for overridden values.
func TestM02E02ProjectOverridesHome(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config with baseline values
	tmpHome := setupHomeWithConfig(t, fullHomeConfig)

	// Create project config that overrides some fields and run fluxid
	projectConfigContent := `agent: opencode
implement_retries: 7
iterations: 15
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)
	// Use dry-run mode (only need init status, not workflow execution)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir, "--fluxid-dry-run")

	// Verify project values override home values
	verifyConfigLine(t, output, "Agent: opencode", "source: project")
	verifyConfigLine(t, output, "Max Implement Retries: 7", "source: project")
	verifyConfigLine(t, output, "Max Review Cycles: 15", "source: project")

	// Verify commit_enabled uses home value (not overridden in project)
	verifyConfigLine(t, output, "Commit Enabled: false", "source: home")
}

// TestM02E02ProjectOnlyConfig validates that project config works when home config
// doesn't exist, with correct source attribution.
func TestM02E02ProjectOnlyConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home directory WITHOUT config file
	tmpHome := t.TempDir()

	// Create temporary project directory with project config
	tmpProjectDir := t.TempDir()
	projectFluxidDir := filepath.Join(tmpProjectDir, ".fluxid")
	if err := os.MkdirAll(projectFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create project .fluxid dir: %v", err)
	}

	// Create project config
	projectConfigContent := `iterations: 12
commit_enabled: false
`
	projectConfigPath := filepath.Join(projectFluxidDir, "config.yaml")
	if err := os.WriteFile(projectConfigPath, []byte(projectConfigContent), 0o644); err != nil {
		t.Fatalf("Failed to write project config: %v", err)
	}

	// Run fluxid from the project directory in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir, "--fluxid-dry-run")

	// Verify project values are used
	verifyConfigLine(t, output, "Max Review Cycles: 12", "source: project")
	verifyConfigLine(t, output, "Commit Enabled: false", "source: project")

	// Verify defaults are used for non-overridden fields
	verifyConfigLine(t, output, "Agent: claude", "source: default")
	verifyConfigLine(t, output, "Max Implement Retries: 3", "source: default")
}

// TestM02E02NoProjectConfigOutsideProject validates that running fluxid outside
// a project directory (no ./.fluxid/config.yaml) uses home config only.
func TestM02E02NoProjectConfigOutsideProject(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config
	homeConfigContent := `iterations: 8
implement_retries: 4
`
	tmpHome := setupHomeWithConfig(t, homeConfigContent)

	// Create temporary directory WITHOUT .fluxid/config.yaml
	tmpWorkDir := t.TempDir()

	// Run fluxid from the directory without project config in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpWorkDir, "--fluxid-dry-run")

	// Verify home values are used (no project overrides)
	verifyConfigLine(t, output, "Max Review Cycles: 8", "source: home")
	verifyConfigLine(t, output, "Max Implement Retries: 4", "source: home")

	// No project source should appear
	if strings.Contains(output, "source: project") {
		t.Errorf("Expected no 'source: project' when no project config exists, got:\n%s", output)
	}
}

// TestM02E02PartialProjectOverride validates that partial project config only
// overrides specified fields, with remaining fields using home or defaults.
func TestM02E02PartialProjectOverride(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config
	tmpHome := setupHomeWithConfig(t, fullHomeConfig)

	// Create project with partial override (only iterations) and run fluxid in dry-run mode
	projectConfigContent := `iterations: 25
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir, "--fluxid-dry-run")

	// Verify project overrides iterations
	verifyConfigLine(t, output, "Max Review Cycles: 25", "source: project")

	// Verify other fields use home config
	verifyConfigLine(t, output, "Agent: claude", "source: home")
	verifyConfigLine(t, output, "Max Implement Retries: 5", "source: home")
	verifyConfigLine(t, output, "Commit Enabled: false", "source: home")
}

// TestM02E02CLIOverridesProjectAndHome validates that CLI flags take precedence
// over both project and home config.
func TestM02E02CLIOverridesProjectAndHome(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config
	tmpHome := setupHomeWithConfig(t, basicHomeConfig)

	// Create project config
	tmpProjectDir := t.TempDir()
	projectFluxidDir := filepath.Join(tmpProjectDir, ".fluxid")
	if err := os.MkdirAll(projectFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create project .fluxid dir: %v", err)
	}

	projectConfigContent := `iterations: 15
implement_retries: 7
`
	projectConfigPath := filepath.Join(projectFluxidDir, "config.yaml")
	if err := os.WriteFile(projectConfigPath, []byte(projectConfigContent), 0o644); err != nil {
		t.Fatalf("Failed to write project config: %v", err)
	}

	// Run fluxid with CLI overrides in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-iterations", "30",
		"--fluxid-implement-retries", "9",
		"--fluxid-dry-run")

	// Verify CLI overrides both project and home
	verifyConfigLine(t, output, "Max Review Cycles: 30", "source: cli")
	verifyConfigLine(t, output, "Max Implement Retries: 9", "source: cli")
}

// TestM02E02InvalidProjectConfig validates that invalid project config
// produces a clear error message.
func TestM02E02InvalidProjectConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpHome := t.TempDir()

	// Create project with invalid config
	tmpProjectDir := t.TempDir()
	projectFluxidDir := filepath.Join(tmpProjectDir, ".fluxid")
	if err := os.MkdirAll(projectFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create project .fluxid dir: %v", err)
	}

	projectConfigContent := `iterations: 0
`
	projectConfigPath := filepath.Join(projectFluxidDir, "config.yaml")
	if err := os.WriteFile(projectConfigPath, []byte(projectConfigContent), 0o644); err != nil {
		t.Fatalf("Failed to write project config: %v", err)
	}

	// Run fluxid expecting error
	errOutput, exitCode := runFluxidInDirExpectError(t, root, tmpHome, tmpProjectDir)

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", exitCode)
	}

	// Verify error message mentions project configuration
	if !strings.Contains(errOutput, "Error loading project configuration") {
		t.Errorf("Expected error about project configuration, got: %s", errOutput)
	}

	// Verify error mentions positive integer requirement
	if !strings.Contains(errOutput, "positive integer") && !strings.Contains(errOutput, "≥1") {
		t.Errorf("Expected error about positive integer requirement, got: %s", errOutput)
	}
}
