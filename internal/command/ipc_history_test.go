package command

import (
	"fluxid-loop/internal/ipc"
	"os"
	"testing"
)

func TestHandleWriteHistory_Help(t *testing.T) {
	t.Parallel()
	exitCode := handleWriteHistory([]string{"--help"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}
}

func TestHandleWriteHistory_HelpShort(t *testing.T) {
	t.Parallel()
	exitCode := handleWriteHistory([]string{"-h"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}
}

func TestHandleWriteHistory_MissingSessionID(t *testing.T) {
	// Ensure no session ID is set
	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleWriteHistory([]string{"test message"})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when session ID is missing")
	}
}

func TestHandleWriteHistory_MissingMessage(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	exitCode := handleWriteHistory([]string{})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when message is missing")
	}
}

func TestHandleWriteHistory_Success(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	exitCode := handleWriteHistory([]string{"test", "message"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for successful write, got %d", exitCode)
	}
}

func TestHandleIPCWriteHistory_Help(t *testing.T) {
	t.Parallel()
	exitCode := handleIPCWriteHistory([]string{"--help"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}
}

func TestHandleIPCWriteHistory_HelpShort(t *testing.T) {
	t.Parallel()
	exitCode := handleIPCWriteHistory([]string{"-h"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}
}

func TestHandleIPCWriteHistory_MissingMessage(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	exitCode := handleIPCWriteHistory([]string{})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when message is missing")
	}
}

func TestHandleIPCWriteHistory_MissingMessageWithSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	exitCode := handleIPCWriteHistory([]string{"--session", "test-session"})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when message is missing")
	}
}

func TestHandleIPCWriteHistory_Success(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	exitCode := handleIPCWriteHistory([]string{"test", "message"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for successful write, got %d", exitCode)
	}
}

func TestHandleIPCWriteHistory_SuccessWithSessionFlag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	exitCode := handleIPCWriteHistory([]string{"test", "message", "--session", "test-session"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for successful write with session flag, got %d", exitCode)
	}
}

func TestHandleViewHistory_Help(t *testing.T) {
	t.Parallel()
	exitCode := handleViewHistory([]string{"--help"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}
}

func TestHandleViewHistory_HelpShort(t *testing.T) {
	t.Parallel()
	exitCode := handleViewHistory([]string{"-h"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}
}

func TestHandleViewHistory_EmptyHistory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	exitCode := handleViewHistory([]string{})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for empty history, got %d", exitCode)
	}
}

func TestHandleViewHistory_WithSessionFlag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	exitCode := handleViewHistory([]string{"--session", "test-session"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for view with session flag, got %d", exitCode)
	}
}

func TestHandleViewHistory_AfterWrite(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	// Write a history entry first
	writeExitCode := handleWriteHistory([]string{"test message"})
	if writeExitCode != 0 {
		t.Fatalf("Failed to write history: exit code %d", writeExitCode)
	}

	// Now view the history
	exitCode := handleViewHistory([]string{})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for view after write, got %d", exitCode)
	}
}

func TestHandleWriteHistory_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	// Write multiple entries
	messages := []string{"first message", "second message", "third message"}
	for _, msg := range messages {
		exitCode := handleWriteHistory([]string{msg})
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for writing '%s', got %d", msg, exitCode)
		}
	}
}

func TestExecute_WriteHistoryCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test --write-history through Execute
	os.Args = []string{"fluxid", "--write-history", "test message"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --write-history via Execute, got %d", exitCode)
	}
}

func TestExecute_IPCWriteHistory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test ipc write-history through Execute
	os.Args = []string{"fluxid", "ipc", "write-history", "test message"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for ipc write-history via Execute, got %d", exitCode)
	}
}

func TestExecute_IPCViewHistory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test ipc view-history through Execute
	os.Args = []string{"fluxid", "ipc", "view-history"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for ipc view-history via Execute, got %d", exitCode)
	}
}

func TestHandleViewHistory_WithHelp(t *testing.T) {
	t.Parallel()
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "ipc", "view-history", "--help"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for view-history --help, got %d", exitCode)
	}
}

func TestHandleViewHistory_MissingSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "") // No session ID

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "ipc", "view-history"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for missing session ID")
	}
}

func TestHandleViewHistory_WithSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Write some history first
	testSessionID := "specific-session-123"
	if err := ipc.WriteHistoryEntry(testSessionID, "test history entry"); err != nil {
		t.Fatal(err)
	}

	os.Args = []string{"fluxid", "ipc", "view-history", "--session", testSessionID}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for view-history with --session, got %d", exitCode)
	}
}
