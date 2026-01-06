//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"errors"
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

// Test session ID constant to avoid goconst violation.
const testPathsSessionID = "550e8400-e29b-41d4-a716-446655440000"

func TestResolveSessionPath_WithRootOverride(t *testing.T) {
	tmpDir := t.TempDir()

	sessionID := testPathsSessionID
	customRoot := filepath.Join(tmpDir, "custom-root")

	path, err := storage.ResolveSessionPath(sessionID, "report.yaml", customRoot)
	if err != nil {
		t.Fatalf("Expected no error with root override, got: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestResolveSessionPath_EmptyFilename(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testPathsSessionID

	path, err := storage.ResolveSessionPath(sessionID, "", "")
	if err != nil {
		t.Fatalf("Expected no error for empty filename, got: %v", err)
	}

	// Should return session directory
	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestValidateSessionID_EmptyString(t *testing.T) {
	err := storage.ValidateSessionID("")
	if err == nil {
		t.Error("Expected error for empty session ID")
	}
}

func TestValidateSessionID_UppercaseUUID(t *testing.T) {
	uppercaseUUID := "550E8400-E29B-41D4-A716-446655440000"
	// UUIDs are case-insensitive, should be accepted
	err := storage.ValidateSessionID(uppercaseUUID)
	if err != nil {
		t.Errorf("Expected no error for uppercase UUID, got: %v", err)
	}
}

func TestValidateSessionID_WithBraces(t *testing.T) {
	uuidWithBraces := "{550e8400-e29b-41d4-a716-446655440000}"
	err := storage.ValidateSessionID(uuidWithBraces)
	if err == nil {
		t.Error("Expected error for UUID with braces")
	}
}

func TestGetSessionRoot_NoEnvVars(t *testing.T) {
	// Unset both XDG_DATA_HOME and HOME
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "")

	// GetSessionRoot falls back to current directory when no env vars
	root, err := storage.GetSessionRoot("")
	if err != nil {
		t.Errorf("Expected no error (fallback to cwd), got: %v", err)
	}
	if !filepath.IsAbs(root) {
		t.Error("Expected absolute path")
	}
}

func TestGetSessionRoot_RelativeOverride(t *testing.T) {
	relativeRoot := "relative/path"

	// Should convert to absolute
	root, err := storage.GetSessionRoot(relativeRoot)
	if err != nil {
		t.Fatalf("Expected no error for relative override, got: %v", err)
	}

	if !filepath.IsAbs(root) {
		t.Error("Expected absolute path even with relative override")
	}
}

func TestEnsureFileExists_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "existing.yaml")

	// Create file first
	if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should not error on existing file
	err := storage.EnsureFileExists(filePath)
	if err != nil {
		t.Errorf("Expected no error for existing file, got: %v", err)
	}
}

func TestEnsureFileExists_InvalidPath(t *testing.T) {
	// Try to create file in non-existent directory without creating parent
	invalidPath := "/nonexistent/directory/file.yaml"

	err := storage.EnsureFileExists(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestIsPathValidationError(t *testing.T) {
	t.Parallel()

	// Test with nil
	if storage.IsPathValidationError(nil) {
		t.Error("nil should not be a path validation error")
	}

	// Test with regular error
	regularErr := errors.New("regular error") //nolint:err113 // Test error, not a sentinel error
	if storage.IsPathValidationError(regularErr) {
		t.Error("regular error should not be a path validation error")
	}
}

func TestResolveSessionPath_MultipleSlashes(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testPathsSessionID
	filename := "subdir//file.yaml"

	_, err := storage.ResolveSessionPath(sessionID, filename, "")
	// Should handle path normalization
	if err != nil {
		t.Errorf("Expected no error for path with multiple slashes, got: %v", err)
	}
}
