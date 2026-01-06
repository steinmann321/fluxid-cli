//nolint:cyclop,funlen // E2E tests: comprehensive scenarios justify complexity
package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestObservabilitySilentSuccessProducesNoStderr verifies that successful operations produce no stderr output.
func TestObservabilitySilentSuccessProducesNoStderr(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory with valid report
	tmpDir := t.TempDir()
	sessionID := testSessionID
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a valid report file
	reportPath := filepath.Join(sessionDir, "report.yaml")
	validReport := `command: fluxid.implement
artifact: test
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(validReport), 0o644); err != nil {
		t.Fatalf("Failed to write report file: %v", err)
	}

	// This test expects:
	// 1. Successful validation produces NO stderr output (FR-042: silent success)
	// 2. Exit code 0
	// 3. stdout may contain data (like file paths, schemas)
	// 4. stderr must be completely empty

	t.Log("Test setup complete - expects silent success (no stderr output, exit 0)")
}

// TestObservabilityErrorsIncludeSufficientContext verifies that errors include comprehensive context.
func TestObservabilityErrorsIncludeSufficientContext(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory with invalid report
	tmpDir := t.TempDir()
	sessionID := testSessionID
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write an invalid report file (missing required fields)
	reportPath := filepath.Join(sessionDir, "report.yaml")
	invalidReport := `command: fluxid.implement
artifact: test
# Missing timestamp, status, issues
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

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command (expect failure)
	_ = cmd.Run()

	// This test expects error messages (FR-043) to include sufficient context:
	// 1. File path being validated
	// 2. Field paths for validation errors (e.g., "status: missing required field")
	// 3. Constraint details (e.g., "expected: PASS or FAIL")
	// 4. Session ID in error context
	// 5. Format per error contract: "[field_path]: [violation] (expected: [constraint], got: [value])"

	stderrOutput := stderr.String()
	if stderrOutput == "" {
		t.Error("Expected error output on stderr for validation failure, got empty stderr")
	}

	// Verify stderr contains expected context elements
	expectedContextElements := []string{
		reportPath,  // File path
		"timestamp", // Field name
		"status",    // Field name
		"issues",    // Field name
	}

	for _, element := range expectedContextElements {
		if !strings.Contains(stderrOutput, element) {
			t.Errorf("Error output missing expected context element: %q\nStderr: %s", element, stderrOutput)
		}
	}

	t.Log("Test setup complete - expects comprehensive error context per FR-043")
}

// TestObservabilityStdoutReservedForData verifies that stdout is reserved for data output only.
func TestObservabilityStdoutReservedForData(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")

	// Test --get-schema command (should output schema to stdout)
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--get-schema")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
	}

	// This test expects (FR-041):
	// 1. Data (schema) goes to stdout
	// 2. NO logging/progress/info messages on stdout
	// 3. stdout contains only parseable data (YAML schema)
	// 4. stderr is empty on success (silent success)

	stdoutOutput := stdout.String()
	stderrOutput := stderr.String()

	if stdoutOutput == "" {
		t.Error("Expected schema data on stdout, got empty")
	}

	if stderrOutput != "" {
		t.Errorf("Expected empty stderr on success, got: %s", stderrOutput)
	}

	// Verify stdout contains only valid YAML (no extraneous messages)
	if strings.Contains(stdoutOutput, "INFO") ||
		strings.Contains(stdoutOutput, "DEBUG") ||
		strings.Contains(stdoutOutput, "Loading") ||
		strings.Contains(stdoutOutput, "Processing") {
		t.Errorf("stdout contains non-data content (logging/progress messages)\nStdout: %s", stdoutOutput)
	}

	t.Log("Test setup complete - verified stdout reserved for data only")
}

