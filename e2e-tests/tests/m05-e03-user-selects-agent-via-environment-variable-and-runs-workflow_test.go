package tests

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestM05E03AgentFromEnvironmentVariable validates agent selection via FLUXID_AGENT env var.
func TestM05E03AgentFromEnvironmentVariable(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpDir := t.TempDir()

	// Set FLUXID_AGENT=opencode in environment
	env := []string{"FLUXID_AGENT=opencode"}

	output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, nil)
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	// Verify agent is opencode
	verifyAgent(t, output, "opencode")
	// Verify source is env
	verifySource(t, output, "source: env")
	// Verify workflow completed successfully
	verifyWorkflowSuccess(t, output)
}

// TestM05E03EnvBeatsProjectAndHome validates env variable precedence over config files.
func TestM05E03EnvBeatsProjectAndHome(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Set up home config with claude
	tmpHome := t.TempDir()
	homeConfigDir := setupConfigDir(t, tmpHome)
	writeConfigFile(t, filepath.Join(homeConfigDir, "config.yaml"), "claude")

	// Set up project config with codex
	tmpDir := t.TempDir()
	projectConfigDir := setupConfigDir(t, tmpDir)
	writeConfigFile(t, filepath.Join(projectConfigDir, "config.yaml"), "codex")

	// Set FLUXID_AGENT=opencode - should override both configs
	env := []string{"FLUXID_AGENT=opencode"}

	output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, nil)
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	// Verify env wins with opencode
	verifyAgent(t, output, "opencode")
	verifySource(t, output, "source: env")
}

// TestM05E03CLIBeatsEnv validates CLI flag precedence over environment variable.
func TestM05E03CLIBeatsEnv(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpDir := t.TempDir()

	// Set FLUXID_AGENT=codex
	env := []string{"FLUXID_AGENT=codex"}

	// But also pass --claude flag - CLI should win
	args := []string{"--claude"}

	output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, args)
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	// Verify CLI wins with claude
	verifyAgent(t, output, "claude")
	verifySource(t, output, "source: cli")
}

// TestM05E03FullPrecedenceChain validates complete precedence: CLI > env > project > home > default.
func TestM05E03FullPrecedenceChain(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tests := []struct {
		name           string
		homeAgent      string
		projectAgent   string
		envAgent       string
		cliAgent       string
		expectedAgent  string
		expectedSource string
	}{
		{
			name:           "CLI beats all",
			homeAgent:      "claude",
			projectAgent:   "codex",
			envAgent:       "opencode",
			cliAgent:       "--claude",
			expectedAgent:  "claude",
			expectedSource: "source: cli",
		},
		{
			name:           "env beats project and home",
			homeAgent:      "claude",
			projectAgent:   "codex",
			envAgent:       "opencode",
			cliAgent:       "",
			expectedAgent:  "opencode",
			expectedSource: "source: env",
		},
		{
			name:           "project beats home",
			homeAgent:      "claude",
			projectAgent:   "codex",
			envAgent:       "",
			cliAgent:       "",
			expectedAgent:  "codex",
			expectedSource: "source: project",
		},
		{
			name:           "home beats default",
			homeAgent:      "opencode",
			projectAgent:   "",
			envAgent:       "",
			cliAgent:       "",
			expectedAgent:  "opencode",
			expectedSource: "source: home",
		},
		{
			name:           "default when nothing set",
			homeAgent:      "",
			projectAgent:   "",
			envAgent:       "",
			cliAgent:       "",
			expectedAgent:  "claude",
			expectedSource: "source: default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpHome := setupTestHome(t, tt.homeAgent)
			tmpDir := setupTestProject(t, tt.projectAgent)
			env := buildTestEnv(tt.envAgent)
			args := buildTestArgs(tt.cliAgent)

			output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, args)
			if err != nil {
				t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
			}

			verifyAgent(t, output, tt.expectedAgent)
			verifySource(t, output, tt.expectedSource)
		})
	}
}

