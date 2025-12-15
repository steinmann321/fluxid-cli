package ipc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSession = "test-session"

func TestWriteAndReadHistory(t *testing.T) {
	t.Parallel()
	sessionID := "test-history-session-123"
	message1 := "First history entry"
	message2 := "Second history entry"
	// Write first entry
	err := WriteHistoryEntry(sessionID, message1)
	if err != nil {
		t.Fatalf("WriteHistoryEntry failed: %v", err)
	}
	// Read history
	history, err := ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}
	if !strings.Contains(history, message1) {
		t.Errorf("History missing first entry, got: %s", history)
	}
	// Verify timestamp format (ISO 8601 with Z suffix)
	if !strings.Contains(history, "[20") || !strings.Contains(history, "Z]") {
		t.Errorf("History missing ISO 8601 timestamp, got: %s", history)
	}
	// Write second entry
	err = WriteHistoryEntry(sessionID, message2)
	if err != nil {
		t.Fatalf("WriteHistoryEntry failed for second entry: %v", err)
	}
	// Read history again
	history, err = ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed after second write: %v", err)
	}
	// Verify both entries are present
	if !strings.Contains(history, message1) {
		t.Errorf("History missing first entry after second write, got: %s", history)
	}
	if !strings.Contains(history, message2) {
		t.Errorf("History missing second entry, got: %s", history)
	}
	_ = ClearHistory(sessionID)
}

func TestWriteHistoryEntryEmptySessionID(t *testing.T) {
	t.Parallel()

	err := WriteHistoryEntry("", "test message")
	if err == nil {
		t.Error("Expected error for empty session ID, got nil")
	}

	if err.Error() != errSessionIDEmpty {
		t.Errorf("Expected 'session ID cannot be empty' error, got: %v", err)
	}
}

func TestWriteHistoryEntryEmptyMessage(t *testing.T) {
	t.Parallel()

	sessionID := testSession
	err := WriteHistoryEntry(sessionID, "")
	if err == nil {
		t.Error("Expected error for empty message, got nil")
	}

	if !strings.Contains(err.Error(), "message cannot be empty") {
		t.Errorf("Expected 'message cannot be empty' error, got: %v", err)
	}
}

func TestReadHistoryEmptySessionID(t *testing.T) {
	t.Parallel()

	_, err := ReadHistory("")
	if err == nil {
		t.Error("Expected error for empty session ID, got nil")
	}

	if err.Error() != errSessionIDEmpty {
		t.Errorf("Expected 'session ID cannot be empty' error, got: %v", err)
	}
}

func TestReadHistoryNonExistent(t *testing.T) {
	t.Parallel()

	sessionID := "non-existent-history-session-999"
	history, err := ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed for non-existent session: %v", err)
	}

	if history != "" {
		t.Errorf("Expected empty string for non-existent history, got: %s", history)
	}
}

func TestClearHistory(t *testing.T) {
	t.Parallel()

	sessionID := "test-clear-history-123"
	message := "Test entry"

	// Write history first
	err := WriteHistoryEntry(sessionID, message)
	if err != nil {
		t.Fatalf("WriteHistoryEntry failed: %v", err)
	}

	// Verify it was written
	history, err := ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}
	if !strings.Contains(history, message) {
		t.Error("History entry was not written")
	}

	// Clear history
	err = ClearHistory(sessionID)
	if err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}

	// Verify it's cleared
	history, err = ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed after clearing: %v", err)
	}
	if history != "" {
		t.Errorf("Expected empty history after clearing, got: %s", history)
	}
}

func TestClearHistoryEmptySession(t *testing.T) {
	t.Parallel()

	err := ClearHistory("")
	if err == nil {
		t.Error("Expected error for empty session ID, got nil")
	}

	if err.Error() != errSessionIDEmpty {
		t.Errorf("Expected 'session ID cannot be empty' error, got: %v", err)
	}
}

func TestClearHistoryNonExistent(t *testing.T) {
	t.Parallel()

	sessionID := "non-existent-history-999"

	// Clearing non-existent history should not error
	err := ClearHistory(sessionID)
	if err != nil {
		t.Errorf("ClearHistory should not error for non-existent history: %v", err)
	}
}

