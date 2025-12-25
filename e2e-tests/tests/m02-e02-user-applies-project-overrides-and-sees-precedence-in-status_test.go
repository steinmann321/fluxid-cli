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

	// v2.0: source tracking removed (Phase 9), commit_enabled removed (Phase 10)
	// Verify project values override home values
	verifyConfigLine(t, output, "Agent: opencode", "")
	verifyConfigLine(t, output, "Max Implement Retries: 7", "")
	verifyConfigLine(t, output, "Max Review Cycles: 15", "")
}

// TestM02E02ProjectOnlyConfig validates that project config works when home config
// doesn't exist, with correct source attribution.
func TestM02E02ProjectOnlyConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home with minimal v2.0 config (need commands section)
	tmpHome := t.TempDir()
	homeFluxidDir := filepath.Join(tmpHome, ".fluxid")
	if err := os.MkdirAll(homeFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create home .fluxid dir: %v", err)
	}
	homeConfigContent := `commands:
  implement: implement.md
  review: review.md
  commit: commit.md
`
	writeHomeConfig(t, homeFluxidDir, homeConfigContent)

	// Create project config with v2.0 commands section
	projectConfigContent := `iterations: 12
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	// Run fluxid from the project directory in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir, "--fluxid-dry-run")

	// v2.0: source tracking removed (Phase 9), commit_enabled removed (Phase 10)
	// Verify project values are used
	verifyConfigLine(t, output, "Max Review Cycles: 12", "")

	// Verify defaults are used for non-overridden fields
	verifyConfigLine(t, output, "Agent: claude", "")
	verifyConfigLine(t, output, "Max Implement Retries: 3", "")
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

	// v2.0: source tracking removed (Phase 9)
	// Verify home values are used (no project overrides)
	verifyConfigLine(t, output, "Max Review Cycles: 8", "")
	verifyConfigLine(t, output, "Max Implement Retries: 4", "")
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

	// v2.0: source tracking removed (Phase 9), commit_enabled removed (Phase 10)
	// Verify project overrides iterations
	verifyConfigLine(t, output, "Max Review Cycles: 25", "")

	// Verify other fields use home config
	verifyConfigLine(t, output, "Agent: claude", "")
	verifyConfigLine(t, output, "Max Implement Retries: 5", "")
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

	// Create project config with v2.0 commands section
	projectConfigContent := `iterations: 15
implement_retries: 7
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	// Run fluxid with CLI overrides in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-iterations=30",
		"--fluxid-implement-retries=9",
		"--fluxid-dry-run")

	// v2.0: source tracking removed (Phase 9)
	// Verify CLI overrides both project and home
	verifyConfigLine(t, output, "Max Review Cycles: 30", "")
	verifyConfigLine(t, output, "Max Implement Retries: 9", "")
}

// TestM02E02InvalidProjectConfig validates that invalid project config
// produces a clear error message.
func TestM02E02InvalidProjectConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create temporary home with minimal v2.0 config (need commands section)
	tmpHome := t.TempDir()
	homeFluxidDir := filepath.Join(tmpHome, ".fluxid")
	if err := os.MkdirAll(homeFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create home .fluxid dir: %v", err)
	}
	homeConfigContent := `commands:
  implement: implement.md
  review: review.md
  commit: commit.md
`
	writeHomeConfig(t, homeFluxidDir, homeConfigContent)

	// Create project with invalid config (iterations: 0) but valid v2.0 commands section
	projectConfigContent := `iterations: 0
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

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
