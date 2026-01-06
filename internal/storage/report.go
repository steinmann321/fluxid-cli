package storage

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Report represents the structure of a report file written by external agents.
// Per data-model.md: Reports capture the outcome of a single workflow phase.
type Report struct {
	Command   string `yaml:"command"`   // Command that generated the report
	Artifact  string `yaml:"artifact"`  // Primary artifact produced
	Timestamp string `yaml:"timestamp"` // ISO 8601 timestamp
	Status    string `yaml:"status"`    // Enum: PASS or FAIL

	// Issues categorizes findings into 5 required categories
	Issues Issues `yaml:"issues"`

	// Optional fields
	NextSteps []string `yaml:"next_steps,omitempty"` // Recommended follow-up actions
	Summary   string   `yaml:"summary,omitempty"`    // Human-readable summary
}

// Issues represents the categorized issues in a report.
// All 5 categories are required and must be present (can be empty arrays).
type Issues struct {
	Blockers     []Issue `yaml:"blockers"`     // Must fix immediately - blocks success
	Defects      []Issue `yaml:"defects"`      // Must fix - prevents correct operation
	Concerns     []Issue `yaml:"concerns"`     // Should fix - potential problems
	Observations []Issue `yaml:"observations"` // FYI - no action needed
	Enhancements []Issue `yaml:"enhancements"` // Nice to have - optional improvements
}

// Issue represents a single issue entry with optional structured fields.
type Issue struct {
	Message    string `yaml:"message"`              // Human-readable description (required)
	Location   string `yaml:"location,omitempty"`   // Where the issue is (file:line, component name)
	Code       string `yaml:"code,omitempty"`       // Machine-readable error code
	Suggestion string `yaml:"suggestion,omitempty"` // How to fix this issue
	Reference  string `yaml:"reference,omitempty"`  // Link to documentation or more info
}

// ReadReport reads and parses a report file for the given session.
//
// This function:
// 1. Validates the session ID
// 2. Resolves the report file path
// 3. Checks file size (must be < 10MB per FR-024)
// 4. Validates YAML security (no anchors, aliases, merge keys per FR-011)
// 5. Parses YAML into Report structure
//
// Per FR-001, FR-002, FR-003: Report files contain status and categorized issues.
//
// Parameters:
//   - sessionID: The session identifier (must be valid UUID)
//   - sessionRoot: Optional session root override (from FLUXID_SESSION_ROOT in CLI layer)
//
// Returns:
//   - Parsed Report structure
//   - Error if validation or parsing fails
//
//nolint:cyclop,err113 // Validation function: complexity and context-specific errors required
func ReadReport(sessionID string, sessionRoot string) (*Report, error) {
	// Step 1: Resolve report file path
	filePath, err := ResolveSessionPath(sessionID, "report.yaml", sessionRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve report path: %w", err)
	}

	// Step 2: Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("[file]: file not found (expected: file at %s, got: file does not exist)", filePath)
		}
		return nil, fmt.Errorf("failed to stat report file: %w", err)
	}

	// Step 3: Check file size (< 10MB per FR-024)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("[file]: file size exceeds limit (expected: file size < 10MB, got: %.2f MB)",
			float64(fileInfo.Size())/(bytesPerKB*bytesPerKB))
	}

	// Step 4: Validate YAML security (reject anchors, aliases, merge keys)
	if err := ValidateYAMLSecurity(filePath); err != nil {
		return nil, err
	}

	// Step 5: Read file content
	content, err := os.ReadFile(filePath) // #nosec G304 -- Path pre-validated by ResolveSessionPath
	if err != nil {
		return nil, fmt.Errorf("failed to read report file: %w", err)
	}

	// Check for empty file
	if len(content) == 0 {
		return nil, errors.New("[file]: empty file (expected: non-empty YAML document, got: 0 bytes)")
	}

	// Step 6: Parse YAML
	var report Report
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true) // Strict decoding - reject unknown fields in nested structures

	if err := decoder.Decode(&report); err != nil {
		return nil, fmt.Errorf("[file]: malformed YAML (expected: valid YAML document, got: parse error: %w)", err)
	}

	// Step 7: Basic validation of required fields
	// More comprehensive validation happens in ValidateReport()
	if report.Command == "" {
		return nil, errors.New("command: missing required field (expected: non-empty string, got: empty)")
	}
	if report.Artifact == "" {
		return nil, errors.New("artifact: missing required field (expected: non-empty string, got: empty)")
	}
	if report.Status == "" {
		return nil, errors.New("status: missing required field (expected: PASS or FAIL, got: empty)")
	}

	return &report, nil
}

// WriteReport writes a report YAML string to the session's report file.
//
// This function:
// 1. Validates the session ID
// 2. Resolves the report file path (creates session directory if needed)
// 3. Writes the YAML content to report.yaml
//
// Per FR-001: Agents write reports via file-based interface.
//
// Parameters:
//   - sessionID: The session identifier (must be valid UUID)
//   - reportYAML: The YAML content to write
//
// Returns:
//   - Error if validation or writing fails
func WriteReport(sessionID string, reportYAML string) error {
	// Step 1: Resolve report file path (creates directory if needed)
	filePath, err := ResolveSessionPath(sessionID, "report.yaml", "")
	if err != nil {
		return fmt.Errorf("failed to resolve report path: %w", err)
	}

	// Step 2: Write report to file
	if err := os.WriteFile(filePath, []byte(reportYAML), filePermission); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}
