package tests

import (
	"strings"
	"testing"
)

// TestM05E04PrecedenceChainForAgent validates the complete precedence chain
// for agent selection: CLI > env > project > home > default.
func TestM05E04PrecedenceChainForAgent(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	t.Run("CLI overrides env", func(t *testing.T) {
		t.Parallel()

		tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "", "")

		envVars := map[string]string{
			"FLUXID_AGENT": "codex",
		}

		// CLI flag should override env
		output := runFluxidInDirWithEnvAndArgs(t, root, tmpHome, tmpProjectDir, envVars, "--opencode")

		expectedLine := "Agent: opencode (source: cli)"
		if !strings.Contains(output, expectedLine) {
			t.Errorf("Expected %q in output (CLI should override env), got:\n%s", expectedLine, output)
		}
	})

	t.Run("env overrides project config", func(t *testing.T) {
		t.Parallel()

		tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "", "codex")

		// Env should override project config
		envVars := map[string]string{
			"FLUXID_AGENT": "opencode",
		}
		output := runFluxidInDirWithEnv(t, root, tmpHome, tmpProjectDir, envVars)

		expectedLine := "Agent: opencode (source: env)"
		if !strings.Contains(output, expectedLine) {
			t.Errorf("Expected %q in output (env should override project), got:\n%s", expectedLine, output)
		}
	})

	t.Run("project config overrides home config", func(t *testing.T) {
		t.Parallel()

		tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "claude", "codex")

		output := runFluxidInDirWithOutput(t, root, tmpHome, tmpProjectDir)

		// Project should override home
		if !strings.Contains(output, "Agent: codex") {
			t.Errorf("Expected codex (project config should override home), got:\n%s", output)
		}

		if !strings.Contains(output, "(source: project") {
			t.Errorf("Expected source: project, got:\n%s", output)
		}
	})
}

// TestM05E04FormattingConsistentAcrossSources validates that the agent line
// format is consistent regardless of the source.
//
//nolint:funlen // E2E test with format consistency checks
func TestM05E04FormattingConsistentAcrossSources(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// All sources should use format: "Agent: <name> (source: <source>)"
	tests := []struct {
		name         string
		setupFunc    func(*testing.T) (string, string, map[string]string, []string)
		expectedLine string
	}{
		{
			name: "CLI format",
			setupFunc: func(t *testing.T) (string, string, map[string]string, []string) {
				t.Helper()
				tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "", "")
				return tmpHome, tmpProjectDir, nil, []string{"--codex"}
			},
			expectedLine: "Agent: codex (source: cli)",
		},
		{
			name: "env format",
			setupFunc: func(t *testing.T) (string, string, map[string]string, []string) {
				t.Helper()
				tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "", "")
				return tmpHome, tmpProjectDir, map[string]string{"FLUXID_AGENT": "codex"}, nil
			},
			expectedLine: "Agent: codex (source: env)",
		},
		{
			name: "default format",
			setupFunc: func(t *testing.T) (string, string, map[string]string, []string) {
				t.Helper()
				tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "", "")
				return tmpHome, tmpProjectDir, nil, nil
			},
			expectedLine: "Agent: claude (source: default)",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpHome, tmpProjectDir, envVars, args := testCase.setupFunc(t)

			var output string
			switch {
			case len(args) > 0:
				output = runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir, args...)
			case len(envVars) > 0:
				output = runFluxidInDirWithEnv(t, root, tmpHome, tmpProjectDir, envVars)
			default:
				output = runFluxidInDirWithOutput(t, root, tmpHome, tmpProjectDir)
			}

			if !strings.Contains(output, testCase.expectedLine) {
				t.Errorf("Expected consistent format %q, got:\n%s", testCase.expectedLine, output)
			}
		})
	}
}
