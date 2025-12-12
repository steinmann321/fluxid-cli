//nolint:paralleltest // CLI tests manipulate os.Args
package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHandleWriteReportSuccess(t *testing.T) {
	// Set up environment
	sessionID := "write-test-session"
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Create valid report
	validReport := `command: test
artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	// Capture stdin, stdout, stderr
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdinR, stdinW, _ := os.Pipe()
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdin = stdinR
	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Write report to stdin
	_, _ = stdinW.WriteString(validReport)
	_ = stdinW.Close()

	// Run handleWriteReport
	exitCode := handleWriteReport([]string{})

	// Close writers and restore
	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read outputs
	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = stdoutBuf.ReadFrom(stdoutR)
	_, _ = stderrBuf.ReadFrom(stderrR)

	// Check exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderrBuf.String())
	}

	// Check success message
	stderrOutput := stderrBuf.String()
	if !strings.Contains(stderrOutput, "success") {
		t.Errorf("Expected success message in stderr, got: %s", stderrOutput)
	}
}

func TestHandleWriteReportMissingSession(t *testing.T) {
	// Ensure no session ID is set
	_ = os.Unsetenv("FLUXID_SESSION_ID")

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := handleWriteReport([]string{})

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing session, got %d", exitCode)
	}
}

func TestHandleWriteReportInvalidYAML(t *testing.T) {
	// Set up environment
	sessionID := "invalid-yaml-session"
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Create invalid YAML
	invalidYAML := `{this is not valid yaml`

	// Capture stdin, stderr
	oldStdin := os.Stdin
	oldStderr := os.Stderr

	stdinR, stdinW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdin = stdinR
	os.Stderr = stderrW

	// Write invalid YAML to stdin
	_, _ = stdinW.WriteString(invalidYAML)
	_ = stdinW.Close()

	// Run handleWriteReport
	exitCode := handleWriteReport([]string{})

	// Close and restore
	_ = stderrW.Close()
	os.Stdin = oldStdin
	os.Stderr = oldStderr

	// Read stderr
	var stderrBuf bytes.Buffer
	_, _ = stderrBuf.ReadFrom(stderrR)

	// Check exit code
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid YAML, got %d", exitCode)
	}
}

func TestHandleWriteReportHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := handleWriteReport([]string{"--help"})

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}

	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected help output to contain Usage, got: %s", output)
	}
}
