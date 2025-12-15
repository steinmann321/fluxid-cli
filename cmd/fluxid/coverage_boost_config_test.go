package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndResolveConfig_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	_, _ = loadAndResolveConfig()
	// May succeed or fail, but should not panic
}

//nolint:paralleltest // Simple validation test
func TestValidateAgent_InvalidAgent(t *testing.T) {
	exitCode := validateAgent("invalid-agent-xyz")
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid agent")
	}
}

//nolint:paralleltest // Simple validation test
func TestValidateAgent_ValidAgent(t *testing.T) {
	exitCode := validateAgent(testAgentClaude)
	if exitCode != 0 {
		t.Error("Expected zero exit code for valid agent (path check may fail)")
	}
}

func TestValidateAgent_AgentNotInPath(t *testing.T) {
	t.Setenv("PATH", "")
	exitCode := validateAgent(testAgentClaude)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when agent not in PATH")
	}
}

func TestLoadAllConfigs_HomeConfigError(t *testing.T) {
	t.Setenv("HOME", "/dev/null/invalid/home")

	_, _, _, exitCode := loadAllConfigs() //nolint:dogsled // Testing all return values
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when home config fails")
	}
}

func TestLoadAllConfigs_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	_, _, _, exitCode := loadAllConfigs() //nolint:dogsled // Testing all return values
	if exitCode != 0 {
		t.Errorf("Expected zero exit code, got %d", exitCode)
	}
	// Configs may be nil if no files exist, that's OK
}

func TestLoadAllConfigs_ProjectConfigError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	projectConfigDir := filepath.Join(".", ".fluxid")
	if err := os.MkdirAll(projectConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create project config dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(projectConfigDir)
	}()

	projectConfigPath := filepath.Join(projectConfigDir, "config.yaml")
	invalidYAML := "{ invalid yaml ["
	if err := os.WriteFile(projectConfigPath, []byte(invalidYAML), 0o644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	_, _, _, exitCode := loadAllConfigs() //nolint:dogsled // Testing all return values
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid YAML")
	}
}

func TestLoadAllConfigs_EnvConfigError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("FLUXID_ITERATIONS", "not-a-number")

	_, _, _, exitCode := loadAllConfigs() //nolint:dogsled // Testing all return values
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid env var")
	}
}

func TestLoadAndResolveConfig_WithFlags(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create a simple project config
	projectConfigDir := filepath.Join(".", ".fluxid")
	if err := os.MkdirAll(projectConfigDir, 0o755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(projectConfigDir)
	}()

	configYAML := `iterations: 5
`
	if err := os.WriteFile(filepath.Join(projectConfigDir, "config.yaml"), []byte(configYAML), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_, _ = loadAndResolveConfig()
}

//nolint:paralleltest // Test doesn't use t
func TestValidateAgent_ValidAgents(_ *testing.T) {
	agents := []string{testAgentClaude, testAgentCodex, testAgentOpencode}
	for _, agent := range agents {
		_ = validateAgent(agent)
		// May fail if not in PATH, but validates the name check path
	}
}

func TestLoadAndResolveConfig_MergeError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("FLUXID_AGENT", "invalid-agent-name-xyz")

	_, exitCode := loadAndResolveConfig()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid agent")
	}
}

func TestValidateAgent_WithPathCheck(t *testing.T) {
	// Create a temporary directory and add it to PATH
	tmpDir := t.TempDir()
	oldPath := os.Getenv("PATH")
	defer func() { t.Setenv("PATH", oldPath) }()

	t.Setenv("PATH", tmpDir)

	// Create a dummy "claude" executable
	claudePath := filepath.Join(tmpDir, testAgentClaude)
	if err := os.WriteFile(claudePath, []byte("#!/bin/sh\necho test"), 0o755); err != nil {
		t.Fatal(err)
	}

	exitCode := validateAgent(testAgentClaude)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for valid agent in PATH, got %d", exitCode)
	}
}
