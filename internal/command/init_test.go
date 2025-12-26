package command

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetInitTargetDir_NoArgs(t *testing.T) {
	t.Parallel()
	targetDir, err := getInitTargetDir([]string{})
	if err != nil {
		t.Fatalf("getInitTargetDir failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	if targetDir != homeDir {
		t.Errorf("Expected home dir %s, got %s", homeDir, targetDir)
	}
}

func TestGetInitTargetDir_WithPath(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "project")

	targetDir, err := getInitTargetDir([]string{testPath})
	if err != nil {
		t.Fatalf("getInitTargetDir failed: %v", err)
	}

	// Should create the directory
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Error("Target directory not created")
	}

	// Should be absolute path
	if !filepath.IsAbs(targetDir) {
		t.Error("Target directory is not absolute path")
	}
}

func TestGetInitTargetDir_ExistingPath(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	targetDir, err := getInitTargetDir([]string{tmpDir})
	if err != nil {
		t.Fatalf("getInitTargetDir failed: %v", err)
	}

	// Should work with existing directory
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Error("Target directory doesn't exist")
	}

	// Should be the same directory (absolute)
	absTemp, _ := filepath.Abs(tmpDir)
	if targetDir != absTemp {
		t.Errorf("Expected %s, got %s", absTemp, targetDir)
	}
}

func TestGetInitTargetDir_TooManyArgs(t *testing.T) {
	t.Parallel()
	_, err := getInitTargetDir([]string{"path1", "path2"})
	if err == nil {
		t.Error("Expected error for too many args")
	}
}

func TestGetInitTargetDir_RelativePath(t *testing.T) {
	t.Parallel()
	// Test with relative path - should be converted to absolute
	targetDir, err := getInitTargetDir([]string{"."})
	if err != nil {
		t.Fatalf("getInitTargetDir failed: %v", err)
	}

	if !filepath.IsAbs(targetDir) {
		t.Error("Relative path not converted to absolute")
	}
}
