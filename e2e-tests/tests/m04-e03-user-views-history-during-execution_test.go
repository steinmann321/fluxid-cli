//nolint:paralleltest // E2E tests with subprocess execution
package tests

import (
	"context"
	"fluxid-loop/internal/ipc"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// TestViewHistoryBasic tests the basic flow:
// set FLUXID_SESSION_ID → write two entries → run fluxid ipc view-history → verify output.
//
//nolint:cyclop // E2E test with history setup and output validation
func TestViewHistoryBasic(t *testing.T) {
	sessionID := "test-session-view-history-basic"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write two history entries using different methods
	// Entry 1: via --write-history flag
	message1 := "First decision: use FIFO eviction"
	cmd1 := exec.CommandContext(context.Background(), fluxidBin, "--write-history", message1)
	cmd1.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	output1, err := cmd1.CombinedOutput()
	if err != nil {
		t.Fatalf("--write-history failed: %v\nOutput: %s", err, output1)
	}

	// Entry 2: via ipc write-history
	message2 := "Second decision: implement buffer cap"
	cmd2 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", message2)
	cmd2.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	output2, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history failed: %v\nOutput: %s", err, output2)
	}

	// View history
	viewCmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
	viewCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	viewOutput, err := viewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc view-history failed: %v\nOutput: %s", err, viewOutput)
	}

	outputStr := string(viewOutput)

	// Verify both messages are present
	if !strings.Contains(outputStr, message1) {
		t.Errorf("History missing first message, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, message2) {
		t.Errorf("History missing second message, got: %s", outputStr)
	}

	// Verify we have exactly 2 lines
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 history lines, got %d: %v", len(lines), lines)
	}

	// Verify ISO 8601 timestamp format for each line: [YYYY-MM-DDTHH:MM:SSZ] message
	timestampRegex := regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z\] .+$`)
	for i, line := range lines {
		if !timestampRegex.MatchString(line) {
			t.Errorf("Line %d does not match ISO 8601 format [YYYY-MM-DDTHH:MM:SSZ] message: %s", i+1, line)
		}
	}

	// Verify chronological order (first entry comes first)
	if !strings.Contains(lines[0], message1) {
		t.Errorf("Expected first line to contain '%s', got: %s", message1, lines[0])
	}
	if !strings.Contains(lines[1], message2) {
		t.Errorf("Expected second line to contain '%s', got: %s", message2, lines[1])
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestViewHistoryWithSessionFlag tests using the --session flag instead of environment variable.
func TestViewHistoryWithSessionFlag(t *testing.T) {
	sessionID := "test-session-view-history-flag"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entry
	message := "Decision recorded via flag"
	writeCmd := exec.CommandContext(
		context.Background(),
		fluxidBin, "ipc", "write-history", message, "--session", sessionID,
	)
	// Filter out FLUXID_SESSION_ID from environment to verify --session flag works
	var filteredEnv []string
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "FLUXID_SESSION_ID=") {
			filteredEnv = append(filteredEnv, env)
		}
	}
	writeCmd.Env = filteredEnv

	output, err := writeCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history with --session failed: %v\nOutput: %s", err, output)
	}

	// View history with --session flag
	viewCmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history", "--session", sessionID)
	viewCmd.Env = filteredEnv

	viewOutput, err := viewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc view-history with --session failed: %v\nOutput: %s", err, viewOutput)
	}

	outputStr := string(viewOutput)

	// Verify message is present
	if !strings.Contains(outputStr, message) {
		t.Errorf("History missing message, got: %s", outputStr)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestViewHistoryEmpty tests viewing history when no entries exist.
func TestViewHistoryEmpty(t *testing.T) {
	sessionID := "test-session-view-history-empty"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// View history (no entries written)
	viewCmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
	viewCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	viewOutput, err := viewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc view-history failed: %v\nOutput: %s", err, viewOutput)
	}

	outputStr := string(viewOutput)

	// Verify output is empty (or just whitespace)
	if strings.TrimSpace(outputStr) != "" {
		t.Errorf("Expected empty history output, got: %s", outputStr)
	}
}

// TestViewHistoryWithoutSessionID tests error handling when session ID is not provided.
func TestViewHistoryWithoutSessionID(t *testing.T) {
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Attempt to view history without session ID
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
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

// TestViewHistoryHelp tests the --help flag for ipc view-history.
func TestViewHistoryHelp(t *testing.T) {
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Request help
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history", "--help")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc view-history --help failed: %v\nOutput: %s", err, output)
	}

	// Verify help text
	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage: fluxid ipc view-history") {
		t.Errorf("Expected usage in help text, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "ISO 8601") {
		t.Errorf("Expected ISO 8601 mention in help text, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "--session") {
		t.Errorf("Expected --session flag mention in help text, got: %s", outputStr)
	}
}

// TestViewHistoryMultipleInvocations tests that viewing history multiple times returns consistent results.
func TestViewHistoryMultipleInvocations(t *testing.T) {
	sessionID := "test-session-view-history-multiple-invocations"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entries
	messages := []string{
		"Entry 1",
		"Entry 2",
		"Entry 3",
	}

	for _, message := range messages {
		cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", message)
		cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("ipc write-history failed for message '%s': %v\nOutput: %s", message, err, output)
		}
	}

	// View history multiple times
	for viewIndex := 0; viewIndex < 3; viewIndex++ {
		viewCmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
		viewCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

		viewOutput, err := viewCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("ipc view-history invocation %d failed: %v\nOutput: %s", viewIndex+1, err, viewOutput)
		}

		outputStr := string(viewOutput)

		// Verify all messages are present
		for _, message := range messages {
			if !strings.Contains(outputStr, message) {
				t.Errorf("Invocation %d missing message '%s', got: %s", viewIndex+1, message, outputStr)
			}
		}

		// Verify we have exactly 3 lines
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		if len(lines) != len(messages) {
			t.Errorf("Invocation %d expected %d history lines, got %d", viewIndex+1, len(messages), len(lines))
		}
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestViewHistoryWithUTF8Characters tests handling of UTF-8 characters in history.
func TestViewHistoryWithUTF8Characters(t *testing.T) {
	sessionID := "test-session-view-history-utf8"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write history entry with UTF-8 characters
	message := "Decision: 日本語のテスト with émojis 🎉"
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", message)
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history failed: %v\nOutput: %s", err, output)
	}

	// View history
	viewCmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
	viewCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	viewOutput, err := viewCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc view-history failed: %v\nOutput: %s", err, viewOutput)
	}

	outputStr := string(viewOutput)

	// Verify full message with UTF-8 characters is preserved
	if !strings.Contains(outputStr, message) {
		t.Errorf("History missing UTF-8 message, got: %s", outputStr)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestViewHistoryZeroExitCode tests that successful history viewing returns exit code 0.
func TestViewHistoryZeroExitCode(t *testing.T) {
	sessionID := "test-session-view-history-exit-code"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write a history entry
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", "test message")
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ipc write-history failed: %v", err)
	}

	// View history
	viewCmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
	viewCmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	err = viewCmd.Run()
	if err != nil {
		t.Errorf("Expected exit code 0, got error: %v", err)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}
