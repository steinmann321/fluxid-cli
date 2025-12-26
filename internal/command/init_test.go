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

func TestHandleInit_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	exitCode := handleInit([]string{tmpDir})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify .fluxid directory was created
	fluxidDir := filepath.Join(tmpDir, ".fluxid")
	if _, err := os.Stat(fluxidDir); os.IsNotExist(err) {
		t.Error(".fluxid directory not created")
	}
}

func TestHandleInit_Help(t *testing.T) {
	t.Parallel()

	exitCode := handleInit([]string{"--help"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help, got %d", exitCode)
	}

	exitCode = handleInit([]string{"-h"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}
}

func TestHandleInit_AlreadyExists(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Initialize once
	exitCode := handleInit([]string{tmpDir})
	if exitCode != 0 {
		t.Fatalf("First init failed with exit code %d", exitCode)
	}

	// Try to initialize again - should fail
	exitCode = handleInit([]string{tmpDir})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when .fluxid already exists")
	}
}

func TestHandleInit_InvalidPath(t *testing.T) {
	t.Parallel()

	// Try to initialize to an invalid path
	exitCode := handleInit([]string{"/dev/null/invalid/path"})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid path")
	}
}

func TestHandleInit_TooManyArgs(t *testing.T) {
	t.Parallel()

	exitCode := handleInit([]string{"path1", "path2"})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for too many args")
	}
}

func TestPrintInitHelp(t *testing.T) {
	t.Parallel()

	// Just call it to ensure it doesn't panic
	// We can't easily capture the output without refactoring
	// but we can at least ensure it runs
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printInitHelp panicked: %v", r)
		}
	}()

	printInitHelp()
}
