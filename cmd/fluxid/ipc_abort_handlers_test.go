//nolint:paralleltest // IPC tests use os.Stdout/Stderr
package main

import (
	"bytes"
	"fluxid-loop/internal/ipc"
	"io"
	"os"
	"strings"
	"testing"
)

func TestHandleAbort_WithHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := handleAbort([]string{"--help"})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "abort") {
		t.Error("Expected help text for abort")
	}
}

func TestHandleAbort_MissingSession(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleAbort([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing session, got %d", exitCode)
	}
}

func TestHandleAbort_Success(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-abort-session"

	exitCode := handleAbort([]string{"--session", sessionID})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify abort flag was set
	aborted, err := ipc.CheckAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("Failed to check abort flag: %v", err)
	}
	if !aborted {
		t.Error("Expected abort flag to be set")
	}
}

func TestHandleAbort_WithSessionFlag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Capture stderr to verify message
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := handleAbort([]string{"--session", "explicit-session-id"})

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "Abort requested") {
		t.Error("Expected success message about abort request")
	}
}
