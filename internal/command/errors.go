package command

import (
	"fluxid-cli/internal/storage"
	"fmt"
	"io"
	"os"
)

// Exit codes per FR-006, FR-007, FR-040.
const (
	ExitSuccess          = 0 // Successful operation
	ExitValidationFailed = 1 // Validation failed (schema violations)
	ExitOperationalError = 2 // File system error (file not found, permission denied, schema load failure)
	ExitConfigError      = 3 // Configuration error (missing session ID, invalid session ID)
	ExitInternalError    = 4 // Internal error (unexpected failure)
)

// ErrorWriter writes errors to stderr in a consistent format.
// Per FR-041: Errors go to stderr with sufficient context.
type ErrorWriter struct {
	Stderr io.Writer
}

// NewErrorWriter creates an ErrorWriter that writes to os.Stderr.
func NewErrorWriter() *ErrorWriter {
	return &ErrorWriter{Stderr: os.Stderr}
}

// WriteError writes an error to stderr with proper formatting and returns the appropriate exit code.
//
// Per FR-040, FR-041, FR-042, FR-043:
// - Errors written to stderr (not stdout)
// - Include sufficient context (file paths, field paths, constraints, session IDs)
// - Silent success (no stderr output on success)
// - Clear, instructive error messages
//
// Returns the exit code that the command should use.
func (w *ErrorWriter) WriteError(err error, context string) int {
	if err == nil {
		// Silent success - no output to stderr
		return ExitSuccess
	}

	// Determine error type and appropriate exit code
	var exitCode int
	var formattedError string

	switch {
	case storage.IsValidationError(err):
		// Validation errors (exit 1)
		exitCode = ExitValidationFailed
		formattedError = w.formatValidationError(err, context)

	case storage.IsSecurityError(err):
		// Security errors are validation failures (exit 1)
		exitCode = ExitValidationFailed
		formattedError = err.Error()

	case storage.IsPathValidationError(err):
		// Path validation errors are configuration errors (exit 3)
		exitCode = ExitConfigError
		formattedError = err.Error()

	case isOperationalError(err):
		// Operational errors (file not found, permission denied, etc.) - exit 2
		exitCode = ExitOperationalError
		formattedError = w.formatOperationalError(err, context)

	default:
		// Internal/unexpected errors - exit 4
		exitCode = ExitInternalError
		formattedError = w.formatInternalError(err, context)
	}

	// Write to stderr
	_, _ = fmt.Fprintln(w.Stderr, formattedError)

	return exitCode
}

// formatValidationError formats validation errors with field paths and constraints.
//
// Format per Validation Contract:
// "[field_path]: [violation] (expected: [constraint], got: [value])".
// Per FR-043: Include file path for sufficient context.
func (w *ErrorWriter) formatValidationError(err error, context string) string {
	// ValidationErrors and ValidationError already have proper formatting
	// from storage.ValidationError.Error() method
	// If context includes file path, include it for FR-043
	if context != "" {
		return fmt.Sprintf("%s\n%s", context, err.Error())
	}
	return err.Error()
}

// formatOperationalError formats operational errors (file not found, permission denied, etc.)
func (w *ErrorWriter) formatOperationalError(err error, context string) string {
	// Operational errors already include sufficient context in their messages
	// (e.g., "[file]: file not found (expected: file at /path, got: file does not exist)")
	if context != "" {
		return fmt.Sprintf("%s: %v", context, err)
	}
	return err.Error()
}

// formatInternalError formats unexpected internal errors.
func (w *ErrorWriter) formatInternalError(err error, context string) string {
	if context != "" {
		return fmt.Sprintf("internal error in %s: %v", context, err)
	}
	return fmt.Sprintf("internal error: %v", err)
}

// isOperationalError checks if an error is an operational error.
// Operational errors include: file not found, permission denied, size limits, etc.
func isOperationalError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for operational error indicators
	return contains(errStr, "file not found") ||
		contains(errStr, "permission denied") ||
		contains(errStr, "empty file") ||
		contains(errStr, "file size exceeds") ||
		contains(errStr, "failed to read") ||
		contains(errStr, "failed to write") ||
		contains(errStr, "failed to stat") ||
		contains(errStr, "failed to create")
}

// contains is a helper for case-insensitive substring check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// WriteSuccess writes nothing to stderr (silent success per FR-042).
// This function exists for API consistency but does nothing.
func (w *ErrorWriter) WriteSuccess() {
	// Silent success - no output to stderr
	// Per FR-042: Successful validation produces no stderr output
}

// WriteData writes data to stdout.
// Per FR-041: stdout is reserved for data output only (not logging or errors).
//
// Parameters:
//   - data: The data to write (typically schema YAML, file paths, etc.)
func WriteData(data string) {
	_, _ = fmt.Fprintln(os.Stdout, data)
}

// WriteDataBytes writes binary data to stdout.
func WriteDataBytes(data []byte) {
	_, _ = os.Stdout.Write(data)
}
