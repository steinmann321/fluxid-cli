package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//nolint:cyclop // Test requires multiple assertions
func TestCopyAssetsToDir_VerifyPermissions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Verify directories are created with correct permissions
	fluxidDir := filepath.Join(tmpDir, ".fluxid")
	info, err := os.Stat(fluxidDir)
	if err != nil {
		t.Fatalf("Failed to stat .fluxid directory: %v", err)
	}
	if !info.IsDir() {
		t.Error(".fluxid is not a directory")
	}

	commandsDir := filepath.Join(fluxidDir, "commands")
	info, err = os.Stat(commandsDir)
	if err != nil {
		t.Fatalf("Failed to stat commands directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("commands is not a directory")
	}

	templatesDir := filepath.Join(fluxidDir, "templates")
	info, err = os.Stat(templatesDir)
	if err != nil {
		t.Fatalf("Failed to stat templates directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("templates is not a directory")
	}

	// Verify config file exists
	configPath := filepath.Join(fluxidDir, "config.yaml")
	info, err = os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config.yaml: %v", err)
	}
	if info.IsDir() {
		t.Error("config.yaml is a directory, expected file")
	}
	if info.Size() == 0 {
		t.Error("config.yaml is empty")
	}
}

func TestCopyAssetsToDir_CreatesAllDirectories(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Check all expected directories exist
	dirs := []string{
		filepath.Join(tmpDir, ".fluxid"),
		filepath.Join(tmpDir, ".fluxid", "commands"),
		filepath.Join(tmpDir, ".fluxid", "templates"),
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Directory does not exist: %s: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("Path is not a directory: %s", dir)
		}
	}
}

func TestCopyAssetsToDir_WritesConfigFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".fluxid", "config.yaml")

	// Read config and verify placeholders were replaced
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)
	// Verify placeholders were replaced (should NOT contain {{FLUXID_DIR}})
	if strings.Contains(contentStr, "{{FLUXID_DIR}}") {
		t.Error("Config still contains unreplaced placeholder {{FLUXID_DIR}}")
	}

	// Verify it contains absolute paths
	absFluxidPath := filepath.Join(tmpDir, ".fluxid")
	if !strings.Contains(contentStr, absFluxidPath) {
		t.Errorf("Config should contain absolute path %s", absFluxidPath)
	}
}

func TestCopyAssetsToDir_CopiesTemplateFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	templatesDir := filepath.Join(tmpDir, ".fluxid", "templates")

	// Check both template files
	files := []string{"report-schema.yaml", "report-example.yaml"}
	for _, file := range files {
		path := filepath.Join(templatesDir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Template file missing: %s: %v", file, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("Template file is empty: %s", file)
		}
	}
}

func TestCopyAssetsToDir_CopiesCommandFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	commandsDir := filepath.Join(tmpDir, ".fluxid", "commands")

	// Check all 6 command files
	files := []string{
		"fluxid.commit.md",
		"fluxid.implement.md",
		"fluxid.implement-cli.md",
		"fluxid.implement-e2e.md",
		"fluxid.review-implementation.md",
		"fluxid.review-implementation-e2e.md",
	}

	for _, file := range files {
		path := filepath.Join(commandsDir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Command file missing: %s: %v", file, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("Command file is empty: %s", file)
		}
	}
}

//nolint:cyclop,funlen // Test verification requires checking all files
func TestCopyAssetsToDir_VerifyAllFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Verify all 6 command files are copied
	commandsDir := filepath.Join(tmpDir, ".fluxid", "commands")
	expectedFiles := map[string]bool{
		"fluxid.commit.md":                    false,
		"fluxid.commit-speckit.md":            false,
		"fluxid.implement-cli.md":             false,
		"fluxid.implement-e2e.md":             false,
		"fluxid.implement.md":                 false,
		"fluxid.implement-speckit.md":         false,
		"fluxid.review-implementation-e2e.md": false,
		"fluxid.review-implementation.md":     false,
		"fluxid.review-speckit.md":            false,
	}

	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		t.Fatalf("Failed to read commands directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			if _, exists := expectedFiles[entry.Name()]; exists {
				expectedFiles[entry.Name()] = true
			} else {
				t.Errorf("Unexpected file in commands directory: %s", entry.Name())
			}
		}
	}

	for file, found := range expectedFiles {
		if !found {
			t.Errorf("Expected command file not found: %s", file)
		}
	}

	// Verify template files
	templatesDir := filepath.Join(tmpDir, ".fluxid", "templates")
	templateFiles := map[string]bool{
		"report-schema.yaml":  false,
		"report-example.yaml": false,
	}

	entries, err = os.ReadDir(templatesDir)
	if err != nil {
		t.Fatalf("Failed to read templates directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			if _, exists := templateFiles[entry.Name()]; exists {
				templateFiles[entry.Name()] = true
			} else {
				t.Errorf("Unexpected file in templates directory: %s", entry.Name())
			}
		}
	}

	for file, found := range templateFiles {
		if !found {
			t.Errorf("Expected template file not found: %s", file)
		}
	}
}

