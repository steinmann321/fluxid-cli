//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"fluxid-cli/internal/storage"
	"fmt"
	"os"
	"testing"
)

// Test FIFO eviction behavior

//nolint:cyclop,perfsprint,varnamelen,funlen // Test function with large file generation
func TestReadHistory_FIFOEvictionLargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440401"

	// Create history file that exceeds 10MB
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}

	// Create many history events to exceed 10MB
	// Each event is ~200 bytes, so we need ~50,000 events
	// But to keep test fast, we'll create fewer events with longer summaries
	var historyYAML string
	eventCount := 10000

	for i := 0; i < eventCount; i++ {
		// Each event with ~1KB summary = 10,000 events * 1KB = ~10MB
		summary := fmt.Sprintf("Event %d: ", i)
		// Add ~1KB of padding to summary
		for j := 0; j < 100; j++ {
			summary += "This is a long summary with many details about the event execution. "
		}

		historyYAML += fmt.Sprintf(`- timestamp: "2025-12-13T10:%02d:%02d.000Z"
  step: "step-%d"
  status: "SUCCESS"
  summary: "%s"
  details: "Additional details for event %d"
`, i/3600, (i%3600)/60, i, summary, i)
	}

	// Write large history file
	if err := os.WriteFile(historyPath, []byte(historyYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Verify file size exceeds 10MB
	fileInfo, err := os.Stat(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Size() <= 10*1024*1024 {
		t.Skipf("Test file size is %d bytes, expected > 10MB for FIFO eviction test", fileInfo.Size())
	}

	// ReadHistory should trigger FIFO eviction
	history, err := storage.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	// After eviction, history should have fewer entries (70% retained)
	// Original: 10000 events, after eviction: ~7000 events
	if len(history) >= eventCount {
		t.Errorf("Expected history to be truncated, got %d entries (same as original %d)", len(history), eventCount)
	}

	// Verify oldest entries were removed (FIFO)
	// The first event in truncated history should NOT be event 0
	if len(history) > 0 {
		if history[0].Step == "step-0" {
			t.Error("Expected oldest events to be removed, but step-0 is still present")
		}
	}

	// Verify file was rewritten with truncated history
	fileInfo2, err := os.Stat(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo2.Size() >= fileInfo.Size() {
		t.Error("Expected file size to decrease after FIFO eviction")
	}
}

func TestPerformFIFOEviction_EdgeCases(t *testing.T) {
	t.Skip("performFIFOEviction is an internal function, tested via ReadHistory")
}
