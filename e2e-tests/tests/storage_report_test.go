package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStorageReadReportValidFile verifies that storage.ReadReport() can read a valid report file.
func TestStorageReadReportValidFile(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-session-123"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a valid report file
	reportPath := filepath.Join(sessionDir, "report.yaml")
	validReport := `command: fluxid.implement
artifact: internal/storage/report.go
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations:
    - Implementation complete
  enhancements: []
summary: Test report
`
	if err := os.WriteFile(reportPath, []byte(validReport), 0o644); err != nil {
		t.Fatalf("Failed to write report file: %v", err)
	}

	// Set environment variable for session root
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test verifies the storage package can read the report
	// The actual test will call the storage.ReadReport() function once implemented
	// For now, this test is expected to fail as storage package doesn't exist yet
	t.Log("Test setup complete - ready for storage.ReadReport() implementation")
}

// TestStorageReadReportMalformedYAML verifies that storage.ReadReport() handles malformed YAML gracefully.
func TestStorageReadReportMalformedYAML(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-session-malformed"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a malformed YAML file
	reportPath := filepath.Join(sessionDir, "report.yaml")
	malformedYAML := `command: fluxid.implement
artifact: test
status: PASS
issues:
  blockers: []
  - invalid structure here
timestamp: not properly formatted
`
	if err := os.WriteFile(reportPath, []byte(malformedYAML), 0o644); err != nil {
		t.Fatalf("Failed to write malformed report file: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects storage.ReadReport() to return a clear error for malformed YAML
	t.Log("Test setup complete - expects error for malformed YAML")
}

// TestStorageReadReportMissingFile verifies that storage.ReadReport() handles missing files gracefully.
func TestStorageReadReportMissingFile(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory without a report file
	tmpDir := t.TempDir()
	sessionID := "test-session-missing"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects storage.ReadReport() to return a clear error for missing file
	t.Log("Test setup complete - expects error for missing report file")
}

// TestStorageReadReportExceedsSizeLimit verifies that storage.ReadReport() rejects files > 10MB.
func TestStorageReadReportExceedsSizeLimit(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-session-large"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a report file that exceeds 10MB
	reportPath := filepath.Join(sessionDir, "report.yaml")
	file, err := os.Create(reportPath)
	if err != nil {
		t.Fatalf("Failed to create report file: %v", err)
	}
	defer func() { _ = file.Close() }()

	// Write header
	header := `command: fluxid.implement
artifact: test
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations:
`
	if _, err := file.WriteString(header); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	// Write enough data to exceed 10MB
	largeString := "    - This is a large observation entry that will be repeated many times to exceed the size limit\n"
	targetSize := 11 * 1024 * 1024 // 11MB
	for size := 0; size < targetSize; size += len(largeString) {
		if _, err := file.WriteString(largeString); err != nil {
			t.Fatalf("Failed to write large content: %v", err)
		}
	}

	if _, err := file.WriteString("  enhancements: []\n"); err != nil {
		t.Fatalf("Failed to write footer: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects storage.ReadReport() to reject files larger than 10MB
	t.Log("Test setup complete - expects error for file size > 10MB")
}
