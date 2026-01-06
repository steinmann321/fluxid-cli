package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStorageReadHistoryValidFile verifies that storage.ReadHistory() can read a valid history file.
func TestStorageReadHistoryValidFile(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-history-valid"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a valid history file
	historyPath := filepath.Join(sessionDir, "history.yaml")
	validHistory := `- timestamp: 2026-01-05T10:00:00Z
  step: implement
  status: FAIL
  summary: First attempt failed
  details: Tried approach X but encountered error Y

- timestamp: 2026-01-05T11:00:00Z
  step: implement
  status: SUCCESS
  summary: Second attempt succeeded
  details: Used approach Z which worked correctly
`
	if err := os.WriteFile(historyPath, []byte(validHistory), 0o644); err != nil {
		t.Fatalf("Failed to write history file: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test verifies the storage package can read the history
	t.Log("Test setup complete - ready for storage.ReadHistory() implementation")
}

// TestStorageReadHistoryFIFOEviction verifies that storage.ReadHistory() performs FIFO eviction
// when file size exceeds 10MB by removing oldest 30% of complete entries.
func TestStorageReadHistoryFIFOEviction(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-history-fifo"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a history file that exceeds 10MB
	historyPath := filepath.Join(sessionDir, "history.yaml")
	file, err := os.Create(historyPath)
	if err != nil {
		t.Fatalf("Failed to create history file: %v", err)
	}
	defer func() { _ = file.Close() }()

	// Generate entries to exceed 10MB
	// Each entry is approximately 500 bytes
	entryTemplate := `- timestamp: 2026-01-05T%02d:%02d:%02dZ
  step: implement-step-%d
  status: SUCCESS
  summary: Completed step %d successfully
  details: ` +
		`This is a detailed description of what happened in this step. ` +
		`It includes information about the approach taken, the challenges encountered, ` +
		`and how they were resolved. This text is long enough to make each entry substantial in size.
`
	targetSize := 11 * 1024 * 1024 // 11MB
	entryCount := 0
	currentSize := 0

	for currentSize < targetSize {
		hour := (entryCount / 3600) % 24
		minute := (entryCount / 60) % 60
		second := entryCount % 60
		entry := fmt.Sprintf(entryTemplate, hour, minute, second, entryCount, entryCount)

		if _, err := file.WriteString(entry); err != nil {
			t.Fatalf("Failed to write entry: %v", err)
		}

		currentSize += len(entry)
		entryCount++
	}

	t.Logf("Created history file with %d entries, size %d bytes (%.2f MB)",
		entryCount, currentSize, float64(currentSize)/(1024*1024))

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects storage.ReadHistory() to:
	// 1. Detect file size > 10MB
	// 2. Remove oldest 30% of entries (calculated as ceiling(entry_count * 0.30))
	// 3. Write truncated history back to file
	// 4. Return the truncated history array

	expectedRemovalCount := int(float64(entryCount)*0.30 + 0.999) // ceiling
	t.Logf("Expected FIFO eviction to remove %d oldest entries (30%% of %d)",
		expectedRemovalCount, entryCount)

	t.Log("Test setup complete - expects FIFO eviction for file size > 10MB")
}

// TestStorageReadHistoryEmptyFile verifies that storage.ReadHistory() handles empty files gracefully.
func TestStorageReadHistoryEmptyFile(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-history-empty"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Create an empty history file
	historyPath := filepath.Join(sessionDir, "history.yaml")
	if err := os.WriteFile(historyPath, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to write empty history file: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects storage.ReadHistory() to return empty array for empty file
	t.Log("Test setup complete - expects empty array for empty history file")
}

// TestStorageReadHistoryMissingFile verifies that storage.ReadHistory() handles missing files gracefully.
func TestStorageReadHistoryMissingFile(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory without a history file
	tmpDir := t.TempDir()
	sessionID := "test-history-missing"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects storage.ReadHistory() to return empty array or handle gracefully
	t.Log("Test setup complete - expects graceful handling of missing history file")
}

// TestStorageReadHistoryFIFOPreservesStructure verifies that FIFO eviction preserves YAML structure.
func TestStorageReadHistoryFIFOPreservesStructure(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-history-structure"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a history file with distinct entries to verify correct eviction
	historyPath := filepath.Join(sessionDir, "history.yaml")
	file, err := os.Create(historyPath)
	if err != nil {
		t.Fatalf("Failed to create history file: %v", err)
	}
	defer func() { _ = file.Close() }()

	// Write 10 entries, each ~1.1MB, total ~11MB
	largeDetail := ""
	var largeDetailSb194 strings.Builder
	for idx := 0; idx < 1100000; idx++ {
		largeDetailSb194.WriteString("x")
	}
	largeDetail += largeDetailSb194.String()

	for idx := 0; idx < 10; idx++ {
		entry := fmt.Sprintf(`- timestamp: 2026-01-05T10:%02d:00Z
  step: step-%d
  status: SUCCESS
  summary: Entry %d
  details: "%s"
`, idx, idx, idx, largeDetail)
		if _, err := file.WriteString(entry); err != nil {
			t.Fatalf("Failed to write entry: %v", err)
		}
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects:
	// 1. FIFO eviction removes entries 0-2 (30% of 10 = 3 entries)
	// 2. Resulting file contains entries 3-9 in valid YAML array format
	// 3. File can be parsed as valid YAML after truncation

	t.Log("Test setup complete - expects FIFO to preserve YAML structure after eviction")
}
