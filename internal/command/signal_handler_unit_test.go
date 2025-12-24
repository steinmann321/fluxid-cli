package command

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"syscall"
	"testing"
)

var errAbortSetterFailed = errors.New("abort setter failed")

func TestCleanupAllSignalHandlers(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Reset signal count
	signalCount.Store(0)

	// Create multiple signal handlers
	cleanup1 := setupSignalHandler("test-session-1")
	cleanup2 := setupSignalHandler("test-session-2")
	cleanup3 := setupSignalHandler("test-session-3")

	// Add to cleanup list
	t.Cleanup(cleanup1)
	t.Cleanup(cleanup2)
	t.Cleanup(cleanup3)

	// Call cleanupAllSignalHandlers
	cleanupAllSignalHandlers()

	// Calling again should be safe (idempotent)
	cleanupAllSignalHandlers()
}

// TestSignalHandler_HandleSignal_FirstSignal tests the first signal handling.
func TestSignalHandler_HandleSignal_FirstSignal(t *testing.T) {
	t.Parallel()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	abortCalled := false
	handler := &signalHandler{
		sessionID: "test-session-first-signal",
		abortSetter: func(_ string) error {
			abortCalled = true
			return nil
		},
		exitFunc: func(_ int) {
			t.Error("Exit should not be called on first signal")
		},
		logger: logger,
	}

	// First signal should return true (continue processing)
	shouldContinue := handler.handleSignal(syscall.SIGINT, 1)
	if !shouldContinue {
		t.Error("Expected handleSignal to return true for first signal")
	}

	if !abortCalled {
		t.Error("Expected abortSetter to be called")
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "requesting graceful abort") {
		t.Errorf("Expected log output to contain 'requesting graceful abort', got: %s", logOutput)
	}
}

// TestSignalHandler_HandleSignal_SecondSignal tests the second signal handling.
func TestSignalHandler_HandleSignal_SecondSignal(t *testing.T) {
	t.Parallel()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	exitCalled := false
	exitCode := 0
	handler := &signalHandler{
		sessionID: "test-session-second-signal",
		abortSetter: func(_ string) error {
			t.Error("AbortSetter should not be called on second signal")
			return nil
		},
		exitFunc: func(code int) {
			exitCalled = true
			exitCode = code
		},
		logger: logger,
	}

	// Second signal should return false (stop processing)
	shouldContinue := handler.handleSignal(syscall.SIGTERM, 2)
	if shouldContinue {
		t.Error("Expected handleSignal to return false for second signal")
	}

	if !exitCalled {
		t.Error("Expected exitFunc to be called")
	}

	if exitCode != exitCodeInterrupted {
		t.Errorf("Expected exit code %d, got %d", exitCodeInterrupted, exitCode)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "forcing immediate exit") {
		t.Errorf("Expected log output to contain 'forcing immediate exit', got: %s", logOutput)
	}
}

// TestSignalHandler_HandleSignal_AbortSetterError tests error handling in abortSetter.
func TestSignalHandler_HandleSignal_AbortSetterError(t *testing.T) {
	t.Parallel()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	handler := &signalHandler{
		sessionID: "test-session-error",
		abortSetter: func(_ string) error {
			return errAbortSetterFailed
		},
		exitFunc: func(_ int) {
			t.Error("Exit should not be called on first signal")
		},
		logger: logger,
	}

	// Should handle error gracefully and still return true
	shouldContinue := handler.handleSignal(syscall.SIGINT, 1)
	if !shouldContinue {
		t.Error("Expected handleSignal to return true even with abort setter error")
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "failed to set abort flag") {
		t.Errorf("Expected log output to contain error warning, got: %s", logOutput)
	}
}