func TestGetDefaultConfig_Content(t *testing.T) {
	t.Parallel()

	config := GetDefaultConfig()

	// Verify it contains the expected e2e command references
	if !containsStr(config, "implement") {
		t.Error("Config does not contain 'implement'")
	}
	if !containsStr(config, "review") {
		t.Error("Config does not contain 'review'")
	}
	if !containsStr(config, "commit") {
		t.Error("Config does not contain 'commit'")
	}
	if !containsStr(config, "fluxid.implement-e2e.md") {
		t.Error("Config does not contain default e2e implement command")
	}
	if !containsStr(config, "fluxid.review-implementation-e2e.md") {
		t.Error("Config does not contain default e2e review command")
	}
}

//nolint:cyclop,funlen // Test verification requires multiple file checks
func TestCopyAssetsToDir_VerifyFileContents(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := CopyAssetsToDir(tmpDir)
	if err != nil {
		t.Fatalf("CopyAssetsToDir failed: %v", err)
	}

	// Read each command file and verify it has content
	commandsDir := filepath.Join(tmpDir, ".fluxid", "commands")
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		t.Fatalf("Failed to read commands directory: %v", err)
	}

	if len(entries) != 9 {
		t.Errorf("Expected 9 command files, got %d", len(entries))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			t.Errorf("Found directory in commands: %s", entry.Name())
			continue
		}

		path := filepath.Join(commandsDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read %s: %v", entry.Name(), err)
			continue
		}

		if len(content) < 10 {
			t.Errorf("File %s seems too small (%d bytes)", entry.Name(), len(content))
		}

		// Verify it's a markdown file
		if len(entry.Name()) < 3 || entry.Name()[len(entry.Name())-3:] != ".md" {
			t.Errorf("Command file %s doesn't have .md extension", entry.Name())
		}
	}

	// Read template files and verify content
	templatesDir := filepath.Join(tmpDir, ".fluxid", "templates")
	entries, err = os.ReadDir(templatesDir)
	if err != nil {
		t.Fatalf("Failed to read templates directory: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 template files, got %d", len(entries))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			t.Errorf("Found directory in templates: %s", entry.Name())
			continue
		}

		path := filepath.Join(templatesDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read %s: %v", entry.Name(), err)
			continue
		}

		if len(content) < 10 {
			t.Errorf("File %s seems too small (%d bytes)", entry.Name(), len(content))
		}

		// Verify it's a YAML file
		if len(entry.Name()) < 5 || entry.Name()[len(entry.Name())-5:] != ".yaml" {
			t.Errorf("Template file %s doesn't have .yaml extension", entry.Name())
		}
	}
}

func TestCopyAssetsToDir_MultipleInvocations(t *testing.T) {
	t.Parallel()

	// Test that calling CopyAssetsToDir multiple times in different dirs works
	for invocation := 0; invocation < 3; invocation++ {
		tmpDir := t.TempDir()

		_, err := CopyAssetsToDir(tmpDir)
		if err != nil {
			t.Errorf("Invocation %d failed: %v", invocation, err)
		}

		// Verify creation
		fluxidDir := filepath.Join(tmpDir, ".fluxid")
		if _, err := os.Stat(fluxidDir); err != nil {
			t.Errorf("Invocation %d: .fluxid not created: %v", invocation, err)
		}
	}
}
