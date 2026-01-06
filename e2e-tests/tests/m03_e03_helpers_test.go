package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runInvalidReportTest runs a test for an invalid report and returns the error output.
//
//nolint:unused // Helper function for future test scenarios
func runInvalidReportTest(t *testing.T, sessionID, invalidReport string) string {
	t.Helper()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")

	cmd := exec.CommandContext(t.Context(), binPath, "ipc", "write-report")
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)
	cmd.Stdin = strings.NewReader(invalidReport)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("Expected write-report to fail, but it succeeded\nStdout: %s", stdout.String())
	}

	return stderr.String()
}

// writeValidReport writes a valid report for testing purposes.
//
//nolint:unused // Helper function for future test scenarios
func writeValidReport(t *testing.T, binPath, sessionID, report string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), binPath, "ipc", "write-report")
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)
	cmd.Stdin = strings.NewReader(report)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to write valid report: %v\nStderr: %s", err, stderr.String())
	}
}

// readReport reads the current report for a session.
//
//nolint:unused // Helper function for future test scenarios
func readReport(t *testing.T, binPath, sessionID string) string {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), binPath, "ipc", "read-report")
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("read-report failed: %v\nStderr: %s", err, stderr.String())
	}

	return stdout.String()
}
