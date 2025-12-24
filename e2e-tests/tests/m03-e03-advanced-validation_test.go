package tests

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestM03E03InvalidReportMissingIssueCategories verifies that a report
// with incomplete issue categories is rejected with specific diagnostics.
//
//nolint:funlen // E2E test with comprehensive validation checks
func TestM03E03InvalidReportMissingIssueCategories(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		report          string
		missingCategory string
	}{
		{
			name: "missing blockers",
			report: `command: test
artifact: test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  defects: []
  concerns: []
  observations: []
  enhancements: []
`,
			missingCategory: "blockers",
		},
		{
			name: "missing defects",
			report: `command: test
artifact: test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  concerns: []
  observations: []
  enhancements: []
`,
			missingCategory: "defects",
		},
		{
			name: "missing concerns",
			report: `command: test
artifact: test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  observations: []
  enhancements: []
`,
			missingCategory: "concerns",
		},
		{
			name: "missing observations",
			report: `command: test
artifact: test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  enhancements: []
`,
			missingCategory: "observations",
		},
		{
			name: "missing enhancements",
			report: `command: test
artifact: test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
`,
			missingCategory: "enhancements",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sessionID := "test-invalid-category-" + testCase.missingCategory

			errOutput := runInvalidReportTest(t, sessionID, testCase.report)

			// Verify error mentions the missing category
			if !strings.Contains(errOutput, testCase.missingCategory) {
				t.Errorf("Expected error to mention missing category %q, got:\n%s", testCase.missingCategory, errOutput)
			}
		})
	}
}

// TestM03E03InvalidReportNotStoredOnFailure verifies that when validation
// fails, no report is stored and read-report returns previous/empty state.
func TestM03E03InvalidReportNotStoredOnFailure(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	sessionID := "test-no-store-on-failure"

	// First, write a valid report
	validReport := `command: test-initial
artifact: initial-report
timestamp: 2025-12-12T09:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	writeValidReport(t, binPath, sessionID, validReport)

	// Now attempt to write an invalid report
	invalidReport := `command: test-invalid
artifact: should-not-be-stored
timestamp: 2025-12-12T10:00:00Z
status: INVALID_VALUE
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	errOutput := runInvalidReportTest(t, sessionID, invalidReport)
	if !strings.Contains(errOutput, "status") {
		t.Errorf("Expected validation error for invalid status, got:\n%s", errOutput)
	}

	// Read the report - should still be the valid one, not the invalid one
	readOutput := readReport(t, binPath, sessionID)

	// Verify we still have the original valid report
	if !strings.Contains(readOutput, "initial-report") {
		t.Errorf("Expected to read original valid report, got:\n%s", readOutput)
	}

	// Verify we don't have the invalid report data
	if strings.Contains(readOutput, "should-not-be-stored") {
		t.Errorf("Invalid report was stored despite validation failure:\n%s", readOutput)
	}
}

// TestM03E03InvalidReportMultipleErrors verifies that validation
// reports all errors at once, not just the first one.
func TestM03E03InvalidReportMultipleErrors(t *testing.T) {
	t.Parallel()

	invalidReport := `artifact: test-artifact
timestamp: 2025-12-12T10:00:00Z
status: WRONG
issues:
  blockers: []
  defects: []
`

	errOutput := runInvalidReportTest(t, "test-multiple-errors", invalidReport)

	// Should mention missing 'command'
	if !strings.Contains(errOutput, "command") {
		t.Errorf("Expected error to mention missing 'command', got:\n%s", errOutput)
	}

	// Should mention invalid 'status'
	if !strings.Contains(errOutput, "status") {
		t.Errorf("Expected error to mention invalid 'status', got:\n%s", errOutput)
	}

	// Should mention missing issue categories (concerns, observations, enhancements)
	hasIssueCategoryError := strings.Contains(errOutput, "concerns") ||
		strings.Contains(errOutput, "observations") ||
		strings.Contains(errOutput, "enhancements")
	if !hasIssueCategoryError {
		t.Errorf("Expected error to mention missing issue categories, got:\n%s", errOutput)
	}
}

// TestM03E03ValidationErrorsAreDetailed verifies that validation
// errors provide specific guidance on what's wrong and how to fix it.
func TestM03E03ValidationErrorsAreDetailed(t *testing.T) {
	t.Parallel()

	invalidReport := `command: test
artifact: test
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	errOutput := runInvalidReportTest(t, "test-detailed-errors", invalidReport)

	// Verify error is detailed and specific
	if !strings.Contains(errOutput, "timestamp") {
		t.Errorf("Expected error to mention 'timestamp', got:\n%s", errOutput)
	}

	// Verify error indicates what's required
	hasRequiredIndicator := strings.Contains(errOutput, "required") ||
		strings.Contains(errOutput, "missing") ||
		strings.Contains(errOutput, "must")
	if !hasRequiredIndicator {
		t.Errorf("Expected error to indicate field is required, got:\n%s", errOutput)
	}

	// Verify error output is structured (contains validation context)
	hasValidationContext := strings.Contains(errOutput, "validation") ||
		strings.Contains(errOutput, "Error")
	if !hasValidationContext {
		t.Errorf("Expected error to provide validation context, got:\n%s", errOutput)
	}
}
