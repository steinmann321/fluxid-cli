package tests

import (
	"strings"
	"testing"
)

// TestM06E06DryRunEffectiveConfigurationDisplay validates that the effective
// configuration is clearly displayed with sources.
func TestM06E06DryRunEffectiveConfigurationDisplay(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpProjectDir := t.TempDir()

	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-dry-run",
		"--claude",
		"--fluxid-iterations", "15",
		"--fluxid-implement-retries", "4",
	)

	if !strings.Contains(output, "=== fluxid Workflow Initialization ===") {
		t.Errorf("Missing initialization section showing effective configuration:\n%s", output)
	}

	expectedFields := []string{
		"Agent:",
		"source:",
		"Max Review Cycles:",
		"Max Implement Retries:",
		"Commit Enabled:",
		"Session ID:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field %q in effective configuration display, got:\n%s", field, output)
		}
	}

	sourceCount := strings.Count(output, "source:")
	if sourceCount < 4 {
		t.Errorf("Expected at least 4 source annotations, got %d:\n%s", sourceCount, output)
	}
}

// TestM06E06DryRunExitCode validates dry-run exits with 0 on successful simulation.
func TestM06E06DryRunExitCode(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	homeConfigContent := `iterations: 10
`
	tmpHome := setupHomeWithConfig(t, homeConfigContent)

	projectConfigContent := `implement_retries: 3
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	output := runFluxidInDirWithEnvAndArgs(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_AGENT": "claude",
		},
		"--fluxid-dry-run",
		"--fluxid-commit-enabled",
	)

	if !strings.Contains(output, "=== End Simulation ===") {
		t.Errorf("Expected simulation to complete successfully:\n%s", output)
	}
}
