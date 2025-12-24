package tests

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestM05E02AgentFromHomeConfig validates agent selection from home config.
func TestM05E02AgentFromHomeConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome := t.TempDir()
	homeConfigDir := setupConfigDir(t, tmpHome)
	homeConfigPath := filepath.Join(homeConfigDir, "config.yaml")
	writeConfigFile(t, homeConfigPath, "codex")

	// Use minimal iterations for fast test execution
	output, err := runFluxidWithConfig(t, root, "", tmpHome, nil, []string{"--fluxid-iterations", "1"})
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	verifyAgentAndSource(t, output, "codex", "home", homeConfigPath)
	verifyWorkflowSuccess(t, output)
}

// TestM05E02AgentFromProjectConfig validates agent selection from project config.
func TestM05E02AgentFromProjectConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpDir := t.TempDir()
	projectConfigDir := setupConfigDir(t, tmpDir)
	projectConfigPath := filepath.Join(projectConfigDir, "config.yaml")
	writeConfigFile(t, projectConfigPath, "opencode")

	// Use minimal iterations for fast test execution
	output, err := runFluxidWithConfig(t, root, tmpDir, "", nil, []string{"--fluxid-iterations", "1"})
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	verifyAgentAndSource(t, output, "opencode", "project", projectConfigPath)
	verifyWorkflowSuccess(t, output)
}

// TestM05E02ProjectOverridesHome validates project config takes precedence over home.
func TestM05E02ProjectOverridesHome(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

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

	verifyAgentAndSource(t, output, "codex", "project", projectConfigPath)
}

// TestM05E02PrecedenceChain validates full precedence: CLI > env > project > home > default.
//
//nolint:funlen // E2E test with config precedence validation
func TestM05E02PrecedenceChain(t *testing.T) {
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
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpHome := setupTestHome(t, testCase.homeAgent)
			tmpDir := setupTestProject(t, testCase.projectAgent)
			env := buildTestEnv(testCase.envAgent)
			args := buildTestArgs(testCase.cliAgent)

			// Add dry-run flag to args (only need to verify agent selection)
			args = append(args, "--fluxid-dry-run")

			output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, env, args)
			if err != nil {
				t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
			}

			verifyAgent(t, output, testCase.expectedAgent)
			verifySource(t, output, testCase.expectedSource)
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

	// Use minimal iterations for fast test execution
	output, err := runFluxidWithConfig(t, root, tmpDir, tmpHome, nil, []string{"--fluxid-iterations", "1"})
	if err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output)
	}

	verifyAgent(t, output, "claude")
	verifySource(t, output, "source: default")
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

func buildTestEnv(envAgent string) []string {
	if envAgent != "" {
		return []string{"FLUXID_AGENT=" + envAgent}
	}
	return nil
}

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

func verifySource(t *testing.T, output, expectedSource string) {
	t.Helper()
	if !strings.Contains(output, expectedSource) {
		t.Errorf("Expected %s, got:\n%s", expectedSource, output)
	}
}

func verifyAgentAndSource(t *testing.T, output, agent, source, configPath string) {
	t.Helper()
	verifyAgent(t, output, agent)
	verifySource(t, output, "source: "+source)
	if !strings.Contains(output, configPath) {
		t.Errorf("Expected config path %s in output, got:\n%s", configPath, output)
	}
}

func verifyWorkflowSuccess(t *testing.T, output string) {
	t.Helper()
	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Expected successful completion, got:\n%s", output)
	}
}
