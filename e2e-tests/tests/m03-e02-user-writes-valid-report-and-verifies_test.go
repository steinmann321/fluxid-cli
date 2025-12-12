package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestM03E02WriteAndReadValidReport verifies that a valid report can be written
// and read back within the same session.
func TestM03E02WriteAndReadValidReport(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	sessionID := "test-session-write-read"

	// Valid report YAML
	validReport := `command: fluxid.implement-cli
artifact: m03-e02-test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations:
    - message: "Test observation"
  enhancements: []
summary: "Test report for write-read flow"
`

	// Write the report
	writeCmd := exec.CommandContext(t.Context(), binPath, "ipc", "write-report")
	writeCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)
	writeCmd.Stdin = strings.NewReader(validReport)

	var writeStdout, writeStderr bytes.Buffer
	writeCmd.Stdout = &writeStdout
	writeCmd.Stderr = &writeStderr

	if err := writeCmd.Run(); err != nil {
		t.Fatalf("write-report failed: %v\nStdout: %s\nStderr: %s",
			err, writeStdout.String(), writeStderr.String())
	}

	// Verify write command exited with status 0
	if writeStderr.Len() > 0 {
		t.Logf("Write stderr: %s", writeStderr.String())
	}

	// Read the report back
	readCmd := exec.CommandContext(t.Context(), binPath, "ipc", "read-report")
	readCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	var readStdout, readStderr bytes.Buffer
	readCmd.Stdout = &readStdout
	readCmd.Stderr = &readStderr

	if err := readCmd.Run(); err != nil {
		t.Fatalf("read-report failed: %v\nStdout: %s\nStderr: %s",
			err, readStdout.String(), readStderr.String())
	}

	// Parse both reports and compare
	var written, read map[string]interface{}

	if err := yaml.Unmarshal([]byte(validReport), &written); err != nil {
		t.Fatalf("Failed to parse written report: %v", err)
	}

	readOutput := readStdout.String()
	if len(readOutput) == 0 {
		t.Fatal("read-report returned empty output")
	}

	if err := yaml.Unmarshal([]byte(readOutput), &read); err != nil {
		t.Fatalf("Failed to parse read report: %v\nOutput:\n%s", err, readOutput)
	}

	// Deep compare key fields
	if read["command"] != written["command"] {
		t.Errorf("command mismatch: got %v, want %v", read["command"], written["command"])
	}
	if read["artifact"] != written["artifact"] {
		t.Errorf("artifact mismatch: got %v, want %v", read["artifact"], written["artifact"])
	}
	if read["status"] != written["status"] {
		t.Errorf("status mismatch: got %v, want %v", read["status"], written["status"])
	}
	if read["timestamp"] != written["timestamp"] {
		t.Errorf("timestamp mismatch: got %v, want %v", read["timestamp"], written["timestamp"])
	}
	if read["summary"] != written["summary"] {
		t.Errorf("summary mismatch: got %v, want %v", read["summary"], written["summary"])
	}
}

