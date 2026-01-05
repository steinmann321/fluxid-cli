package assets

import (
	"os"
	"path/filepath"
	"testing"
)

//nolint:cyclop,funlen // Test validation requires multiple assertions
func TestCopyAssetsToDir_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	counts, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Verify counts
	if counts.Commands != 9 {
		t.Errorf("Expected 9 command files, got %d", counts.Commands)
	}
	if counts.Templates != 3 {
		t.Errorf("Expected 3 template files, got %d", counts.Templates)
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
	// We expect 9 command files (implement, review, commit variants + speckit variants)
	if len(entries) != 9 {
		t.Errorf("Expected 9 command files, got %d", len(entries))
	}

	// Verify templates directory with files
	templatesDir := filepath.Join(fluxidDir, "templates")
	entries, err = os.ReadDir(templatesDir)
	if err != nil {
		t.Fatalf("Failed to read templates dir: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 template files, got %d", len(entries))
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

	historySchema := filepath.Join(templatesDir, "history-schema.yaml")
	if _, err := os.Stat(historySchema); os.IsNotExist(err) {
		t.Error("history-schema.yaml not created")
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

func TestCopyAssetsToDir_FileContents(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Verify command files have content
	commandsDir := filepath.Join(tmpDir, ".fluxid", "commands")
	commandFiles := []string{
		"fluxid.implement.md",
		"fluxid.implement-cli.md",
		"fluxid.implement-e2e.md",
		"fluxid.review-implementation.md",
		"fluxid.review-implementation-e2e.md",
		"fluxid.commit.md",
	}

	for _, file := range commandFiles {
		path := filepath.Join(commandsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read %s: %v", file, err)
			continue
		}
		if len(content) == 0 {
			t.Errorf("Command file %s is empty", file)
		}
	}

	// Verify template files have content
	templatesDir := filepath.Join(tmpDir, ".fluxid", "templates")
	templateFiles := []string{"report-schema.yaml", "report-example.yaml", "history-schema.yaml"}

	for _, file := range templateFiles {
		path := filepath.Join(templatesDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read %s: %v", file, err)
			continue
		}
		if len(content) == 0 {
			t.Errorf("Template file %s is empty", file)
		}
	}
}

func TestCopyAssetsToDir_ConfigContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Read and verify config content
	configPath := filepath.Join(tmpDir, ".fluxid", "config.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.yaml: %v", err)
	}

	configStr := string(content)

	// Check for expected e2e default commands
	expectedStrings := []string{
		"fluxid.implement-e2e.md",
		"fluxid.review-implementation-e2e.md",
		"fluxid.commit.md",
	}

	for _, expected := range expectedStrings {
		if !containsStr(configStr, expected) {
			t.Errorf("Config does not contain expected string: %s", expected)
		}
	}
}
