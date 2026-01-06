package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// UUID v4 regex pattern for validation.
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// PathValidationError represents a path validation error.
type PathValidationError struct {
	Field      string
	Violation  string
	Constraint string
	Value      string
}

// Error implements the error interface for PathValidationError.
func (e *PathValidationError) Error() string {
	return fmt.Sprintf("%s: %s (expected: %s, got: %q)",
		e.Field, e.Violation, e.Constraint, e.Value)
}

// ValidateSessionID validates that a session ID is a valid UUID format.
// Per FR-012, FR-026: Session IDs must be valid UUIDs to prevent path traversal.
//
// Returns PathValidationError if invalid, nil otherwise.
func ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return &PathValidationError{
			Field:      "FLUXID_SESSION_ID",
			Violation:  "environment variable not set",
			Constraint: "valid UUID",
			Value:      "<not set>",
		}
	}

	// Check for path traversal characters
	if strings.Contains(sessionID, "..") || strings.Contains(sessionID, "/") || strings.Contains(sessionID, "\\") {
		return &PathValidationError{
			Field:      "session_id",
			Violation:  "path traversal not allowed",
			Constraint: "valid UUID v4 (no path separators or .. sequences)",
			Value:      sessionID,
		}
	}

	// Validate UUID v4 format (lowercase hex with dashes)
	if !uuidPattern.MatchString(strings.ToLower(sessionID)) {
		return &PathValidationError{
			Field:      "session_id",
			Violation:  "invalid format",
			Constraint: "valid UUID v4 (format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)",
			Value:      sessionID,
		}
	}

	return nil
}

// GetSessionRoot returns the session root directory, using the following precedence:
// 1. rootOverride parameter (if provided)
// 2. .fluxid/sessions/ in current working directory
// 3. os.TempDir()/fluxid/ as fallback (cross-platform)
//
// Per FR-010, FR-022: Session files must be scoped to session-specific directories.
// Note: Callers in CLI layer should pass os.Getenv("FLUXID_SESSION_ROOT") as rootOverride.
func GetSessionRoot(rootOverride string) (string, error) {
	// Check override parameter first (set by CLI layer from FLUXID_SESSION_ROOT env var)
	if rootOverride != "" {
		absRoot, err := filepath.Abs(rootOverride)
		if err != nil {
			return "", fmt.Errorf("failed to resolve session root override: %w", err)
		}
		return absRoot, nil
	}

	// Try .fluxid/sessions/ in current working directory
	cwd, err := os.Getwd()
	if err == nil {
		cwdRoot := filepath.Join(cwd, ".fluxid", "sessions")
		// Check if it exists or can be created
		if _, statErr := os.Stat(cwdRoot); statErr == nil || os.IsNotExist(statErr) {
			return cwdRoot, nil
		}
	}

	// Fall back to OS temp directory (cross-platform)
	tempRoot := filepath.Join(os.TempDir(), "fluxid")
	return tempRoot, nil
}

// ResolveSessionPath resolves and validates a session-specific file path.
//
// This function:
// 1. Validates the session ID (must be valid UUID)
// 2. Resolves the session root directory
// 3. Constructs the session directory path
// 4. Resolves symlinks to prevent escape via symbolic links
// 5. Verifies the resolved path is within the session root
// 6. Appends the filename to get the full file path
//
// Per FR-012, FR-026: Path validation prevents path traversal attacks.
//
// Parameters:
//   - sessionID: The session identifier (must be valid UUID)
//   - filename: The file name to append (e.g., "report.yaml", "history.yaml")
//   - rootOverride: Optional session root override (from FLUXID_SESSION_ROOT env var in CLI layer)
//
// Returns:
//   - Absolute path to the file
//   - PathValidationError if validation fails, other error if resolution fails
func ResolveSessionPath(sessionID, filename, rootOverride string) (string, error) {
	// Step 1: Validate session ID
	if err := ValidateSessionID(sessionID); err != nil {
		return "", err
	}

	// Step 2: Get session root
	sessionRoot, err := GetSessionRoot(rootOverride)
	if err != nil {
		return "", fmt.Errorf("failed to get session root: %w", err)
	}

	// Step 3: Construct session directory path
	sessionDir := filepath.Join(sessionRoot, sessionID)

	// Step 4: Ensure session directory exists (create if needed)
	if err := os.MkdirAll(sessionDir, directoryPermission); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	// Step 5: Construct full file path
	filePath := filepath.Join(sessionDir, filename)

	// Step 6: Resolve symlinks for security validation
	// If file doesn't exist, EvalSymlinks fails, so we evaluate the directory only
	resolvedDir, err := filepath.EvalSymlinks(sessionDir)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}
	if err == nil {
		// Directory resolved, now verify it's within session root
		// Resolve session root symlinks too for proper comparison (e.g., macOS /var -> /private/var)
		resolvedRoot, rootErr := filepath.EvalSymlinks(sessionRoot)
		if rootErr != nil {
			return "", fmt.Errorf("failed to resolve session root symlinks: %w", rootErr)
		}

		// Check if resolved directory is within resolved session root
		if !strings.HasPrefix(resolvedDir, resolvedRoot) {
			return "", &PathValidationError{
				Field:      "path",
				Violation:  "path traversal not allowed",
				Constraint: "path within " + resolvedRoot,
				Value:      resolvedDir,
			}
		}
	}

	return filePath, nil
}

// EnsureFileExists creates an empty file if it doesn't exist.
// Used for history files that start empty and grow via appends.
//
// Returns error if file creation fails, nil if file exists or is created successfully.
func EnsureFileExists(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, nothing to do
		return nil
	} else if !os.IsNotExist(err) {
		// Some other error (permission denied, etc.)
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	// File doesn't exist, create empty file
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, filePermission) // #nosec G304 -- Path pre-validated by caller
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return nil
}

// IsPathValidationError checks if an error is a PathValidationError.
func IsPathValidationError(err error) bool {
	pathValidationError := &PathValidationError{} //nolint:exhaustruct
	ok := errors.As(err, &pathValidationError)
	return ok
}
