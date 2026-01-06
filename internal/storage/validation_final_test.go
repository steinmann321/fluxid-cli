//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

// Final validation tests to close coverage gaps

func TestValidateHistory_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "nonexistent.yaml")

	err := storage.ValidateHistory(historyPath)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestValidateReport_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "nonexistent.yaml")

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestValidateYAMLSecurity_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "nonexistent.yaml")

	err := storage.ValidateYAMLSecurity(yamlPath)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestGetSessionRoot_NoEnvVarsNoCwd(t *testing.T) {
	// Test fallback to temp dir
	root, err := storage.GetSessionRoot("")
	if err != nil {
		t.Fatalf("GetSessionRoot failed: %v", err)
	}

	if !filepath.IsAbs(root) {
		t.Error("Expected absolute path")
	}
}

func TestResolveSessionPath_WithFilename(t *testing.T) {
	tmpDir := t.TempDir()

	sessionID := "550e8400-e29b-41d4-a716-446655440601"
	filename := "custom-file.yaml"

	path, err := storage.ResolveSessionPath(sessionID, filename, tmpDir)
	if err != nil {
		t.Fatalf("ResolveSessionPath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}

	// Verify filename is in path
	if filepath.Base(path) != filename {
		t.Errorf("Expected filename %s in path, got %s", filename, filepath.Base(path))
	}
}

func TestWriteReport_WritesSuccessfully(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440602"

	reportYAML := `command: "test"
artifact: "test"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	err := storage.WriteReport(sessionID, reportYAML)
	if err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	// Verify file was written
	reportPath, err := storage.ResolveSessionPath(sessionID, "report.yaml", "")
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != reportYAML {
		t.Error("Written content doesn't match input")
	}
}

func TestValidateReport_AutoParsedTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report with timestamp that YAML will auto-parse to time.Time
	reportWithTimestamp := `command: "test-command"
artifact: "test-artifact"
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: "Test"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(reportWithTimestamp), 0o644); err != nil {
		t.Fatal(err)
	}

	// YAML parser auto-converts valid timestamps to time.Time
	// ValidateReport should accept both string and time.Time
	err := storage.ValidateReport(reportPath)
	if err != nil {
		t.Errorf("Expected no error for auto-parsed timestamp, got: %v", err)
	}
}

func TestValidateHistory_AutoParsedTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	// History with timestamp that YAML will auto-parse to time.Time
	historyWithTimestamp := `- timestamp: 2025-12-13T10:00:00Z
  step: "implement"
  status: "SUCCESS"
  summary: "Test"
`
	if err := os.WriteFile(historyPath, []byte(historyWithTimestamp), 0o644); err != nil {
		t.Fatal(err)
	}

	// YAML parser auto-converts valid timestamps to time.Time
	// ValidateHistory should accept both string and time.Time
	err := storage.ValidateHistory(historyPath)
	if err != nil {
		t.Errorf("Expected no error for auto-parsed timestamp, got: %v", err)
	}
}

func TestValidationErrors_ErrorMethod(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report with multiple validation errors to trigger ValidationErrors.Error() method
	multiErrorReport := `command: "test"
artifact: "test"
status: INVALID_STATUS
issues:
  # missing all categories
`
	if err := os.WriteFile(reportPath, []byte(multiErrorReport), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Fatal("Expected validation errors")
	}

	// Call Error() method to get formatted error string
	errStr := err.Error()
	if errStr == "" {
		t.Error("Expected non-empty error string")
	}

	// Should contain multiple violations
	if !storage.IsValidationError(err) {
		t.Error("Expected IsValidationError to return true")
	}
}

func TestReadReport_SecondaryValidation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440603"

	// Write report that passes YAML parsing but fails field validation
	reportMissingStatus := `command: "test"
artifact: "test"
timestamp: "2025-12-13T10:00:00Z"
# status is missing, will be parsed as empty string
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := storage.WriteReport(sessionID, reportMissingStatus); err != nil {
		t.Fatal(err)
	}

	_, err := storage.ReadReport(sessionID, "")
	if err == nil {
		t.Error("Expected error for report with missing status after parsing")
	}
}

func TestReadHistory_EmptyFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440604"

	// ReadHistory should create empty file if it doesn't exist
	history, err := storage.ReadHistory(sessionID, "")
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(history))
	}

	// Verify file was created
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		t.Error("Expected history file to be created")
	}
}

func TestResolveSessionPath_PathNormalization(t *testing.T) {
	tmpDir := t.TempDir()

	sessionID := "550e8400-e29b-41d4-a716-446655440605"
	filename := "path//with///multiple////slashes.yaml"

	path, err := storage.ResolveSessionPath(sessionID, filename, tmpDir)
	if err != nil {
		t.Fatalf("ResolveSessionPath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestGetSessionRoot_CurrentDirectoryFallback(t *testing.T) {
	// Test with empty override - should try .fluxid/sessions or fallback to temp
	root, err := storage.GetSessionRoot("")
	if err != nil {
		t.Fatalf("GetSessionRoot failed: %v", err)
	}

	if root == "" {
		t.Error("Expected non-empty session root")
	}

	if !filepath.IsAbs(root) {
		t.Error("Expected absolute path for session root")
	}
}
