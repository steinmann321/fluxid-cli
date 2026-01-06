//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"errors"
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateReport_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	validReport := `command: "test-command"
artifact: "test-artifact"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: "Test passed"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - "Continue"
`
	if err := os.WriteFile(reportPath, []byte(validReport), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err != nil {
		t.Errorf("Expected no error for valid report, got: %v", err)
	}
}

func TestValidateReport_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "nonexistent.yaml")

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestValidateReport_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	malformedYAML := "command: [invalid yaml"
	if err := os.WriteFile(reportPath, []byte(malformedYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for malformed YAML")
	}
}

func TestValidateReport_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	incompleteReport := `command: "test-command"
artifact: "test-artifact"
`
	if err := os.WriteFile(reportPath, []byte(incompleteReport), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for missing required fields")
	}
}

func TestValidateReport_InvalidStatus(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	invalidStatusReport := `command: "test-command"
artifact: "test-artifact"
timestamp: "2025-12-13T10:00:00Z"
status: INVALID_STATUS
summary: "Test"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps: []
`
	if err := os.WriteFile(reportPath, []byte(invalidStatusReport), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for invalid status enum")
	}
}

func TestValidateReport_WrongFieldTypes(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	wrongTypesReport := `command: 123
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
next_steps: []
`
	if err := os.WriteFile(reportPath, []byte(wrongTypesReport), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Error("Expected error for wrong field types")
	}
}

func TestValidateHistory_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	validHistory := `- timestamp: "2025-12-13T10:00:00Z"
  step: "implement"
  status: "SUCCESS"
  summary: "Implementation completed"
  details: "Details here"
`
	if err := os.WriteFile(historyPath, []byte(validHistory), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err != nil {
		t.Errorf("Expected no error for valid history, got: %v", err)
	}
}

func TestValidateHistory_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	if err := os.WriteFile(historyPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err != nil {
		t.Errorf("Expected no error for empty history file, got: %v", err)
	}
}

func TestValidateHistory_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	incompleteHistory := `- timestamp: "2025-12-13T10:00:00Z"
  step: "implement"
`
	if err := os.WriteFile(historyPath, []byte(incompleteHistory), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err == nil {
		t.Error("Expected error for missing required fields")
	}
}

func TestValidateHistory_InvalidStatus(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	invalidStatusHistory := `- timestamp: "2025-12-13T10:00:00Z"
  step: "implement"
  status: "INVALID"
  summary: "Test"
`
	if err := os.WriteFile(historyPath, []byte(invalidStatusHistory), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err == nil {
		t.Error("Expected error for invalid status enum")
	}
}

func TestIsValidationError(t *testing.T) {
	t.Parallel()

	// Test with nil
	if storage.IsValidationError(nil) {
		t.Error("nil should not be a validation error")
	}

	// Test with regular error
	regularErr := errors.New("regular error") //nolint:err113 // Test error, not a sentinel error
	if storage.IsValidationError(regularErr) {
		t.Error("regular error should not be a validation error")
	}
}
