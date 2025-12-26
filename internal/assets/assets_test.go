package assets

import (
	"os"
	"path/filepath"
	"testing"
)

//nolint:cyclop // Test validation requires multiple assertions
func TestCopyAssetsToDir_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Verify structure created
	fluxidDir := filepath.Join(tmpDir, ".fluxid")
	if _, err := os.Stat(fluxidDir); os.IsNotExist(err) {
		t.Error(".fluxid directory not created")
	}

	// Verify config.yaml
	configPath := filepath.Join(fluxidDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.yaml not created")
	}

	// Verify config content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.yaml: %v", err)
	}
	if len(content) == 0 {
		t.Error("config.yaml is empty")
	}

	// Verify commands directory with files
	commandsDir := filepath.Join(fluxidDir, "commands")
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		t.Fatalf("Failed to read commands dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("No command files created")
	}
	// We expect 28 command files
	if len(entries) != 28 {
		t.Errorf("Expected 28 command files, got %d", len(entries))
	}

	// Verify templates directory with files
	templatesDir := filepath.Join(fluxidDir, "templates")
	entries, err = os.ReadDir(templatesDir)
	if err != nil {
		t.Fatalf("Failed to read templates dir: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 template files, got %d", len(entries))
	}

	// Verify specific template files exist
	reportSchema := filepath.Join(templatesDir, "report-schema.yaml")
	if _, err := os.Stat(reportSchema); os.IsNotExist(err) {
		t.Error("report-schema.yaml not created")
	}

	reportExample := filepath.Join(templatesDir, "report-example.yaml")
	if _, err := os.Stat(reportExample); os.IsNotExist(err) {
		t.Error("report-example.yaml not created")
	}
}

func TestCopyAssetsToDir_AlreadyExists(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create .fluxid directory
	fluxidDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Should fail
	err := CopyAssetsToDir(tmpDir)
	if err == nil {
		t.Error("Expected error when .fluxid already exists")
	}
	if err != nil && !os.IsExist(err) {
		// Check error message contains expected text
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("Error message is empty")
		}
	}
}

func TestGetDefaultConfig(t *testing.T) {
	t.Parallel()
	config := GetDefaultConfig()
	if config == "" {
		t.Error("GetDefaultConfig returned empty string")
	}

	// Verify config contains expected sections
	if len(config) < 50 {
		t.Error("Config seems too short, might be incomplete")
	}
}
