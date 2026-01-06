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

// TestReportValidateValidFile tests validation of a valid report file.
func TestReportValidateValidFile(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := testSessionID
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a valid report file
	reportPath := filepath.Join(sessionDir, "report.yaml")
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
summary: Valid report test
`
	if err := os.WriteFile(reportPath, []byte(validReport), 0o644); err != nil {
		t.Fatalf("Failed to write report file: %v", err)
	}

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--validate")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Should succeed with exit 0
	if err := cmd.Run(); err != nil {
		t.Fatalf("Validation should succeed, got error: %v\nStderr: %s", err, stderr.String())
	}

	// Verify silent success (no stderr output per FR-042)
	if stderr.Len() > 0 {
		t.Errorf("Expected no stderr output for valid report, got: %s", stderr.String())
	}

	// Verify no stdout output (validation doesn't output data)
	if stdout.Len() > 0 {
		t.Errorf("Expected no stdout output for validation, got: %s", stdout.String())
	}
}

// TestReportValidateMissingRequiredField tests validation error for missing required fields.
func TestReportValidateMissingRequiredField(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := testSessionID2
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a report missing required field (timestamp)
	reportPath := filepath.Join(sessionDir, "report.yaml")
	invalidReport := `command: fluxid.implement
artifact: test
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(invalidReport), 0o644); err != nil {
		t.Fatalf("Failed to write report file: %v", err)
	}

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--validate")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	// Should fail with exit 1 (validation failure)
	if err == nil {
		t.Fatal("Expected validation to fail for missing required field")
	}

	exitErr := &exec.ExitError{} //nolint:exhaustruct
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got: %d", exitErr.ExitCode())
		}
	}
	// Verify error message includes field name
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "timestamp") {
		t.Errorf("Expected error to mention 'timestamp', got: %s", stderrOutput)
	}
}

// TestReportValidateInvalidStatusValue tests validation error for invalid enum value.
func TestReportValidateInvalidStatusValue(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := testSessionID3
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a report with invalid status value
	reportPath := filepath.Join(sessionDir, "report.yaml")
	invalidReport := `command: fluxid.implement
artifact: test
timestamp: 2026-01-05T14:32:10Z
status: INVALID
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(invalidReport), 0o644); err != nil {
		t.Fatalf("Failed to write report file: %v", err)
	}

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--validate")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should fail with exit 1 (validation failure)
	if err == nil {
		t.Fatal("Expected validation to fail for invalid status value")
	}

	exitErr := &exec.ExitError{} //nolint:exhaustruct
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got: %d", exitErr.ExitCode())
		}
	}

	// Verify error message includes enum values
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "PASS") && !strings.Contains(stderrOutput, "FAIL") {
		t.Errorf("Expected error to mention PASS or FAIL, got: %s", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "INVALID") {
		t.Errorf("Expected error to mention invalid value 'INVALID', got: %s", stderrOutput)
	}
}

// TestReportValidateFileNotFound tests error handling when report file doesn't exist.
func TestReportValidateFileNotFound(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := "123e4567-e89b-12d3-a456-426614174000"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Don't create report file

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--validate")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should fail with exit 2 (operational error)
	if err == nil {
		t.Fatal("Expected validation to fail for missing file")
	}

	exitErr := &exec.ExitError{} //nolint:exhaustruct
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() != 2 {
			t.Errorf("Expected exit code 2 for file not found, got: %d", exitErr.ExitCode())
		}
	}

	// Verify error message mentions file not found
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "file not found") && !strings.Contains(stderrOutput, "does not exist") {
		t.Errorf("Expected error to mention file not found, got: %s", stderrOutput)
	}
}

// TestReportValidateSchemaMismatch tests validation errors for various schema mismatches.
//
//nolint:funlen // E2E test: comprehensive validation requires multiple test cases
func TestReportValidateSchemaMismatch(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	testCases := []struct {
		name           string
		report         string
		expectedErrors []string
	}{
		{
			name: "wrong_data_type",
			report: `command: fluxid.implement
artifact: test
timestamp: 12345
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`,
			expectedErrors: []string{"timestamp", "wrong type"},
		},
		{
			name: "missing_issues_category",
			report: `command: fluxid.implement
artifact: test
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
`,
			expectedErrors: []string{"concerns", "observations", "enhancements"},
		},
		{
			name: "invalid_yaml_structure",
			report: `command: fluxid.implement
artifact: test
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues: "not an object"
`,
			expectedErrors: []string{"issues", "wrong type"},
		},
	}

	for _, testCase := range testCases {
		// capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			sessionID := testSessionID
			sessionDir := filepath.Join(tmpDir, sessionID)
			if err := os.MkdirAll(sessionDir, 0o755); err != nil {
				t.Fatalf("Failed to create session directory: %v", err)
			}

			reportPath := filepath.Join(sessionDir, "report.yaml")
			if err := os.WriteFile(reportPath, []byte(testCase.report), 0o644); err != nil {
				t.Fatalf("Failed to write report file: %v", err)
			}

			binPath := filepath.Join(root, "bin", "fluxid")
			cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--validate")
			cmd.Env = append(os.Environ(),
				"FLUXID_SESSION_ROOT="+tmpDir,
				"FLUXID_SESSION_ID="+sessionID,
			)

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Should fail with exit 1 (validation failure)
			if err == nil {
				t.Fatal("Expected validation to fail")
			}

			// Verify error message includes expected strings
			stderrOutput := stderr.String()
			for _, expectedErr := range testCase.expectedErrors {
				if !strings.Contains(stderrOutput, expectedErr) {
					t.Errorf("Expected error to contain '%s', got: %s", expectedErr, stderrOutput)
				}
			}
		})
	}
}

// TestReportValidateMultipleErrors tests that all validation errors are reported at once.
func TestReportValidateMultipleErrors(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := testSessionID
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a report with multiple errors
	reportPath := filepath.Join(sessionDir, "report.yaml")
	invalidReport := `artifact: test
status: INVALID
issues:
  blockers: []
`
	// Missing: command, timestamp
	// Invalid: status value
	// Missing: defects, concerns, observations, enhancements

	if err := os.WriteFile(reportPath, []byte(invalidReport), 0o644); err != nil {
		t.Fatalf("Failed to write report file: %v", err)
	}

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--validate")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err == nil {
		t.Fatal("Expected validation to fail")
	}

	stderrOutput := stderr.String()

	// Verify multiple errors are reported
	expectedErrors := []string{"command", "timestamp", "status"}
	for _, expectedErr := range expectedErrors {
		if !strings.Contains(stderrOutput, expectedErr) {
			t.Errorf("Expected error output to contain '%s', got: %s", expectedErr, stderrOutput)
		}
	}
}
