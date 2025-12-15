//nolint:paralleltest // E2E tests with subprocess execution
package tests

import (
	"context"
	"fluxid-loop/internal/ipc"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// TestHistoryErrorMissingSessionID tests that missing FLUXID_SESSION_ID yields clear error.
// Success criteria: Missing or invalid FLUXID_SESSION_ID yields clear error
// [Test: unset env; verify message and non-zero exit].
func TestHistoryErrorMissingSessionID(t *testing.T) {
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Test 1: ipc view-history without FLUXID_SESSION_ID
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
	// Explicitly set env WITHOUT FLUXID_SESSION_ID
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Verify non-zero exit code
	if err == nil {
		t.Errorf("Expected non-zero exit code for missing session ID, got success")
	}

	// Verify error message is clear and actionable
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected 'Error:' prefix in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "session ID") {
		t.Errorf("Expected 'session ID' in error message, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "FLUXID_SESSION_ID") {
		t.Errorf("Expected 'FLUXID_SESSION_ID' in error message for remediation, got: %s", outputStr)
	}

	// Test 2: ipc write-history without FLUXID_SESSION_ID
	cmd2 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", "test message")
	cmd2.Env = []string{"PATH=" + os.Getenv("PATH")}

	output2, err2 := cmd2.CombinedOutput()
	outputStr2 := string(output2)

	// Verify non-zero exit code
	if err2 == nil {
		t.Errorf("Expected non-zero exit code for missing session ID, got success")
	}

	// Verify error message is clear
	if !strings.Contains(outputStr2, "Error:") {
		t.Errorf("Expected 'Error:' prefix in output, got: %s", outputStr2)
	}
	if !strings.Contains(outputStr2, "session ID") {
		t.Errorf("Expected 'session ID' in error message, got: %s", outputStr2)
	}

	// Test 3: --write-history without FLUXID_SESSION_ID
	cmd3 := exec.CommandContext(context.Background(), fluxidBin, "--write-history", "test message")
	cmd3.Env = []string{"PATH=" + os.Getenv("PATH")}

	output3, err3 := cmd3.CombinedOutput()
	outputStr3 := string(output3)

	// Verify non-zero exit code
	if err3 == nil {
		t.Errorf("Expected non-zero exit code for missing session ID, got success")
	}

	// Verify error message is clear
	if !strings.Contains(outputStr3, "Error:") {
		t.Errorf("Expected 'Error:' prefix in output, got: %s", outputStr3)
	}
	if !strings.Contains(outputStr3, "FLUXID_SESSION_ID") {
		t.Errorf("Expected 'FLUXID_SESSION_ID' in error message, got: %s", outputStr3)
	}
}

// TestHistoryErrorEmptyMessage tests that empty message is rejected with validation feedback.
// Success criteria: Empty message rejected with validation feedback
// [Test: `ipc write-history ""` returns validation error].
func TestHistoryErrorEmptyMessage(t *testing.T) {
	sessionID := "test-session-empty-message"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Test 1: ipc write-history with no message argument
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history")
	cmd.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Verify non-zero exit code
	if err == nil {
		t.Errorf("Expected non-zero exit code for missing message, got success")
	}

	// Verify error message mentions the message requirement
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected 'Error:' prefix in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "message") {
		t.Errorf("Expected 'message' in error message, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Usage:") || !strings.Contains(outputStr, "write-history") {
		t.Errorf("Expected usage hint in error message, got: %s", outputStr)
	}

	// Test 2: --write-history with no message argument
	cmd2 := exec.CommandContext(context.Background(), fluxidBin, "--write-history")
	cmd2.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))

	output2, err2 := cmd2.CombinedOutput()
	outputStr2 := string(output2)

	// Verify non-zero exit code
	if err2 == nil {
		t.Errorf("Expected non-zero exit code for missing message, got success")
	}

	// Verify error message
	if !strings.Contains(outputStr2, "Error:") {
		t.Errorf("Expected 'Error:' prefix in output, got: %s", outputStr2)
	}
	if !strings.Contains(outputStr2, "message") {
		t.Errorf("Expected 'message' in error message, got: %s", outputStr2)
	}
}

// TestHistoryErrorConsistentFormat tests that errors are plain text and consistent in format.
// Success criteria: Errors are plain text and consistent in format [Test: regex match across errors].
func TestHistoryErrorConsistentFormat(t *testing.T) {
	sessionID := "test-session-error-format"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Collect various error messages
	errorMessages := []string{}

	// Error 1: Missing session ID
	cmd1 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
	cmd1.Env = []string{"PATH=" + os.Getenv("PATH")}
	output1, _ := cmd1.CombinedOutput()
	errorMessages = append(errorMessages, string(output1))

	// Error 2: Missing message
	cmd2 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history")
	cmd2.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))
	output2, _ := cmd2.CombinedOutput()
	errorMessages = append(errorMessages, string(output2))

	// Error 3: Invalid session flag value
	cmd3 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history", "--session")
	cmd3.Env = []string{"PATH=" + os.Getenv("PATH")}
	output3, _ := cmd3.CombinedOutput()
	errorMessages = append(errorMessages, string(output3))

	// Verify all errors follow consistent format
	// All should start with "Error:" prefix
	errorPattern := regexp.MustCompile(`^Error:`)

	for i, msg := range errorMessages {
		if !errorPattern.MatchString(msg) {
			t.Errorf("Error message %d does not start with 'Error:', got: %s", i+1, msg)
		}

		// Verify errors are plain text (no HTML, JSON, etc.)
		if strings.Contains(msg, "<html>") || strings.Contains(msg, "\"error\":") || strings.Contains(msg, "{\"") {
			t.Errorf("Error message %d appears to contain structured data, expected plain text: %s", i+1, msg)
		}

		// Verify errors are concise (not overly verbose)
		lines := strings.Split(strings.TrimSpace(msg), "\n")
		if len(lines) > 5 {
			t.Errorf("Error message %d is too verbose (%d lines), expected concise error", i+1, len(lines))
		}
	}
}