// TestM05E03InvalidEnvValue validates error handling for invalid FLUXID_AGENT values.
func TestM05E03InvalidEnvValue(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpHome := t.TempDir()
	tmpDir := t.TempDir()

	// FLUXID_AGENT is checked by LoadEnvConfig but empty strings pass through
	// since the check is `if val := envGetter.Getenv("FLUXID_AGENT"); val != "" {`
	// So we can't test empty string rejection via env var.
	// Test that an unsupported agent value fails validation with a clear error.

	env := []string{"FLUXID_AGENT=nonexistent-agent-xyz"}

	output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, nil)
	if err == nil {
		t.Fatal("Expected error with unsupported agent, but succeeded")
	}

	// Should get unsupported agent error (not PATH error, since validation happens first)
	if !strings.Contains(output, "unsupported agent") {
		t.Errorf("Expected 'unsupported agent' error, got: %s", output)
	}

	// Should list supported agents
	if !strings.Contains(output, "claude") || !strings.Contains(output, "codex") || !strings.Contains(output, "opencode") {
		t.Errorf("Expected error to list supported agents, got: %s", output)
	}
}

// TestM05E03PathResolutionWithEnv validates PATH resolution when agent specified via env.
func TestM05E03PathResolutionWithEnv(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpDir := t.TempDir()

	t.Run("agent found in PATH", func(t *testing.T) {
		t.Parallel()
		env := []string{"FLUXID_AGENT=codex"}

		output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, nil)
		if err != nil {
			t.Fatalf("Expected success when codex is in PATH, got: %v\nOutput:\n%s", err, output)
		}

		verifyAgent(t, output, "codex")
		verifySource(t, output, "source: env")
	})

	t.Run("agent not in PATH", func(t *testing.T) {
		t.Parallel()
		// Use a valid agent name (codex) but with limited PATH that doesn't include it
		env := []string{
			"FLUXID_AGENT=codex",
			"PATH=/usr/bin:/bin", // Limited PATH without our stubs
		}

		output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, nil)
		if err == nil {
			t.Fatal("Expected error when agent not in PATH, but succeeded")
		}

		if !strings.Contains(output, "not found in PATH") {
			t.Errorf("Expected 'not found in PATH' error, got: %s", output)
		}

		// Verify helpful suggestion is present
		if !strings.Contains(output, "which") {
			t.Errorf("Expected helpful 'which' suggestion in error, got: %s", output)
		}
	})
}

// TestM05E03OrchestrationMatchesBaseline validates workflow orchestration with env agent.
func TestM05E03OrchestrationMatchesBaseline(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpDir := t.TempDir()

	env := []string{"FLUXID_AGENT=opencode"}

	output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, nil)
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	// Verify phase execution order matches baseline
	if !strings.Contains(output, "Review Cycle 1/20") {
		t.Errorf("Missing review cycle indicator")
	}

	if !strings.Contains(output, "Starting phase: implement") {
		t.Errorf("Missing implement phase")
	}

	if !strings.Contains(output, "Starting phase: review") {
		t.Errorf("Missing review phase")
	}

	// Verify completion
	verifyWorkflowSuccess(t, output)
}

// TestM05E03InitializationStatusDisplay validates initialization output shows env source.
func TestM05E03InitializationStatusDisplay(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpDir := t.TempDir()

	env := []string{"FLUXID_AGENT=codex"}

	output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, nil)
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	// Verify initialization section exists
	if !strings.Contains(output, "=== fluxid Workflow Initialization ===") {
		t.Errorf("Missing initialization header")
	}

	// Verify agent and source are displayed correctly
	expectedAgentLine := "Agent: codex (source: env)"
	if !strings.Contains(output, expectedAgentLine) {
		t.Errorf("Expected agent line %q in output, got:\n%s", expectedAgentLine, output)
	}

	// Verify session ID is shown
	if !strings.Contains(output, "Session ID:") {
		t.Errorf("Missing session ID in initialization")
	}
}
