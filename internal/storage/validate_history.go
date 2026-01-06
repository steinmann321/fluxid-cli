//nolint:gocognit // Validation logic: comprehensive event validation
package storage

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ValidateHistory validates a history file against the embedded YAML schema.
// ValidateHistory validates a history file against the embedded YAML schema.
//
//nolint:err113 // Validation errors provide context-specific messages per FR requirements
func ValidateHistory(filePath string) error {
	// Step 1: Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("[file]: file not found (expected: file at %s, got: file does not exist)", filePath)
		}
		if os.IsPermission(err) {
			return fmt.Errorf(
				"[file]: permission denied (expected: read permission on %s, got: permission denied - "+
					"check file ownership and permissions)", filePath)
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Empty history file is valid (no events yet)
	if fileInfo.Size() == 0 {
		return nil
	}

	// Step 2: Validate YAML security
	if err := ValidateYAMLSecurity(filePath); err != nil {
		return err
	}

	// Step 3: Read and parse YAML
	content, err := os.ReadFile(filePath) // #nosec G304 -- Path pre-validated by ResolveSessionPath
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var events []interface{}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&events); err != nil {
		return fmt.Errorf("[file]: malformed YAML (expected: valid YAML array, got: parse error: %w)", err)
	}

	// Step 4: Validate history structure
	var errors ValidationErrors

	for i, event := range events {
		errors = append(errors, validateHistoryEvent(event, i)...)
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateHistoryEvent validates a single history event.
//
//nolint:cyclop,funlen // Validation function: comprehensive field checks justify complexity and length
func validateHistoryEvent(event interface{}, index int) ValidationErrors {
	var errors ValidationErrors

	// Event must be an object
	eventMap, ok := event.(map[string]interface{})
	if !ok {
		errors = append(errors, ValidationError{
			Field:      fmt.Sprintf("events[%d]", index),
			Violation:  "wrong type",
			Constraint: "object",
			Value:      fmt.Sprintf("%T", event),
		})
		return errors
	}

	// Required fields per history-schema.yaml
	requiredFields := []string{"timestamp", "step", "status", "summary"}
	for _, field := range requiredFields {
		if _, exists := eventMap[field]; !exists {
			errors = append(errors, ValidationError{
				Field:      fmt.Sprintf("events[%d].%s", index, field),
				Violation:  "missing required field",
				Constraint: "must be present",
				Value:      "<not present>",
			})
		}
	}

	// Validate field types (excluding timestamp which can be string or time.Time from YAML parser)
	stringFields := []string{"step", "status", "summary"}
	for _, field := range stringFields {
		if val, exists := eventMap[field]; exists {
			if strVal, ok := val.(string); !ok {
				errors = append(errors, ValidationError{
					Field:      fmt.Sprintf("events[%d].%s", index, field),
					Violation:  "wrong type",
					Constraint: "string",
					Value:      fmt.Sprintf("%T", val),
				})
			} else if strVal == "" {
				errors = append(errors, ValidationError{
					Field:      fmt.Sprintf("events[%d].%s", index, field),
					Violation:  "empty value",
					Constraint: "non-empty string",
					Value:      "\"\"",
				})
			}
		}
	}

	// Validate timestamp: accept both string and time.Time (YAML auto-parses RFC3339 to time.Time)
	if val, exists := eventMap["timestamp"]; exists {
		switch typedVal := val.(type) {
		case string:
			// Empty string check
			if typedVal == "" {
				errors = append(errors, ValidationError{
					Field:      fmt.Sprintf("events[%d].timestamp", index),
					Violation:  "empty value",
					Constraint: "non-empty RFC3339 timestamp",
					Value:      "\"\"",
				})
			} else {
				// If it's a non-empty string, validate RFC3339 format
				if _, err := time.Parse(time.RFC3339, typedVal); err != nil {
					errors = append(errors, ValidationError{
						Field:      fmt.Sprintf("events[%d].timestamp", index),
						Violation:  "invalid format",
						Constraint: "ISO 8601 format (e.g., 2026-01-05T12:00:00Z)",
						Value:      typedVal,
					})
				}
			}
		case time.Time:
			// YAML auto-parsed it - this is valid, no error
		default:
			// Neither string nor time.Time - invalid type
			errors = append(errors, ValidationError{
				Field:      fmt.Sprintf("events[%d].timestamp", index),
				Violation:  "wrong type",
				Constraint: "string or RFC3339 timestamp",
				Value:      fmt.Sprintf("%T", val),
			})
		}
	}

	// Validate status enum: SUCCESS or FAIL
	if val, exists := eventMap["status"]; exists {
		if strVal, ok := val.(string); ok {
			if strVal != "SUCCESS" && strVal != "FAIL" {
				errors = append(errors, ValidationError{
					Field:      fmt.Sprintf("events[%d].status", index),
					Violation:  "invalid value",
					Constraint: "SUCCESS or FAIL",
					Value:      strVal,
				})
			}
		}
	}

	// Validate optional details field if present
	if val, exists := eventMap["details"]; exists {
		if _, ok := val.(string); !ok {
			errors = append(errors, ValidationError{
				Field:      fmt.Sprintf("events[%d].details", index),
				Violation:  "wrong type",
				Constraint: "string",
				Value:      fmt.Sprintf("%T", val),
			})
		}
	}

	return errors
}

// IsValidationError checks if an error is a ValidationError or ValidationErrors.
