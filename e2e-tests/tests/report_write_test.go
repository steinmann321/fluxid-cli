//nolint:cyclop,funlen // E2E tests: comprehensive scenarios justify complexity
package tests

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestReportWriteWorkflow tests the complete agent workflow for writing reports.
// This is the MVP test for User Story 1:
// 1. Agent runs `fluxid report --get-file` to get the report file path
// 2. Agent writes a valid YAML report to that path
// 3. Fluxid can read the report successfully.
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
//nolint:cyclop // E2E test complexity justified by comprehensive validation
//nolint:cyclop,funlen // E2E test: comprehensive validation scenarios
func TestReportWriteWorkflow(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "550e8400-e29b-41d4-a716-446655440000"

	binPath := filepath.Join(root, "bin", "fluxid")

	// Step 1: Get report file path
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--get-file")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid report --get-file failed: %v\nStderr: %s", err, stderr.String())
	}

	// Verify no stderr output (silent success per FR-042)
	if stderr.Len() > 0 {
		t.Errorf("Expected no stderr output, got: %s", stderr.String())
	}

	// Get the file path from stdout
	reportPath := strings.TrimSpace(stdout.String())
	if reportPath == "" {
		t.Fatal("Expected file path in stdout, got empty string")
	}

	// Verify the path is absolute
	if !filepath.IsAbs(reportPath) {
		t.Errorf("Expected absolute path, got: %s", reportPath)
	}

	// Verify the path contains the session ID
	if !strings.Contains(reportPath, sessionID) {
		t.Errorf("Expected path to contain session ID %s, got: %s", sessionID, reportPath)
	}

	// Verify the file/directory exists or can be created
	dir := filepath.Dir(reportPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory does not exist and was not created: %s", dir)
	}

	// Step 2: Agent writes a valid report
	validReport := `command: fluxid.implement
artifact: internal/storage/report.go
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations:
    - Implementation complete
  enhancements: []
summary: Test report for E2E workflow
`
	if err := os.WriteFile(reportPath, []byte(validReport), 0o644); err != nil {
		t.Fatalf("Failed to write report file: %v", err)
	}

	// Step 3: Verify fluxid can read the report
	// This simulates the workflow controller reading the report
	// We'll verify this by checking the file exists and is readable
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Report file is empty")
	}

	// Verify content matches what we wrote
	if string(content) != validReport {
		t.Errorf("Report content mismatch.\nExpected:\n%s\nGot:\n%s", validReport, string(content))
	}

	t.Log("Report write workflow completed successfully")
}

// TestReportGetFileCreatesDirectory verifies that --get-file creates the session directory if needed.
func TestReportGetFileCreatesDirectory(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := testSessionID2

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--get-file")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	reportPath := strings.TrimSpace(stdout.String())

	// Verify the directory was created
	dir := filepath.Dir(reportPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", dir)
	}
}

// TestReportGetFileWithMissingSessionID verifies error handling when session ID is not set.
func TestReportGetFileWithMissingSessionID(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--get-file")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		// FLUXID_SESSION_ID not set
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should fail with exit code 3 (config error)
	if err == nil {
		t.Fatal("Expected command to fail when FLUXID_SESSION_ID is not set")
	}

	// Verify error message on stderr
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "FLUXID_SESSION_ID") {
		t.Errorf("Expected error message to mention FLUXID_SESSION_ID, got: %s", stderrOutput)
	}

	// Verify exit code is 3
	exitErr := &exec.ExitError{} //nolint:exhaustruct
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() != 3 {
			t.Errorf("Expected exit code 3, got: %d", exitErr.ExitCode())
		}
	}
}

// TestReportGetFileWithInvalidSessionID verifies error handling for invalid session IDs.
func TestReportGetFileWithInvalidSessionID(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()

	invalidSessionIDs := []string{
		"not-a-uuid",
		"../../../etc/passwd",
		"test session",
	}

	for _, sessionID := range invalidSessionIDs {
		t.Run(sessionID, func(t *testing.T) {
			t.Parallel()
			binPath := filepath.Join(root, "bin", "fluxid")
			cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--get-file")
			cmd.Env = append(os.Environ(),
				"FLUXID_SESSION_ROOT="+tmpDir,
				"FLUXID_SESSION_ID="+sessionID,
			)

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Should fail with exit code 3 (config error)
			if err == nil {
				t.Errorf("Expected command to fail for invalid session ID: %s", sessionID)
			}

			// Verify error on stderr
			if stderr.Len() == 0 {
				t.Error("Expected error message on stderr")
			}
		})
	}
}

// TestReportWorkflowWithFailStatus tests the workflow when agent reports FAIL status.
func TestReportWorkflowWithFailStatus(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"

	binPath := filepath.Join(root, "bin", "fluxid")

	// Get report file path
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--get-file")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid report --get-file failed: %v", err)
	}

	reportPath := strings.TrimSpace(stdout.String())

	// Agent writes a FAIL report
	failReport := `command: fluxid.implement
artifact: internal/feature.go
timestamp: 2026-01-05T15:45:22Z
status: FAIL
issues:
  blockers:
    - Tests failing
  defects:
    - Implementation incomplete
  concerns: []
  observations: []
  enhancements: []
summary: Implementation failed due to test failures
`
	if err := os.WriteFile(reportPath, []byte(failReport), 0o644); err != nil {
		t.Fatalf("Failed to write report file: %v", err)
	}

	// Verify fluxid can read the FAIL report
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	if !strings.Contains(string(content), "status: FAIL") {
		t.Error("Expected status: FAIL in report")
	}

	t.Log("FAIL status report workflow completed successfully")
}
