package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintTextToWriter_StartupBannerFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	//nolint:exhaustruct // Optional fields intentionally omitted for test simplicity
	status := InitializationStatus{
		Version:             "0.1.0+abc1234",
		SessionID:           "banner-test-session",
		Agent:               "claude",
		MaxReviewCycles:     20,
		MaxImplementRetries: 100,
		MaxCommitRetries:    50,
		TaskFile:            "/path/to/task.md",
	}

	PrintTextToWriter(&buf, status)
	output := buf.String()

	// Verify banner structure
	lines := strings.Split(output, "\n")

	// Check header line
	if !strings.Contains(lines[0], "=== fluxid Workflow Initialization ===") {
		t.Errorf("Expected banner header, got: %q", lines[0])
	}

	// Check version is on second line (index 1) and properly formatted
	if !strings.HasPrefix(lines[1], "Version: ") {
		t.Errorf("Expected 'Version:' on second line, got: %q", lines[1])
	}
	if !strings.Contains(lines[1], "0.1.0+abc1234") {
		t.Errorf("Expected version '0.1.0+abc1234' in line: %q", lines[1])
	}

	// Check agent is on third line
	if !strings.HasPrefix(lines[2], "Agent: ") {
		t.Errorf("Expected 'Agent:' on third line, got: %q", lines[2])
	}

	// Verify all config values are present
	expectedFields := map[string]string{
		"Version:":               "0.1.0+abc1234",
		"Agent:":                 "claude",
		"Session ID:":            "banner-test-session",
		"Max Review Cycles:":     "20",
		"Max Implement Retries:": "100",
		"Max Commit Retries:":    "50",
		"Task File:":             "/path/to/task.md",
	}

	verifyExpectedFieldsInOutput(t, output, expectedFields)

	// Check footer line
	if !strings.Contains(output, "======================================") {
		t.Error("Expected banner footer line")
	}

	// Verify ordering: Version should come before Agent
	versionIdx := strings.Index(output, "Version:")
	agentIdx := strings.Index(output, "Agent:")
	if versionIdx >= agentIdx {
		t.Error("Expected Version to appear before Agent in output")
	}
}

func TestPrintTextToWriter_VersionInBannerWithCommitHash(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	status := InitializationStatus{
		Version:             "1.2.3-next+abc123",
		SessionID:           "version-test",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		MaxCommitRetries:    100,
		TaskFile:            "",
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	PrintTextToWriter(&buf, status)
	output := buf.String()

	// Verify version with commit hash is displayed
	if !strings.Contains(output, "Version: 1.2.3-next+abc123") {
		t.Errorf("Expected version with commit hash in output, got:\n%s", output)
	}

	// Verify version line comes immediately after header
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Fatal("Output has less than 2 lines")
	}
	if !strings.Contains(lines[0], "=== fluxid Workflow Initialization ===") {
		t.Errorf("Expected header on first line, got: %q", lines[0])
	}
	if !strings.Contains(lines[1], "Version: 1.2.3-next+abc123") {
		t.Errorf("Expected version on second line, got: %q", lines[1])
	}
}

// verifyExpectedFieldsInOutput checks that all expected fields and values appear in the output.
func verifyExpectedFieldsInOutput(t *testing.T, output string, expectedFields map[string]string) {
	t.Helper()

	for field, value := range expectedFields {
		expectedLine := field + " " + value
		if !strings.Contains(output, expectedLine) {
			t.Errorf("Expected to find '%s' in output", expectedLine)
		}
	}
}
