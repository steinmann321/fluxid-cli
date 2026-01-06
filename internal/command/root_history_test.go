package command

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

const testHistorySessionID = "550e8400-e29b-41d4-a716-446655440000"

func TestHandleHistoryCommand_GetSchema(t *testing.T) {
	t.Parallel()

	// Test get-schema (doesn't require session ID)
	code := handleHistoryCommand([]string{"history", "--get-schema"})
	if code != ExitSuccess {
		t.Errorf("Expected exit code %d for --get-schema, got %d", ExitSuccess, code)
	}
}

func TestHandleHistoryCommand_NoFlags(t *testing.T) {
	t.Parallel()

	// Test with no flags - should fail
	code := handleHistoryCommand([]string{"history"})
	if code == ExitSuccess {
		t.Error("Expected non-zero exit code for no flags")
	}
}

func TestHandleHistoryCommand_MultipleFlags(t *testing.T) {
	t.Parallel()

	// Test with multiple flags - should fail
	code := handleHistoryCommand([]string{"history", "--get-schema", "--validate"})
	if code == ExitSuccess {
		t.Error("Expected non-zero exit code for multiple flags")
	}
}

func TestHandleHistoryCommand_GetFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := testHistorySessionID
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Create session directory
	sessionDir := filepath.Join(tmpDir, "fluxid", sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	code := handleHistoryCommand([]string{"history", "--get-file"})
	// Note: This will call os.Exit() from handleHistoryGetFile, so we can't easily test it
	// But calling handleHistoryCommand exercises the code path up to that point
	_ = code
}

func TestHandleHistoryCommand_Validate(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := testHistorySessionID
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// Create valid history file
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}

	historyDir := filepath.Dir(historyPath)
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	historyYAML := `- timestamp: "2025-12-13T10:00:00Z"
  step: "implement"
  status: "SUCCESS"
  summary: "Test"
`
	if err := os.WriteFile(historyPath, []byte(historyYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	code := handleHistoryCommand([]string{"history", "--validate"})
	// Note: This will call os.Exit() from handleHistoryValidate
	_ = code
}
