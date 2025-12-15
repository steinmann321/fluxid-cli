package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// setupM05TestEnv creates a test environment with optional home and project configs.
func setupM05TestEnv(t *testing.T, _ string, homeAgent, projectAgent string) (tmpHome, tmpProjectDir string) {
	t.Helper()

	tmpHome = t.TempDir()
	tmpProjectDir = t.TempDir()

	// Create home config if specified
	if homeAgent != "" {
		homeConfigContent := "agent: " + homeAgent + "\n"
		homeConfigPath := filepath.Join(tmpHome, ".fluxid", "config.yaml")
		if err := os.MkdirAll(filepath.Dir(homeConfigPath), 0o755); err != nil {
			t.Fatalf("Failed to create home .fluxid dir: %v", err)
		}
		if err := os.WriteFile(homeConfigPath, []byte(homeConfigContent), 0o644); err != nil {
			t.Fatalf("Failed to write home config: %v", err)
		}
	}

	// Create project config if specified
	if projectAgent != "" {
		projectConfigContent := "agent: " + projectAgent + "\n"
		projectConfigPath := filepath.Join(tmpProjectDir, ".fluxid", "config.yaml")
		if err := os.MkdirAll(filepath.Dir(projectConfigPath), 0o755); err != nil {
			t.Fatalf("Failed to create project .fluxid dir: %v", err)
		}
		if err := os.WriteFile(projectConfigPath, []byte(projectConfigContent), 0o644); err != nil {
			t.Fatalf("Failed to write project config: %v", err)
		}
	}

	return tmpHome, tmpProjectDir
}

// verifyAgentInInitSection checks that the agent line appears in the initialization section.
func verifyAgentInInitSection(t *testing.T, output, agentPattern string) {
	t.Helper()

	initIdx := indexOfString(output, "=== fluxid Workflow Initialization ===")
	agentIdx := indexOfString(output, agentPattern)
	workflowIdx := indexOfString(output, "Review Cycle")

	if initIdx == -1 {
		t.Fatalf("Initialization section not found in output")
	}
	if agentIdx == -1 {
		t.Fatalf("Agent line %q not found in output", agentPattern)
	}
	if workflowIdx == -1 {
		t.Fatalf("Workflow section not found in output")
	}

	if agentIdx < initIdx || agentIdx > workflowIdx {
		t.Errorf("Agent line not in initialization section (init=%d, agent=%d, workflow=%d)",
			initIdx, agentIdx, workflowIdx)
	}
}

// verifyAgentBeforeWorkflow checks that the agent line appears before the workflow section.
func verifyAgentBeforeWorkflow(t *testing.T, output, agentPattern string) {
	t.Helper()

	agentIdx := indexOfString(output, agentPattern)
	workflowIdx := indexOfString(output, "Review Cycle")

	if agentIdx == -1 {
		t.Fatalf("Agent line %q not found in output", agentPattern)
	}
	if workflowIdx == -1 {
		t.Fatalf("Workflow section not found in output")
	}

	if agentIdx >= workflowIdx {
		t.Errorf("Agent line must appear before workflow (agent=%d, workflow=%d)",
			agentIdx, workflowIdx)
	}
}

// indexOfString returns the index of substr in s, or -1 if not found.
func indexOfString(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
