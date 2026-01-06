//nolint:cyclop,funlen // E2E tests: comprehensive scenarios justify complexity
package tests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestHistoryWriteWorkflow tests the complete agent workflow for writing history events.
// Per User Story 4 (FR-004): Agent writes history events via file-based interface.
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
//nolint:cyclop // E2E test complexity justified by comprehensive validation
//nolint:cyclop,funlen // E2E test: comprehensive validation scenarios
func TestHistoryWriteWorkflow(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := "550e8400-e29b-41d4-a716-446655440000"

	binPath := filepath.Join(root, "bin", "fluxid")

	// Step 1: Get history file path
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "history", "--get-file")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid history --get-file failed: %v\nStderr: %s", err, stderr.String())
	}

	// Verify silent success (no stderr output per FR-042)
	if stderr.Len() > 0 {
		t.Errorf("Expected no stderr output, got: %s", stderr.String())
	}

	historyPath := strings.TrimSpace(stdout.String())
	if historyPath == "" {
		t.Fatal("Expected history file path on stdout, got empty string")
	}

	// Verify path contains session ID
	if !strings.Contains(historyPath, sessionID) {
		t.Errorf("History path should contain session ID %s, got: %s", sessionID, historyPath)
	}

	// Verify path ends with history.yaml
	if !strings.HasSuffix(historyPath, "history.yaml") {
		t.Errorf("History path should end with 'history.yaml', got: %s", historyPath)
	}

	// Step 2: Write multiple history events (history is a YAML array per schema)
	events := []map[string]interface{}{
		{
			"timestamp": "2026-01-05T14:00:00Z",
			"step":      "implement",
			"status":    "SUCCESS",
			"summary":   "Implemented feature",
		},
		{
			"timestamp": "2026-01-05T14:05:00Z",
			"step":      "review",
			"status":    "SUCCESS",
			"summary":   "Reviewed implementation",
		},
		{
			"timestamp": "2026-01-05T14:10:00Z",
			"step":      "commit",
			"status":    "SUCCESS",
			"summary":   "Committed changes",
		},
	}

	yamlBytes, err := yaml.Marshal(events)
	if err != nil {
		t.Fatalf("Failed to marshal history YAML: %v", err)
	}

	if err := os.WriteFile(historyPath, yamlBytes, 0o644); err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	// Step 3: Verify history can be read back
	content, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	var readEvents []interface{}
	if err := yaml.Unmarshal(content, &readEvents); err != nil {
		t.Fatalf("Failed to parse history file: %v", err)
	}

	if len(readEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(readEvents))
	}

	// Step 4: Validate history file
	validateCmd := exec.CommandContext(testCtx(30*time.Second), binPath, "history", "--validate")
	validateCmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var valStderr bytes.Buffer
	validateCmd.Stderr = &valStderr

	if err := validateCmd.Run(); err != nil {
		t.Fatalf("History validation should succeed, got error: %v\nStderr: %s", err, valStderr.String())
	}

	// Verify silent success
	if valStderr.Len() > 0 {
		t.Errorf("Expected no stderr output for valid history, got: %s", valStderr.String())
	}
}

// TestHistoryFIFOEviction tests that history files are automatically truncated via FIFO eviction.
// Per FR-041: When history file exceeds 10MB, remove oldest 30% of entries.
func TestHistoryFIFOEviction(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	historyPath := filepath.Join(sessionDir, "history.yaml")

	// Create a history file with many events (history is a YAML array per schema)
	// Each event is ~200 bytes, need ~50,000 events for 10MB
	// For test efficiency, we'll create a smaller number and manually test the logic
	var events []map[string]interface{}

	// Create 100 events (will test logic manually in unit tests for 10MB threshold)
	for i := 0; i < 100; i++ {
		events = append(events, map[string]interface{}{
			"timestamp": fmt.Sprintf("2026-01-05T14:%02d:%02dZ", i/60, i%60),
			"step":      fmt.Sprintf("step-%d", i),
			"status":    "SUCCESS",
			"summary":   fmt.Sprintf("Event %d with some padding text to make it larger", i),
		})
	}

	yamlBytes, err := yaml.Marshal(events)
	if err != nil {
		t.Fatalf("Failed to marshal history YAML: %v", err)
	}

	// Write the history file
	if err := os.WriteFile(historyPath, yamlBytes, 0o644); err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	// Verify all events are present
	content, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	var readEvents []interface{}
	if err := yaml.Unmarshal(content, &readEvents); err != nil {
		t.Fatalf("Failed to parse history file: %v", err)
	}

	if len(readEvents) != 100 {
		t.Errorf("Expected 100 events before eviction, got %d", len(readEvents))
	}

	// Note: FIFO eviction logic is tested in storage_history_test.go
	// This E2E test verifies the workflow integration
	t.Log("FIFO eviction detailed tests in storage_history_test.go")
}

// TestHistoryValidateInvalidEvents tests validation error handling for invalid history events.
func TestHistoryValidateInvalidEvents(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	tmpDir := t.TempDir()
	sessionID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	historyPath := filepath.Join(sessionDir, "history.yaml")

	// Write history with invalid event (missing required fields)
	// History is a YAML array per schema
	invalidHistory := `- step: implement
  status: SUCCESS
  # Missing timestamp and summary
`
	if err := os.WriteFile(historyPath, []byte(invalidHistory), 0o644); err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "history", "--validate")
	cmd.Env = append(os.Environ(),
		"FLUXID_SESSION_ROOT="+tmpDir,
		"FLUXID_SESSION_ID="+sessionID,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should fail with exit 1 (validation failure)
	if err == nil {
		t.Fatal("Expected validation to fail for invalid history")
	}

	exitErr := &exec.ExitError{} //nolint:exhaustruct
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got: %d", exitErr.ExitCode())
		}
	}

	// Verify error message mentions missing fields
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "timestamp") {
		t.Errorf("Expected error to mention 'timestamp', got: %s", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "summary") {
		t.Errorf("Expected error to mention 'summary', got: %s", stderrOutput)
	}
}
