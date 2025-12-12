package tests

import (
	"fmt"
	"strings"
	"testing"
)

// TestM02E04EnvOverridesProjectAndHome validates that environment variables
// override both project and home config with correct source tracking.
func TestM02E04EnvOverridesProjectAndHome(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config
	tmpHome := setupHomeWithConfig(t, fullHomeConfig)

	// Create project config
	projectConfigContent := `agent: opencode
iterations: 15
implement_retries: 7
commit_enabled: false
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	// Run with environment variables that override both
	output := runFluxidInDirWithEnv(t, root, tmpHome, tmpProjectDir, map[string]string{
		"FLUXID_AGENT":             "codex",
		"FLUXID_ITERATIONS":        "25",
		"FLUXID_IMPLEMENT_RETRIES": "9",
		"FLUXID_COMMIT_ENABLED":    "false",
	})

	// Verify environment variables override project and home
	if !strings.Contains(output, "Agent: codex (source: env)") {
		t.Errorf("Expected Agent: codex (source: env), got:\n%s", output)
	}

	if !strings.Contains(output, "Max Review Cycles: 25 (source: env)") {
		t.Errorf("Expected Max Review Cycles: 25 (source: env), got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 9 (source: env)") {
		t.Errorf("Expected Max Implement Retries: 9 (source: env), got:\n%s", output)
	}

	if !strings.Contains(output, "Commit Enabled: false (source: env)") {
		t.Errorf("Expected Commit Enabled: false (source: env), got:\n%s", output)
	}
}

// TestM02E04CLIOverridesEnvProjectAndHome validates that CLI flags override
// environment variables, project, and home config.
func TestM02E04CLIOverridesEnvProjectAndHome(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config
	tmpHome := setupHomeWithConfig(t, basicHomeConfig)

	// Create project config
	projectConfigContent := `iterations: 15
implement_retries: 7
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	// Run with env vars and CLI flags - CLI should win
	output := runFluxidInDirWithEnvAndArgs(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_ITERATIONS":        "20",
			"FLUXID_IMPLEMENT_RETRIES": "8",
		},
		"--fluxid-iterations", "30",
		"--fluxid-implement-retries", "12",
	)

	// Verify CLI flags override everything
	if !strings.Contains(output, "Max Review Cycles: 30 (source: cli)") {
		t.Errorf("Expected Max Review Cycles: 30 (source: cli), got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 12 (source: cli)") {
		t.Errorf("Expected Max Implement Retries: 12 (source: cli), got:\n%s", output)
	}
}

// TestM02E04NoCommitFlag validates that --fluxid-no-commit sets commit_enabled=false.
func TestM02E04NoCommitFlag(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config with commit enabled
	homeConfigContent := `commit_enabled: true
`
	tmpHome := setupHomeWithConfig(t, homeConfigContent)
	tmpProjectDir := t.TempDir()

	// Run with --fluxid-no-commit flag
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir,
		"--fluxid-no-commit",
	)

	// Verify --fluxid-no-commit sets commit_enabled=false
	if !strings.Contains(output, "Commit Enabled: false (source: cli)") {
		t.Errorf("Expected Commit Enabled: false (source: cli), got:\n%s", output)
	}
}

// TestM02E04NoCommitFlagOverridesEnvAndConfig validates --fluxid-no-commit
// overrides environment variable and config files.
func TestM02E04NoCommitFlagOverridesEnvAndConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config with commit enabled
	homeConfigContent := `commit_enabled: true
`
	tmpHome := setupHomeWithConfig(t, homeConfigContent)
	tmpProjectDir := t.TempDir()

	// Run with env var and --fluxid-no-commit - CLI should win
	output := runFluxidInDirWithEnvAndArgs(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_COMMIT_ENABLED": "true",
		},
		"--fluxid-no-commit",
	)

	// Verify CLI flag overrides env var
	if !strings.Contains(output, "Commit Enabled: false (source: cli)") {
		t.Errorf("Expected Commit Enabled: false (source: cli), got:\n%s", output)
	}
}

