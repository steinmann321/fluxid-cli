package ipc

import (
	"strings"
	"testing"
)

func TestValidateReportSuccess(t *testing.T) {
	t.Parallel()

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

	err := ValidateReport(validReport)
	if err != nil {
		t.Errorf("Expected valid report to pass, got error: %v", err)
	}
}

func TestValidateReportWithIssues(t *testing.T) {
	t.Parallel()

	reportWithIssues := `command: test
artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
status: FAIL
issues:
  blockers:
    - message: "Critical bug found"
      location: "file.go:123"
  defects:
    - message: "Minor issue"
  concerns: []
  observations:
    - message: "Code could be improved"
      suggestion: "Use better naming"
  enhancements: []
`

	err := ValidateReport(reportWithIssues)
	if err != nil {
		t.Errorf("Expected valid report with issues to pass, got error: %v", err)
	}
}

func TestValidateReportEmpty(t *testing.T) {
	t.Parallel()

	err := ValidateReport("")
	if err == nil {
		t.Error("Expected error for empty report, got nil")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error about empty report, got: %v", err)
	}
}

func TestValidateReportMissingCommand(t *testing.T) {
	t.Parallel()

	report := `artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for missing command, got nil")
	}

	if !strings.Contains(err.Error(), "command") {
		t.Errorf("Expected error about missing command, got: %v", err)
	}
}

func TestValidateReportMissingArtifact(t *testing.T) {
	t.Parallel()

	report := `command: test
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for missing artifact, got nil")
	}

	if !strings.Contains(err.Error(), "artifact") {
		t.Errorf("Expected error about missing artifact, got: %v", err)
	}
}

func TestValidateReportMissingTimestamp(t *testing.T) {
	t.Parallel()

	report := `command: test
artifact: test.txt
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for missing timestamp, got nil")
	}

	if !strings.Contains(err.Error(), "timestamp") {
		t.Errorf("Expected error about missing timestamp, got: %v", err)
	}
}

func TestValidateReportMissingStatus(t *testing.T) {
	t.Parallel()

	report := `command: test
artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for missing status, got nil")
	}

	if !strings.Contains(err.Error(), "status") {
		t.Errorf("Expected error about missing status, got: %v", err)
	}
}

func TestValidateReportInvalidStatus(t *testing.T) {
	t.Parallel()

	report := `command: test
artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
status: INVALID
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for invalid status, got nil")
	}

	if !strings.Contains(err.Error(), "status") || !strings.Contains(err.Error(), "INVALID") {
		t.Errorf("Expected error about invalid status, got: %v", err)
	}
}

func TestValidateReportMissingIssues(t *testing.T) {
	t.Parallel()

	report := `command: test
artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
status: PASS
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for missing issues, got nil")
	}

	if !strings.Contains(err.Error(), "issues") {
		t.Errorf("Expected error about missing issues, got: %v", err)
	}
}

func TestValidateReportMissingIssueCategory(t *testing.T) {
	t.Parallel()

	report := `command: test
artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for missing issue category, got nil")
	}

	if !strings.Contains(err.Error(), "enhancements") {
		t.Errorf("Expected error about missing enhancements category, got: %v", err)
	}
}

func TestValidateReportIssuesMustBeObject(t *testing.T) {
	t.Parallel()

	report := `command: test
artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues: []
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for issues being array instead of object, got nil")
	}

	// The YAML unmarshaler will fail, producing a "YAML" error
	if !strings.Contains(err.Error(), "YAML") {
		t.Errorf("Expected error about invalid YAML structure, got: %v", err)
	}
}

func TestValidateReportIssueMissingMessage(t *testing.T) {
	t.Parallel()

	report := `command: test
artifact: test.txt
timestamp: 2025-12-12T10:00:00Z
status: FAIL
issues:
  blockers:
    - location: "file.go:123"
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for issue missing message, got nil")
	}

	if !strings.Contains(err.Error(), "message") {
		t.Errorf("Expected error about missing message, got: %v", err)
	}
}

func TestValidateReportInvalidYAML(t *testing.T) {
	t.Parallel()

	report := `command: test
artifact: test.txt
timestamp: [invalid
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}

	if !strings.Contains(err.Error(), "YAML") {
		t.Errorf("Expected error about invalid YAML, got: %v", err)
	}
}

func TestValidateReportMultipleErrors(t *testing.T) {
	t.Parallel()

	report := `timestamp: 2025-12-12T10:00:00Z
status: INVALID
`

	err := ValidateReport(report)
	if err == nil {
		t.Error("Expected error for multiple validation failures, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "command") {
		t.Error("Expected error to mention missing command")
	}

	if !strings.Contains(errMsg, "artifact") {
		t.Error("Expected error to mention missing artifact")
	}

	if !strings.Contains(errMsg, "status") {
		t.Error("Expected error to mention invalid status")
	}
}

//nolint:funlen // Unit test with extensive issue structure validation
func TestValidateIssuesStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yaml        string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid issues structure",
			yaml: `issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []`,
			expectError: false,
			errorMsg:    "",
		},
		{
			name:        "missing issues field",
			yaml:        `command: test`,
			expectError: true,
			errorMsg:    "issues",
		},
		{
			name:        "issues is not an object",
			yaml:        `issues: []`,
			expectError: true,
			errorMsg:    "object",
		},
		{
			name: "missing blockers category",
			yaml: `issues:
  defects: []
  concerns: []
  observations: []
  enhancements: []`,
			expectError: true,
			errorMsg:    "blockers",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			errors := validateIssuesStructure(testCase.yaml)

			hasError := len(errors) > 0
			if hasError != testCase.expectError {
				t.Errorf("Expected error: %v, got errors: %v", testCase.expectError, errors)
			}

			if testCase.expectError && len(errors) > 0 {
				allErrors := strings.Join(errors, " ")
				if !strings.Contains(allErrors, testCase.errorMsg) {
					t.Errorf("Expected error to contain %q, got: %v", testCase.errorMsg, errors)
				}
			}
		})
	}
}
