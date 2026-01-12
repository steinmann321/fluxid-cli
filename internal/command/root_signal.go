package command

import (
	"fluxid-cli/internal/process"
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

// handleSignal processes a single signal and immediately exits.
func (h *signalHandler) handleSignal(sig os.Signal, _ int32) bool {
	// Immediately exit on any signal - no graceful shutdown
	h.logger.Printf("\nReceived signal %v - terminating immediately", sig)

	// Kill all active child processes and their subprocesses before exiting
	process.KillAll()

	h.exitFunc(exitCodeInterrupted)
	return false
}

// setupSignalHandler installs a signal handler that immediately exits on SIGINT/SIGTERM.
// The handler forces immediate termination without graceful shutdown.
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
