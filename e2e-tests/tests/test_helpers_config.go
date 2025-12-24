package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	permDir  = 0o755 // Directory permissions for test fixtures
	permFile = 0o644 // File permissions for test fixtures
)

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

// writeConfigFile writes a config file with the given agent name.
func writeConfigFile(t *testing.T, configPath, agent string) {
	t.Helper()
	config := fmt.Sprintf("agent: %s\n", agent)
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
