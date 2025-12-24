//nolint:paralleltest // E2E tests with subprocess execution
package tests

import (
	"fluxid-loop/internal/ipc"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestIPCWriteHistoryConcurrentWrites tests concurrent writes to verify no corruption.
//
//nolint:cyclop // E2E test with concurrent operations and race condition checks
func TestIPCWriteHistoryConcurrentWrites(t *testing.T) {
	sessionID := "test-session-ipc-concurrent-writes"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Number of concurrent writes
	numWrites := 10
	var waitGroup sync.WaitGroup

	// Spawn multiple parallel writes
	for writeIndex := 0; writeIndex < numWrites; writeIndex++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()

			message := fmt.Sprintf("Concurrent write %d", index)
			cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin, "ipc", "write-history", message)
			cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Concurrent write %d failed: %v\nOutput: %s", index, err, output)
			}
		}(writeIndex)
	}

	// Wait for all writes to complete
	waitGroup.Wait()

	// Read history and verify completeness
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}

	// Verify all messages are present
	for i := 0; i < numWrites; i++ {
		expectedMessage := fmt.Sprintf("Concurrent write %d", i)
		if !strings.Contains(history, expectedMessage) {
			t.Errorf("History missing concurrent write %d, got: %s", i, history)
		}
	}

	// Verify we have exactly numWrites lines
	lines := strings.Split(strings.TrimSpace(history), "\n")
	if len(lines) != numWrites {
		t.Errorf("Expected %d history lines, got %d\nHistory: %s", numWrites, len(lines), history)
	}

	// Verify no partial writes or interleaving corruption
	for _, line := range lines {
		// Each line should start with [ and contain ]
		if !strings.HasPrefix(line, "[") {
			t.Errorf("Line doesn't start with '[': %s", line)
		}
		if !strings.Contains(line, "]") {
			t.Errorf("Line doesn't contain ']': %s", line)
		}
		// Verify the line contains "Concurrent write" followed by a number
		if !strings.Contains(line, "Concurrent write") {
			t.Errorf("Line doesn't contain expected message: %s", line)
		}
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestIPCWriteHistorySessionIsolation tests that different sessions have isolated history.
func TestIPCWriteHistorySessionIsolation(t *testing.T) {
	sessionID1 := "test-session-ipc-isolation-1"
	sessionID2 := "test-session-ipc-isolation-2"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write to first session
	message1 := "Message for session 1"
	cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin, "ipc", "write-history", message1)
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID1)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history failed: %v\nOutput: %s", err, output)
	}

	// Write to second session
	message2 := "Message for session 2"
	cmd = exec.CommandContext(testCtx(30*time.Second), fluxidBin, "ipc", "write-history", message2)
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID2)

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history failed: %v\nOutput: %s", err, output)
	}

	// Verify first session only has its message
	history1, err := ipc.ReadHistory(sessionID1)
	if err != nil {
		t.Fatalf("Failed to read history for session 1: %v", err)
	}
	if !strings.Contains(history1, message1) {
		t.Errorf("Session 1 history missing message, got: %s", history1)
	}
	if strings.Contains(history1, message2) {
		t.Errorf("Session 1 history leaked message from session 2, got: %s", history1)
	}

	// Verify second session only has its message
	history2, err := ipc.ReadHistory(sessionID2)
	if err != nil {
		t.Fatalf("Failed to read history for session 2: %v", err)
	}
	if !strings.Contains(history2, message2) {
		t.Errorf("Session 2 history missing message, got: %s", history2)
	}
	if strings.Contains(history2, message1) {
		t.Errorf("Session 2 history leaked message from session 1, got: %s", history2)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID1)
	_ = ipc.ClearHistory(sessionID2)
}

// TestIPCWriteHistoryZeroExitCode tests that successful history writes return exit code 0.
func TestIPCWriteHistoryZeroExitCode(t *testing.T) {
	sessionID := "test-session-ipc-exit-code"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entry
	cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin, "ipc", "write-history", "test message")
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	err := cmd.Run()
	if err != nil {
		t.Errorf("Expected exit code 0, got error: %v", err)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestIPCWriteHistoryUTF8Support tests UTF-8 messages without truncation.
func TestIPCWriteHistoryUTF8Support(t *testing.T) {
	sessionID := "test-session-ipc-utf8"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Test various UTF-8 messages
	messages := []string{
		"Simple ASCII message",
		"Japanese: 日本語でメッセージ",
		"Emoji: 🚀 🎉 ✨",
		"Chinese: 中文信息",
		"Arabic: رسالة عربية",
		"Mixed: Hello 世界 🌍",
	}

	for _, message := range messages {
		cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin, "ipc", "write-history", message)
		cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("ipc write-history failed for UTF-8 message '%s': %v\nOutput: %s", message, err, output)
		}
	}

	// Read history
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}

	// Verify all messages are present without truncation
	for _, message := range messages {
		if !strings.Contains(history, message) {
			t.Errorf("History missing UTF-8 message '%s', got: %s", message, history)
		}
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}
