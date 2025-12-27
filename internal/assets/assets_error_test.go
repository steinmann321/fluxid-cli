package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyAssetsToDir_AlreadyExists(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create .fluxid directory
	fluxidDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Should fail
	_, err := CopyAssetsToDir(tmpDir)
	if err == nil {
		t.Error("Expected error when .fluxid already exists")
	}
	if err != nil && !os.IsExist(err) {
		// Check error message contains expected text
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("Error message is empty")
		}
	}
}

func TestCopyAssetsToDir_InvalidPath(t *testing.T) {
	t.Parallel()

	// Try to copy to a path that doesn't exist and can't be created
	_, err := CopyAssetsToDir("/dev/null/invalid/path")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

func TestCopyAssetsToDir_ReadOnlyParent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a subdirectory and make it read-only
	roDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(roDir, 0o755); err != nil {
		t.Fatalf("Failed to create readonly dir: %v", err)
	}

	// Make it read-only (no write permission)
	if err := os.Chmod(roDir, 0o555); err != nil {
		t.Fatalf("Failed to chmod readonly dir: %v", err)
	}

	// Restore permissions after test
	defer func() {
		_ = os.Chmod(roDir, 0o755)
	}()

	// Try to copy assets to a subdirectory of the read-only directory
	_, err := CopyAssetsToDir(roDir)
	if err == nil {
		t.Error("Expected error when parent directory is read-only, got nil")
	}
}

func TestCopyAssetsToDir_FileExistsWhereDirectoryShouldBe(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a file named .fluxid (blocking directory creation)
	blockingFile := filepath.Join(tmpDir, ".fluxid")
	if err := os.WriteFile(blockingFile, []byte("blocker"), 0o644); err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}

	// Try to copy assets - should fail because .fluxid is a file, not a directory
	_, err := CopyAssetsToDir(tmpDir)
	if err == nil {
		t.Error("Expected error when .fluxid is a file, got nil")
	}

	// Error should mention that .fluxid already exists
	if err != nil && !containsStr(err.Error(), ".fluxid") {
		t.Errorf("Error message should mention .fluxid, got: %v", err)
	}
}
