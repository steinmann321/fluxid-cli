package command

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

func TestNewHistoryCommand(t *testing.T) {
	t.Parallel()

	cmd := NewHistoryCommand()
	if cmd == nil {
		t.Fatal("Expected non-nil history command")
	}

	if cmd.Use != "history" {
		t.Errorf("Expected command use 'history', got %s", cmd.Use)
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

func TestHistoryCommand_MutuallyExclusiveFlags(t *testing.T) {
	t.Parallel()

	cmd := NewHistoryCommand()

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

func TestHistoryCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewHistoryCommand()

	// Execute with no flags should fail
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no flags are set")
	}
}

func TestHandleHistoryGetFile_MissingSessionID(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")

	writer := NewErrorWriter()
	err := handleHistoryGetFile(writer)
	if err == nil {
		t.Error("Expected error for missing session ID")
	}
}

func TestHandleHistoryGetFile_InvalidSessionID(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "invalid")
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	writer := NewErrorWriter()
	err := handleHistoryGetFile(writer)
	if err == nil {
		t.Error("Expected error for invalid session ID")
	}
}

func TestHandleHistoryGetFile_Success(t *testing.T) {
	sessionID := "880e8400-e29b-41d4-a716-446655440003"
	t.Setenv("FLUXID_SESSION_ID", sessionID)
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	writer := NewErrorWriter()
	err := handleHistoryGetFile(writer)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify directory and file were created
	sessionDir := filepath.Join(tmpDir, sessionID)
	historyFile := filepath.Join(sessionDir, "history.yaml")
	if _, statErr := os.Stat(historyFile); os.IsNotExist(statErr) {
		t.Error("Expected history file to be created")
	}
}

func TestHandleHistoryValidate_MissingSessionID(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "")

	writer := NewErrorWriter()
	err := handleHistoryValidate(writer)
	if err == nil {
		t.Error("Expected error for missing session ID")
	}
}

func TestHandleHistoryValidate_InvalidSessionID(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "invalid")
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	writer := NewErrorWriter()
	err := handleHistoryValidate(writer)
	if err == nil {
		t.Error("Expected error for invalid session ID")
	}
}

func TestHandleHistoryValidate_ValidHistory(t *testing.T) {
	sessionID := "990e8400-e29b-41d4-a716-446655440004"
	t.Setenv("FLUXID_SESSION_ID", sessionID)
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	// Create valid history file
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	validHistory := `[]`
	if err := os.WriteFile(historyPath, []byte(validHistory), 0o644); err != nil {
		t.Fatal(err)
	}

	writer := NewErrorWriter()
	err = handleHistoryValidate(writer)
	if err != nil {
		t.Errorf("Expected no error for valid history, got: %v", err)
	}
}

func TestHandleHistoryValidate_InvalidHistory(t *testing.T) {
	sessionID := "aa0e8400-e29b-41d4-a716-446655440005"
	t.Setenv("FLUXID_SESSION_ID", sessionID)
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	// Create invalid history (not a YAML array)
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	invalidHistory := `invalid: yaml`
	if err := os.WriteFile(historyPath, []byte(invalidHistory), 0o644); err != nil {
		t.Fatal(err)
	}

	writer := NewErrorWriter()
	err = handleHistoryValidate(writer)
	if err == nil {
		t.Error("Expected error for invalid history")
	}
}

func TestHandleHistoryGetSchema_Success(t *testing.T) {
	t.Parallel()

	writer := NewErrorWriter()
	err := handleHistoryGetSchema(writer)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
