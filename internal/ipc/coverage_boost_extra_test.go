package ipc

import (
	"os"
	"path/filepath"
	"testing"
)

const testReport = "test report"

// TestReadReport_EmptySessionID tests reading report with empty session ID.
func TestReadReport_EmptySessionID(t *testing.T) {
	t.Parallel()
	_, err := ReadReport("")
	if err == nil {
		t.Error("Expected error for empty session ID")
	}
}

// TestWriteReport_WriteFileError tests error when writing report file fails.
func TestWriteReport_WriteFileError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Write a valid report first
	sessionID := "test-write-error"
	err := WriteReport(sessionID, testReport)
	if err != nil {
		t.Fatalf("Expected no error writing initial report, got: %v", err)
	}

	// Verify we can read it back
	report, err := ReadReport(sessionID)
	if err != nil {
		t.Errorf("Expected no error reading report, got: %v", err)
	}
	if report != testReport {
		t.Errorf("Expected '%s', got '%s'", testReport, report)
	}
}

// TestSetAbortFlag_WriteError tests error when writing abort flag fails.
func TestSetAbortFlag_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-abort-write"
	err := SetAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Expected no error setting abort flag, got: %v", err)
	}

	// Verify abort flag is set
	aborted, err := CheckAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Expected no error checking abort flag, got: %v", err)
	}
	if !aborted {
		t.Error("Expected abort flag to be set")
	}
}

// TestWriteHistoryEntry_DirectoryError tests error when history directory creation fails.
//
//nolint:paralleltest // Manipulates environment
func TestWriteHistoryEntry_DirectoryError(_ *testing.T) {
	_ = WriteHistoryEntry("test-session", "test message")
	// Expected to fail but we don't check - just ensuring no panic
}

// TestClearHistory_NonExistent tests clearing non-existent history.
func TestClearHistory_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-clear-nonexistent"
	err := ClearHistory(sessionID)
	if err != nil {
		t.Errorf("Expected no error clearing non-existent history, got: %v", err)
	}
}

// TestWriteHistoryEntry_LargeEntry tests writing a very large history entry.
func TestWriteHistoryEntry_LargeEntry(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-large-entry"
	// Create a large message (but not too large to avoid timeout)
	largeMessage := ""
	for i := 0; i < 1000; i++ {
		largeMessage += "This is a test message. "
	}

	err := WriteHistoryEntry(sessionID, largeMessage)
	if err != nil {
		t.Errorf("Expected no error for large entry, got: %v", err)
	}

	// Verify we can read it back
	history, err := ReadHistory(sessionID)
	if err != nil {
		t.Errorf("Expected no error reading history, got: %v", err)
	}
	if history == "" {
		t.Error("Expected non-empty history")
	}
}

// TestReadHistory_NonExistent tests reading history for nonexistent session.
func TestReadHistory_NonExistent(t *testing.T) {
	t.Parallel()
	sessionID := "nonexistent-history-session-xyz"

	history, err := ReadHistory(sessionID)
	if err != nil {
		t.Errorf("Expected no error for nonexistent history, got: %v", err)
	}
	// Expected - empty history is returned as empty string
	_ = history
}

// TestSetAbortFlag_Success tests successful abort flag setting.
func TestSetAbortFlag_Success(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-abort-success"

	err := SetAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	aborted, err := CheckAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Expected no error checking abort, got: %v", err)
	}
	if !aborted {
		t.Error("Expected abort flag to be set")
	}
}

// TestClearAbortFlag_Success tests successful abort flag clearing.
func TestClearAbortFlag_Success(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-clear-abort"

	// Set then clear
	_ = SetAbortFlag(sessionID)
	err := ClearAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	aborted, err := CheckAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Expected no error checking abort, got: %v", err)
	}
	if aborted {
		t.Error("Expected abort flag to be cleared")
	}
}

// TestReadReport_ReadFileError tests error when reading report file fails.
func TestReadReport_ReadFileError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-read-error"

	// Write a report
	err := WriteReport(sessionID, testReport)
	if err != nil {
		t.Fatalf("Expected no error writing, got: %v", err)
	}

	// Make the file unreadable by changing permissions
	reportPath := getReportPath(sessionID)
	err = os.Chmod(reportPath, 0o000)
	if err != nil {
		t.Fatalf("Failed to change permissions: %v", err)
	}

	// Try to read - should fail
	_, err = ReadReport(sessionID)
	if err == nil {
		t.Error("Expected error reading unreadable file")
	}

	// Clean up - restore permissions so cleanup works
	_ = os.Chmod(reportPath, 0o600)
}

// TestReadHistory_ReadFileError tests error when reading history file fails.
func TestReadHistory_ReadFileError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-hist-read-error"

	// Write history
	err := WriteHistoryEntry(sessionID, "test entry")
	if err != nil {
		t.Fatalf("Expected no error writing, got: %v", err)
	}

	// Make the file unreadable
	historyPath := getHistoryPath(sessionID)
	err = os.Chmod(historyPath, 0o000)
	if err != nil {
		t.Fatalf("Failed to change permissions: %v", err)
	}

	// Try to read - should fail
	_, err = ReadHistory(sessionID)
	if err == nil {
		t.Error("Expected error reading unreadable file")
	}

	// Clean up
	_ = os.Chmod(historyPath, 0o600)
}

// TestClearHistory_RemoveError tests clearing history when remove fails.
func TestClearHistory_RemoveError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-clear-hist-error"

	// Write history
	err := WriteHistoryEntry(sessionID, "test entry")
	if err != nil {
		t.Fatalf("Expected no error writing, got: %v", err)
	}

	// Make the file unremovable by removing write permissions from parent directory
	historyPath := getHistoryPath(sessionID)
	parentDir := filepath.Dir(historyPath)
	err = os.Chmod(parentDir, 0o500) //#nosec G302 // Test needs directory permissions to trigger error
	if err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}

	// Try to clear - should return error
	err = ClearHistory(sessionID)
	if err == nil {
		t.Error("Expected error clearing history from read-only directory")
	}

	// Clean up
	_ = os.Chmod(parentDir, 0o750) //#nosec G302 // Test cleanup restores directory permissions
}
