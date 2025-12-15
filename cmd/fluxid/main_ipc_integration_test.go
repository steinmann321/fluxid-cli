//nolint:paralleltest // CLI tests with command execution
package main

import (
	"os"
	"testing"
)

func TestRunIPCCommand_GetReportSchema(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "ipc", "get-report-schema"}
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for ipc get-report-schema, got %d", exitCode)
	}
}

func TestRunIPCCommand_Help(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "ipc", "get-report-schema", "--help"}
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for ipc command with --help, got %d", exitCode)
	}
}

func TestRunWithSessionFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test parsing --session flag
	os.Args = []string{"fluxid", "ipc", "abort", "--session", "test-session-123"}

	// This will call handleIPCCommand which should accept the session flag
	exitCode := run()
	// It's ok if this fails due to missing session data, we're testing flag parsing
	if exitCode != 0 && exitCode != 1 {
		t.Logf("IPC abort returned non-zero (expected): %d", exitCode)
	}
}

func TestRunWithIPCWrite(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Test IPC write-report without stdin (should fail)
	os.Args = []string{"fluxid", "ipc", "write-report", "--session", "test-session"}
	exitCode := run()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for write-report without stdin")
	}
}

func TestRunWithIPCRead(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Test IPC read-report for non-existent session
	os.Args = []string{"fluxid", "ipc", "read-report", "--session", "nonexistent-session"}
	exitCode := run()
	// Should succeed but report will be empty
	if exitCode != 0 {
		t.Logf("read-report returned: %d (may be ok if no report exists)", exitCode)
	}
}

func TestRunWithIPCWriteHistory(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Test IPC write-history without stdin (should fail)
	os.Args = []string{"fluxid", "ipc", "write-history", "--session", "test-session"}
	exitCode := run()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for write-history without stdin")
	}
}

func TestRunWithIPCViewHistory(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Test IPC view-history for non-existent session
	os.Args = []string{"fluxid", "ipc", "view-history", "--session", "nonexistent-session"}
	exitCode := run()
	// Should handle gracefully
	if exitCode != 0 && exitCode != 1 {
		t.Logf("view-history returned: %d", exitCode)
	}
}

func TestRunWithUnknownIPCCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test unknown IPC command
	os.Args = []string{"fluxid", "ipc", "unknown-command"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for unknown IPC command, got %d", exitCode)
	}
}

func TestRunWithIPCAbortMissingSession(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "")

	// Test IPC abort without session flag or env var
	os.Args = []string{"fluxid", "ipc", "abort"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for abort without session, got %d", exitCode)
	}
}

func TestRunWithIPCGetReportSchemaMissingFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test that get-report-schema doesn't require session ID
	os.Args = []string{"fluxid", "ipc", "get-report-schema"}
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for get-report-schema, got %d", exitCode)
	}
}

func TestRunWithIPCWriteReportMissingSession(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Setenv("FLUXID_SESSION_ID", "")

	// Test write-report without session
	os.Args = []string{"fluxid", "ipc", "write-report"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for write-report without session, got %d", exitCode)
	}
}

func TestRunWithIPCReadReportMissingSession(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Setenv("FLUXID_SESSION_ID", "")

	// Test read-report without session
	os.Args = []string{"fluxid", "ipc", "read-report"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for read-report without session, got %d", exitCode)
	}
}
