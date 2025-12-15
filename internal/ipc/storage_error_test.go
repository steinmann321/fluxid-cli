//nolint:paralleltest // IPC storage tests require sequential execution
package ipc

import (
	"os"
	"testing"
)

// TestWriteReport_ErrorPaths tests error scenarios in WriteReport.
func TestWriteReport_ErrorPaths(t *testing.T) {
	err := WriteReport("", "test: data")
	if err == nil {
		t.Error("Expected error when writing with empty session ID")
	}
}

// TestSetAbortFlag_ErrorPaths tests error scenarios in SetAbortFlag.
func TestSetAbortFlag_ErrorPaths(t *testing.T) {
	err := SetAbortFlag("")
	if err == nil {
		t.Error("Expected error when setting abort flag with empty session ID")
	}
}

// TestCheckAbortFlag_ErrorPaths tests error scenarios in CheckAbortFlag.
func TestCheckAbortFlag_ErrorPaths(t *testing.T) {
	_, err := CheckAbortFlag("")
	if err == nil {
		t.Error("Expected error when checking abort flag with empty session ID")
	}
}

// TestClearAbortFlag_ErrorPaths tests error scenarios in ClearAbortFlag.
func TestClearAbortFlag_ErrorPaths(t *testing.T) {
	err := ClearAbortFlag("")
	if err == nil {
		t.Error("Expected error when clearing abort flag with empty session ID")
	}
}

// TestReadReport_NonExistent tests reading a non-existent report.
func TestReadReport_NonExistent(t *testing.T) {
	sessionID := "nonexistent-session-xyz-123"

	content, err := ReadReport(sessionID)
	if err != nil {
		t.Errorf("Unexpected error when reading non-existent report: %v", err)
	}
	if content != "" {
		t.Error("Expected empty string for non-existent report")
	}
}

// TestWriteHistoryEntry_ErrorPaths tests error scenarios in WriteHistoryEntry.
func TestWriteHistoryEntry_ErrorPaths(t *testing.T) {
	err := WriteHistoryEntry("", "test message")
	if err == nil {
		t.Error("Expected error when writing history with empty session ID")
	}
}

// TestClearHistory_ErrorPaths tests error scenarios in ClearHistory.
func TestClearHistory_ErrorPaths(t *testing.T) {
	err := ClearHistory("")
	if err == nil {
		t.Error("Expected error when clearing history with empty session ID")
	}

	sessionID := "nonexistent-history-xyz-789"
	err = ClearHistory(sessionID)
	// Should not error if history doesn't exist.
	if err != nil {
		t.Errorf("Unexpected error when clearing non-existent history: %v", err)
	}
}

// TestReadHistory_ErrorPaths tests error scenarios in ReadHistory.
func TestReadHistory_ErrorPaths(t *testing.T) {
	_, err := ReadHistory("")
	if err == nil {
		t.Error("Expected error when reading history with empty session ID")
	}

	sessionID := "nonexistent-read-history-456"
	entries, err := ReadHistory(sessionID)
	if err != nil {
		t.Errorf("Unexpected error when reading non-existent history: %v", err)
	}
	if len(entries) != 0 {
		t.Error("Expected empty entries for non-existent history")
	}
}

// TestWriteReport_Success tests successful report writing.
func TestWriteReport_Success(t *testing.T) {
	sessionID := "test-write-success"
	reportContent := "command: test\nstatus: PASS"

	err := WriteReport(sessionID, reportContent)
	if err != nil {
		t.Errorf("Unexpected error writing report: %v", err)
	}

	// Verify we can read it back.
	content, err := ReadReport(sessionID)
	if err != nil {
		t.Errorf("Unexpected error reading report: %v", err)
	}
	if content != reportContent {
		t.Errorf("Expected %q, got %q", reportContent, content)
	}

	// Clean up.
	reportPath := getReportPath(sessionID)
	_ = os.Remove(reportPath)
}

// TestCheckAbortFlag_SetAndCheck tests setting and checking abort flag.
func TestCheckAbortFlag_SetAndCheck(t *testing.T) {
	sessionID := "test-abort-set-check"

	// Initially should not be aborted.
	aborted, err := CheckAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Unexpected error checking abort flag: %v", err)
	}
	if aborted {
		t.Error("Expected abort flag to be false initially")
	}

	// Set the abort flag.
	err = SetAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Unexpected error setting abort flag: %v", err)
	}

	// Now should be aborted.
	aborted, err = CheckAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Unexpected error checking abort flag after set: %v", err)
	}
	if !aborted {
		t.Error("Expected abort flag to be true after setting")
	}

	// Clean up.
	_ = ClearAbortFlag(sessionID)
}
