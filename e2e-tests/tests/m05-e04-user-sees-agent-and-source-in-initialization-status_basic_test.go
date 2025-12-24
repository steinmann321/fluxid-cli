package tests

import (
	"strings"
	"testing"
)

// TestM05E04AgentAndSourceDisplayedViaCLI validates that agent selection
// via CLI flag shows "Agent: <name> (source: cli)" in initialization status.
func TestM05E04AgentAndSourceDisplayedViaCLI(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tests := []struct {
		name      string
		flag      string
		wantAgent string
	}{
		{
			name:      "claude via CLI",
			flag:      "--claude",
			wantAgent: "claude",
		},
		{
			name:      "codex via CLI",
			flag:      "--codex",
			wantAgent: "codex",
		},
		{
			name:      "opencode via CLI",
			flag:      "--opencode",
			wantAgent: "opencode",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "", "")

			output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir, testCase.flag)

			// Verify agent is displayed with source: cli
			expectedLine := "Agent: " + testCase.wantAgent + " (source: cli)"
			if !strings.Contains(output, expectedLine) {
				t.Errorf("Expected %q in output, got:\n%s", expectedLine, output)
			}

			verifyAgentInInitSection(t, output, expectedLine)
		})
	}
}

// TestM05E04AgentAndSourceDisplayedViaEnv validates that agent selection
// via environment variable shows "Agent: <name> (source: env)" in initialization status.
func TestM05E04AgentAndSourceDisplayedViaEnv(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tests := []struct {
		name      string
		envAgent  string
		wantAgent string
	}{
		{
			name:      "claude via env",
			envAgent:  "claude",
			wantAgent: "claude",
		},
		{
			name:      "codex via env",
			envAgent:  "codex",
			wantAgent: "codex",
		},
		{
			name:      "opencode via env",
			envAgent:  "opencode",
			wantAgent: "opencode",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "", "")

			envVars := map[string]string{
				"FLUXID_AGENT": testCase.envAgent,
			}
			output := runFluxidInDirWithEnv(t, root, tmpHome, tmpProjectDir, envVars)

			// Verify agent is displayed with source: env
			expectedLine := "Agent: " + testCase.wantAgent + " (source: env)"
			if !strings.Contains(output, expectedLine) {
				t.Errorf("Expected %q in output, got:\n%s", expectedLine, output)
			}

			verifyAgentBeforeWorkflow(t, output, expectedLine)
		})
	}
}

// TestM05E04AgentAndSourceDisplayedViaHomeConfig validates that agent selection
// via home config shows "Agent: <name> (source: home (<path>))" in initialization status.
func TestM05E04AgentAndSourceDisplayedViaHomeConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tests := []struct {
		name      string
		agent     string
		wantAgent string
	}{
		{
			name:      "claude via home config",
			agent:     "claude",
			wantAgent: "claude",
		},
		{
			name:      "codex via home config",
			agent:     "codex",
			wantAgent: "codex",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpHome, tmpProjectDir := setupM05TestEnv(t, root, testCase.agent, "")

			output := runFluxidInDirWithOutput(t, root, tmpHome, tmpProjectDir)

			// Verify agent is displayed with source: home (<path>)
			if !strings.Contains(output, "Agent: "+testCase.wantAgent) {
				t.Errorf("Agent name not found in output:\n%s", output)
			}

			if !strings.Contains(output, "(source: home") {
				t.Errorf("Source 'home' not found in output:\n%s", output)
			}

			agentPattern := "Agent: " + testCase.wantAgent
			verifyAgentBeforeWorkflow(t, output, agentPattern)
		})
	}
}

// TestM05E04AgentAndSourceDisplayedViaProjectConfig validates that agent selection
// via project config shows "Agent: <name> (source: project (<path>))" in initialization status.
func TestM05E04AgentAndSourceDisplayedViaProjectConfig(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tests := []struct {
		name      string
		agent     string
		wantAgent string
	}{
		{
			name:      "claude via project config",
			agent:     "claude",
			wantAgent: "claude",
		},
		{
			name:      "opencode via project config",
			agent:     "opencode",
			wantAgent: "opencode",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "", testCase.agent)

			output := runFluxidInDirWithOutput(t, root, tmpHome, tmpProjectDir)

			// Verify agent is displayed with source: project (<path>)
			if !strings.Contains(output, "Agent: "+testCase.wantAgent) {
				t.Errorf("Agent name not found in output:\n%s", output)
			}

			if !strings.Contains(output, "(source: project") {
				t.Errorf("Source 'project' not found in output:\n%s", output)
			}

			agentPattern := "Agent: " + testCase.wantAgent
			verifyAgentBeforeWorkflow(t, output, agentPattern)
		})
	}
}

// TestM05E04AgentAndSourceDisplayedViaDefault validates that when no agent
// is configured, the default agent is shown with "source: default".
func TestM05E04AgentAndSourceDisplayedViaDefault(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tmpHome, tmpProjectDir := setupM05TestEnv(t, root, "", "")

	output := runFluxidInDirWithOutput(t, root, tmpHome, tmpProjectDir)

	// Verify default agent (claude) is displayed with source: default
	expectedLine := "Agent: claude (source: default)"
	if !strings.Contains(output, expectedLine) {
		t.Errorf("Expected %q in output, got:\n%s", expectedLine, output)
	}

	verifyAgentBeforeWorkflow(t, output, expectedLine)
}
