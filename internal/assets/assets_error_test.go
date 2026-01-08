package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyAssetsToDir_AlreadyExists(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create .fluxid directory
	fluxidDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Should fail
	_, err := CopyAssetsToDir(tmpDir)
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

func TestCopyAssetsToDir_InvalidPath(t *testing.T) {
	t.Parallel()

	// Try to copy to a path that doesn't exist and can't be created
	_, err := CopyAssetsToDir("/dev/null/invalid/path")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

func TestCopyAssetsToDir_ReadOnlyParent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a subdirectory and make it read-only
	roDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(roDir, 0o755); err != nil {
		t.Fatalf("Failed to create readonly dir: %v", err)
	}

	// Make it read-only (no write permission)
	if err := os.Chmod(roDir, 0o555); err != nil {
		t.Fatalf("Failed to chmod readonly dir: %v", err)
	}

	// Restore permissions after test
	defer func() {
		_ = os.Chmod(roDir, 0o755)
	}()

	// Try to copy assets to a subdirectory of the read-only directory
	_, err := CopyAssetsToDir(roDir)
	if err == nil {
		t.Error("Expected error when parent directory is read-only, got nil")
	}
}

func TestCopyAssetsToDir_FileExistsWhereDirectoryShouldBe(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a file named .fluxid (blocking directory creation)
	blockingFile := filepath.Join(tmpDir, ".fluxid")
	if err := os.WriteFile(blockingFile, []byte("blocker"), 0o644); err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}

	// Try to copy assets - should fail because .fluxid is a file, not a directory
	_, err := CopyAssetsToDir(tmpDir)
	if err == nil {
		t.Error("Expected error when .fluxid is a file, got nil")
	}

	// Error should mention that .fluxid already exists
	if err != nil && !containsStr(err.Error(), ".fluxid") {
		t.Errorf("Error message should mention .fluxid, got: %v", err)
	}
}

func TestReplacePlaceholdersBasic(t *testing.T) {
	t.Parallel()

	content := "path: {{FLUXID_DIR}}/commands"
	result := replacePlaceholders(content, "/test/path/.fluxid")

	// Verify placeholder was replaced
	if containsStr(result, "{{FLUXID_DIR}}") {
		t.Error("Placeholder was not replaced")
	}

	// Verify result contains the path
	if !containsStr(result, "/commands") {
		t.Errorf("Expected path to contain /commands, got: %s", result)
	}
}

func TestReplacePlaceholdersMultiple(t *testing.T) {
	t.Parallel()

	content := `
implement: {{FLUXID_DIR}}/commands/impl.md
review: {{FLUXID_DIR}}/commands/review.md
commit: {{FLUXID_DIR}}/commands/commit.md
`
	result := replacePlaceholders(content, "/home/user/.fluxid")

	// Verify all placeholders were replaced
	if containsStr(result, "{{FLUXID_DIR}}") {
		t.Error("Some placeholders were not replaced")
	}

	// Verify all paths are present
	if !containsStr(result, "/commands/impl.md") {
		t.Error("impl.md path missing")
	}
	if !containsStr(result, "/commands/review.md") {
		t.Error("review.md path missing")
	}
	if !containsStr(result, "/commands/commit.md") {
		t.Error("commit.md path missing")
	}
}

func TestReplacePlaceholdersRelativePath(t *testing.T) {
	t.Parallel()

	// Test with relative path - should convert to absolute
	content := "path: {{FLUXID_DIR}}"
	result := replacePlaceholders(content, ".fluxid")

	// Verify placeholder was replaced
	if containsStr(result, "{{FLUXID_DIR}}") {
		t.Error("Placeholder was not replaced")
	}

	// Result should not be exactly the input
	if result == content {
		t.Error("Content was not modified")
	}
}

func TestGetDefaultConfigStructure(t *testing.T) {
	t.Parallel()

	config := GetDefaultConfig()

	// Verify config has expected YAML keys (uncommented ones)
	expectedKeys := []string{
		"commands:",
		"implement:",
		"review:",
		"commit:",
	}

	for _, key := range expectedKeys {
		if !containsStr(config, key) {
			t.Errorf("Config missing expected key: %s", key)
		}
	}

	// Verify config has expected active fields with defaults
	requiredFields := map[string]string{
		"agent:":             "claude",
		"implement_retries:": "3",
		"commit_retries:":    "100",
		"iterations:":        "20",
	}

	for field, defaultValue := range requiredFields {
		if !containsStr(config, field) {
			t.Errorf("Config missing expected field: %s", field)
		}
		if !containsStr(config, defaultValue) {
			t.Errorf("Config missing expected default value: %s", defaultValue)
		}
	}
}

func TestGetDefaultConfigPlaceholders(t *testing.T) {
	t.Parallel()

	config := GetDefaultConfig()

	// Verify config contains placeholders
	if !containsStr(config, "{{FLUXID_DIR}}") {
		t.Error("Config does not contain {{FLUXID_DIR}} placeholder")
	}

	// Verify placeholders appear in command paths
	if !containsStr(config, "{{FLUXID_DIR}}/commands/") {
		t.Error("Config does not contain command path placeholders")
	}
}

func TestCopyAssetsToDir_CommandsCreated(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	counts, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Verify command files were created
	if counts.Commands == 0 {
		t.Error("No command files were copied")
	}

	// Verify commands directory exists
	commandsDir := filepath.Join(tmpDir, ".fluxid", "commands")
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		t.Error("Commands directory was not created")
	}

	// Verify at least one command file exists
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		t.Fatalf("Failed to read commands dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("No command files in commands directory")
	}
}

func TestCopyAssetsToDir_TemplatesCreated(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	counts, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Verify template files were created
	if counts.Templates == 0 {
		t.Error("No template files were copied")
	}

	// Verify templates directory exists
	templatesDir := filepath.Join(tmpDir, ".fluxid", "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		t.Error("Templates directory was not created")
	}

	// Verify template files exist
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		t.Fatalf("Failed to read templates dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("No template files in templates directory")
	}
}

func TestCopyAssetsToDir_ConfigYamlCreated(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Verify config.yaml exists
	configPath := filepath.Join(tmpDir, ".fluxid", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.yaml was not created")
	}

	// Verify config.yaml has content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.yaml: %v", err)
	}
	if len(content) == 0 {
		t.Error("config.yaml is empty")
	}

	// Verify placeholders were replaced in config
	configStr := string(content)
	if containsStr(configStr, "{{FLUXID_DIR}}") {
		t.Error("config.yaml still contains unreplaced placeholders")
	}
}
