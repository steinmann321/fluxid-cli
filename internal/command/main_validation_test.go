//nolint:paralleltest // Tests manipulate environment
package command

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

//nolint:funlen,gocognit,cyclop // Table-driven test with comprehensive test cases
func TestLoadAllConfigs(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
		t.Setenv("XDG_DATA_HOME", tmpDir)

		// Create valid config files in the expected location
		configDir := filepath.Join(tmpDir, ".fluxid")
		commandsDir := filepath.Join(configDir, "commands")
		if err := os.MkdirAll(commandsDir, 0o755); err != nil {
			t.Fatal(err)
		}

		homeConfig := `agent: claude
max_review_cycles: 10
`
		if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(homeConfig), 0o644); err != nil {
			t.Fatal(err)
		}

		home, _, exitCode := loadAllConfigs(nil)
		if exitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", exitCode)
		}
		if home == nil {
			t.Error("Expected home config to be loaded")
		}
		// Project config can be nil if no project config is set
	})

	t.Run("home config load error", func(t *testing.T) {
		// Set an invalid HOME path
		t.Setenv("HOME", "/dev/null/invalid/path/no/way/this/exists")
		t.Setenv("XDG_CONFIG_HOME", "/dev/null/invalid")

		_, _, exitCode := loadAllConfigs(nil)
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 for home config error, got %d", exitCode)
		}
	})

	t.Run("project config with valid directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
		t.Setenv("XDG_DATA_HOME", tmpDir)

		// Create home config in .fluxid directory
		configDir := filepath.Join(tmpDir, ".fluxid")
		commandsDir := filepath.Join(configDir, "commands")
		if err := os.MkdirAll(commandsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("agent: claude\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		home, _, exitCode := loadAllConfigs(nil)
		if exitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", exitCode)
		}
		if home == nil {
			t.Error("Expected home config to be loaded")
		}
		// Project config can be nil if no project config is set
	})

	t.Run("custom config path provided", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
		t.Setenv("XDG_DATA_HOME", tmpDir)

		// Create a custom config file
		customConfigPath := filepath.Join(tmpDir, "custom-config.yaml")
		customConfig := `agent: codex
iterations: 25
implement_retries: 5
commands:
  implement: /tmp/implement.md
  review: /tmp/review.md
  commit: /tmp/commit.md
`
		if err := os.WriteFile(customConfigPath, []byte(customConfig), 0o644); err != nil {
			t.Fatal(err)
		}

		// Create the command files referenced in the config
		for _, file := range []string{"/tmp/implement.md", "/tmp/review.md", "/tmp/commit.md"} {
			if err := os.WriteFile(file, []byte("# Command"), 0o644); err != nil {
				t.Fatal(err)
			}
			defer func(f string) { _ = os.Remove(f) }(file)
		}

		_, projectConfig, exitCode := loadAllConfigs(&customConfigPath)
		if exitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", exitCode)
		}
		if projectConfig == nil {
			t.Error("Expected project config to be loaded from custom config")
			return
		}
		if projectConfig.Agent == nil || *projectConfig.Agent != "codex" {
			t.Errorf("Expected agent to be 'codex', got %v", projectConfig.Agent)
		}
		if projectConfig.Iterations == nil || *projectConfig.Iterations != 25 {
			t.Errorf("Expected iterations to be 25, got %v", projectConfig.Iterations)
		}
	})

	t.Run("custom config path invalid", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		invalidPath := "/nonexistent/config.yaml"
		_, _, exitCode := loadAllConfigs(&invalidPath)
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 for invalid custom config path, got %d", exitCode)
		}
	})
}

// TestValidateAgent tests the validateAgent function for various scenarios.
func TestValidateAgent(t *testing.T) {
	t.Run("invalid agent name", func(t *testing.T) {
		exitCode := validateAgent("invalid-agent-xyz")
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 for invalid agent, got %d", exitCode)
		}
	})

	t.Run("valid agent not in PATH", func(t *testing.T) {
		// Set PATH to empty to ensure claude is not found
		t.Setenv("PATH", "")
		exitCode := validateAgent("claude")
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 when agent not in PATH, got %d", exitCode)
		}
	})

	t.Run("valid agent in PATH", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create an executable file with a valid agent name
		execPath := filepath.Join(tmpDir, "claude")
		if err := os.WriteFile(execPath, []byte("#!/bin/sh\necho test"), 0o755); err != nil {
			t.Fatal(err)
		}

		// Add to PATH
		t.Setenv("PATH", tmpDir)
		exitCode := validateAgent("claude")
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for valid agent in PATH, got %d", exitCode)
		}
	})

	t.Run("agent exists but not executable", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a non-executable file
		nonExecPath := filepath.Join(tmpDir, "claude")
		if err := os.WriteFile(nonExecPath, []byte("#!/bin/sh\necho test"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Add to PATH
		t.Setenv("PATH", tmpDir)
		exitCode := validateAgent("claude")
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 for non-executable agent, got %d", exitCode)
		}
	})
}

// TestLoadAndResolveConfig tests additional error paths.
func TestLoadAndResolveConfig(t *testing.T) {
	t.Run("load config error propagates", func(t *testing.T) {
		// Set invalid paths to trigger config load error
		t.Setenv("HOME", "/dev/null/invalid/home")
		t.Setenv("XDG_CONFIG_HOME", "/dev/null/invalid/config")

		_, exitCode := loadAndResolveConfig()
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 for config load error, got %d", exitCode)
		}
	})

	t.Run("successful config resolution", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
		t.Setenv("XDG_DATA_HOME", tmpDir)

		// Create valid config in .fluxid directory with commands section
		configDir := filepath.Join(tmpDir, ".fluxid")
		commandsDir := filepath.Join(configDir, "commands")
		if err := os.MkdirAll(commandsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		configContent := fmt.Sprintf(`agent: claude
commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, commandsDir, commandsDir, commandsDir)
		if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
			t.Fatal(err)
		}

		// Create command files in commandsDir (not configDir)
		if err := os.WriteFile(filepath.Join(commandsDir, "implement.md"), []byte("# Implement"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(commandsDir, "review.md"), []byte("# Review"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(commandsDir, "commit.md"), []byte("# Commit"), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, exitCode := loadAndResolveConfig()
		if exitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", exitCode)
		}
		if cfg.Agent == "" {
			t.Error("Expected resolved config to have agent set")
		}
	})
}
