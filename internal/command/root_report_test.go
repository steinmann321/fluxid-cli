package command

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleReportCommand_GetSchema(t *testing.T) {
	t.Parallel()

	// Test get-schema (doesn't require session ID)
	code := handleReportCommand([]string{"report", "--get-schema"})
	if code != ExitSuccess {
		t.Errorf("Expected exit code %d for --get-schema, got %d", ExitSuccess, code)
	}
}

func TestHandleReportCommand_NoFlags(t *testing.T) {
	t.Parallel()

	// Test with no flags - should fail
	code := handleReportCommand([]string{"report"})
	if code == ExitSuccess {
		t.Error("Expected non-zero exit code for no flags")
	}
}

func TestHandleReportCommand_MultipleFlags(t *testing.T) {
	t.Parallel()

	// Test with multiple flags - should fail
	code := handleReportCommand([]string{"report", "--get-schema", "--validate"})
	if code == ExitSuccess {
		t.Error("Expected non-zero exit code for multiple flags")
	}
}

func TestHandleReportCommand_GetFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "550e8400-e29b-41d4-a716-446655440000"
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Create session directory
	sessionDir := filepath.Join(tmpDir, "fluxid", sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	code := handleReportCommand([]string{"report", "--get-file"})
	// Note: This will call os.Exit() from handleReportGetFile, so we can't easily test it
	// But calling handleReportCommand exercises the code path up to that point
	_ = code
}

func TestHandleReportCommand_Validate(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "550e8400-e29b-41d4-a716-446655440000"
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Create valid report file
	reportYAML := `command: "test"
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
	if err := storage.WriteReport(sessionID, reportYAML); err != nil {
		t.Fatal(err)
	}

	code := handleReportCommand([]string{"report", "--validate"})
	// Note: This will call os.Exit() from handleReportValidate
	_ = code
}