func TestGetHistoryPath(t *testing.T) {
	t.Parallel()

	sessionID := testSession
	path := getHistoryPath(sessionID)

	expectedSuffix := filepath.Join("fluxid-reports", sessionID+".history")
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

func TestFormatISO8601(t *testing.T) {
	t.Parallel()

	timestamp := formatISO8601()

	// Verify format: YYYY-MM-DDTHH:MM:SSZ
	if len(timestamp) != 20 {
		t.Errorf("Expected timestamp length 20, got %d: %s", len(timestamp), timestamp)
	}

	if !strings.HasSuffix(timestamp, "Z") {
		t.Errorf("Expected timestamp to end with 'Z', got: %s", timestamp)
	}

	if !strings.Contains(timestamp, "T") {
		t.Errorf("Expected timestamp to contain 'T', got: %s", timestamp)
	}

	// Verify it starts with a valid year (20XX)
	if !strings.HasPrefix(timestamp, "20") {
		t.Errorf("Expected timestamp to start with '20', got: %s", timestamp)
	}
}

func TestWriteHistoryEntry_FIFOEviction(t *testing.T) {
	t.Parallel()

	sessionID := "test-eviction-session-456"

	// Write large messages that will trigger eviction
	// Each message should be > 11MB to ensure we exceed 32MB with 3 messages
	largeMessage := strings.Repeat("A", 12*1024*1024) // 12MB

	// Write first message
	err := WriteHistoryEntry(sessionID, largeMessage)
	if err != nil {
		t.Fatalf("WriteHistoryEntry failed for first message: %v", err)
	}

	// Write second message
	err = WriteHistoryEntry(sessionID, largeMessage)
	if err != nil {
		t.Fatalf("WriteHistoryEntry failed for second message: %v", err)
	}

	// Write third message - this should trigger eviction of first
	err = WriteHistoryEntry(sessionID, largeMessage)
	if err != nil {
		t.Fatalf("WriteHistoryEntry failed for third message: %v", err)
	}

	// Read history - should be under 32MB with oldest evicted
	history, err := ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	historySize := len([]byte(history))
	if historySize > maxHistorySize {
		t.Errorf("History size %d exceeds max %d after eviction", historySize, maxHistorySize)
	}

	// Verify history is not empty
	if history == "" {
		t.Error("History should not be empty after eviction")
	}

	// Clean up
	_ = ClearHistory(sessionID)
}

func TestWriteReport_DirPermissions(t *testing.T) {
	t.Parallel()

	sessionID := "test-report-permissions"
	reportYAML := "test: report"

	// Write report
	err := WriteReport(sessionID, reportYAML)
	if err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	// Verify file has correct permissions (0600)
	reportPath := getReportPath(sessionID)
	info, err := os.Stat(reportPath)
	if err != nil {
		t.Fatalf("Stat failed for report file: %v", err)
	}

	perm := info.Mode().Perm()
	expectedPerm := os.FileMode(0o600)
	if perm != expectedPerm {
		t.Errorf("Expected report file permissions %o, got %o", expectedPerm, perm)
	}
}

func TestReadReport_StatError(t *testing.T) {
	t.Parallel()

	sessionID := "test-read-stat-error"

	// Reading non-existent report should return empty string
	report, err := ReadReport(sessionID)
	if err != nil {
		t.Errorf("ReadReport should not error for non-existent report: %v", err)
	}
	if report != "" {
		t.Errorf("Expected empty report for non-existent session, got: %s", report)
	}
}

func TestCheckAbortFlag_StatError(t *testing.T) {
	t.Parallel()

	sessionID := "test-abort-stat-error"

	// Checking non-existent abort flag should return false
	aborted, err := CheckAbortFlag(sessionID)
	if err != nil {
		t.Errorf("CheckAbortFlag should not error for non-existent flag: %v", err)
	}
	if aborted {
		t.Error("Expected aborted to be false for non-existent flag")
	}
}

func TestWriteHistoryEntry_EmptyHistory(t *testing.T) {
	t.Parallel()

	sessionID := "test-empty-history"
	message := "First message in empty history"

	// Write to empty history
	err := WriteHistoryEntry(sessionID, message)
	if err != nil {
		t.Fatalf("WriteHistoryEntry failed: %v", err)
	}

	// Verify it was written
	history, err := ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	if !strings.Contains(history, message) {
		t.Errorf("History missing message: %s", history)
	}

	_ = ClearHistory(sessionID)
}

func TestWriteReport_EmptySessionID(t *testing.T) {
	t.Parallel()

	err := WriteReport("", "test: data")
	if err == nil {
		t.Error("Expected error for empty session ID")
	}

	if !strings.Contains(err.Error(), "session ID cannot be empty") {
		t.Errorf("Expected 'session ID cannot be empty' error, got: %v", err)
	}
}

func TestSetAbortFlag_EmptySessionID(t *testing.T) {
	t.Parallel()

	err := SetAbortFlag("")
	if err == nil {
		t.Error("Expected error for empty session ID")
	}

	if !strings.Contains(err.Error(), "session ID cannot be empty") {
		t.Errorf("Expected 'session ID cannot be empty' error, got: %v", err)
	}
}

func TestCheckAbortFlag_EmptySessionID(t *testing.T) {
	t.Parallel()

	_, err := CheckAbortFlag("")
	if err == nil {
		t.Error("Expected error for empty session ID")
	}

	if !strings.Contains(err.Error(), "session ID cannot be empty") {
		t.Errorf("Expected 'session ID cannot be empty' error, got: %v", err)
	}
}

func TestClearAbortFlag_EmptySessionID(t *testing.T) {
	t.Parallel()

	err := ClearAbortFlag("")
	if err == nil {
		t.Error("Expected error for empty session ID")
	}

	if !strings.Contains(err.Error(), "session ID cannot be empty") {
		t.Errorf("Expected 'session ID cannot be empty' error, got: %v", err)
	}
}
