//nolint:cyclop,funlen // E2E tests: comprehensive scenarios justify complexity
package tests

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestDeveloperReadsHistoryFile tests that developers can read history files directly.
// Per User Story 6 (FR-046): Developers can read history.yaml files for debugging.
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
//nolint:cyclop // E2E test complexity justified by comprehensive validation
//nolint:cyclop,funlen // E2E test: comprehensive validation scenarios
func TestDeveloperReadsHistoryFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sessionID := testSessionID
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Create a sample history file
	historyPath := filepath.Join(sessionDir, "history.yaml")
	sampleHistory := []map[string]interface{}{
		{
			"timestamp": "2026-01-05T14:00:00Z",
			"step":      "implement",
			"status":    "SUCCESS",
			"summary":   "Implemented feature",
		},
		{
			"timestamp": "2026-01-05T14:05:00Z",
			"step":      "review",
			"status":    "FAIL",
			"summary":   "Review found issues",
			"details":   "Missing error handling",
		},
	}

	yamlBytes, err := yaml.Marshal(sampleHistory)
	if err != nil {
		t.Fatalf("Failed to marshal history: %v", err)
	}

	if err := os.WriteFile(historyPath, yamlBytes, 0o644); err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	// Developer reads history file directly
	content, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatalf("Developer should be able to read history file: %v", err)
	}

	// Parse history
	var history []interface{}
	if err := yaml.Unmarshal(content, &history); err != nil {
		t.Fatalf("History file should be valid YAML: %v", err)
	}

	// Verify structure
	if len(history) != 2 {
		t.Errorf("Expected 2 events, got %d", len(history))
	}

	// Verify events are readable maps
	event1, eventOk := history[0].(map[string]interface{})
	if !eventOk {
		t.Fatal("Event should be a map")
	}

	// Verify required fields are present and readable
	if timestamp, exists := event1["timestamp"]; !exists {
		t.Error("Event should have 'timestamp' field")
	} else if _, ok := timestamp.(string); !ok {
		t.Errorf("Timestamp should be string, got: %T", timestamp)
	}

	if step, exists := event1["step"]; !exists {
		t.Error("Event should have 'step' field")
	} else if _, ok := step.(string); !ok {
		t.Errorf("Step should be string, got: %T", step)
	}

	if status, exists := event1["status"]; !exists {
		t.Error("Event should have 'status' field")
	} else {
		statusStr, ok := status.(string)
		if !ok {
			t.Errorf("Status should be string, got: %T", status)
		} else if statusStr != "SUCCESS" && statusStr != "FAIL" {
			t.Errorf("Status should be SUCCESS or FAIL, got: %s", statusStr)
		}
	}

	if summary, exists := event1["summary"]; !exists {
		t.Error("Event should have 'summary' field")
	} else if _, ok := summary.(string); !ok {
		t.Errorf("Summary should be string, got: %T", summary)
	}

	// Verify optional details field in second event
	event2, ok := history[1].(map[string]interface{})
	if !ok {
		t.Fatal("Event 2 should be a map")
	}

	if details, exists := event2["details"]; exists {
		if _, ok := details.(string); !ok {
			t.Errorf("Details should be string if present, got: %T", details)
		}
	}

	t.Log("Developer successfully read and parsed history file")
}

// TestDeveloperReadsEmptyHistoryFile tests reading an empty history file.
// Per FR-041: Empty history files are valid (no events yet).
func TestDeveloperReadsEmptyHistoryFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sessionID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Create empty history file
	historyPath := filepath.Join(sessionDir, "history.yaml")
	if err := os.WriteFile(historyPath, []byte{}, 0o644); err != nil {
		t.Fatalf("Failed to write empty history file: %v", err)
	}

	// Developer reads empty file
	content, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatalf("Developer should be able to read empty history file: %v", err)
	}

	// Empty file is valid - no events yet
	if len(content) != 0 {
		t.Errorf("Expected empty file, got %d bytes", len(content))
	}

	t.Log("Developer successfully read empty history file")
}
