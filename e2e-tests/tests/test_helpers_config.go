package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	permDir  = 0o755 // Directory permissions for test fixtures
	permFile = 0o644 // File permissions for test fixtures
)

// filterSessionIDFromEnv removes FLUXID_SESSION_ID from the environment slice.
// This is used in tests to ensure that child processes generate fresh session IDs
// rather than inheriting the parent process's session ID.
func filterSessionIDFromEnv(env []string) []string {
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, "FLUXID_SESSION_ID=") {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// setupConfigDir creates a .fluxid config directory and returns its path.
func setupConfigDir(t *testing.T, baseDir string) string {
	t.Helper()
	fluxidDir := filepath.Join(baseDir, ".fluxid")
	// #nosec G301 -- Test fixture directory with standard permissions
	if err := os.MkdirAll(fluxidDir, permDir); err != nil {
		t.Fatalf("Failed to create .fluxid directory: %v", err)
	}
	return fluxidDir
}

// writeConfigFile writes a config file with the given agent name and absolute paths.
func writeConfigFile(t *testing.T, configPath, agent string) {
	t.Helper()
	// Create command files in the same directory as the config
	configDir := filepath.Dir(configPath)
	createCommandFiles(t, configDir)

	// Create v2.0 config with commands section using absolute paths
	config := fmt.Sprintf(`agent: %s
commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, agent, configDir, configDir, configDir)
	// #nosec G306 -- Test fixture file with standard permissions
	if err := os.WriteFile(configPath, []byte(config), permFile); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
}

// writeRawConfigFile writes raw config content to the specified path.
func writeRawConfigFile(t *testing.T, configPath, content string) {
	t.Helper()
	// #nosec G306 -- Test fixture file with standard permissions
	if err := os.WriteFile(configPath, []byte(content), permFile); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
}

// setupConfigWithCommands creates a v2.0 config with required commands section and command files.
// Returns the config directory path.
//
//nolint:unparam // agent parameter maintained for API flexibility
func setupConfigWithCommands(t *testing.T, baseDir, agent string) string {
	t.Helper()
	configDir := setupConfigDir(t, baseDir)

	// Create command files first
	createCommandFiles(t, configDir)

	// Create config with commands section using absolute paths
	configContent := fmt.Sprintf(`agent: %s
commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, agent, configDir, configDir, configDir)
	configPath := filepath.Join(configDir, "config.yaml")
	writeRawConfigFile(t, configPath, configContent)

	return configDir
}

// createCommandFiles creates the required command files in the specified directory.
func createCommandFiles(t *testing.T, dir string) {
	t.Helper()
	commandFiles := map[string]string{
		"implement.md": "# Implement\nImplement the required changes.",
		"review.md":    "# Review\nReview the implementation.",
		"commit.md":    "# Commit\nCreate a commit with changes.",
	}

	for filename, content := range commandFiles {
		path := filepath.Join(dir, filename)
		// #nosec G306 -- Test fixture file with standard permissions
		if err := os.WriteFile(path, []byte(content), permFile); err != nil {
			t.Fatalf("Failed to write command file %s: %v", filename, err)
		}
	}
}

// setupHomeWithConfigAndCommands creates a home directory with v2.0 config and command files.
