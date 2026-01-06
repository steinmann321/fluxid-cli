package storage

import (
	"bytes"
	"errors"
	"fmt"
	"os"
)

// SecurityError represents a YAML security validation error.
type SecurityError struct {
	FilePath  string
	Feature   string // "anchor", "alias", or "merge key"
	Line      int
	Character string // The actual character found (&, *, or <<)
}

// Error implements the error interface for SecurityError.
func (e *SecurityError) Error() string {
	return fmt.Sprintf(
		"[file]: YAML %s not allowed (expected: no anchors, aliases, or merge keys, got: %s '%s' at line %d)",
		e.Feature, e.Feature, e.Character, e.Line,
	)
}

// ValidateYAMLSecurity checks a YAML file for dangerous features that could enable
// complexity attacks (billion laughs, recursive references, etc.).
//
// This function rejects:
// - Anchors (&) - allow defining reusable nodes
// - Aliases (*) - allow referencing anchors
// - Merge keys (<<) - allow merging maps
//
// These features are disabled because:
// 1. Billion laughs attack: Exponential expansion via nested aliases
// 2. Memory exhaustion: Recursive references
// 3. Complexity attacks: Deeply nested merge operations
//
// Per FR-011, FR-025: Security constraints require rejecting these features.
//
// Returns SecurityError if dangerous features are detected, nil otherwise.
func ValidateYAMLSecurity(filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath) // #nosec G304 -- Path pre-validated by caller
	if err != nil {
		return fmt.Errorf("failed to read file for security validation: %w", err)
	}

	// Check for dangerous YAML features by scanning for specific character patterns
	lines := bytes.Split(content, []byte("\n"))

	for lineNum, line := range lines {
		lineStr := string(line)

		// Check for anchors (&)
		// Anchors appear as "key: &anchor_name value" or "- &anchor_name value"
		if bytes.Contains(line, []byte("&")) {
			return &SecurityError{
				FilePath:  filePath,
				Feature:   "anchor",
				Line:      lineNum + 1, // Line numbers are 1-indexed for user display
				Character: "&",
			}
		}

		// Check for aliases (*)
		// Aliases appear as "key: *anchor_name" or "- *anchor_name"
		if bytes.Contains(line, []byte("*")) {
			return &SecurityError{
				FilePath:  filePath,
				Feature:   "alias",
				Line:      lineNum + 1,
				Character: "*",
			}
		}

		// Check for merge keys (<<)
		// Merge keys appear as "<<: *anchor" to merge another map
		if bytes.Contains(line, []byte("<<")) {
			return &SecurityError{
				FilePath:  filePath,
				Feature:   "merge key",
				Line:      lineNum + 1,
				Character: "<<",
			}
		}

		// Note: We intentionally do NOT parse the YAML here to avoid triggering
		// the very attacks we're trying to prevent. Simple byte-level scanning
		// is safer and faster for security validation.

		// This approach may have false positives (e.g., anchor character in a string),
		// but security constraints prefer rejecting edge cases over allowing attacks.
		// Per constitution principle VII: Fail fast with clear diagnostics.
		// Users can avoid these characters in YAML content, or escape them in strings
		// if needed (though for report/history files, these characters shouldn't appear
		// in normal usage).
		_ = lineStr // Keep for potential future detailed error messages
	}

	return nil
}

// IsSecurityError checks if an error is a SecurityError.
func IsSecurityError(err error) bool {
	securityError := &SecurityError{} //nolint:exhaustruct
	ok := errors.As(err, &securityError)
	return ok
}
