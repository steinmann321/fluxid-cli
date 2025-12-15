package main

import (
	"os"
	"testing"
)

func TestHandleGetReportSchema_HelpShortFlag(t *testing.T) {
	t.Parallel()
	exitCode := handleGetReportSchema([]string{"-h"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h flag, got %d", exitCode)
	}
}

func TestHandleWriteReport_HelpShortFlag(t *testing.T) {
	t.Parallel()
	exitCode := handleWriteReport([]string{"-h"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h flag, got %d", exitCode)
	}
}

func TestHandleReadReport_HelpShortFlag(t *testing.T) {
	t.Parallel()
	exitCode := handleReadReport([]string{"-h"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h flag, got %d", exitCode)
	}
}

func TestHandleWriteReport_StdinReadError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-stdin-error")

	// Close stdin to cause read error
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r
	_ = w.Close() // Close write end to simulate EOF

	_ = handleWriteReport([]string{})
	// Should succeed with empty input (EOF)
}

func TestHandleGetReportSchema_NoArgs(t *testing.T) {
	t.Parallel()
	exitCode := handleGetReportSchema([]string{})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestHandleWriteReport_NoSessionID(t *testing.T) {
	oldEnv := os.Getenv("FLUXID_SESSION_ID")
	defer func() {
		if oldEnv != "" {
			t.Setenv("FLUXID_SESSION_ID", oldEnv)
		} else {
			_ = os.Unsetenv("FLUXID_SESSION_ID")
		}
	}()

	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleWriteReport([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 when no session ID, got %d", exitCode)
	}
}

func TestHandleReadReport_NoSessionID(t *testing.T) {
	oldEnv := os.Getenv("FLUXID_SESSION_ID")
	defer func() {
		if oldEnv != "" {
			t.Setenv("FLUXID_SESSION_ID", oldEnv)
		} else {
			_ = os.Unsetenv("FLUXID_SESSION_ID")
		}
	}()

	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleReadReport([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 when no session ID, got %d", exitCode)
	}
}

func TestHandleAbort_NoSessionID(t *testing.T) {
	oldEnv := os.Getenv("FLUXID_SESSION_ID")
	defer func() {
		if oldEnv != "" {
			t.Setenv("FLUXID_SESSION_ID", oldEnv)
		} else {
			_ = os.Unsetenv("FLUXID_SESSION_ID")
		}
	}()

	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleAbort([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 when no session ID, got %d", exitCode)
	}
}

func TestHandleViewHistory_NoSessionID(t *testing.T) {
	oldEnv := os.Getenv("FLUXID_SESSION_ID")
	defer func() {
		if oldEnv != "" {
			t.Setenv("FLUXID_SESSION_ID", oldEnv)
		} else {
			_ = os.Unsetenv("FLUXID_SESSION_ID")
		}
	}()

	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleViewHistory([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 when no session ID, got %d", exitCode)
	}
}
