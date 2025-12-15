//nolint:paralleltest // E2E tests with subprocess execution
package tests

import (
	"context"
	"fluxid-loop/internal/ipc"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestWriteHistoryWithSessionID tests the basic flow:
// set FLUXID_SESSION_ID → run fluxid --write-history "message" → verify confirmation and entry.
func TestWriteHistoryWithSessionID(t *testing.T) {
	sessionID := "test-session-write-history-basic"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entry
	message := "First note"
	cmd := exec.CommandContext(context.Background(), fluxidBin, "--write-history", message)
	cmd.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--write-history failed: %v\nOutput: %s", err, output)
	}

	// Verify confirmation message
	outputStr := string(output)
	if !strings.Contains(outputStr, "History entry recorded for session") {
		t.Errorf("Expected confirmation message, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, sessionID) {
		t.Errorf("Expected session ID in output, got: %s", outputStr)
	}

	// Verify entry was written to history
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}

	if !strings.Contains(history, message) {
		t.Errorf("History missing message, got: %s", history)
	}

	// Verify timestamp format (ISO 8601 with Z suffix)
	if !strings.Contains(history, "[20") || !strings.Contains(history, "Z]") {
		t.Errorf("History missing ISO 8601 timestamp, got: %s", history)
	}

	// Verify exact format: [YYYY-MM-DDTHH:MM:SSZ] message
	lines := strings.Split(strings.TrimSpace(history), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 history line, got %d", len(lines))
	}

	line := lines[0]
	if !strings.HasPrefix(line, "[") {
		t.Errorf("Expected line to start with '[', got: %s", line)
	}

	// Extract timestamp
	closeBracket := strings.Index(line, "]")
	if closeBracket == -1 {
		t.Errorf("Expected closing bracket in timestamp, got: %s", line)
	} else {
		timestamp := line[1:closeBracket]
		// Verify ISO 8601 format: YYYY-MM-DDTHH:MM:SSZ
		if len(timestamp) != 20 {
			t.Errorf("Expected timestamp length 20, got %d: %s", len(timestamp), timestamp)
		}
		if !strings.HasSuffix(timestamp, "Z") {
			t.Errorf("Expected timestamp to end with 'Z', got: %s", timestamp)
		}
		if !strings.Contains(timestamp, "T") {
			t.Errorf("Expected timestamp to contain 'T', got: %s", timestamp)
		}
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestWriteHistoryMultipleEntries tests appending multiple history entries.
func TestWriteHistoryMultipleEntries(t *testing.T) {
	sessionID := "test-session-write-history-multiple"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	messages := []string{
		"First note",
		"Second note",
		"Third note with UTF-8 characters: 日本語",
	}

	// Write multiple entries
	for _, message := range messages {
		cmd := exec.CommandContext(context.Background(), fluxidBin, "--write-history", message)
		cmd.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("--write-history failed for message '%s': %v\nOutput: %s", message, err, output)
		}
	}

	// Read history
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}

	// Verify all messages are present
	for _, message := range messages {
		if !strings.Contains(history, message) {
			t.Errorf("History missing message '%s', got: %s", message, history)
		}
	}

	// Verify we have exactly 3 lines
	lines := strings.Split(strings.TrimSpace(history), "\n")
	if len(lines) != len(messages) {
		t.Errorf("Expected %d history lines, got %d", len(messages), len(lines))
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestWriteHistoryWithoutSessionID tests error handling when FLUXID_SESSION_ID is not set.
func TestWriteHistoryWithoutSessionID(t *testing.T) {
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Attempt to write without session ID
	cmd := exec.CommandContext(context.Background(), fluxidBin, "--write-history", "test message")
	// Explicitly clear FLUXID_SESSION_ID from environment
	cmd.Env = []string{}

	output, err := cmd.CombinedOutput()

	// Expect non-zero exit code
	if err == nil {
		t.Error("Expected error when FLUXID_SESSION_ID not set, got nil")
	}

	// Verify error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "FLUXID_SESSION_ID environment variable not set") {
		t.Errorf("Expected clear error message about missing session ID, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "active session context") {
		t.Errorf("Expected context explanation in error, got: %s", outputStr)
	}
}

// TestWriteHistoryWithoutMessage tests error handling when message is missing.
func TestWriteHistoryWithoutMessage(t *testing.T) {
	sessionID := "test-session-write-history-no-message"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Attempt to write without message
	cmd := exec.CommandContext(context.Background(), fluxidBin, "--write-history")
	cmd.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))

	output, err := cmd.CombinedOutput()

	// Expect non-zero exit code
	if err == nil {
		t.Error("Expected error when message is missing, got nil")
	}

	// Verify error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "--write-history requires a message argument") {
		t.Errorf("Expected clear error message about missing message, got: %s", outputStr)
	}
}

// TestWriteHistoryMultiWordMessage tests handling of messages with spaces.
func TestWriteHistoryMultiWordMessage(t *testing.T) {
	sessionID := "test-session-write-history-multi-word"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entry with multiple words
	message := "This is a longer message with multiple words"
	cmd := exec.CommandContext(context.Background(), fluxidBin, "--write-history", message)
	cmd.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--write-history failed: %v\nOutput: %s", err, output)
	}

	// Verify full message was written
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}

	if !strings.Contains(history, message) {
		t.Errorf("History missing full message, got: %s", history)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestWriteHistoryHelp tests the --help flag for --write-history.
func TestWriteHistoryHelp(t *testing.T) {
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Request help
	cmd := exec.CommandContext(context.Background(), fluxidBin, "--write-history", "--help")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--write-history --help failed: %v\nOutput: %s", err, output)
	}

	// Verify help text
	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage: fluxid --write-history <message>") {
		t.Errorf("Expected usage in help text, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "ISO 8601") {
		t.Errorf("Expected ISO 8601 mention in help text, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "FLUXID_SESSION_ID") {
		t.Errorf("Expected FLUXID_SESSION_ID mention in help text, got: %s", outputStr)
	}
}

// TestWriteHistoryNoFilePersistence tests that new sessions start with empty history.
func TestWriteHistoryNoFilePersistence(t *testing.T) {
	sessionID1 := "test-session-no-persist-1"
	sessionID2 := "test-session-no-persist-2"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write to first session
	message1 := "Message for session 1"
	cmd := exec.CommandContext(context.Background(), fluxidBin, "--write-history", message1)
	cmd.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID1))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--write-history failed: %v\nOutput: %s", err, output)
	}

	// Verify first session has entry
	history1, err := ipc.ReadHistory(sessionID1)
	if err != nil {
		t.Fatalf("Failed to read history for session 1: %v", err)
	}
	if !strings.Contains(history1, message1) {
		t.Errorf("Session 1 history missing message, got: %s", history1)
	}

	// Verify second session is empty
	history2, err := ipc.ReadHistory(sessionID2)
	if err != nil {
		t.Fatalf("Failed to read history for session 2: %v", err)
	}
	if history2 != "" {
		t.Errorf("Expected empty history for session 2, got: %s", history2)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID1)
}

// TestWriteHistoryZeroExitCode tests that successful history writes return exit code 0.
func TestWriteHistoryZeroExitCode(t *testing.T) {
	sessionID := "test-session-exit-code"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entry
	cmd := exec.CommandContext(context.Background(), fluxidBin, "--write-history", "test message")
	cmd.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))

	err := cmd.Run()
	if err != nil {
		t.Errorf("Expected exit code 0, got error: %v", err)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}
