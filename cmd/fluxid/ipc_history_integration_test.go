package main

import (
	"fluxid-loop/internal/ipc"
	"strings"
	"testing"
)

func TestHandleWriteHistory_MissingEnvSession(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleWriteHistory([]string{"test message"})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing session, got %d", exitCode)
	}
}

func TestHandleWriteHistory_NoMessage(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	exitCode := handleWriteHistory([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing message, got %d", exitCode)
	}
}

func TestHandleWriteHistory_WithMessage(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-history-write"
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	exitCode := handleWriteHistory([]string{"Test", "history", "entry"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify history
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}
	if !strings.Contains(history, "Test history entry") {
		t.Error("Expected history to contain written message")
	}
}

func TestHandleViewHistory_MissingEnvSession(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleViewHistory([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing session, got %d", exitCode)
	}
}

func TestHandleViewHistory_EmptyHistory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "empty-history-session"
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	exitCode := handleViewHistory([]string{})
	// Should succeed but show no history
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 even with empty history, got %d", exitCode)
	}
}

func TestHandleWriteHistory_HelpFlag(t *testing.T) {
	t.Parallel()
	exitCode := handleWriteHistory([]string{"--help"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help flag, got %d", exitCode)
	}
}

func TestHandleIPCWriteHistory_HelpFlag(t *testing.T) {
	t.Parallel()
	exitCode := handleIPCWriteHistory([]string{"--help"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help flag, got %d", exitCode)
	}
}

func TestHandleIPCWriteHistory_MissingSession(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleIPCWriteHistory([]string{"test message"})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing session, got %d", exitCode)
	}
}

func TestHandleIPCWriteHistory_NoMessage(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	exitCode := handleIPCWriteHistory([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing message, got %d", exitCode)
	}
}

func TestHandleIPCWriteHistory_WithSessionFlag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-ipc-history"

	exitCode := handleIPCWriteHistory([]string{"Test", "message", "--session", sessionID})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify history
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}
	if !strings.Contains(history, "Test message") {
		t.Error("Expected history to contain written message")
	}
}

func TestHandleIPCWriteHistory_SessionFlagAtEnd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-ipc-history-2"

	exitCode := handleIPCWriteHistory([]string{"Another", "test", "--session", sessionID})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify history
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}
	if !strings.Contains(history, "Another test") {
		t.Error("Expected history to contain written message")
	}
}

func TestHandleViewHistory_HelpFlag(t *testing.T) {
	t.Parallel()
	exitCode := handleViewHistory([]string{"--help"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help flag, got %d", exitCode)
	}
}

func TestHandleViewHistory_WithSessionFlag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-view-session"

	// Write some history first
	_ = ipc.WriteHistoryEntry(sessionID, "Test entry")

	exitCode := handleViewHistory([]string{"--session", sessionID})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestHandleViewHistory_WithHistory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-view-with-history"
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Write some history entries
	_ = ipc.WriteHistoryEntry(sessionID, "First entry")
	_ = ipc.WriteHistoryEntry(sessionID, "Second entry")

	exitCode := handleViewHistory([]string{})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
