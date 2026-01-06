package command

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

const (
	exitCodeInterrupted = 130 // Exit code for SIGINT/SIGTERM user interrupt
)

//nolint:gochecknoglobals // Global state needed for signal handler cleanup in tests
var (
	signalCleanups []func()
	cleanupMutex   sync.Mutex
)

//nolint:gochecknoglobals // Signal handling requires global state for goroutine coordination
var signalCount atomic.Int32

// signalHandler defines the interface for handling signals.
type signalHandler struct {
	sessionID   string
	abortSetter func(string) error
	exitFunc    func(int)
	logger      *log.Logger
}

// handleSignal processes a single signal and returns true if processing should continue.
func (h *signalHandler) handleSignal(sig os.Signal, count int32) bool {
	if count == 1 {
		// First signal: set abort flag for graceful shutdown
		h.logger.Printf("\nReceived signal %v - requesting graceful abort...", sig)
		h.logger.Println("Workflow will exit after current phase completes.")
		h.logger.Println("Press Ctrl+C again to force immediate exit.")

		if err := h.abortSetter(h.sessionID); err != nil {
			h.logger.Printf("Warning: failed to set abort flag: %v", err)
		}
		return true
	}
	// Second signal: force immediate exit
	h.logger.Printf("\nReceived signal %v again - forcing immediate exit", sig)
	h.exitFunc(exitCodeInterrupted)
	return false
}

// setupSignalHandler installs a signal handler that sets the abort flag on SIGINT/SIGTERM.
// On the first signal, it sets the abort flag for graceful shutdown.
// On the second signal, it forces immediate exit.
func setupSignalHandler(sessionID string) func() {
	sigChan := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Abort flag functionality removed per 001-report-history-refactor
	// Abort mechanism is out of scope and requires separate evaluation
	noopAbortSetter := func(string) error { return nil }

	handler := &signalHandler{
		sessionID:   sessionID,
		abortSetter: noopAbortSetter,
		exitFunc:    os.Exit,
		logger:      log.Default(),
	}

	go func() {
		for {
			select {
			case sig := <-sigChan:
				count := signalCount.Add(1)
				if !handler.handleSignal(sig, count) {
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Return cleanup function to stop signal handling and close done channel
	// Use sync.Once to ensure cleanup only runs once even if called multiple times
	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			signal.Stop(sigChan)
			close(done)
		})
	}

	// Track cleanup for tests (store the idempotent cleanup function)
	cleanupMutex.Lock()
	signalCleanups = append(signalCleanups, cleanup)
	cleanupMutex.Unlock()

	return cleanup
}

// cleanupAllSignalHandlers cleans up all signal handlers. Used in tests to prevent goroutine leaks.
// Each cleanup function uses sync.Once internally, so it's safe to call multiple times.
func cleanupAllSignalHandlers() {
	cleanupMutex.Lock()
	defer cleanupMutex.Unlock()

	// Call all cleanup functions (they're idempotent thanks to sync.Once)
	for _, cleanup := range signalCleanups {
		cleanup()
	}
	signalCleanups = nil
}
