//nolint:paralleltest // E2E tests with subprocess execution
package tests

import (
	"context"
	"fluxid-loop/internal/ipc"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestIPCWriteHistoryBasic tests the basic flow:
// set FLUXID_SESSION_ID → run fluxid ipc write-history "message" → verify confirmation and entry.
//
//nolint:cyclop,funlen // E2E test with IPC command execution and history validation
func TestIPCWriteHistoryBasic(t *testing.T) {
	sessionID := "test-session-ipc-write-history-basic"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entry via IPC
	message := "Decision: adopt FIFO eviction"
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", message)
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history failed: %v\nOutput: %s", err, output)
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

// TestIPCWriteHistoryWithSessionFlag tests using the --session flag instead of environment variable.
func TestIPCWriteHistoryWithSessionFlag(t *testing.T) {
	sessionID := "test-session-ipc-write-history-flag"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entry via IPC with --session flag
	message := "Testing session flag"
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", message, "--session", sessionID)
	// Filter out FLUXID_SESSION_ID from environment to verify --session flag works
	var filteredEnv []string
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "FLUXID_SESSION_ID=") {
			filteredEnv = append(filteredEnv, env)
		}
	}
	cmd.Env = filteredEnv

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history with --session failed: %v\nOutput: %s", err, output)
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

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestIPCWriteHistoryMultipleEntries tests appending multiple history entries.
func TestIPCWriteHistoryMultipleEntries(t *testing.T) {
	sessionID := "test-session-ipc-write-history-multiple"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	messages := []string{
		"First decision",
		"Second decision",
		"Third decision with UTF-8: 日本語",
	}

	// Write multiple entries
	for _, message := range messages {
		cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", message)
		cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("ipc write-history failed for message '%s': %v\nOutput: %s", message, err, output)
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

// TestIPCWriteHistoryWithoutSessionID tests error handling when session ID is not provided.
func TestIPCWriteHistoryWithoutSessionID(t *testing.T) {
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Attempt to write without session ID
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", "test message")
	// Explicitly clear FLUXID_SESSION_ID from environment
	cmd.Env = []string{}

	output, err := cmd.CombinedOutput()

	// Expect non-zero exit code
	if err == nil {
		t.Error("Expected error when session ID not provided, got nil")
	}

	// Verify error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "session ID not provided") {
		t.Errorf("Expected clear error message about missing session ID, got: %s", outputStr)
	}
}

// TestIPCWriteHistoryWithoutMessage tests error handling when message is missing.
func TestIPCWriteHistoryWithoutMessage(t *testing.T) {
	sessionID := "test-session-ipc-write-history-no-message"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Attempt to write without message
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history")
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	output, err := cmd.CombinedOutput()

	// Expect non-zero exit code
	if err == nil {
		t.Error("Expected error when message is missing, got nil")
	}

	// Verify error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "write-history requires a message argument") {
		t.Errorf("Expected clear error message about missing message, got: %s", outputStr)
	}
}

// TestIPCWriteHistoryMultiWordMessage tests handling of messages with spaces.
func TestIPCWriteHistoryMultiWordMessage(t *testing.T) {
	sessionID := "test-session-ipc-write-history-multi-word"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entry with multiple words
	message := "This is a longer decision with multiple words"
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", message)
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history failed: %v\nOutput: %s", err, output)
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

// TestIPCWriteHistoryHelp tests the --help flag for ipc write-history.
func TestIPCWriteHistoryHelp(t *testing.T) {
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Request help
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", "--help")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history --help failed: %v\nOutput: %s", err, output)
	}

	// Verify help text
	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage: fluxid ipc write-history <message>") {
		t.Errorf("Expected usage in help text, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "ISO 8601") {
		t.Errorf("Expected ISO 8601 mention in help text, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "--session") {
		t.Errorf("Expected --session flag mention in help text, got: %s", outputStr)
	}
}