// TestM02E04InvalidEnvIterations validates that invalid FLUXID_ITERATIONS
// produces a clear error message with source identification.
func TestM02E04InvalidEnvIterations(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpHome := t.TempDir()
	tmpProjectDir := t.TempDir()

	// Run with invalid env var
	errOutput, exitCode := runFluxidInDirWithEnvExpectError(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_ITERATIONS": "-5",
		},
	)

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", exitCode)
	}

	// Verify error mentions environment configuration and validation
	if !strings.Contains(errOutput, "Error loading environment configuration") {
		t.Errorf("Expected error about environment configuration, got: %s", errOutput)
	}

	if !strings.Contains(errOutput, "FLUXID_ITERATIONS") {
		t.Errorf("Expected error to mention FLUXID_ITERATIONS, got: %s", errOutput)
	}

	if !strings.Contains(errOutput, "positive integer") && !strings.Contains(errOutput, "≥1") {
		t.Errorf("Expected error about positive integer requirement, got: %s", errOutput)
	}
}

// TestM02E04InvalidEnvCommitEnabled validates that invalid FLUXID_COMMIT_ENABLED
// produces a clear error message.
func TestM02E04InvalidEnvCommitEnabled(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpHome := t.TempDir()
	tmpProjectDir := t.TempDir()

	// Run with invalid env var
	errOutput, exitCode := runFluxidInDirWithEnvExpectError(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_COMMIT_ENABLED": "maybe",
		},
	)

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", exitCode)
	}

	// Verify error mentions environment configuration and valid values
	if !strings.Contains(errOutput, "Error loading environment configuration") {
		t.Errorf("Expected error about environment configuration, got: %s", errOutput)
	}

	if !strings.Contains(errOutput, "FLUXID_COMMIT_ENABLED") {
		t.Errorf("Expected error to mention FLUXID_COMMIT_ENABLED, got: %s", errOutput)
	}

	if !strings.Contains(errOutput, "true/false") {
		t.Errorf("Expected error to mention valid values (true/false), got: %s", errOutput)
	}
}

// TestM02E04CommitEnabledTrueFalseVariations validates different true/false
// values for FLUXID_COMMIT_ENABLED environment variable.
func TestM02E04CommitEnabledTrueFalseVariations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"true", "true", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := getProjectRoot(t)
			buildFluxid(t, root)
			createStubClaude(t, root)

			tmpHome := t.TempDir()
			tmpProjectDir := t.TempDir()

			output := runFluxidInDirWithEnv(t, root, tmpHome, tmpProjectDir, map[string]string{
				"FLUXID_COMMIT_ENABLED": tc.envValue,
			})

			expectedOutput := fmt.Sprintf("Commit Enabled: %v (source: env)", tc.expected)
			if !strings.Contains(output, expectedOutput) {
				t.Errorf("Expected '%s', got:\n%s", expectedOutput, output)
			}
		})
	}
}

// TestM02E04FullPrecedenceChain validates the complete precedence chain:
// defaults → home → project → env → CLI.
func TestM02E04FullPrecedenceChain(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create home config
	homeConfigContent := `agent: home-agent
iterations: 10
implement_retries: 3
commit_enabled: false
`
	tmpHome := setupHomeWithConfig(t, homeConfigContent)

	// Create project config that overrides some home values
	projectConfigContent := `agent: project-agent
iterations: 20
`
	tmpProjectDir := createProjectWithConfig(t, projectConfigContent)

	// Run with env vars that override some project values, and CLI that overrides env
	output := runFluxidInDirWithEnvAndArgs(t, root, tmpHome, tmpProjectDir,
		map[string]string{
			"FLUXID_ITERATIONS": "30",
		},
		"--fluxid-iterations", "40",
	)

	// Verify precedence chain:
	// - Agent: project overrides home
	// - Iterations: CLI overrides env, which would override project
	// - ImplementRetries: home (not overridden)
	// - CommitEnabled: home (not overridden)

	if !strings.Contains(output, "Agent: project-agent (source: project)") {
		t.Errorf("Expected Agent from project, got:\n%s", output)
	}

	if !strings.Contains(output, "Max Review Cycles: 40 (source: cli)") {
		t.Errorf("Expected iterations from CLI, got:\n%s", output)
	}

	if !strings.Contains(output, "Max Implement Retries: 3 (source: home)") {
		t.Errorf("Expected implement_retries from home, got:\n%s", output)
	}

	if !strings.Contains(output, "Commit Enabled: false (source: home)") {
		t.Errorf("Expected commit_enabled from home, got:\n%s", output)
	}
}
