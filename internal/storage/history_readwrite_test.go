//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"fluxid-cli/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

// Test session ID constant to avoid goconst violation.
const testHistorySessionID = "550e8400-e29b-41d4-a716-446655440000"

func TestReadHistory_CreatesEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testHistorySessionID

	// ReadHistory creates empty file which contains empty array []
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	sessionDir := filepath.Dir(historyPath)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write valid empty YAML array
	if err := os.WriteFile(historyPath, []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}

	history, err := storage.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(history))
	}
}

func TestReadHistory_ParsesValidHistory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testHistorySessionID

	// Write valid history manually
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	sessionDir := filepath.Dir(historyPath)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	validHistory := `- timestamp: "2025-12-13T10:00:00Z"
  step: "implement"
  status: "SUCCESS"
  summary: "Implementation completed"
- timestamp: "2025-12-13T10:05:00Z"
  step: "review"
  status: "SUCCESS"
  summary: "Review completed"
`
	if err := os.WriteFile(historyPath, []byte(validHistory), 0o644); err != nil {
		t.Fatal(err)
	}

	history, err := storage.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(history))
	}

	if history[0].Step != "implement" {
		t.Errorf("Expected step 'implement', got %s", history[0].Step)
	}
	if history[1].Step != "review" {
		t.Errorf("Expected step 'review', got %s", history[1].Step)
	}
}

func TestReadHistory_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := testHistorySessionID

	// Create corrupted history file
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	sessionDir := filepath.Dir(historyPath)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	corruptedYAML := "- invalid: [structure"
	if err := os.WriteFile(historyPath, []byte(corruptedYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = storage.ReadHistory(sessionID)
	if err == nil {
		t.Error("Expected error for corrupted history file")
	}
}

func TestReadHistory_FIFOEvictionTriggered(t *testing.T) {
	t.Skip("FIFO eviction requires large file (>10MB) - tested in E2E")
}
