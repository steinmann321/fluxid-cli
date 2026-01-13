//nolint:paralleltest // Output tests with log capture
package command

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
)

func TestPrintWorkflowSuccess(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	sessionID := "test-success-session-12345"
	printWorkflowSuccess(sessionID)

	output := buf.String()

	// Verify all expected strings are in the output
	expectedStrings := []string{
		"=== Workflow Completion Summary ===",
		"Session ID: " + sessionID,
		"Status: SUCCESS",
		"All workflow loops completed.",
		"===================================",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but got:\n%s", expected, output)
		}
	}
}

func TestPrintWorkflowSuccessWithDifferentSessionID(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	sessionID := "different-session-id-xyz"
	printWorkflowSuccess(sessionID)

	output := buf.String()

	// Verify the specific session ID appears in the output
	if !strings.Contains(output, sessionID) {
		t.Errorf("Expected output to contain session ID %q, but got:\n%s", sessionID, output)
	}

	// Verify success status
	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Expected output to contain 'Status: SUCCESS', but got:\n%s", output)
	}
}

var errTestWorkflow = errors.New("test workflow error")

func TestPrintWorkflowError(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	reader, writer, _ := os.Pipe()
	os.Stderr = writer
	defer func() {
		os.Stderr = oldStderr
	}()

	sessionID := "test-error-session-12345"
	err := errTestWorkflow
	exitCode := 42

	printWorkflowError(err, exitCode, sessionID)

	_ = writer.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(reader)
	output := buf.String()

	// Verify all expected strings are in the output
	expectedStrings := []string{
		"=== Workflow Aborted ===",
		"Agent execution failed: test workflow error",
		"Exit code: 42",
		"Next steps:",
		"Session ID: " + sessionID,
		"========================",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but got:\n%s", expected, output)
		}
	}
}
