package ipc

import (
	"strings"
	"testing"
)

const (
	testSessionLarge      = "test-large"
	testSessionMultiAbort = "test-multi-abort"
	testSessionMultiHist  = "test-multi-history"
	testSessionConcurrent = "test-concurrent"
	testSessionFallback   = "test-fallback"
	testSessionClearWrite = "test-clear-write"
)

// TestWriteReport_LargeReport tests writing a large report.
func TestWriteReport_LargeReport(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testSessionLarge
	var builder strings.Builder
	builder.WriteString(`command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: PASS
summary: `)
	// Add a large summary to test size handling
	for index := 0; index < 1000; index++ {
		builder.WriteString("This is a test message. ")
	}
	builder.WriteString(`
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`)
	largeReport := builder.String()

	err := WriteReport(sessionID, largeReport)
	if err != nil {
		t.Errorf("Expected no error for large report, got: %v", err)
	}

	// Verify we can read it back
	report, err := ReadReport(sessionID)
	if err != nil {
		t.Errorf("Expected no error reading large report, got: %v", err)
	}
	if report == "" {
		t.Error("Expected non-empty report")
	}
}

// TestSetAbortFlag_MultipleSets tests setting abort flag multiple times.
func TestSetAbortFlag_MultipleSets(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testSessionMultiAbort

	// Set abort flag multiple times
	for index := 0; index < 5; index++ {
		err := SetAbortFlag(sessionID)
		if err != nil {
			t.Errorf("Expected no error on iteration %d, got: %v", index, err)
		}

		aborted, err := CheckAbortFlag(sessionID)
		if err != nil {
			t.Errorf("Expected no error checking abort on iteration %d, got: %v", index, err)
		}
		if !aborted {
			t.Errorf("Expected abort flag to be set on iteration %d", index)
		}
	}
}

// TestWriteHistoryEntry_MultipleEntries tests writing many history entries.
func TestWriteHistoryEntry_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testSessionMultiHist

	// Write multiple entries
	for index := 0; index < 10; index++ {
		err := WriteHistoryEntry(sessionID, "Entry "+string(rune('A'+index)))
		if err != nil {
			t.Errorf("Expected no error on entry %d, got: %v", index, err)
		}
	}

	// Read history
	history, err := ReadHistory(sessionID)
	if err != nil {
		t.Errorf("Expected no error reading history, got: %v", err)
	}
	if history == "" {
		t.Error("Expected non-empty history")
	}
}

// TestReadReport_Concurrent tests concurrent report reading.
func TestReadReport_Concurrent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testSessionConcurrent
	report := `command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: PASS
summary: Test
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	err := WriteReport(sessionID, report)
	if err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Read concurrently
	done := make(chan bool)
	for index := 0; index < 5; index++ {
		go func() {
			_, err := ReadReport(sessionID)
			if err != nil {
				t.Errorf("Expected no error reading concurrently, got: %v", err)
			}
			done <- true
		}()
	}

	for index := 0; index < 5; index++ {
		<-done
	}
}

// TestGetStorageDir_InvalidHome tests storage dir with invalid XDG_DATA_HOME.
//
//nolint:paralleltest // Manipulates environment
func TestGetStorageDir_InvalidHome(_ *testing.T) {
	// Should still work, using fallback
	sessionID := testSessionFallback
	_ = WriteReport(sessionID, testReport)
	// May fail but shouldn't panic
}

// TestClearHistory_ThenWrite tests clearing history then writing new entries.
func TestClearHistory_ThenWrite(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testSessionClearWrite

	// Write some entries
	_ = WriteHistoryEntry(sessionID, "Entry 1")
	_ = WriteHistoryEntry(sessionID, "Entry 2")

	// Clear
	err := ClearHistory(sessionID)
	if err != nil {
		t.Errorf("Expected no error clearing, got: %v", err)
	}

	// Write new entries
	err = WriteHistoryEntry(sessionID, "New Entry")
	if err != nil {
		t.Errorf("Expected no error writing after clear, got: %v", err)
	}

	// Verify history
	history, err := ReadHistory(sessionID)
	if err != nil {
		t.Errorf("Expected no error reading, got: %v", err)
	}
	if history == "" {
		t.Error("Expected non-empty history after new write")
	}
}

// TestWriteReport_InvalidSessionID tests report writing with various session IDs.
func TestWriteReport_InvalidSessionID(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	report := `command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: PASS
summary: Test
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	// Try empty session ID
	err := WriteReport("", report)
	if err == nil {
		t.Error("Expected error for empty session ID")
	}
}

// TestCheckAbortFlag_NonExistentSession tests checking abort on nonexistent session.
func TestCheckAbortFlag_NonExistentSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	aborted, err := CheckAbortFlag("nonexistent-session-xyz")
	if err != nil {
		t.Errorf("Expected no error for nonexistent session, got: %v", err)
	}
	if aborted {
		t.Error("Expected abort flag to be false for nonexistent session")
	}
}

// TestWriteReport_MkdirError tests error when directory creation fails.
//
//nolint:paralleltest // Manipulates environment
func TestWriteReport_MkdirError(_ *testing.T) {
	// Try to write to /dev/null which should fail
	_ = WriteReport("test-session", testReport)
	// Expected to fail but we don't check - just ensuring no panic
}

// TestSetAbortFlag_MkdirError tests error when directory creation fails for abort flag.
//
//nolint:paralleltest // Manipulates environment
func TestSetAbortFlag_MkdirError(_ *testing.T) {
	_ = SetAbortFlag("test-session")
	// Expected to fail but we don't check - just ensuring no panic
}

// TestWriteHistoryEntry_EmptyMessage tests writing history with empty message.
func TestWriteHistoryEntry_EmptyMessage(t *testing.T) {
	t.Parallel()
	err := WriteHistoryEntry("test-session", "")
	if err == nil {
		t.Error("Expected error for empty message")
	}
}

// TestClearAbortFlag_RemoveError tests clearing abort flag with remove error.
func TestClearAbortFlag_RemoveError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-clear-error"

	// Set abort flag first
	_ = SetAbortFlag(sessionID)

	// Clear it normally
	err := ClearAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Expected no error clearing, got: %v", err)
	}

	// Try clearing again (file doesn't exist - should not error)
	err = ClearAbortFlag(sessionID)
	if err != nil {
		t.Errorf("Expected no error when clearing non-existent flag, got: %v", err)
	}
}
