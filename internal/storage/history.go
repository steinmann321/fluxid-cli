// Package storage provides file-based storage operations for reports and history.
package storage

import (
	"bytes"
	"fmt"
	"math"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	// File size and eviction constants.
	bytesPerKB          = 1024
	evictionPercent     = 0.30
	yamlIndentSpaces    = 2
	filePermission      = 0o644
	directoryPermission = 0o755
)

// HistoryEvent represents a single event from a workflow step.
// Per data-model.md: History records workflow events to prevent repeating failed approaches.
type HistoryEvent struct {
	Timestamp string `yaml:"timestamp"`         // ISO 8601 timestamp (UTC)
	Step      string `yaml:"step"`              // Workflow step name
	Status    string `yaml:"status"`            // Enum: SUCCESS or FAIL
	Summary   string `yaml:"summary"`           // Brief outcome description
	Details   string `yaml:"details,omitempty"` // Optional: detailed approach and failure reason
}

// History represents the array of historical workflow events.
type History []HistoryEvent

// ReadHistory reads and parses a history file for the given session.
//
// This function implements FIFO eviction per FR-022:
// 1. If file size > 10MB, remove oldest 30% of entries
// 2. Eviction uses ceiling function: ceiling(entry_count * evictionPercent)
// 3. Minimum 1 entry removed when entry_count >= 4
// 4. Only whole event objects are removed (preserve YAML structure)
//
// Steps:
// 1. Validates the session ID
// 2. Resolves the history file path
// 3. Creates empty file if it doesn't exist
// 4. Checks file size
// 5. If > 10MB, performs FIFO eviction (removes oldest entries)
// 6. Validates YAML security (no anchors, aliases, merge keys per FR-011)
// 7. Parses YAML into History array
//
// Per FR-019, FR-020, FR-021: History files track workflow events across sessions.
//
// Parameters:
//   - sessionID: The session identifier (must be valid UUID)
//
// Returns:
//   - Parsed History array (may be empty if file is new or after eviction)
//   - Error if validation or parsing fails
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
func ReadHistory(sessionID string) (History, error) {
	// Step 1: Resolve history file path
	filePath, err := ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve history path: %w", err)
	}

	// Step 2: Ensure file exists (create if needed)
	if err := EnsureFileExists(filePath); err != nil {
		return nil, fmt.Errorf("failed to ensure history file exists: %w", err)
	}

	// Step 3: Check file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat history file: %w", err)
	}

	const maxFileSize = 10 * 1024 * 1024 // 10MB
	needsEviction := fileInfo.Size() > maxFileSize

	// Step 4: Read file content
	content, err := os.ReadFile(filePath) // #nosec G304 -- Path pre-validated by ResolveSessionPath
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	// Handle empty file
	if len(content) == 0 {
		return History{}, nil
	}

	// Step 5: Validate YAML security (reject anchors, aliases, merge keys)
	if err := ValidateYAMLSecurity(filePath); err != nil {
		return nil, err
	}

	// Step 6: Parse YAML
	var history History
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true) // Strict decoding

	if err := decoder.Decode(&history); err != nil {
		// If decode fails and file is not empty, it's malformed
		if len(content) > 0 {
			return nil, fmt.Errorf("[file]: malformed YAML (expected: valid YAML array, got: parse error: %w)", err)
		}
		// Empty content decodes to empty array
		return History{}, nil
	}

	// Step 7: Perform FIFO eviction if file is too large
	if needsEviction && len(history) > 0 {
		evictedHistory, evictErr := performFIFOEviction(history, filePath)
		if evictErr != nil {
			return nil, fmt.Errorf("failed to perform FIFO eviction: %w", evictErr)
		}
		history = evictedHistory
	}

	return history, nil
}

// performFIFOEviction removes the oldest 30% of entries from history and writes back to file.
//
// Eviction calculation per FR-022:
// - Remove oldest 30% of entries (from array index 0)
// - Use ceiling function: ceiling(entry_count * evictionPercent)
// - Example: 10 entries -> remove ceiling(10 * evictionPercent) = ceiling(3.0) = 3 entries
// - Example: 5 entries -> remove ceiling(5 * evictionPercent) = ceiling(1.5) = 2 entries
// - Minimum 1 entry removed when entry_count >= 4
//
// Parameters:
//   - history: The current history array
//   - filePath: The path to the history file (for writing back truncated version)
//
// Returns:
//   - Truncated history array
//   - Error if eviction or write-back fails
func performFIFOEviction(history History, filePath string) (History, error) {
	if len(history) == 0 {
		return history, nil
	}

	// Calculate number of entries to remove (30% using ceiling)
	entryCount := len(history)
	removeCount := int(math.Ceil(float64(entryCount) * evictionPercent))

	// Ensure we remove at least 1 entry if we have 4 or more entries
	if entryCount >= 4 && removeCount < 1 {
		removeCount = 1
	}

	// Ensure we don't remove more entries than we have
	if removeCount >= entryCount {
		// Edge case: would remove all entries
		// Keep at least the most recent entry
		removeCount = entryCount - 1
		if removeCount < 0 {
			removeCount = 0
		}
	}

	// Remove oldest entries (from index 0)
	truncatedHistory := history[removeCount:]

	// Write truncated history back to file
	if err := writeHistoryToFile(truncatedHistory, filePath); err != nil {
		return nil, fmt.Errorf("failed to write truncated history: %w", err)
	}

	return truncatedHistory, nil
}

// writeHistoryToFile writes a history array to a file in YAML format.
//
// Parameters:
//   - history: The history array to write
//   - filePath: The path to the history file
//
// Returns:
//   - Error if write fails, nil otherwise
func writeHistoryToFile(history History, filePath string) error {
	// Marshal history to YAML
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(yamlIndentSpaces) // Use 2-space indentation

	if err := encoder.Encode(history); err != nil {
		return fmt.Errorf("failed to marshal history to YAML: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close YAML encoder: %w", err)
	}

	// Write to file (overwrite existing)
	if err := os.WriteFile(filePath, buf.Bytes(), filePermission); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}
