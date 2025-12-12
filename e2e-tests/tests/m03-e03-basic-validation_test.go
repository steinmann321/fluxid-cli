package tests

import (
	"strings"
	"testing"
)

// TestM03E03InvalidReportMissingCommand verifies that a report missing
// the required 'command' field is rejected with a clear diagnostic.
func TestM03E03InvalidReportMissingCommand(t *testing.T) {
	t.Parallel()

	invalidReport := `artifact: test-artifact
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	errOutput := runInvalidReportTest(t, "test-invalid-missing-command", invalidReport)

	// Verify error mentions 'command'
	if !strings.Contains(errOutput, "command") {
		t.Errorf("Expected error to mention 'command', got:\n%s", errOutput)
	}

	// Verify error indicates it's missing or required
	hasMissingOrRequired := strings.Contains(errOutput, "missing") || strings.Contains(errOutput, "required")
	if !hasMissingOrRequired {
		t.Errorf("Expected error to indicate field is missing/required, got:\n%s", errOutput)
	}
}

// TestM03E03InvalidReportMissingStatus verifies that a report missing
// the required 'status' field is rejected with a clear diagnostic.
func TestM03E03InvalidReportMissingStatus(t *testing.T) {
	t.Parallel()

	invalidReport := `command: test-command
artifact: test-artifact
timestamp: 2025-12-12T10:00:00Z
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	errOutput := runInvalidReportTest(t, "test-invalid-missing-status", invalidReport)

	// Verify error mentions 'status'
	if !strings.Contains(errOutput, "status") {
		t.Errorf("Expected error to mention 'status', got:\n%s", errOutput)
	}

	// Verify error indicates it's missing or required
	hasMissingOrRequired := strings.Contains(errOutput, "missing") || strings.Contains(errOutput, "required")
	if !hasMissingOrRequired {
		t.Errorf("Expected error to indicate field is missing/required, got:\n%s", errOutput)
	}
}

// TestM03E03InvalidReportBadStatusEnum verifies that a report with
// an invalid status value (not PASS or FAIL) is rejected.
func TestM03E03InvalidReportBadStatusEnum(t *testing.T) {
	t.Parallel()

	invalidReport := `command: test-command
artifact: test-artifact
timestamp: 2025-12-12T10:00:00Z
status: INVALID_STATUS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	errOutput := runInvalidReportTest(t, "test-invalid-bad-status", invalidReport)

	// Verify error mentions 'status'
	if !strings.Contains(errOutput, "status") {
		t.Errorf("Expected error to mention 'status', got:\n%s", errOutput)
	}

	// Verify error mentions valid values (PASS/FAIL)
	mentionsPASS := strings.Contains(errOutput, "PASS")
	mentionsFAIL := strings.Contains(errOutput, "FAIL")
	if !mentionsPASS || !mentionsFAIL {
		t.Errorf("Expected error to mention valid values PASS and FAIL, got:\n%s", errOutput)
	}
}

// TestM03E03InvalidReportMissingIssues verifies that a report missing
// the entire 'issues' object is rejected.
func TestM03E03InvalidReportMissingIssues(t *testing.T) {
	t.Parallel()

	invalidReport := `command: test-command
artifact: test-artifact
timestamp: 2025-12-12T10:00:00Z
status: PASS
`

	errOutput := runInvalidReportTest(t, "test-invalid-missing-issues", invalidReport)

	// Verify error mentions 'issues'
	if !strings.Contains(errOutput, "issues") {
		t.Errorf("Expected error to mention 'issues', got:\n%s", errOutput)
	}
}

// TestM03E03InvalidReportMalformedYAML verifies that completely
// malformed YAML is rejected with a clear diagnostic.
func TestM03E03InvalidReportMalformedYAML(t *testing.T) {
	t.Parallel()

	malformedReport := `command: test
  artifact: test-bad-indent
status: PASS
  timestamp: 2025-12-12T10:00:00Z
`

	errOutput := runInvalidReportTest(t, "test-invalid-malformed-yaml", malformedReport)

	// Verify error indicates YAML parsing issue
	hasYAMLError := strings.Contains(errOutput, "YAML") || strings.Contains(errOutput, "yaml") ||
		strings.Contains(errOutput, "parse") || strings.Contains(errOutput, "invalid")
	if !hasYAMLError {
		t.Errorf("Expected error to indicate YAML parsing issue, got:\n%s", errOutput)
	}
}

// TestM03E03InvalidReportEmptyInput verifies that empty input
// is rejected with a clear diagnostic.
func TestM03E03InvalidReportEmptyInput(t *testing.T) {
	t.Parallel()

	errOutput := runInvalidReportTest(t, "test-empty-input", "")

	// Verify error indicates empty input
	hasEmptyError := strings.Contains(errOutput, "empty") ||
		strings.Contains(errOutput, "no content") ||
		strings.Contains(errOutput, "missing")
	if !hasEmptyError {
		t.Errorf("Expected error to indicate empty input, got:\n%s", errOutput)
	}
}