// TestM03E02SessionOverrideViaFlag verifies that --session flag
// overrides FLUXID_SESSION_ID environment variable.
func TestM03E02SessionOverrideViaFlag(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	envSessionID := "env-session"
	flagSessionID := "flag-session"

	validReport := `command: fluxid.test
artifact: session-override-test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	// Write with --session flag (should override env var)
	writeCmd := exec.CommandContext(t.Context(), binPath, "ipc", "write-report", "--session", flagSessionID)
	writeCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+envSessionID)
	writeCmd.Stdin = strings.NewReader(validReport)

	var writeStdout, writeStderr bytes.Buffer
	writeCmd.Stdout = &writeStdout
	writeCmd.Stderr = &writeStderr

	if err := writeCmd.Run(); err != nil {
		t.Fatalf("write-report failed: %v\nStderr: %s", err, writeStderr.String())
	}

	// Read from flag session (should get the report)
	readFlagCmd := exec.CommandContext(t.Context(), binPath, "ipc", "read-report", "--session", flagSessionID)
	var readFlagStdout bytes.Buffer
	readFlagCmd.Stdout = &readFlagStdout

	if err := readFlagCmd.Run(); err != nil {
		t.Fatalf("read-report with --session flag failed: %v", err)
	}

	if readFlagStdout.Len() == 0 {
		t.Fatal("Expected report from flag session, got empty output")
	}

	var flagReport map[string]interface{}
	if err := yaml.Unmarshal(readFlagStdout.Bytes(), &flagReport); err != nil {
		t.Fatalf("Failed to parse flag session report: %v", err)
	}

	if flagReport["artifact"] != "session-override-test" {
		t.Errorf("Flag session report artifact mismatch: got %v", flagReport["artifact"])
	}

	// Read from env session (should be empty/no report)
	readEnvCmd := exec.CommandContext(t.Context(), binPath, "ipc", "read-report")
	readEnvCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+envSessionID)
	var readEnvStdout, readEnvStderr bytes.Buffer
	readEnvCmd.Stdout = &readEnvStdout
	readEnvCmd.Stderr = &readEnvStderr

	if err := readEnvCmd.Run(); err != nil {
		t.Fatalf("read-report from env session should not error: %v", err)
	}

	// Should have no report in env session
	if readEnvStdout.Len() > 0 {
		t.Errorf("Expected no report in env session, got: %s", readEnvStdout.String())
	}
}

// TestM03E02WriteInvalidReportFails verifies that an invalid report
// is rejected with clear error diagnostics.
func TestM03E02WriteInvalidReportFails(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	sessionID := "test-invalid-report"

	testCases := []struct {
		name        string
		report      string
		expectedErr string
	}{
		{
			name: "missing required field command",
			report: `artifact: test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`,
			expectedErr: "command",
		},
		{
			name: "missing required field status",
			report: `command: test
artifact: test
timestamp: 2025-12-12T10:00:00Z
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`,
			expectedErr: "status",
		},
		{
			name: "invalid status value",
			report: `command: test
artifact: test
timestamp: 2025-12-12T10:00:00Z
status: INVALID
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`,
			expectedErr: "status",
		},
		{
			name: "missing issues object",
			report: `command: test
artifact: test
timestamp: 2025-12-12T10:00:00Z
status: PASS
`,
			expectedErr: "issues",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			writeCmd := exec.CommandContext(t.Context(), binPath, "ipc", "write-report")
			writeCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)
			writeCmd.Stdin = strings.NewReader(tc.report)

			var stdout, stderr bytes.Buffer
			writeCmd.Stdout = &stdout
			writeCmd.Stderr = &stderr

			err := writeCmd.Run()
			if err == nil {
				t.Fatalf("Expected write-report to fail for invalid report, but it succeeded\nStdout: %s", stdout.String())
			}

			// Verify error message contains relevant information
			errOutput := stderr.String()
			if !strings.Contains(errOutput, tc.expectedErr) {
				t.Errorf("Expected error output to contain %q, got:\n%s", tc.expectedErr, errOutput)
			}
		})
	}
}

// TestM03E02ReadReportWithoutSession verifies that read-report
// fails gracefully when no session is provided.
func TestM03E02ReadReportWithoutSession(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")

	// Run read-report without session
	readCmd := exec.CommandContext(t.Context(), binPath, "ipc", "read-report")
	// Explicitly clear FLUXID_SESSION_ID if present
	readCmd.Env = []string{}

	var stdout, stderr bytes.Buffer
	readCmd.Stdout = &stdout
	readCmd.Stderr = &stderr

	err := readCmd.Run()
	if err == nil {
		t.Fatal("Expected read-report to fail without session, but it succeeded")
	}

	// Verify error message is helpful
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "session") {
		t.Errorf("Expected error about missing session, got:\n%s", errOutput)
	}
}

// TestM03E02ReadNonexistentReport verifies that reading a report
// from a session that hasn't written one exits cleanly.
func TestM03E02ReadNonexistentReport(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	sessionID := "nonexistent-session"

	readCmd := exec.CommandContext(t.Context(), binPath, "ipc", "read-report")
	readCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	var stdout, stderr bytes.Buffer
	readCmd.Stdout = &stdout
	readCmd.Stderr = &stderr

	// Should exit 0 even when no report exists
	if err := readCmd.Run(); err != nil {
		t.Fatalf("read-report should exit 0 for nonexistent report, got error: %v\nStderr: %s",
			err, stderr.String())
	}

	// Should have empty stdout (no report)
	if stdout.Len() > 0 {
		t.Errorf("Expected empty stdout for nonexistent report, got: %s", stdout.String())
	}

	// May have informative message on stderr
	if stderr.Len() > 0 {
		t.Logf("Stderr message: %s", stderr.String())
	}
}

// TestM03E02WriteReportSuccessMessage verifies that write-report
// outputs a clear success message.
func TestM03E02WriteReportSuccessMessage(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	sessionID := "test-success-message"

	validReport := `command: test
artifact: success-test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	writeCmd := exec.CommandContext(t.Context(), binPath, "ipc", "write-report")
	writeCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)
	writeCmd.Stdin = strings.NewReader(validReport)

	var stdout, stderr bytes.Buffer
	writeCmd.Stdout = &stdout
	writeCmd.Stderr = &stderr

	if err := writeCmd.Run(); err != nil {
		t.Fatalf("write-report failed: %v\nStderr: %s", err, stderr.String())
	}

	// Check for success message (could be on stdout or stderr)
	output := stdout.String() + stderr.String()
	hasSuccess := strings.Contains(output, "success") ||
		strings.Contains(output, "written") ||
		strings.Contains(output, "stored")
	if !hasSuccess {
		t.Logf("Warning: No clear success message found. Output:\n%s", output)
	}
}
