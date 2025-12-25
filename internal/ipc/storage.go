// Package ipc provides inter-process communication commands for fluxid.
package ipc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	permDir  = 0o750 // Directory permissions for IPC storage
	permFile = 0o600 // File permissions for IPC storage (private files)
)

var (
	errSessionIDEmpty = errors.New("session ID cannot be empty")
	errMessageEmpty   = errors.New("message cannot be empty")
)

// getReportPath returns the file path for storing a session's report.
// Reports are stored in <temp-dir>/fluxid-reports/<session-id>.yaml
// where <temp-dir> is os.TempDir() (e.g., /tmp on Unix, %TEMP% on Windows).
func getReportPath(sessionID string) string {
	dir := filepath.Join(os.TempDir(), "fluxid-reports")
	return filepath.Join(dir, sessionID+".yaml")
}

// WriteReport stores a report for the given session ID.
// Reports are stored as YAML files in <temp-dir>/fluxid-reports/
// where <temp-dir> is os.TempDir() (e.g., /tmp on Unix, %TEMP% on Windows).
func WriteReport(sessionID string, reportYAML string) error {
	if sessionID == "" {
		return errSessionIDEmpty
	}

	// Ensure reports directory exists
	dir := filepath.Join(os.TempDir(), "fluxid-reports")
	if err := os.MkdirAll(dir, permDir); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Write report to file
	reportPath := getReportPath(sessionID)
	if err := os.WriteFile(reportPath, []byte(reportYAML), permFile); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// ReadReport retrieves the report for the given session ID.
// Returns an empty string if no report exists for the session.
func ReadReport(sessionID string) (string, error) {
	if sessionID == "" {
		return "", errSessionIDEmpty
	}

	reportPath := getReportPath(sessionID)

	// Check if report file exists
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		return "", nil
	}

	// Read report from file
	data, err := os.ReadFile(reportPath) // #nosec G304 - reportPath is constructed from validated sessionID
	if err != nil {
		return "", fmt.Errorf("failed to read report file: %w", err)
	}

	return string(data), nil
}

// getAbortFlagPath returns the file path for storing a session's abort flag.
// Abort flags are stored in <temp-dir>/fluxid-reports/<session-id>.abort
// where <temp-dir> is os.TempDir() (e.g., /tmp on Unix, %TEMP% on Windows).
func getAbortFlagPath(sessionID string) string {
	dir := filepath.Join(os.TempDir(), "fluxid-reports")
	return filepath.Join(dir, sessionID+".abort")
}

// SetAbortFlag sets the abort flag for the given session ID.
// This signals that the workflow should gracefully terminate.
func SetAbortFlag(sessionID string) error {
	if sessionID == "" {
		return errSessionIDEmpty
	}

	// Ensure reports directory exists
	dir := filepath.Join(os.TempDir(), "fluxid-reports")
	if err := os.MkdirAll(dir, permDir); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Write abort flag file
	abortPath := getAbortFlagPath(sessionID)
	if err := os.WriteFile(abortPath, []byte("abort"), permFile); err != nil {
		return fmt.Errorf("failed to write abort flag: %w", err)
	}

	return nil
}

// CheckAbortFlag checks if the abort flag is set for the given session ID.
// Returns true if abort was requested, false otherwise.
func CheckAbortFlag(sessionID string) (bool, error) {
	if sessionID == "" {
		return false, errSessionIDEmpty
	}

	abortPath := getAbortFlagPath(sessionID)

	// Check if abort flag file exists
	if _, err := os.Stat(abortPath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check abort flag: %w", err)
	}

	return true, nil
}

// ClearAbortFlag removes the abort flag for the given session ID.
// This is primarily for testing and cleanup.
func ClearAbortFlag(sessionID string) error {
	if sessionID == "" {
		return errSessionIDEmpty
	}

	abortPath := getAbortFlagPath(sessionID)

	// Remove abort flag file if it exists
	if err := os.Remove(abortPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear abort flag: %w", err)
	}

	return nil
}

// formatISO8601 returns the current time in ISO 8601 format with Z suffix (UTC).
func formatISO8601() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// getHistoryPath returns the file path for storing a session's history.
// History entries are stored in <temp-dir>/fluxid-reports/<session-id>.history
// where <temp-dir> is os.TempDir() (e.g., /tmp on Unix, %TEMP% on Windows).
func getHistoryPath(sessionID string) string {
	dir := filepath.Join(os.TempDir(), "fluxid-reports")
	return filepath.Join(dir, sessionID+".history")
}

// maxHistorySize is the maximum size of history per session in bytes (32MB).
const maxHistorySize = 32 * 1024 * 1024

// acquireFileLock acquires an exclusive lock on a file for inter-process synchronization.
// Returns the locked file descriptor which must be closed by the caller.
func acquireFileLock(lockPath string) (*os.File, error) {
	// Create lock file if it doesn't exist
	// #nosec G304 - lockPath is constructed internally from validated sessionID
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, permFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Acquire exclusive lock (blocking)
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		_ = lockFile.Close()
		return nil, fmt.Errorf("failed to acquire file lock: %w", err)
	}

	return lockFile, nil
}

