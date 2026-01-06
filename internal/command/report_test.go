package command

import (
	"errors"
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

// Test session ID constant to avoid goconst violation.
const testReportSessionID = "550e8400-e29b-41d4-a716-446655440000"

func TestNewReportCommand(t *testing.T) {
	t.Parallel()

	cmd := NewReportCommand()
	if cmd == nil {
		t.Fatal("Expected non-nil report command")
	}

	if cmd.Use != "report" {
		t.Errorf("Expected command use 'report', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty short description")
	}

	if cmd.Long == "" {
		t.Error("Expected non-empty long description")
	}

	// Verify flags are registered
	if cmd.Flags().Lookup("get-file") == nil {
		t.Error("Expected --get-file flag to be registered")
	}

	if cmd.Flags().Lookup("validate") == nil {
		t.Error("Expected --validate flag to be registered")
	}

	if cmd.Flags().Lookup("get-schema") == nil {
		t.Error("Expected --get-schema flag to be registered")
	}
}

func TestReportCommand_MutuallyExclusiveFlags(t *testing.T) {
	t.Parallel()

	cmd := NewReportCommand()

	// Test that multiple operation flags fail
	if err := cmd.Flags().Set("get-file", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("validate", "true"); err != nil {
		t.Fatal(err)
	}

	// Execute should fail with multiple flags
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when multiple flags are set")
	}
}

func TestReportCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewReportCommand()

	// Execute with no flags should fail
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no flags are set")
	}
}

func TestHandleReportGetFile_MissingSessionID(t *testing.T) {
	// Don't set FLUXID_SESSION_ID
	t.Setenv("FLUXID_SESSION_ID", "")

	err := handleReportGetFile()
	if err == nil {
		t.Error("Expected error for missing session ID")
	}

	// Should be PathValidationError
	var pathErr *storage.PathValidationError
	if !asPathValidationError(err, &pathErr) {
		t.Errorf("Expected PathValidationError, got: %T", err)
	}
}

func TestHandleReportGetFile_InvalidSessionID(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "not-a-uuid")
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	err := handleReportGetFile()
	if err == nil {
		t.Error("Expected error for invalid session ID")
	}
}

func TestHandleReportGetFile_Success(t *testing.T) {
	sessionID := testReportSessionID
	t.Setenv("FLUXID_SESSION_ID", sessionID)
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	err := handleReportGetFile()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify directory was created
	sessionDir := filepath.Join(tmpDir, sessionID)
	if _, statErr := os.Stat(sessionDir); os.IsNotExist(statErr) {
		t.Error("Expected session directory to be created")
	}
}

func TestHandleReportValidate_MissingSessionID(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")

	err := handleReportValidate()
	if err == nil {
		t.Error("Expected error for missing session ID")
	}

	var pathErr *storage.PathValidationError
	if !asPathValidationError(err, &pathErr) {
		t.Errorf("Expected PathValidationError, got: %T", err)
	}
}

func TestHandleReportValidate_InvalidSessionID(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "invalid")
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	err := handleReportValidate()
	if err == nil {
		t.Error("Expected error for invalid session ID")
	}
}

func TestHandleReportValidate_ValidReport(t *testing.T) {
	sessionID := "660e8400-e29b-41d4-a716-446655440001"
	t.Setenv("FLUXID_SESSION_ID", sessionID)
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	// Create valid report
	reportPath, err := storage.ResolveSessionPath(sessionID, "report.yaml", tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	validReport := `command: "test"
artifact: "test"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(validReport), 0o644); err != nil {
		t.Fatal(err)
	}

	err = handleReportValidate()
	if err != nil {
		t.Errorf("Expected no error for valid report, got: %v", err)
	}
}

func TestHandleReportValidate_InvalidReport(t *testing.T) {
	sessionID := "770e8400-e29b-41d4-a716-446655440002"
	t.Setenv("FLUXID_SESSION_ID", sessionID)
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	// Create invalid report (missing required field)
	reportPath, err := storage.ResolveSessionPath(sessionID, "report.yaml", tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	invalidReport := `status: PASS`
	if err := os.WriteFile(reportPath, []byte(invalidReport), 0o644); err != nil {
		t.Fatal(err)
	}

	err = handleReportValidate()
	if err == nil {
		t.Error("Expected error for invalid report")
	}
}

func TestHandleReportGetSchema_Success(t *testing.T) {
	t.Parallel()

	err := handleReportGetSchema()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// Helper function to check if error is PathValidationError.
func asPathValidationError(err error, target **storage.PathValidationError) bool {
	if err == nil {
		return false
	}
	return errors.As(err, target)
}
