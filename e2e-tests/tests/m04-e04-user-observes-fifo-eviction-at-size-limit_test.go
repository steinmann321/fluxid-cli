//nolint:paralleltest // E2E tests with subprocess execution
package tests

import (
	"fluxid-cli/internal/ipc"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestFIFOEvictionAtSizeLimit tests that history evicts oldest entries when exceeding 32MB.
// This test writes 34MB of data (34 entries × ~1MB each) and verifies FIFO eviction.
//
//nolint:cyclop,funlen // E2E test with large data writes and FIFO eviction validation
func TestFIFOEvictionAtSizeLimit(t *testing.T) {
	sessionID := "test-session-fifo-eviction-size-limit"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Clean up any existing history
	_ = ipc.ClearHistory(sessionID)

	// Create entries that total > 32MB
	// Each entry is ~1MB (1,000,000 bytes) + timestamp (~25 bytes) + newline (1 byte) = 1,000,026 bytes
	// 32MB = 33,554,432 bytes
	// 34 entries = 34,000,884 bytes > 32MB, should trigger eviction
	entrySize := 1_000_000
	numEntries := 34

	for entryIndex := 0; entryIndex < numEntries; entryIndex++ {
		// Generate large message with identifiable prefix
		prefix := fmt.Sprintf("ENTRY-%03d:", entryIndex)
		message := prefix + strings.Repeat("X", entrySize-len(prefix))

		// Write via IPC library directly (CLI args too long)
		err := ipc.WriteHistoryEntry(sessionID, message)
		if err != nil {
			t.Fatalf("Failed to write entry %d: %v", entryIndex, err)
		}
	}

	// View history
	viewCmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin, "ipc", "view-history")
	viewCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	viewOutput, err := viewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc view-history failed: %v\nOutput: %s", err, viewOutput)
	}

	outputStr := string(viewOutput)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")

	// Verify total size is under 32MB
	totalSize := len([]byte(outputStr))
	maxSize := 32 * 1024 * 1024
	if totalSize > maxSize {
		t.Errorf("History size %d exceeds 32MB limit %d", totalSize, maxSize)
	}

	// Verify oldest entry was evicted (only ENTRY-000 should be gone with 34 entries)
	// 34 entries = 34,000,884 bytes, exceeds 32MB by 446,452 bytes
	// This requires evicting only 1 entry (~1MB each)
	if strings.Contains(outputStr, "ENTRY-000:") {
		t.Error("Oldest entry ENTRY-000 should have been evicted")
	}

	// ENTRY-001 should still be present (only 1 entry evicted)
	if !strings.Contains(outputStr, "ENTRY-001:") {
		t.Error("ENTRY-001 should still be present after minimal eviction")
	}

	// Verify newest entries remain
	if !strings.Contains(outputStr, "ENTRY-031:") {
		t.Error("Recent entry ENTRY-031 should be present")
	}
	if !strings.Contains(outputStr, "ENTRY-032:") {
		t.Error("Recent entry ENTRY-032 should be present")
	}
	if !strings.Contains(outputStr, "ENTRY-033:") {
		t.Error("Newest entry ENTRY-033 should be present")
	}

	// Verify chronological order is maintained
	// Find first and last retained entries
	firstLine := lines[0]
	lastLine := lines[len(lines)-1]

	// Extract entry IDs
	if !strings.Contains(firstLine, "ENTRY-") || !strings.Contains(lastLine, "ENTRY-033:") {
		t.Errorf("Chronological order broken: first=%s, last=%s", firstLine, lastLine)
	}

	// Verify we have 33 entries remaining (34 - 1 evicted)
	if len(lines) != 33 {
		t.Errorf("Expected 33 entries after eviction (34 - 1), got %d", len(lines))
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestFIFOEvictionBoundary tests eviction at the exact boundary (just over 32MB).
func TestFIFOEvictionBoundary(t *testing.T) {
	sessionID := "test-session-fifo-eviction-boundary"
	setupReportDir(t)

	// Clean up any existing history
	_ = ipc.ClearHistory(sessionID)

	// Write entries that exceed 32MB
	// Each entry is ~1MB + timestamp (~25 bytes) + newline = 1,000,026 bytes
	// 32MB = 33,554,432 bytes, so 34 entries = 34,000,884 bytes (exceeds limit)
	// 33 entries = 33,000,858 bytes (UNDER limit - won't trigger eviction)
	// Use 35 entries to guarantee eviction
	entrySize := 1_000_000
	numEntries := 35 // Enough to trigger eviction

	for i := 0; i < numEntries; i++ {
		message := fmt.Sprintf("BOUNDARY-%03d:", i) + strings.Repeat("Y", entrySize-len(fmt.Sprintf("BOUNDARY-%03d:", i)))
		err := ipc.WriteHistoryEntry(sessionID, message)
		if err != nil {
			t.Fatalf("Failed to write entry %d: %v", i, err)
		}
	}

	// Read history
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}

	// Verify eviction occurred
	if strings.Contains(history, "BOUNDARY-000:") {
		t.Error("Oldest entry BOUNDARY-000 should have been evicted")
	}

	// Verify newest entry is present
	if !strings.Contains(history, fmt.Sprintf("BOUNDARY-%03d:", numEntries-1)) {
		t.Error("Newest entry should be present")
	}

	// Verify size is under limit
	historySize := len([]byte(history))
	maxSize := 32 * 1024 * 1024
	if historySize > maxSize {
		t.Errorf("History size %d exceeds 32MB limit %d", historySize, maxSize)
	}

	// Verify chronological order is preserved
	lines := strings.Split(strings.TrimSpace(history), "\n")
	if len(lines) < 2 {
		t.Fatal("Expected at least 2 lines after eviction")
	}

	// Check that retained entries are in order
	for lineIndex := 1; lineIndex < len(lines); lineIndex++ {
		prevLine := lines[lineIndex-1]
		currLine := lines[lineIndex]

		// Extract timestamps (assumes ISO 8601 format: [YYYY-MM-DDTHH:MM:SSZ])
		// Timestamps should be increasing
		if prevLine > currLine { // Lexicographic comparison works for ISO 8601
			t.Errorf("Chronological order violated: line %d (%s) > line %d (%s)",
				lineIndex-1, prevLine[:30], lineIndex, currLine[:30])
		}
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestFIFOEvictionMultibyteUTF8 tests that UTF-8 multibyte characters are handled correctly
// and size is measured in bytes, not characters.
func TestFIFOEvictionMultibyteUTF8(t *testing.T) {
	sessionID := "test-session-fifo-eviction-utf8"
	setupReportDir(t)

	// Clean up any existing history
	_ = ipc.ClearHistory(sessionID)

	// UTF-8 multibyte test: 日本語 (9 bytes) + 🚀 (4 bytes) = 13 bytes total
	// This tests that size counting is accurate for multibyte sequences
	utf8Chars := "日本語🚀"
	utf8ByteSize := len([]byte(utf8Chars)) // Should be 13 bytes

	if utf8ByteSize != 13 {
		t.Fatalf("UTF-8 test setup error: expected 13 bytes, got %d", utf8ByteSize)
	}

	// Create entries with multibyte characters to exceed 32MB
	entrySize := 1_000_000
	numEntries := 34

	for entryIndex := 0; entryIndex < numEntries; entryIndex++ {
		// Include UTF-8 chars in each entry
		prefix := fmt.Sprintf("UTF8-%03d-%s:", entryIndex, utf8Chars)
		padding := strings.Repeat("Z", entrySize-len([]byte(prefix)))
		message := prefix + padding

		err := ipc.WriteHistoryEntry(sessionID, message)
		if err != nil {
			t.Fatalf("Failed to write UTF-8 entry %d: %v", entryIndex, err)
		}
	}

	// Read history
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}

	// Verify size is measured in bytes
	historySize := len([]byte(history))
	maxSize := 32 * 1024 * 1024
	if historySize > maxSize {
		t.Errorf("History size %d exceeds 32MB limit %d", historySize, maxSize)
	}

	// Verify no split multibyte sequences (all UTF-8 chars should be intact)
	lines := strings.Split(strings.TrimSpace(history), "\n")
	for i, line := range lines {
		if !strings.Contains(line, utf8Chars) {
			t.Errorf("Line %d missing UTF-8 characters '%s': %s", i, utf8Chars, line[:50])
		}
	}

	// Verify lines are valid UTF-8
	for i, line := range lines {
		if !utf8Valid(line) {
			t.Errorf("Line %d contains invalid UTF-8", i)
		}
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestFIFOEvictionViaIPCCommand tests eviction using CLI commands (where feasible).
func TestFIFOEvictionViaIPCCommand(t *testing.T) {
	sessionID := "test-session-fifo-eviction-ipc-command"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Clean up any existing history
	_ = ipc.ClearHistory(sessionID)

	// Write large entries via IPC library (CLI args too long for 1MB messages)
	// Then verify view-history shows eviction results
	entrySize := 1_000_000
	numEntries := 34

	for i := 0; i < numEntries; i++ {
		message := fmt.Sprintf("CMD-%03d:", i) + strings.Repeat("W", entrySize-len(fmt.Sprintf("CMD-%03d:", i)))
		err := ipc.WriteHistoryEntry(sessionID, message)
		if err != nil {
			t.Fatalf("Failed to write entry %d via IPC: %v", i, err)
		}
	}

	// View history via CLI command
	viewCmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin, "ipc", "view-history")
	viewCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	viewOutput, err := viewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc view-history failed: %v\nOutput: %s", err, viewOutput)
	}

	outputStr := string(viewOutput)

	// Verify eviction occurred
	if strings.Contains(outputStr, "CMD-000:") {
		t.Error("Oldest entry CMD-000 should have been evicted")
	}

	// Verify newest entries remain
	if !strings.Contains(outputStr, "CMD-033:") {
		t.Error("Newest entry CMD-033 should be present")
	}

	// Verify size under limit
	totalSize := len([]byte(outputStr))
	maxSize := 32 * 1024 * 1024
	if totalSize > maxSize {
		t.Errorf("History size %d exceeds 32MB limit %d", totalSize, maxSize)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// utf8Valid checks if a string contains valid UTF-8.
func utf8Valid(s string) bool {
	// Go strings are always valid UTF-8 by construction,
	// but we verify no corruption occurred during file operations
	for _, r := range s {
		if r == '\uFFFD' { // Unicode replacement character indicates invalid UTF-8
			return false
		}
	}
	return true
}
