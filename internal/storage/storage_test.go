//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

// Basic integration tests for storage package
// These tests verify the main API contracts without making assumptions about internal behavior

const (
	validReportYAML = `command: "test-command"
artifact: "test-artifact"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: "Test passed"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - "Continue"
`
	// Test session ID constant to avoid goconst violation.
	testStorageSessionID = "550e8400-e29b-41d4-a716-446655440000"
)

func TestReportWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testStorageSessionID

	// Write report
	err := storage.WriteReport(sessionID, validReportYAML)
	if err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	// Read it back
	report, err := storage.ReadReport(sessionID)
	if err != nil {
		t.Fatalf("ReadReport failed: %v", err)
	}

	if report == nil {
		t.Fatal("Expected non-nil report")
	}

	if report.Status != "PASS" {
		t.Errorf("Expected status PASS, got %s", report.Status)
	}
}

func TestHistoryRead(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testStorageSessionID

	// Pre-create empty history file
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	sessionDir := filepath.Dir(historyPath)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(historyPath, []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Read empty history
	history, err := storage.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(history))
	}
}

func TestResolveSessionPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testStorageSessionID

	path, err := storage.ResolveSessionPath(sessionID, "report.yaml", "")
	if err != nil {
		t.Fatalf("ResolveSessionPath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestValidateSessionID(t *testing.T) {
	validUUID := testStorageSessionID
	err := storage.ValidateSessionID(validUUID)
	if err != nil {
		t.Errorf("Expected no error for valid UUID, got: %v", err)
	}

	invalidUUID := "not-a-uuid"
	err = storage.ValidateSessionID(invalidUUID)
	if err == nil {
		t.Error("Expected error for invalid UUID")
	}
}

func TestGetSchemas(t *testing.T) {
	t.Parallel()

	reportSchema := storage.GetReportSchema()
	if reportSchema == "" {
		t.Error("Expected non-empty report schema")
	}

	historySchema := storage.GetHistorySchema()
	if historySchema == "" {
		t.Error("Expected non-empty history schema")
	}
}

func TestEnsureFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.yaml")

	err := storage.EnsureFileExists(filePath)
	if err != nil {
		t.Fatalf("EnsureFileExists failed: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File was not created")
	}
}

func TestGetSessionRoot(t *testing.T) {
	tmpDir := t.TempDir()

	// With override
	root, err := storage.GetSessionRoot(tmpDir)
	if err != nil {
		t.Fatalf("GetSessionRoot failed: %v", err)
	}

	if !filepath.IsAbs(root) {
		t.Error("Expected absolute path")
	}
}
