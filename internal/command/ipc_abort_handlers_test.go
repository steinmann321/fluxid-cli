//nolint:paralleltest // IPC tests use os.Stdout/Stderr
package command

import (
	"bytes"
	"fluxid-cli/internal/ipc"
	"io"
	"os"
	"strings"
	"testing"
)

func TestHandleAbort_WithHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleAbort([]string{"--help"})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

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
	readPipe, writePipe, _ := os.Pipe()
	os.Stderr = writePipe

	exitCode := handleAbort([]string{"--session", "explicit-session-id"})

	_ = writePipe.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "Abort requested") {
		t.Error("Expected success message about abort request")
	}
}

func TestHandleAbort_WithShortHelp(t *testing.T) {
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleAbort([]string{"-h"})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h flag, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "abort") {
		t.Error("Expected help text for abort")
	}
}
