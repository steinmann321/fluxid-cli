//nolint:paralleltest // Tests modify environment variables
package storage_test

import (
	"fluxid-cli/internal/storage"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Edge case tests for FIFO eviction logic

//nolint:cyclop,perfsprint // Test functions with skipped large file generation
func TestReadHistory_FIFOWithSmallHistory(t *testing.T) {
	t.Skip("Large file tests are too slow - FIFO eviction covered by E2E tests")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440501"

	// Create history file with only 2 events that exceeds 10MB (large summaries)
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}

	// Each event with ~6MB summary to exceed 10MB total with just 2 events
	// This tests the edge case: removeCount >= entryCount
	largeSummary1 := "Event 1: "
	for i := 0; i < 600000; i++ {
		largeSummary1 += "This is event 1 with a very long summary to make the file exceed 10MB. "
	}

	largeSummary2 := "Event 2: "
	for i := 0; i < 600000; i++ {
		largeSummary2 += "This is event 2 with a very long summary to make the file exceed 10MB. "
	}

	historyYAML := fmt.Sprintf(`- timestamp: "2025-12-13T10:00:00Z"
  step: "step-1"
  status: "SUCCESS"
  summary: "%s"
- timestamp: "2025-12-13T10:01:00Z"
  step: "step-2"
  status: "SUCCESS"
  summary: "%s"
`, largeSummary1, largeSummary2)

	if err := os.WriteFile(historyPath, []byte(historyYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Verify file size exceeds 10MB
	fileInfo, err := os.Stat(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Size() <= 10*1024*1024 {
		t.Skipf("Test file size is %d bytes, expected > 10MB", fileInfo.Size())
	}

	// ReadHistory should handle edge case: keep at least 1 entry
	history, err := storage.ReadHistory(sessionID, "")
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	// Should have at least 1 entry (edge case: can't remove all)
	if len(history) < 1 {
		t.Error("Expected at least 1 entry after eviction")
	}

	// Should have removed oldest entry
	if len(history) > 0 && history[0].Step == "step-1" {
		t.Error("Expected oldest entry (step-1) to be removed")
	}
}

//nolint:perfsprint // Test functions with skipped large file generation
func TestReadHistory_FIFOWithExactlyOneEvent(t *testing.T) {
	t.Skip("Large file tests are too slow - FIFO eviction covered by E2E tests")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440502"

	// Create history file with exactly 1 event that exceeds 10MB
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}

	// Single event with >10MB summary
	largeSummary := "Event 1: "
	for i := 0; i < 1200000; i++ {
		largeSummary += "This is a single event with a very long summary exceeding 10MB. "
	}

	historyYAML := fmt.Sprintf(`- timestamp: "2025-12-13T10:00:00Z"
  step: "only-step"
  status: "SUCCESS"
  summary: "%s"
`, largeSummary)

	if err := os.WriteFile(historyPath, []byte(historyYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Verify file size exceeds 10MB
	fileInfo, err := os.Stat(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Size() <= 10*1024*1024 {
		t.Skipf("Test file size is %d bytes, expected > 10MB", fileInfo.Size())
	}

	// ReadHistory should handle edge case: single event, can't remove it
	history, err := storage.ReadHistory(sessionID, "")
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	// Should keep the single event (edge case: removeCount would be 1, but entryCount is 1)
	if len(history) != 1 {
		t.Errorf("Expected 1 entry after eviction attempt, got %d", len(history))
	}
}

//nolint:cyclop,perfsprint,varnamelen // Test functions with skipped large file generation
func TestReadHistory_FIFOWithFourEvents(t *testing.T) {
	t.Skip("Large file tests are too slow - FIFO eviction covered by E2E tests")
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "550e8400-e29b-41d4-a716-446655440503"

	// Create history file with exactly 4 events that exceeds 10MB
	// This tests the edge case: entryCount >= 4 && removeCount < 1
	// (though with 30% eviction, 4 * 0.3 = 1.2, ceil = 2, so this path might not trigger)
	// Let's create this to ensure we test the condition
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", "")
	if err != nil {
		t.Fatal(err)
	}

	// Each event with ~3MB summary to exceed 10MB total with 4 events
	var historyYAML string
	for i := 1; i <= 4; i++ {
		largeSummary := fmt.Sprintf("Event %d: ", i)
		for j := 0; j < 300000; j++ {
			largeSummary += "Long summary text for event. "
		}
		historyYAML += fmt.Sprintf(`- timestamp: "2025-12-13T10:00:%02dZ"
  step: "step-%d"
  status: "SUCCESS"
  summary: "%s"
`, i, i, largeSummary)
	}

	if err := os.WriteFile(historyPath, []byte(historyYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Verify file size exceeds 10MB
	fileInfo, err := os.Stat(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Size() <= 10*1024*1024 {
		t.Skipf("Test file size is %d bytes, expected > 10MB", fileInfo.Size())
	}

	// ReadHistory should perform eviction
	history, err := storage.ReadHistory(sessionID, "")
	if err != nil {
		t.Fatalf("ReadHistory failed: %v", err)
	}

	// With 4 events and 30% eviction: ceil(4 * 0.3) = 2 events removed
	// Should have 2 events remaining
	if len(history) != 2 {
		t.Errorf("Expected 2 entries after eviction, got %d", len(history))
	}

	// Should have removed oldest 2 entries (step-1 and step-2)
	if len(history) > 0 {
		if history[0].Step == "step-1" || history[0].Step == "step-2" {
			t.Error("Expected oldest entries (step-1, step-2) to be removed")
		}
	}
}

func TestReadHistory_FIFOWriteFailure(t *testing.T) {
	t.Skip("Write failure tests require special permissions setup")
}

func TestValidateHistory_EmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	// Empty YAML array
	emptyArray := "[]"
	if err := os.WriteFile(historyPath, []byte(emptyArray), 0o644); err != nil {
		t.Fatal(err)
	}

	// Empty array should pass validation
	err := storage.ValidateHistory(historyPath)
	if err != nil {
		t.Errorf("Expected no error for empty history array, got: %v", err)
	}
}

func TestValidateReport_EmptyOptionalFields(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.yaml")

	// Report without optional fields (summary, next_steps)
	minimalReport := `command: "test-command"
artifact: "test-artifact"
timestamp: "2025-12-13T10:00:00Z"
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(minimalReport), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateReport(reportPath)
	if err != nil {
		t.Errorf("Expected no error for minimal valid report, got: %v", err)
	}
}

func TestValidateHistory_OptionalDetailsField(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "history.yaml")

	// History event without optional details field
	minimalEvent := `- timestamp: "2025-12-13T10:00:00Z"
  step: "implement"
  status: "SUCCESS"
  summary: "Test"
`
	if err := os.WriteFile(historyPath, []byte(minimalEvent), 0o644); err != nil {
		t.Fatal(err)
	}

	err := storage.ValidateHistory(historyPath)
	if err != nil {
		t.Errorf("Expected no error for minimal valid history, got: %v", err)
	}
}
