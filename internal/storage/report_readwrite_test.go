package storage_test

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteReport_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440001"
	invalidYAML := "command: [invalid"

	// WriteReport doesn't validate content, just writes it
	err := storage.WriteReport(sessionID, invalidYAML)
	if err != nil {
		t.Errorf("WriteReport should succeed (validation happens on read), got: %v", err)
	}
}

func TestWriteReport_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440002"

	// WriteReport doesn't validate content, just writes it
	err := storage.WriteReport(sessionID, "")
	if err != nil {
		t.Errorf("WriteReport should succeed for empty content, got: %v", err)
	}
}

func TestWriteReport_MissingStatus(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440003"
	reportWithoutStatus := `command: "test"
artifact: "test"
timestamp: "2025-12-13T10:00:00Z"
`

	// WriteReport doesn't validate content, just writes it
	err := storage.WriteReport(sessionID, reportWithoutStatus)
	if err != nil {
		t.Errorf("WriteReport should succeed (validation happens on read), got: %v", err)
	}
}

func TestReadReport_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440004"

	// Write a valid report first
	validReport := `command: "test"
artifact: "test"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: "Test"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps: []
`
	if err := storage.WriteReport(sessionID, validReport); err != nil {
		t.Fatal(err)
	}

	// Corrupt the file by overwriting with invalid content
	reportPath, err := storage.ResolveSessionPath(sessionID, "report.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reportPath, []byte("corrupted: [data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Try to read it
	_, err = storage.ReadReport(sessionID, "")
	if err == nil {
		t.Error("Expected error for corrupted file")
	}
}

func TestReadReport_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440005"

	// Write empty file directly
	reportPath, err := storage.ResolveSessionPath(sessionID, "report.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	sessionDir := filepath.Dir(reportPath)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reportPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	// ReadReport should return error for empty file
	_, err = storage.ReadReport(sessionID, "")
	if err == nil {
		t.Error("Expected error for empty file")
	}
}

func TestWriteAndReadReport_FAILStatus(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440006"

	failReport := `command: "test"
artifact: "test"
timestamp: "2025-12-13T10:00:00Z"
status: FAIL
summary: "Test failed"
issues:
  blockers:
    - message: "Critical issue"
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - "Fix the issue"
`

	err := storage.WriteReport(sessionID, failReport)
	if err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	report, err := storage.ReadReport(sessionID, "")
	if err != nil {
		t.Fatalf("ReadReport failed: %v", err)
	}

	if report.Status != "FAIL" {
		t.Errorf("Expected status FAIL, got %s", report.Status)
	}
}

func TestReadReport_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Use unique session ID to avoid conflict with other tests
	sessionID := "550e8400-e29b-41d4-a716-446655440099"

	_, err := storage.ReadReport(sessionID, "")
	// ReadReport returns error for non-existent file (tested in E2E)
	if err == nil {
		t.Error("Expected error for non-existent report")
	}
}

func TestWriteReport_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440007"

	validReport := `command: "test"
artifact: "test"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: "Test"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps: []
`

	err := storage.WriteReport(sessionID, validReport)
	if err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	// Verify report file was created
	reportPath, err := storage.ResolveSessionPath(sessionID, "report.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}
}
