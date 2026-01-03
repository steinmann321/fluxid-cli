package command

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateTaskFile_Success(t *testing.T) {
	t.Parallel()
	// Create a valid task file
	tmpDir := t.TempDir()
	taskPath := filepath.Join(tmpDir, "task.md")
	if err := os.WriteFile(taskPath, []byte("# Task"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateTaskFile(taskPath); err != nil {
		t.Errorf("Expected no error for valid task file, got: %v", err)
	}
}

func TestValidateTaskFile_EmptyPath(t *testing.T) {
	t.Parallel()
	err := validateTaskFile("")
	if err == nil {
		t.Error("Expected error for empty path")
	}
	if !errors.Is(err, errTaskFileRequired) {
		t.Errorf("Expected errTaskFileRequired, got: %v", err)
	}
}

func TestValidateTaskFile_RelativePath(t *testing.T) {
	t.Parallel()
	err := validateTaskFile("relative/path.md")
	if err == nil {
		t.Error("Expected error for relative path")
	}
}

func TestValidateTaskFile_NotFound(t *testing.T) {
	t.Parallel()
	err := validateTaskFile("/nonexistent/path/task.md")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestValidateTaskFile_Directory(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	err := validateTaskFile(tmpDir)
	if err == nil {
		t.Error("Expected error for directory")
	}
}

func TestValidateTaskFile_Unreadable(t *testing.T) {
	t.Parallel()
	// Create a file and make it unreadable
	tmpDir := t.TempDir()
	taskPath := filepath.Join(tmpDir, "task.md")
	if err := os.WriteFile(taskPath, []byte("# Task"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Make it write-only (no read permission)
	if err := os.Chmod(taskPath, 0o200); err != nil {
		t.Fatal(err)
	}
	// Restore permissions after test
	defer func() {
		_ = os.Chmod(taskPath, 0o644)
	}()
	err := validateTaskFile(taskPath)
	if err == nil {
		t.Error("Expected error for unreadable file")
	}
}
