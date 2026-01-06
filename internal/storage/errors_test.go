//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

// Test error formatters for custom error types

func TestPathValidationError_Error(t *testing.T) {
	// Trigger PathValidationError by using invalid session ID
	err := storage.ValidateSessionID("../../../etc/passwd")
	if err == nil {
		t.Fatal("Expected path validation error")
	}

	// Check error message formatting
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}

	// Verify it's a PathValidationError
	if !storage.IsPathValidationError(err) {
		t.Error("Expected IsPathValidationError to return true")
	}
}

func TestSecurityError_Error(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440101"

	// Write report with YAML anchors (security violation)
	reportWithAnchors := `defaults: &defaults
  timeout: 30
config:
  <<: *defaults
  name: test
`
	if err := storage.WriteReport(sessionID, reportWithAnchors); err != nil {
		t.Fatal(err)
	}

	// ReadReport should fail with SecurityError
	_, err := storage.ReadReport(sessionID)
	if err == nil {
		t.Fatal("Expected security error for YAML with anchors")
	}

	// Check error message formatting
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}

	// Verify it's a SecurityError
	if !storage.IsSecurityError(err) {
		t.Error("Expected IsSecurityError to return true")
	}
}

func TestValidationError_Error(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Write report with missing required field for ValidateReport function
	invalidReport := `command: "test"
artifact: "test"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
summary: "Test"
issues:
  blockers: []
  # missing defects, concerns, observations, enhancements
`
	if err := os.WriteFile(reportPath, []byte(invalidReport), 0o644); err != nil {
		t.Fatal(err)
	}

	// ValidateReport should fail with ValidationError
	err := storage.ValidateReport(reportPath)
	if err == nil {
		t.Fatal("Expected validation error for missing issue categories")
	}

	// Check error message formatting
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}

	// Verify it's a ValidationError
	if !storage.IsValidationError(err) {
		t.Error("Expected IsValidationError to return true")
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440103"

	// Write report with multiple validation errors
	invalidReport := `command: "test"
artifact: "test"
# missing timestamp
# missing status
summary: "Test"
issues:
  blockers: []
  defects: []
  # missing concerns, observations, enhancements
next_steps: []
`
	if err := storage.WriteReport(sessionID, invalidReport); err != nil {
		t.Fatal(err)
	}

	// ReadReport should fail with multiple validation errors
	_, err := storage.ReadReport(sessionID)
	if err == nil {
		t.Fatal("Expected validation errors for multiple missing fields")
	}

	// Check error message formatting (should include multiple violations)
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}