// WriteHistoryEntry appends a timestamped history entry for the given session ID.
// The entry is prefixed with an ISO 8601 timestamp in the format [YYYY-MM-DDTHH:MM:SSZ].
// History is stored in-memory (via temp files) and scoped per session.
// If adding the new entry would exceed 32MB, oldest entries are evicted (FIFO) until under limit.
// Uses file-based locking to ensure concurrent writes from multiple processes are safe.
func WriteHistoryEntry(sessionID string, message string) error {
	if err := validateHistoryInput(sessionID, message); err != nil {
		return err
	}

	// Ensure reports directory exists
	dir := filepath.Join(os.TempDir(), "fluxid-reports")
	if err := os.MkdirAll(dir, permDir); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	historyPath := getHistoryPath(sessionID)

	// Acquire file lock for atomic read-modify-write across processes
	lockFile, err := acquireHistoryLock(historyPath)
	if err != nil {
		return err
	}
	defer releaseHistoryLock(lockFile)

	// Prepare new entry
	entry := formatHistoryEntry(message)
	entrySize := len([]byte(entry))

	// Read and process existing history
	finalContent := prepareHistoryContent(historyPath, entry, entrySize)

	// Write atomically
	return writeHistoryFile(historyPath, finalContent)
}

func validateHistoryInput(sessionID, message string) error {
	if sessionID == "" {
		return errSessionIDEmpty
	}
	if message == "" {
		return errMessageEmpty
	}
	return nil
}

func acquireHistoryLock(historyPath string) (*os.File, error) {
	lockPath := historyPath + ".lock"
	return acquireFileLock(lockPath)
}

func releaseHistoryLock(lockFile *os.File) {
	_ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
	_ = lockFile.Close()
}

func formatHistoryEntry(message string) string {
	timestamp := fmt.Sprintf("[%s]", formatISO8601())
	return fmt.Sprintf("%s %s\n", timestamp, message)
}

func prepareHistoryContent(historyPath, entry string, entrySize int) string {
	existingHistory := ""
	// #nosec G304 - historyPath is constructed from validated sessionID
	if data, err := os.ReadFile(historyPath); err == nil {
		existingHistory = string(data)
	}

	existingSize := len([]byte(existingHistory))
	totalSize := existingSize + entrySize

	// If adding new entry exceeds limit, evict oldest entries (FIFO)
	finalHistory := existingHistory
	if totalSize > maxHistorySize {
		finalHistory = evictOldestEntries(existingHistory, existingSize, entrySize)
	}

	return finalHistory + entry
}

func evictOldestEntries(existingHistory string, existingSize, entrySize int) string {
	// Split into lines
	lines := []string{}
	if existingHistory != "" {
		lines = strings.Split(strings.TrimRight(existingHistory, "\n"), "\n")
	}

	// Evict oldest entries (from beginning) until under limit
	currentSize := existingSize
	evictedLines := 0
	for evictedLines < len(lines) && currentSize+entrySize > maxHistorySize {
		lineToEvict := lines[evictedLines]
		lineSize := len([]byte(lineToEvict + "\n"))
		currentSize -= lineSize
		evictedLines++
	}

	// Keep only non-evicted lines
	remainingLines := lines[evictedLines:]
	if len(remainingLines) > 0 {
		return strings.Join(remainingLines, "\n") + "\n"
	}
	return ""
}

func writeHistoryFile(historyPath, content string) error {
	// Write atomically: write to temp file, then rename
	// Use timestamp to make temp file unique for concurrent writes
	tmpFile := fmt.Sprintf("%s.tmp.%d", historyPath, time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, []byte(content), permFile); err != nil {
		return fmt.Errorf("failed to write temporary history file: %w", err)
	}

	if err := os.Rename(tmpFile, historyPath); err != nil {
		_ = os.Remove(tmpFile) // Clean up temp file on error
		return fmt.Errorf("failed to update history file: %w", err)
	}

	return nil
}

// ReadHistory retrieves all history entries for the given session ID.
// Returns an empty string if no history exists for the session.
func ReadHistory(sessionID string) (string, error) {
	if sessionID == "" {
		return "", errSessionIDEmpty
	}

	historyPath := getHistoryPath(sessionID)

	// Check if history file exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return "", nil
	}

	// Read history from file
	data, err := os.ReadFile(historyPath) // #nosec G304 - historyPath is constructed from validated sessionID
	if err != nil {
		return "", fmt.Errorf("failed to read history file: %w", err)
	}

	return string(data), nil
}

// ClearHistory removes all history entries for the given session ID.
// This is primarily for testing and cleanup.
func ClearHistory(sessionID string) error {
	if sessionID == "" {
		return errSessionIDEmpty
	}

	historyPath := getHistoryPath(sessionID)

	// Remove history file if it exists
	if err := os.Remove(historyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	return nil
}
