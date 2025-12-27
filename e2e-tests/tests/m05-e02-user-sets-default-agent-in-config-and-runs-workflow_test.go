package tests

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// TestM05E02AgentFromHomeConfig validates agent selection from home config.
func TestM05E02AgentFromHomeConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root) // Creates stubs for all agents

	tmpHome := t.TempDir()
	homeConfigDir := setupConfigDir(t, tmpHome)
	homeConfigPath := filepath.Join(homeConfigDir, "config.yaml")
	writeConfigFile(t, homeConfigPath, "codex")

	// Use minimal iterations for fast test execution
	output, err := runFluxidWithConfig(t, root, "", tmpHome, nil, []string{"--fluxid-iterations=1"})
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	// v2.0: source tracking removed (Phase 9), just verify agent selection
	verifyAgent(t, output, "codex")
	verifyWorkflowSuccess(t, output)
}

// TestM05E02AgentFromProjectConfig validates agent selection from project config.
func TestM05E02AgentFromProjectConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root) // Creates stubs for all agents

	tmpDir := t.TempDir()
	projectConfigDir := setupConfigDir(t, tmpDir)
	projectConfigPath := filepath.Join(projectConfigDir, "config.yaml")
	writeConfigFile(t, projectConfigPath, "opencode")

	// Use minimal iterations for fast test execution
	output, err := runFluxidWithConfig(t, root, tmpDir, "", nil, []string{"--fluxid-iterations=1"})
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	// v2.0: source tracking removed (Phase 9), just verify agent selection
	verifyAgent(t, output, "opencode")
	verifyWorkflowSuccess(t, output)
}

// TestM05E02ProjectOverridesHome validates project config takes precedence over home.
func TestM05E02ProjectOverridesHome(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root) // Creates stubs for all agents

	tmpHome := t.TempDir()
	homeConfigDir := setupConfigDir(t, tmpHome)
	writeConfigFile(t, filepath.Join(homeConfigDir, "config.yaml"), "claude")

	tmpDir := t.TempDir()
	projectConfigDir := setupConfigDir(t, tmpDir)
	projectConfigPath := filepath.Join(projectConfigDir, "config.yaml")
	writeConfigFile(t, projectConfigPath, "codex")

	// Use dry-run mode (only need to verify agent selection, not workflow execution)
	output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, nil, []string{"--fluxid-dry-run"})
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	// v2.0: source tracking removed (Phase 9), just verify agent selection
	verifyAgent(t, output, "codex")
}

// TestM05E02PrecedenceChain validates full precedence: CLI > project > home > default.
// v2.0: environment variable support removed (Phase 7).
func TestM05E02PrecedenceChain(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	// createStubClaude creates stubs for all agents (claude, codex, opencode, project-agent)
	createStubClaude(t, root)

	tests := []struct {
		name          string
		homeAgent     string
		projectAgent  string
		cliAgent      string
		expectedAgent string
	}{
		{
			name:          "CLI beats all",
			homeAgent:     "codex",
			projectAgent:  "opencode",
			cliAgent:      "--claude",
			expectedAgent: "claude",
		},
		{
			name:          "project beats home",
			homeAgent:     "claude",
			projectAgent:  "codex",
			cliAgent:      "",
			expectedAgent: "codex",
		},
		{
			name:          "home beats default",
			homeAgent:     "opencode",
			projectAgent:  "",
			cliAgent:      "",
			expectedAgent: "opencode",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpHome := setupTestHome(t, testCase.homeAgent)
			tmpDir := setupTestProject(t, testCase.projectAgent)
			args := buildTestArgs(testCase.cliAgent)

			// Add dry-run flag to args (only need to verify agent selection)
			args = append(args, "--fluxid-dry-run")

			output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, nil, args)
			if err != nil {
				t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
			}

			// v2.0: source tracking removed (Phase 9), just verify agent selection
			verifyAgent(t, output, testCase.expectedAgent)
		})
	}
}

// TestM05E02InvalidAgentValue validates empty agent rejection.
func TestM05E02InvalidAgentValue(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	homeConfigDir := setupConfigDir(t, tmpHome)
	writeRawConfigFile(t, filepath.Join(homeConfigDir, "config.yaml"), "agent: ''\n")

	output, err := runFluxidWithConfig(t, root, "", tmpHome, nil, nil)
	if err == nil {
		t.Fatal("Expected error with empty agent, but succeeded")
	}

	if !strings.Contains(output, "agent cannot be empty") {
		t.Errorf("Expected error containing 'agent cannot be empty', got: %s", output)
	}
}

// TestM05E02NoConfigUsesDefault validates default agent when no config exists.
func TestM05E02NoConfigUsesDefault(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	tmpDir := t.TempDir()

	// Need to create a minimal config with commands section for v2.0
	// The agent will use the default (claude) since no agent is specified in config
	configDir := setupConfigDir(t, tmpDir)
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := fmt.Sprintf(`commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, configDir, configDir, configDir)
	writeRawConfigFile(t, configPath, configContent)

	// Create command files
	createCommandFiles(t, configDir)

	// Use minimal iterations for fast test execution
	output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, nil, []string{"--fluxid-iterations=1"})
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	// v2.0: source tracking removed (Phase 9), just verify agent selection
	verifyAgent(t, output, "claude")
	verifyWorkflowSuccess(t, output)
}

// Helper functions for test setup and verification

func setupTestHome(t *testing.T, agent string) string {
	t.Helper()
	tmpHome := t.TempDir()
	if agent != "" {
		homeConfigDir := setupConfigDir(t, tmpHome)
		writeConfigFile(t, filepath.Join(homeConfigDir, "config.yaml"), agent)
	}
	return tmpHome
}

func setupTestProject(t *testing.T, agent string) string {
	t.Helper()
	tmpDir := t.TempDir()
	if agent != "" {
		projectConfigDir := setupConfigDir(t, tmpDir)
		writeConfigFile(t, filepath.Join(projectConfigDir, "config.yaml"), agent)
	}
	return tmpDir
}

// buildTestEnv removed in v2.0 (Phase 7: environment variable support removed)

func buildTestArgs(cliAgent string) []string {
	if cliAgent != "" {
		return []string{cliAgent}
	}
	return nil
}

func verifyAgent(t *testing.T, output, expectedAgent string) {
	t.Helper()
	expectedLine := "Agent: " + expectedAgent
	if !strings.Contains(output, expectedLine) {
		t.Errorf("Expected %s, got:\n%s", expectedLine, output)
	}
}

// verifySource and verifyAgentAndSource removed in v2.0 (Phase 9: source tracking removed)

func verifyWorkflowSuccess(t *testing.T, output string) {
	t.Helper()
	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Expected successful completion, got:\n%s", output)
	}
}
