package command

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

// TestHandleHistoryGetFile_FileCreation tests that history file is created when it doesn't exist.
func TestHandleHistoryGetFile_FileCreation(t *testing.T) {
	sessionID := "bb0e8400-e29b-41d4-a716-446655440010"
	t.Setenv("FLUXID_SESSION_ID", sessionID)
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	writer := NewErrorWriter()

	// Call handleHistoryGetFile - should create empty file
	err := handleHistoryGetFile(writer)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify file was created
	historyPath := filepath.Join(tmpDir, sessionID, "history.yaml")
	if _, statErr := os.Stat(historyPath); os.IsNotExist(statErr) {
		t.Error("Expected history file to be created")
	}
}

// TestHandleHistoryValidate_EmptyFile tests validation of empty history file.
func TestHandleHistoryValidate_EmptyFile(t *testing.T) {
	sessionID := "cc0e8400-e29b-41d4-a716-446655440011"
	t.Setenv("FLUXID_SESSION_ID", sessionID)
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	// Create empty history file
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(historyPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	writer := NewErrorWriter()
	err = handleHistoryValidate(writer)
	if err != nil {
		t.Errorf("Expected no error for empty file, got: %v", err)
	}
}

// TestHandleReportGetFile_DirectoryCreation tests that session directory is created.
func TestHandleReportGetFile_DirectoryCreation(t *testing.T) {
	sessionID := "dd0e8400-e29b-41d4-a716-446655440012"
	t.Setenv("FLUXID_SESSION_ID", sessionID)
	tmpDir := t.TempDir()
	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)

	// Call handleReportGetFile - should create directory
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
