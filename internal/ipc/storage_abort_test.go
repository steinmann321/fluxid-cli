package ipc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetAbortFlag(t *testing.T) {
	t.Parallel()

	sessionID := "test-abort-session-123"

	// Set abort flag
	err := SetAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("SetAbortFlag failed: %v", err)
	}

	// Verify flag file was created
	flagPath := getAbortFlagPath(sessionID)
	if _, err := os.Stat(flagPath); os.IsNotExist(err) {
		t.Errorf("Abort flag file was not created at %s", flagPath)
	}

	// Clean up
	_ = os.Remove(flagPath)
}

func TestSetAbortFlagEmptySession(t *testing.T) {
	t.Parallel()

	err := SetAbortFlag("")
	if err == nil {
		t.Error("Expected error for empty session ID, got nil")
	}

	if !errors.Is(err, errSessionIDEmpty) {
		t.Errorf("Expected errSessionIDEmpty error, got: %v", err)
	}
}

func TestCheckAbortFlag(t *testing.T) {
	t.Parallel()

	sessionID := "test-check-abort-123"

	// Initially, flag should not be set
	aborted, err := CheckAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("CheckAbortFlag failed: %v", err)
	}
	if aborted {
		t.Error("Expected abort flag to be false initially")
	}

	// Set the flag
	err = SetAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("SetAbortFlag failed: %v", err)
	}

	// Now flag should be set
	aborted, err = CheckAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("CheckAbortFlag failed after setting flag: %v", err)
	}
	if !aborted {
		t.Error("Expected abort flag to be true after setting")
	}

	// Clean up
	_ = ClearAbortFlag(sessionID)
}

func TestCheckAbortFlagEmptySession(t *testing.T) {
	t.Parallel()

	_, err := CheckAbortFlag("")
	if err == nil {
		t.Error("Expected error for empty session ID, got nil")
	}

	if !errors.Is(err, errSessionIDEmpty) {
		t.Errorf("Expected errSessionIDEmpty error, got: %v", err)
	}
}

func TestClearAbortFlag(t *testing.T) {
	t.Parallel()

	sessionID := "test-clear-abort-123"

	// Set the flag first
	err := SetAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("SetAbortFlag failed: %v", err)
	}

	// Verify it's set
	aborted, err := CheckAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("CheckAbortFlag failed: %v", err)
	}
	if !aborted {
		t.Error("Expected abort flag to be set before clearing")
	}

	// Clear the flag
	err = ClearAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("ClearAbortFlag failed: %v", err)
	}

	// Verify it's cleared
	aborted, err = CheckAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("CheckAbortFlag failed after clearing: %v", err)
	}
	if aborted {
		t.Error("Expected abort flag to be false after clearing")
	}
}

func TestClearAbortFlagEmptySession(t *testing.T) {
	t.Parallel()

	err := ClearAbortFlag("")
	if err == nil {
		t.Error("Expected error for empty session ID, got nil")
	}

	if !errors.Is(err, errSessionIDEmpty) {
		t.Errorf("Expected errSessionIDEmpty error, got: %v", err)
	}
}

func TestClearAbortFlagNonExistent(t *testing.T) {
	t.Parallel()

	sessionID := "non-existent-abort-session-999"

	// Clearing a non-existent flag should not error
	err := ClearAbortFlag(sessionID)
	if err != nil {
		t.Errorf("ClearAbortFlag should not error for non-existent flag: %v", err)
	}
}

func TestGetAbortFlagPath(t *testing.T) {
	t.Parallel()

	sessionID := "test-session"
	path := getAbortFlagPath(sessionID)

	expectedSuffix := filepath.Join("fluxid-reports", sessionID+".abort")
	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got: %s", path)
	}

	if !strings.HasPrefix(path, os.TempDir()) {
		t.Errorf("Expected path to be in temp dir, got: %s", path)
	}

	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("Expected path to end with %s, got: %s", expectedSuffix, path)
	}
}