// TestHistoryErrorRecovery tests that after fixing issue, commands succeed.
// Success criteria: Recovery works: after fixing issue, commands succeed [Test: set session and retry; success].
func TestHistoryErrorRecovery(t *testing.T) {
	sessionID := "test-session-error-recovery"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Step 1: Fail with missing session ID
	cmd1 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
	cmd1.Env = []string{"PATH=" + os.Getenv("PATH")}
	_, err1 := cmd1.CombinedOutput()
	if err1 == nil {
		t.Errorf("Expected error for missing session ID, got success")
	}

	// Step 2: Fix issue by setting FLUXID_SESSION_ID and retry
	cmd2 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history")
	cmd2.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))
	output2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		t.Fatalf("Expected success after setting session ID, got error: %v\nOutput: %s", err2, output2)
	}

	// Step 3: Fail with missing message
	cmd3 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history")
	cmd3.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))
	_, err3 := cmd3.CombinedOutput()
	if err3 == nil {
		t.Errorf("Expected error for missing message, got success")
	}

	// Step 4: Fix issue by providing message and retry
	message := "Test recovery message"
	cmd4 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "write-history", message)
	cmd4.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID))
	output4, err4 := cmd4.CombinedOutput()
	if err4 != nil {
		t.Fatalf("Expected success after providing message, got error: %v\nOutput: %s", err4, output4)
	}

	// Verify the message was actually written
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		t.Fatalf("Failed to read history: %v", err)
	}
	if !strings.Contains(history, message) {
		t.Errorf("History missing message after recovery, got: %s", history)
	}

	// Clean up
	_ = ipc.ClearHistory(sessionID)
}

// TestHistoryErrorDoesNotAbortWrapper tests that IPC command failures never abort wrapper.
// Success criteria: IPC command failures never abort wrapper
// [Test: induce failures while wrapper runs; wrapper continues].
func TestHistoryErrorDoesNotAbortWrapper(t *testing.T) {
	// This test verifies that history command errors return appropriate exit codes
	// but don't cause catastrophic failures that would abort a wrapper process.
	// We test this by verifying that errors are cleanly returned and the binary
	// exits gracefully with proper exit codes.

	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Test multiple error scenarios in sequence
	testCases := []struct {
		name string
		args []string
		env  []string
	}{
		{
			name: "missing session on view-history",
			args: []string{"ipc", "view-history"},
			env:  []string{"PATH=" + os.Getenv("PATH")},
		},
		{
			name: "missing session on write-history",
			args: []string{"ipc", "write-history", "test"},
			env:  []string{"PATH=" + os.Getenv("PATH")},
		},
		{
			name: "missing message on write-history",
			args: []string{"ipc", "write-history"},
			env:  append(os.Environ(), "FLUXID_SESSION_ID=test-wrapper"),
		},
		{
			name: "invalid session flag",
			args: []string{"ipc", "view-history", "--session"},
			env:  []string{"PATH=" + os.Getenv("PATH")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.CommandContext(context.Background(), fluxidBin, tc.args...)
			cmd.Env = tc.env

			output, err := cmd.CombinedOutput()

			// Verify command exits with error code (not crash/panic)
			if err == nil {
				t.Errorf("Expected non-zero exit code, got success")
			}

			// Verify error output exists and is readable
			outputStr := string(output)
			if len(outputStr) == 0 {
				t.Errorf("Expected error message, got empty output")
			}

			// Verify error is written to stderr (CombinedOutput captures both)
			if !strings.Contains(outputStr, "Error:") {
				t.Errorf("Expected error message to contain 'Error:', got: %s", outputStr)
			}

			// Verify no stack traces or panics (would indicate abort)
			if strings.Contains(outputStr, "panic:") || strings.Contains(outputStr, "goroutine") {
				t.Errorf("Command appears to have panicked instead of returning error: %s", outputStr)
			}
		})
	}
}

// TestHistoryErrorSessionFlagOverride tests --session flag with various error scenarios.
func TestHistoryErrorSessionFlagOverride(t *testing.T) {
	sessionID := "test-session-flag-override"
	setupReportDir(t)

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Test 1: --session flag with no value (should error)
	cmd := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history", "--session")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err == nil {
		t.Errorf("Expected error for --session with no value, got success")
	}

	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected 'Error:' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "--session") || !strings.Contains(outputStr, "value") {
		t.Errorf("Expected error about missing --session value, got: %s", outputStr)
	}

	// Test 2: --session flag with valid value (should succeed)
	cmd2 := exec.CommandContext(context.Background(), fluxidBin, "ipc", "view-history", "--session", sessionID)
	cmd2.Env = []string{"PATH=" + os.Getenv("PATH")} // No FLUXID_SESSION_ID

	output2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		t.Errorf("Expected success with --session flag, got error: %v\nOutput: %s", err2, output2)
	}
}
