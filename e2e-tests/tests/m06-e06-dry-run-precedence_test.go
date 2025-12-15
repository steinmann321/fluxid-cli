package tests

import (
	"strings"
	"testing"
)

// TestM06E06DryRunRespectsAllConfigSources validates that dry-run mode respects
// configuration from all sources (file, env, CLI) with correct precedence.
func TestM06E06DryRunRespectsAllConfigSources(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	homeConfigContent := `agent: claude
iterations: 10
implement_retries: 3
commit_enabled: false
`
	tmpHome := setupHomeWithConfig(t, homeConfigContent)

	projectConfigContent := `agent: opencode
iterations: 15
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	output := runFluxidInDirWithEnvAndArgs(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_AGENT":             "codex",
			"FLUXID_ITERATIONS":        "20",
			"FLUXID_IMPLEMENT_RETRIES": "5",
		},
		"--fluxid-dry-run",
		"--fluxid-iterations", "25",
		"--fluxid-commit-enabled",
	)

	if !strings.Contains(output, "=== DRY RUN MODE - Simulation Only ===") {
		t.Errorf("Missing dry-run header in output:\n%s", output)
	}

	verifyConfigLine(t, output, "Agent: codex", "source: env")
	verifyConfigLine(t, output, "Max Review Cycles: 25", "source: cli")
	verifyConfigLine(t, output, "Max Implement Retries: 5", "source: env")
	verifyConfigLine(t, output, "Commit Enabled: true", "source: cli")
}

// TestM06E06DryRunWithFileOnlyConfig validates dry-run uses file-based config correctly.
func TestM06E06DryRunWithFileOnlyConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	projectConfigContent := `agent: opencode
iterations: 12
implement_retries: 4
commit_enabled: true
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)
	tmpHome := t.TempDir()

	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir, "--fluxid-dry-run")

	verifyConfigLine(t, output, "Agent: opencode", "source: project")
	verifyConfigLine(t, output, "Max Review Cycles: 12", "source: project")
	verifyConfigLine(t, output, "Max Implement Retries: 4", "source: project")
	verifyConfigLine(t, output, "Commit Enabled: true", "source: project")
}

// TestM06E06DryRunWithEnvOnlyConfig validates dry-run uses environment variables correctly.
func TestM06E06DryRunWithEnvOnlyConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpProjectDir := t.TempDir()

	output := runFluxidInDirWithEnvAndArgs(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_AGENT":             "codex",
			"FLUXID_ITERATIONS":        "8",
			"FLUXID_IMPLEMENT_RETRIES": "2",
			"FLUXID_COMMIT_ENABLED":    "true",
		},
		"--fluxid-dry-run",
	)

	verifyConfigLine(t, output, "Agent: codex", "source: env")
	verifyConfigLine(t, output, "Max Review Cycles: 8", "source: env")
	verifyConfigLine(t, output, "Max Implement Retries: 2", "source: env")
	verifyConfigLine(t, output, "Commit Enabled: true", "source: env")
}

// TestM06E06DryRunWithCLIOnlyConfig validates dry-run uses CLI flags correctly.
func TestM06E06DryRunWithCLIOnlyConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpProjectDir := t.TempDir()

	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-dry-run",
		"--claude",
		"--fluxid-iterations", "30",
		"--fluxid-implement-retries", "6",
		"--fluxid-commit-enabled",
	)

	verifyConfigLine(t, output, "Agent: claude", "source: cli")
	verifyConfigLine(t, output, "Max Review Cycles: 30", "source: cli")
	verifyConfigLine(t, output, "Max Implement Retries: 6", "source: cli")
	verifyConfigLine(t, output, "Commit Enabled: true", "source: cli")
}

// TestM06E06DryRunPrecedenceChainComplete validates the complete precedence chain.
func TestM06E06DryRunPrecedenceChainComplete(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	homeConfigContent := `agent: claude
iterations: 5
implement_retries: 2
commit_enabled: false
`
	tmpHome := setupHomeWithConfig(t, homeConfigContent)

	projectConfigContent := `agent: opencode
iterations: 10
implement_retries: 3
commit_enabled: false
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	output := runFluxidInDirWithEnvAndArgs(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_AGENT":             "codex",
			"FLUXID_ITERATIONS":        "20",
			"FLUXID_IMPLEMENT_RETRIES": "8",
			"FLUXID_COMMIT_ENABLED":    "false",
		},
		"--fluxid-dry-run",
		"--fluxid-iterations", "35",
		"--fluxid-commit-enabled",
	)

	verifyConfigLine(t, output, "Agent: codex", "source: env")
	verifyConfigLine(t, output, "Max Review Cycles: 35", "source: cli")
	verifyConfigLine(t, output, "Max Implement Retries: 8", "source: env")
	verifyConfigLine(t, output, "Commit Enabled: true", "source: cli")
}

// TestM06E06DryRunNoCommitFlagOverridesAll validates --fluxid-no-commit precedence.
func TestM06E06DryRunNoCommitFlagOverridesAll(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := setupHomeWithConfig(t, homeConfigCommitEnabled)
	tmpProjectDir := createProjectWithConfig(t, homeConfigCommitEnabled)

	output := runFluxidInDirWithEnvAndArgs(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_AGENT":          "claude",
			"FLUXID_COMMIT_ENABLED": "true",
		},
		"--fluxid-dry-run",
		"--fluxid-no-commit",
	)

	verifyConfigLine(t, output, "Commit Enabled: false", "source: cli")

	if strings.Contains(output, "Phase: commit") {
		t.Errorf("Commit phase should not appear when disabled:\n%s", output)
	}
}
