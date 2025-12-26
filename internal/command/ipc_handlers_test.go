//nolint:paralleltest // IPC tests use os.Stdin/Stdout
package command

import (
	"bytes"
	"fluxid-cli/internal/ipc"
	"io"
	"os"
	"strings"
	"testing"
)

func TestHandleGetReportSchema_WithHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleGetReportSchema([]string{"--help"})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "get-report-schema") {
		t.Error("Expected help text for get-report-schema")
	}
}

func TestHandleGetReportSchema_Success(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleGetReportSchema([]string{})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if buf.Len() == 0 {
		t.Error("Expected schema output, got empty")
	}
}

func TestHandleWriteReport_WithHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleWriteReport([]string{"--help"})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "write-report") {
		t.Error("Expected help text for write-report")
	}
}

func TestHandleWriteReport_MissingSession(t *testing.T) {
	// Unset session ID
	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleWriteReport([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing session, got %d", exitCode)
	}
}

func TestHandleWriteReport_InvalidReport(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", "test-session")

	// Set stdin to invalid report
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	readPipe, writePipe, _ := os.Pipe()
	os.Stdin = readPipe
	_, _ = writePipe.WriteString("invalid: report without required fields\n")
	_ = writePipe.Close()

	exitCode := handleWriteReport([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid report, got %d", exitCode)
	}
}

func TestHandleWriteReport_ValidReport(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-write-session"

	// Set stdin to valid report
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	validReport := `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Test passed
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Continue
`

	readPipe, writePipe, _ := os.Pipe()
	os.Stdin = readPipe
	_, _ = writePipe.WriteString(validReport)
	_ = writePipe.Close()

	exitCode := handleWriteReport([]string{"--session", sessionID})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify report was stored
	storedReport, err := ipc.ReadReport(sessionID)
	if err != nil {
		t.Fatalf("Failed to read stored report: %v", err)
	}
	if storedReport != validReport {
		t.Error("Stored report doesn't match input")
	}
}

func TestHandleReadReport_WithHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleReadReport([]string{"--help"})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "read-report") {
		t.Error("Expected help text for read-report")
	}
}

func TestHandleReadReport_MissingSession(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")

	exitCode := handleReadReport([]string{})
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing session, got %d", exitCode)
	}
}

func TestHandleReadReport_NoReportFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "nonexistent-session"

	// Capture stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe
	os.Stderr = writePipe

	exitCode := handleReadReport([]string{"--session", sessionID})

	_ = writePipe.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 even when no report found, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "No report found") {
		t.Error("Expected message about no report found")
	}
}

func TestHandleReadReport_Success(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-read-session"

	// Write a report first
	testReport := `command: test
artifact: test.txt
timestamp: 2025-12-13T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := ipc.WriteReport(sessionID, testReport); err != nil {
		t.Fatalf("Failed to write test report: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleReadReport([]string{"--session", sessionID})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "status: PASS") {
		t.Error("Expected report content in output")
	}
}

func TestParseSessionFlag_WithFlag(t *testing.T) {
	sessionID, err := parseSessionFlag([]string{"--session", "test-123"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if sessionID != "test-123" {
		t.Errorf("Expected session ID 'test-123', got: %s", sessionID)
	}
}

func TestParseSessionFlag_MissingValue(t *testing.T) {
	_, err := parseSessionFlag([]string{"--session"})
	if err == nil {
		t.Error("Expected error for --session without value")
	}
	if !strings.Contains(err.Error(), "requires a value") {
		t.Errorf("Expected error about missing value, got: %v", err)
	}
}

func TestParseSessionFlag_FromEnv(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "env-session-456")

	sessionID, err := parseSessionFlag([]string{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if sessionID != "env-session-456" {
		t.Errorf("Expected session ID from env 'env-session-456', got: %s", sessionID)
	}
}

func TestParseSessionFlag_NoSessionAvailable(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")

	_, err := parseSessionFlag([]string{})
	if err == nil {
		t.Error("Expected error when no session available")
	}
	if !strings.Contains(err.Error(), "session ID not provided") {
		t.Errorf("Expected error about session not provided, got: %v", err)
	}
}

func TestHandleWriteReportWithSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set stdin to valid report
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	validReport := `command: test
artifact: test.txt
timestamp: 2025-12-13T10:00:00Z
status: FAIL
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	readPipe, writePipe, _ := os.Pipe()
	os.Stdin = readPipe
	_, _ = writePipe.WriteString(validReport)
	_ = writePipe.Close()

	exitCode := handleWriteReport([]string{"--session", "test-session-explicit"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestHandleGetReportSchema_WithShortHelp(t *testing.T) {
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleGetReportSchema([]string{"-h"})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h flag, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "get-report-schema") {
		t.Error("Expected help text for get-report-schema")
	}
}

func TestHandleWriteReport_WithShortHelp(t *testing.T) {
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleWriteReport([]string{"-h"})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h flag, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "write-report") {
		t.Error("Expected help text for write-report")
	}
}

func TestHandleReadReport_WithShortHelp(t *testing.T) {
	oldStdout := os.Stdout
	readPipe, writePipe, _ := os.Pipe()
	os.Stdout = writePipe

	exitCode := handleReadReport([]string{"-h"})

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, readPipe)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h flag, got %d", exitCode)
	}
	if !strings.Contains(buf.String(), "read-report") {
		t.Error("Expected help text for read-report")
	}
}
