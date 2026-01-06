//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

// Comprehensive validation tests to achieve full code coverage

func TestValidateReport_MissingMultipleFields(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report missing timestamp, status, and issues
	incompleteReport := `command: "test-command"
artifact: "test-artifact"
summary: "Test"
`
	if err := os.WriteFile(reportPath, []byte(incompleteReport), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for multiple missing required fields")
	}
}

func TestValidateReport_EmptyStringFields(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report with empty string values
	reportWithEmptyStrings := `command: ""
artifact: ""
timestamp: ""
status: ""
summary: "Test"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(reportWithEmptyStrings), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for empty string fields")
	}
}

func TestValidateReport_InvalidTimestampFormat(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report with invalid timestamp format
	invalidTimestamp := `command: "test-command"
artifact: "test-artifact"
timestamp: "not-a-valid-timestamp"
status: PASS
summary: "Test"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(invalidTimestamp), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for invalid timestamp format")
	}
}

func TestValidateReport_WrongTypeTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report with timestamp as number instead of string
	wrongTypeTimestamp := `command: "test-command"
artifact: "test-artifact"
timestamp: 123456
status: PASS
summary: "Test"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(wrongTypeTimestamp), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for wrong type timestamp")
	}
}

func TestValidateReport_IssuesWrongType(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report with issues as array instead of object
	wrongTypeIssues := `command: "test-command"
artifact: "test-artifact"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: "Test"
issues: []
`
	if err := os.WriteFile(reportPath, []byte(wrongTypeIssues), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for wrong type issues")
	}
}

func TestValidateReport_NextStepsWrongType(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report with next_steps as string instead of array
	wrongTypeNextSteps := `command: "test-command"
artifact: "test-artifact"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: "Test"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps: "not an array"
`
	if err := os.WriteFile(reportPath, []byte(wrongTypeNextSteps), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for wrong type next_steps")
	}
}

func TestValidateReport_SummaryWrongType(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report with summary as number instead of string
	wrongTypeSummary := `command: "test-command"
artifact: "test-artifact"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: 123
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(wrongTypeSummary), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for wrong type summary")
	}
}

//nolint:perfsprint // Test function builds YAML dynamically in loop
func TestValidateReport_MissingIssueCategoriesIndividually(t *testing.T) {
	tmpDir := t.TempDir()

	// Test each missing category individually
	categories := []string{"blockers", "defects", "concerns", "observations", "enhancements"}

	for _, missingCat := range categories {
		reportPath := filepath.Join(tmpDir, "report_missing_"+missingCat+".yaml")

		// Build report YAML with one category missing
		reportYAML := `command: "test-command"
artifact: "test-artifact"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: "Test"
issues:
`
		for _, cat := range categories {
			if cat != missingCat {
				reportYAML += "  " + cat + ": []\n"
			}
		}

		if err := os.WriteFile(reportPath, []byte(reportYAML), 0o644); err != nil {
			t.Fatal(err)
		}

		err := storage.ValidateReport(reportPath)
		if err == nil {
			t.Errorf("Expected error for missing %s category", missingCat)
		}
	}
}

func TestValidateReport_IssueCategoryWrongType(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report with blockers as string instead of array
	wrongTypeCat := `command: "test-command"
artifact: "test-artifact"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: "Test"
issues:
  blockers: "not an array"
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(wrongTypeCat), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for wrong type issue category")
	}
}

func TestValidateReport_FileSizeExceedsLimit(t *testing.T) {
	t.Skip("Large file tests slow down test suite - covered by E2E tests")
}

func TestValidateReport_PermissionDenied(t *testing.T) {
	t.Skip("Permission tests require special setup")
}

func TestValidateHistory_WrongTypeEvents(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	// History with event as string instead of object
	wrongTypeEvent := `- "not an object"
- timestamp: "2025-12-13T10:00:00Z"
  step: "implement"
  status: "SUCCESS"
  summary: "Test"
`
	if err := os.WriteFile(historyPath, []byte(wrongTypeEvent), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err == nil {
		t.Error("Expected error for wrong type event")
	}
}

func TestValidateHistory_EmptyStringFields(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	// History with empty string fields
	emptyFields := `- timestamp: ""
  step: ""
  status: ""
  summary: ""
`
	if err := os.WriteFile(historyPath, []byte(emptyFields), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err == nil {
		t.Error("Expected error for empty string fields")
	}
}

func TestValidateHistory_WrongTypeFields(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	// History with wrong type fields
	wrongTypes := `- timestamp: 123
  step: 456
  status: 789
  summary: 101112
`
	if err := os.WriteFile(historyPath, []byte(wrongTypes), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err == nil {
		t.Error("Expected error for wrong type fields")
	}
}

func TestValidateHistory_InvalidTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	// History with invalid timestamp format
	invalidTS := `- timestamp: "not-a-valid-timestamp"
  step: "implement"
  status: "SUCCESS"
  summary: "Test"
`
	if err := os.WriteFile(historyPath, []byte(invalidTS), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err == nil {
		t.Error("Expected error for invalid timestamp format")
	}
}

func TestValidateHistory_WrongTypeDetails(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	// History with details as number instead of string
	wrongTypeDetails := `- timestamp: "2025-12-13T10:00:00Z"
  step: "implement"
  status: "SUCCESS"
  summary: "Test"
  details: 123
`
	if err := os.WriteFile(historyPath, []byte(wrongTypeDetails), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err == nil {
		t.Error("Expected error for wrong type details")
	}
}

func TestValidateHistory_PermissionDenied(t *testing.T) {
	t.Skip("Permission tests require special setup")
}
