package storage

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError represents a validation error with structured fields.
type ValidationError struct {
	Field      string // Field path (e.g., "status", "issues.blockers")
	Violation  string // What went wrong
	Constraint string // What was expected
	Value      string // What was actually found
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (expected: %s, got: %s)",
		e.Field, e.Violation, e.Constraint, e.Value)
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}

	var buf strings.Builder
	for i, err := range e {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(err.Error())
	}
	return buf.String()
}

// ValidateReport validates a report file against the embedded YAML schema.
//
// This function performs custom YAML validation per FR-024, FR-025, FR-026, FR-027:
// 1. Validates file exists and is readable
// 2. Checks file size < 10MB
// 3. Validates YAML security (no anchors, aliases, merge keys)
// 4. Parses YAML
// 5. Validates required fields
// 6. Validates field types
// 7. Validates enum constraints (status: PASS|FAIL)
// 8. Validates nested structures (issues object)
//
// Error formatting per Validation Contract:
// "[field_path]: [violation] (expected: [constraint], got: [value])"
//
// Exit codes:
// - Function returns ValidationErrors for validation failures (caller should use exit 1)
// - Function returns other error types for operational failures (caller should use exit 2)
//
// Parameters:
//   - filePath: Absolute path to the report file to validate
//
// Returns:
//   - ValidationErrors if validation fails
//   - Other error if operational failure (file not found, permission denied, etc.)
//   - nil if validation succeeds

// IsValidationError checks if an error is a ValidationErrors or ValidationError type.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	validationError := &ValidationError{} //nolint:exhaustruct
	isValidationError := errors.As(err, &validationError)
	var validationErrors ValidationErrors
	isValidationErrors := errors.As(err, &validationErrors)
	return isValidationError || isValidationErrors
}