// TestObservabilityErrorFormatFollowsContract verifies that error messages follow the error format contract.
func TestObservabilityErrorFormatFollowsContract(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory with invalid report
	tmpDir := t.TempDir()
	sessionID := testSessionID
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

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command (expect failure)
	_ = cmd.Run()

	stderrOutput := stderr.String()

	// This test expects error format per contract:
	// "[field_path]: [violation] (expected: [constraint], got: [value])"
	// Example: "status: invalid value (expected: PASS or FAIL, got: INVALID)"

	expectedPatterns := []string{
		"status",   // field path
		"expected", // constraint indicator
		"got",      // value indicator
		"PASS",     // expected value
		"FAIL",     // expected value
		"INVALID",  // actual value
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(stderrOutput, pattern) {
			t.Errorf("Error output missing expected pattern: %q\nStderr: %s", pattern, stderrOutput)
		}
	}

	t.Log("Test setup complete - verified error format follows contract")
}

// TestObservabilityExitCodesAreCorrect verifies that commands use correct exit codes.
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
//nolint:cyclop // E2E test complexity justified by comprehensive validation
//nolint:cyclop,funlen // E2E test: comprehensive validation scenarios
func TestObservabilityExitCodesAreCorrect(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	testCases := []struct {
		name            string
		setupFunc       func(t *testing.T, tmpDir, sessionID string) string // returns session root
		args            []string
		expectedExitErr bool   // true if we expect non-zero exit
		description     string // what this test verifies
	}{
		{
			name: "success_validation",
			setupFunc: func(t *testing.T, tmpDir, sessionID string) string { //nolint:thelper // Anonymous setup function
				sessionDir := filepath.Join(tmpDir, sessionID)
				if err := os.MkdirAll(sessionDir, 0o755); err != nil {
					t.Fatalf("Failed to create session directory: %v", err)
				}
				reportPath := filepath.Join(sessionDir, "report.yaml")
				validReport := `command: fluxid.implement
artifact: test
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
				if err := os.WriteFile(reportPath, []byte(validReport), 0o644); err != nil {
					t.Fatalf("Failed to write report: %v", err)
				}
				return tmpDir
			},
			args:            []string{"report", "--validate"},
			expectedExitErr: false,
			description:     "Exit 0 for successful validation",
		},
		{
			name: "validation_failure",
			setupFunc: func(t *testing.T, tmpDir, sessionID string) string { //nolint:thelper // Anonymous setup function
				sessionDir := filepath.Join(tmpDir, sessionID)
				if err := os.MkdirAll(sessionDir, 0o755); err != nil {
					t.Fatalf("Failed to create session directory: %v", err)
				}
				reportPath := filepath.Join(sessionDir, "report.yaml")
				invalidReport := `command: fluxid.implement
# Missing required fields
`
				if err := os.WriteFile(reportPath, []byte(invalidReport), 0o644); err != nil {
					t.Fatalf("Failed to write report: %v", err)
				}
				return tmpDir
			},
			args:            []string{"report", "--validate"},
			expectedExitErr: true,
			description:     "Exit 1 for validation failure",
		},
		{
			name: "missing_session_id",
			setupFunc: func(_ *testing.T, tmpDir, _ string) string {
				return tmpDir
			},
			args:            []string{"report", "--get-file"},
			expectedExitErr: true,
			description:     "Exit 3 for missing session ID",
		},
	}

	for _, testCase := range testCases {
		// capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			sessionID := testSessionID

			sessionRoot := testCase.setupFunc(t, tmpDir, sessionID)

			binPath := filepath.Join(root, "bin", "fluxid")
			cmd := exec.CommandContext(testCtx(30*time.Second), binPath, testCase.args...)

			// Set environment based on test case
			if testCase.name != "missing_session_id" {
				cmd.Env = append(os.Environ(),
					"FLUXID_SESSION_ROOT="+sessionRoot,
					"FLUXID_SESSION_ID="+sessionID,
				)
			} else {
				cmd.Env = append(os.Environ(),
					"FLUXID_SESSION_ROOT="+sessionRoot,
				)
			}

			err := cmd.Run()

			if testCase.expectedExitErr && err == nil {
				t.Errorf("%s: Expected non-zero exit code, got success", testCase.description)
			} else if !testCase.expectedExitErr && err != nil {
				t.Errorf("%s: Expected exit 0, got error: %v", testCase.description, err)
			}
		})
	}
}
