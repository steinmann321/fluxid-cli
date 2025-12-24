package command

import (
	"bytes"
	"errors"
	"fluxid-loop/internal/ipc"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

var errAbortSetterFailed = errors.New("abort setter failed")

// TestSetupSignalHandler_Coverage tests the setupSignalHandler function for coverage.
// Note: We cannot easily test actual signal delivery in unit tests without
// risking interference with the test runner itself. This test verifies the
// function can be called without panicking and the logic paths are exercised
// via integration tests.
func TestSetupSignalHandler_Coverage(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-signal-coverage"

	// Reset signal count for clean test
	signalCount.Store(0)

	// Just call setupSignalHandler to exercise the code path
	// The goroutine will be started but we won't send actual signals
	// to avoid interfering with the test process
	cleanup := setupSignalHandler(sessionID)
	t.Cleanup(cleanup)

	// Verify that the abort flag is not set initially
	aborted, err := ipc.CheckAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("Failed to check abort flag: %v", err)
	}
	if aborted {
		t.Error("Expected abort flag to be false initially")
	}
}

// TestSetupSignalHandler_InvalidSessionPath tests error handling when
// the session path is invalid.
func TestSetupSignalHandler_InvalidSessionPath(t *testing.T) {
	// Set invalid XDG_DATA_HOME to test error path in SetAbortFlag
	t.Setenv("XDG_DATA_HOME", "/dev/null/invalid/path")
	sessionID := "test-invalid-path"

	// Reset signal count for clean test
	signalCount.Store(0)

	// Should not panic even with invalid path
	cleanup := setupSignalHandler(sessionID)
	t.Cleanup(cleanup)

	// The function should complete without crashing
	// The error path in SetAbortFlag will be logged but not cause failure
}

// TestSetupSignalHandler_CleanupMultipleCalls tests that cleanup can be called multiple times safely.
func TestSetupSignalHandler_CleanupMultipleCalls(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-cleanup-multiple"

	// Reset signal count for clean test
	signalCount.Store(0)

	cleanup := setupSignalHandler(sessionID)

	// Call cleanup multiple times - should not panic
	cleanup()
	cleanup()
	cleanup()
}

// TestSetupSignalHandler_ConcurrentCleanup tests concurrent cleanup calls.
func TestSetupSignalHandler_ConcurrentCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-cleanup-concurrent"

	// Reset signal count for clean test
	signalCount.Store(0)

	cleanup := setupSignalHandler(sessionID)

	// Call cleanup concurrently - should not panic or race
	var waitGroup sync.WaitGroup
	for i := 0; i < 10; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			cleanup()
		}()
	}
	waitGroup.Wait()
}

// TestSetupSignalHandler_SimulatedSignal simulates signal handling by sending to the channel.
func TestSetupSignalHandler_SimulatedSignal(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-simulated-signal-" + time.Now().Format("20060102150405")

	// Reset signal count for clean test
	signalCount.Store(0)

	// Create our own signal channel for testing
	sigChan := make(chan os.Signal, 1)

	// Start a goroutine that mimics the signal handler logic
	done := make(chan bool, 1)
	go func() {
		for range sigChan {
			count := signalCount.Add(1)
			if count == 1 {
				// First signal: set abort flag
				if err := ipc.SetAbortFlag(sessionID); err != nil {
					t.Logf("Warning: failed to set abort flag: %v", err)
				}
				done <- true
				return
			}
		}
	}()

	// Simulate sending a signal
	sigChan <- syscall.SIGINT

	// Wait for the handler to process
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for signal handler")
	}

	close(sigChan)

	// Verify abort flag was set
	aborted, err := ipc.CheckAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("Failed to check abort flag: %v", err)
	}
	if !aborted {
		t.Error("Expected abort flag to be set after simulated signal")
	}

	// Verify signal count
	if signalCount.Load() != 1 {
		t.Errorf("Expected signal count 1, got %d", signalCount.Load())
	}
}

// TestSetupSignalHandler_CheckAbortFlagError tests error handling in abort flag check.
func TestSetupSignalHandler_CheckAbortFlagError(t *testing.T) {
	// Use invalid path to trigger error
	t.Setenv("XDG_DATA_HOME", "/dev/null/invalid")
	sessionID := "test-check-abort-err-" + time.Now().Format("20060102150405")

	signalCount.Store(0)

	// Create our own signal channel for testing
	sigChan := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	go func() {
		for range sigChan {
			count := signalCount.Add(1)
			if count == 1 {
				// This will log a warning but continue
				if err := ipc.SetAbortFlag(sessionID); err != nil {
					// Expected error path
					t.Logf("Expected error setting abort flag: %v", err)
				}
				done <- true
				return
			}
		}
	}()

	sigChan <- syscall.SIGINT

	select {
	case <-done:
		// Success - error was handled
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for signal handler")
	}

	close(sigChan)
}

// TestSetupSignalHandler_SecondSignalSimulated tests the second signal path.
func TestSetupSignalHandler_SecondSignalSimulated(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-second-signal-" + time.Now().Format("20060102150405")

	signalCount.Store(0)

	// Create our own signal channel
	sigChan := make(chan os.Signal, 2)
	firstDone := make(chan bool, 1)
	secondDone := make(chan bool, 1)

	go func() {
		for range sigChan {
			count := signalCount.Add(1)
			if count == 1 {
				// First signal: set abort flag
				if err := ipc.SetAbortFlag(sessionID); err != nil {
					t.Logf("Warning: failed to set abort flag: %v", err)
				}
				firstDone <- true
			} else {
				// Second signal: in real implementation would call os.Exit(130)
				// We just signal completion here for testing
				secondDone <- true
				return
			}
		}
	}()

	// Send first signal
	sigChan <- syscall.SIGINT

	select {
	case <-firstDone:
		// First signal processed
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for first signal")
	}

	// Send second signal
	sigChan <- syscall.SIGTERM

	select {
	case <-secondDone:
		// Second signal processed
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for second signal")
	}

	close(sigChan)

	// Verify signal count is 2
	if signalCount.Load() != 2 {
		t.Errorf("Expected signal count 2, got %d", signalCount.Load())
	}
}

// TestCleanupAllSignalHandlers tests the cleanupAllSignalHandlers function.
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
