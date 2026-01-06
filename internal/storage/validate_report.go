//nolint:funlen // Validation logic: comprehensive checks justify complexity
package storage

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ValidateReport validates a report file against the embedded YAML schema.
// ValidateReport validates a report file against the embedded YAML schema.
//
//nolint:err113 // Validation errors provide context-specific messages per FR requirements
func ValidateReport(filePath string) error {
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

	// Step 2: Check file size
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if fileInfo.Size() > maxFileSize {
		return fmt.Errorf("[file]: file size exceeds limit (expected: file size < 10MB, got: %.2f MB)",
			float64(fileInfo.Size())/(bytesPerKB*bytesPerKB))
	}

	// Check for empty file
	if fileInfo.Size() == 0 {
		return errors.New("[file]: empty file (expected: non-empty YAML document, got: 0 bytes)")
	}

	// Step 3: Validate YAML security
	if err := ValidateYAMLSecurity(filePath); err != nil {
		return err // SecurityError is already properly formatted
	}

	// Step 4: Read and parse YAML
	content, err := os.ReadFile(filePath) // #nosec G304 -- Path pre-validated by ResolveSessionPath
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("[file]: malformed YAML (expected: valid YAML document, got: parse error: %w)", err)
	}

	// Step 5: Validate report structure
	var errors ValidationErrors

	// Required fields per report-schema.yaml
	errors = append(errors, validateReportRequiredFields(data)...)
	errors = append(errors, validateReportFieldTypes(data)...)
	errors = append(errors, validateReportEnumConstraints(data)...)
	errors = append(errors, validateReportIssuesStructure(data)...)

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateReportRequiredFields checks that all required fields are present.
func validateReportRequiredFields(data map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	requiredFields := []string{"command", "artifact", "timestamp", "status", "issues"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			errors = append(errors, ValidationError{
				Field:      field,
				Violation:  "missing required field",
				Constraint: "must be present",
				Value:      "<not present>",
			})
		}
	}

	return errors
}

// validateReportFieldTypes checks that fields have correct types.
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
func validateReportFieldTypes(data map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	// Check string fields (excluding timestamp which can be string or time.Time from YAML parser)
	stringFields := []string{"command", "artifact", "status"}
	for _, field := range stringFields {
		if val, exists := data[field]; exists {
			if strVal, ok := val.(string); !ok {
				errors = append(errors, ValidationError{
					Field:      field,
					Violation:  "wrong type",
					Constraint: "string",
					Value:      fmt.Sprintf("%T", val),
				})
			} else if strVal == "" {
				errors = append(errors, ValidationError{
					Field:      field,
					Violation:  "empty value",
					Constraint: "non-empty string",
					Value:      "\"\"",
				})
			}
		}
	}

	// Validate timestamp: accept both string and time.Time (YAML auto-parses RFC3339 to time.Time)
	if val, exists := data["timestamp"]; exists {
		switch typedVal := val.(type) {
		case string:
			// Empty string check
			if typedVal == "" {
				errors = append(errors, ValidationError{
					Field:      "timestamp",
					Violation:  "empty value",
					Constraint: "non-empty RFC3339 timestamp",
					Value:      "\"\"",
				})
			} else {
				// If it's a non-empty string, validate RFC3339 format
				if _, err := time.Parse(time.RFC3339, typedVal); err != nil {
					errors = append(errors, ValidationError{
						Field:      "timestamp",
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
				Field:      "timestamp",
				Violation:  "wrong type",
				Constraint: "string or RFC3339 timestamp",
				Value:      fmt.Sprintf("%T", val),
			})
		}
	}

	// Check issues is an object
	if val, exists := data["issues"]; exists {
		if _, ok := val.(map[string]interface{}); !ok {
			errors = append(errors, ValidationError{
				Field:      "issues",
				Violation:  "wrong type",
				Constraint: "object with 5 required categories",
				Value:      fmt.Sprintf("%T", val),
			})
		}
	}

	// Check next_steps is array if present (optional field)
	if val, exists := data["next_steps"]; exists {
		if _, ok := val.([]interface{}); !ok {
			errors = append(errors, ValidationError{
				Field:      "next_steps",
				Violation:  "wrong type",
				Constraint: "array of strings",
				Value:      fmt.Sprintf("%T", val),
			})
		}
	}

	// Check summary is string if present (optional field)
	if val, exists := data["summary"]; exists {
		if _, ok := val.(string); !ok {
			errors = append(errors, ValidationError{
				Field:      "summary",
				Violation:  "wrong type",
				Constraint: "string",
				Value:      fmt.Sprintf("%T", val),
			})
		}
	}

	return errors
}

// validateReportEnumConstraints checks enum field constraints.
func validateReportEnumConstraints(data map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	// Validate status enum: PASS or FAIL
	if val, exists := data["status"]; exists {
		if strVal, ok := val.(string); ok {
			if strVal != "PASS" && strVal != "FAIL" {
				errors = append(errors, ValidationError{
					Field:      "status",
					Violation:  "invalid value",
					Constraint: "PASS or FAIL",
					Value:      strVal,
				})
			}
		}
	}

	return errors
}

// validateReportIssuesStructure validates the nested issues object structure.
func validateReportIssuesStructure(data map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	issuesVal, exists := data["issues"]
	if !exists {
		return errors // Already reported as missing required field
	}

	issues, ok := issuesVal.(map[string]interface{})
	if !ok {
		return errors // Already reported as wrong type
	}

	// Required issue categories per report-schema.yaml
	requiredCategories := []string{"blockers", "defects", "concerns", "observations", "enhancements"}
	for _, category := range requiredCategories {
		if _, exists := issues[category]; !exists {
			errors = append(errors, ValidationError{
				Field:      "issues." + category,
				Violation:  "missing required field",
				Constraint: "array (may be empty)",
				Value:      "<not present>",
			})
			continue
		}

		// Validate category is an array
		if _, ok := issues[category].([]interface{}); !ok {
			errors = append(errors, ValidationError{
				Field:      "issues." + category,
				Violation:  "wrong type",
				Constraint: "array",
				Value:      fmt.Sprintf("%T", issues[category]),
			})
		}
	}

	return errors
}

// ValidateHistory validates a history file against the embedded YAML schema.
//
// This function performs custom YAML validation per FR-028, FR-029, FR-030, FR-031:
// 1. Validates file exists and is readable
// 2. Validates YAML security (no anchors, aliases, merge keys)
// 3. Parses YAML as array
// 4. Validates each event has required fields
// 5. Validates field types
// 6. Validates enum constraints (status: SUCCESS|FAIL)
//
// Error formatting per Validation Contract:
// "[field_path]: [violation] (expected: [constraint], got: [value])"
//
// Parameters:
//   - filePath: Absolute path to the history file to validate
//
// Returns:
//   - ValidationErrors if validation fails
//   - Other error if operational failure
//   - nil if validation succeeds
