//nolint:paralleltest // CLI tests manipulate os.Args
package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHandleReadReportSuccess(t *testing.T) {
	// Set up environment
	sessionID := "read-test-session"
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Test that read-report runs without error when no report exists
	// Capture stdout, stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Run handleReadReport
	exitCode := handleReadReport([]string{})

	// Close writers and restore
	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read outputs
	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = stdoutBuf.ReadFrom(stdoutR)
	_, _ = stderrBuf.ReadFrom(stderrR)

	// Check exit code - should be 0 even if no report exists
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderrBuf.String())
	}
}

func TestHandleReadReportMissingSession(t *testing.T) {
	// Ensure no session ID is set
	_ = os.Unsetenv("FLUXID_SESSION_ID")

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := handleReadReport([]string{})

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing session, got %d", exitCode)
	}
}

func TestHandleReadReportHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := handleReadReport([]string{"-h"})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}

	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected help output to contain Usage, got: %s", output)
	}
}
