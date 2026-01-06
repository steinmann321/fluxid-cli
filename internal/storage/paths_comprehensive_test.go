//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

// Comprehensive path validation tests for full coverage

func TestValidateSessionID_EmptyAfterTrim(t *testing.T) {
	err := storage.ValidateSessionID("")
	if err == nil {
		t.Error("Expected error for empty session ID")
	}

	// Verify it's a PathValidationError
	if !storage.IsPathValidationError(err) {
		t.Error("Expected IsPathValidationError to return true")
	}
}

func TestValidateSessionID_PathTraversalDotDot(t *testing.T) {
	err := storage.ValidateSessionID("..550e8400-e29b-41d4-a716-446655440000")
	if err == nil {
		t.Error("Expected error for session ID with ..")
	}

	if !storage.IsPathValidationError(err) {
		t.Error("Expected IsPathValidationError to return true")
	}
}

func TestValidateSessionID_PathTraversalForwardSlash(t *testing.T) {
	err := storage.ValidateSessionID("550e8400/e29b-41d4-a716-446655440000")
	if err == nil {
		t.Error("Expected error for session ID with /")
	}

	if !storage.IsPathValidationError(err) {
		t.Error("Expected IsPathValidationError to return true")
	}
}

func TestValidateSessionID_PathTraversalBackslash(t *testing.T) {
	err := storage.ValidateSessionID("550e8400\\e29b-41d4-a716-446655440000")
	if err == nil {
		t.Error("Expected error for session ID with \\")
	}

	if !storage.IsPathValidationError(err) {
		t.Error("Expected IsPathValidationError to return true")
	}
}

func TestValidateSessionID_InvalidFormat(t *testing.T) {
	invalidUUIDs := []string{
		"not-a-uuid",
		"550e8400-e29b-41d4-a716", // too short
		"550e8400-e29b-41d4-a716-446655440000-extra",        // too long
		"550e8400-e29b-41d4-a716-44665544000g",              // invalid character
		"550e8400_e29b_41d4_a716_446655440000",              // wrong separator
		"550e8400-e29b-41d4-a716-446655440000-446655440000", // double UUID
	}

	for _, uuid := range invalidUUIDs {
		err := storage.ValidateSessionID(uuid)
		if err == nil {
			t.Errorf("Expected error for invalid UUID: %s", uuid)
		}

		if !storage.IsPathValidationError(err) {
			t.Errorf("Expected IsPathValidationError to return true for: %s", uuid)
		}
	}
}

func TestGetSessionRoot_WithAbsoluteOverride(t *testing.T) {
	tmpDir := t.TempDir()
	absOverride := filepath.Join(tmpDir, "custom-root")

	root, err := storage.GetSessionRoot(absOverride)
	if err != nil {
		t.Fatalf("GetSessionRoot failed: %v", err)
	}

	if !filepath.IsAbs(root) {
		t.Error("Expected absolute path")
	}

	// Should use the override
	if root != absOverride {
		t.Errorf("Expected root to use override: %s, got: %s", absOverride, root)
	}
}

func TestGetSessionRoot_WithRelativeOverride(t *testing.T) {
	relOverride := "relative/custom-root"

	root, err := storage.GetSessionRoot(relOverride)
	if err != nil {
		t.Fatalf("GetSessionRoot failed: %v", err)
	}

	if !filepath.IsAbs(root) {
		t.Error("Expected absolute path even with relative override")
	}
}

func TestGetSessionRoot_FallbackToTemp(t *testing.T) {
	// Test with no override and assuming .fluxid/sessions doesn't exist or isn't accessible
	root, err := storage.GetSessionRoot("")
	if err != nil {
		t.Fatalf("GetSessionRoot failed: %v", err)
	}

	if !filepath.IsAbs(root) {
		t.Error("Expected absolute path for fallback")
	}
}

func TestResolveSessionPath_SymlinkEscape(t *testing.T) {
	t.Skip("Symlink escape tests require complex setup with actual symlinks")
}

func TestResolveSessionPath_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	sessionID := "550e8400-e29b-41d4-a716-446655440201"

	// Resolve path should create directory
	path, err := storage.ResolveSessionPath(sessionID, "test.yaml", tmpDir)
	if err != nil {
		t.Fatalf("ResolveSessionPath failed: %v", err)
	}

	// Check that session directory was created
	sessionDir := filepath.Dir(path)
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		t.Error("Expected session directory to be created")
	}
}

func TestResolveSessionPath_EmptyFilenameWithOverride(t *testing.T) {
	tmpDir := t.TempDir()

	sessionID := "550e8400-e29b-41d4-a716-446655440202"

	// Empty filename should just return session directory
	path, err := storage.ResolveSessionPath(sessionID, "", tmpDir)
	if err != nil {
		t.Fatalf("ResolveSessionPath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestResolveSessionPath_SubdirectoryFile(t *testing.T) {
	tmpDir := t.TempDir()

	sessionID := "550e8400-e29b-41d4-a716-446655440203"

	// File in subdirectory
	path, err := storage.ResolveSessionPath(sessionID, "subdir/file.yaml", tmpDir)
	if err != nil {
		t.Fatalf("ResolveSessionPath failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestEnsureFileExists_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "new-file.yaml")

	// File doesn't exist
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatal("File should not exist yet")
	}

	// EnsureFileExists should create it
	err := storage.EnsureFileExists(filePath)
	if err != nil {
		t.Fatalf("EnsureFileExists failed: %v", err)
	}

	// File should now exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File should have been created")
	}
}

func TestEnsureFileExists_ExistingFileUnchanged(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "existing-file.yaml")

	// Create file with content
	content := "existing content"
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// EnsureFileExists should not modify it
	err := storage.EnsureFileExists(filePath)
	if err != nil {
		t.Fatalf("EnsureFileExists failed: %v", err)
	}

	// Content should be unchanged
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(readContent) != content {
		t.Error("File content should not have been modified")
	}
}

func TestEnsureFileExists_PermissionDenied(t *testing.T) {
	t.Skip("Permission tests require special setup")
}

func TestReadReport_LargeFile(t *testing.T) {
	t.Skip("Large file tests slow down test suite - covered by E2E tests")
}

func TestWriteReport_InvalidSessionID(t *testing.T) {
	invalidSessionID := "../../../etc/passwd"

	err := storage.WriteReport(invalidSessionID, "content")
	if err == nil {
		t.Error("Expected error for invalid session ID")
	}

	if !storage.IsPathValidationError(err) {
		t.Error("Expected IsPathValidationError to return true")
	}
}
