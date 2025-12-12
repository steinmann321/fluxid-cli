package ipc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAndReadReport(t *testing.T) {
	t.Parallel()

	sessionID := "test-session-123"
	reportYAML := `command: test
artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	// Write report
	err := WriteReport(sessionID, reportYAML)
	if err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	// Read report back
	readReport, err := ReadReport(sessionID)
	if err != nil {
		t.Fatalf("ReadReport failed: %v", err)
	}

	if readReport != reportYAML {
		t.Errorf("Read report doesn't match written report.\nExpected:\n%s\nGot:\n%s", reportYAML, readReport)
	}

	// Clean up
	reportPath := getReportPath(sessionID)
	_ = os.Remove(reportPath)
}

func TestWriteReportEmptySessionID(t *testing.T) {
	t.Parallel()

	err := WriteReport("", "test report")
	if err == nil {
		t.Error("Expected error for empty session ID, got nil")
	}

	if err.Error() != "session ID cannot be empty" {
		t.Errorf("Expected 'session ID cannot be empty' error, got: %v", err)
	}
}

func TestReadReportEmptySessionID(t *testing.T) {
	t.Parallel()

	_, err := ReadReport("")
	if err == nil {
		t.Error("Expected error for empty session ID, got nil")
	}

	if err.Error() != "session ID cannot be empty" {
		t.Errorf("Expected 'session ID cannot be empty' error, got: %v", err)
	}
}

func TestReadReportNonExistent(t *testing.T) {
	t.Parallel()

	sessionID := "non-existent-session-999"
	report, err := ReadReport(sessionID)
	if err != nil {
		t.Fatalf("ReadReport failed for non-existent session: %v", err)
	}

	if report != "" {
		t.Errorf("Expected empty string for non-existent report, got: %s", report)
	}
}

func TestGetReportPath(t *testing.T) {
	t.Parallel()

	sessionID := "test-session"
	path := getReportPath(sessionID)

	expectedSuffix := filepath.Join("fluxid-reports", sessionID+".yaml")
	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got: %s", path)
	}

	if !strings.HasPrefix(path, os.TempDir()) {
		t.Errorf("Expected path to be in temp dir, got: %s", path)
	}

	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("Expected path to end with %s, got: %s", expectedSuffix, path)
	}
}
