package command

import (
	"bytes"
	"log"
	"strings"
	"syscall"
	"testing"
)

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

// TestSignalHandler_HandleSignal_ImmediateExit tests immediate exit on signal.
func TestSignalHandler_HandleSignal_ImmediateExit(t *testing.T) {
	t.Parallel()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	exitCalled := false
	exitCode := 0
	handler := &signalHandler{
		sessionID: "test-session-immediate-exit",
		abortSetter: func(_ string) error {
			t.Error("AbortSetter should not be called with immediate exit")
			return nil
		},
		exitFunc: func(code int) {
			exitCalled = true
			exitCode = code
		},
		logger: logger,
	}

	// First signal should immediately exit (return false)
	shouldContinue := handler.handleSignal(syscall.SIGINT, 1)
	if shouldContinue {
		t.Error("Expected handleSignal to return false for immediate exit")
	}

	if !exitCalled {
		t.Error("Expected exitFunc to be called on first signal")
	}

	if exitCode != exitCodeInterrupted {
		t.Errorf("Expected exit code %d, got %d", exitCodeInterrupted, exitCode)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "terminating immediately") {
		t.Errorf("Expected log output to contain 'terminating immediately', got: %s", logOutput)
	}
}
