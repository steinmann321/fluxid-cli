package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var errWriteFailed = errors.New("write failed")

// failingWriter always returns an error on Write.
type failingWriter struct{}

func (w *failingWriter) Write(_ []byte) (int, error) {
	return 0, errWriteFailed
}

func TestValidateFormat_Valid(t *testing.T) {
	t.Parallel()
	formats := []string{"text", "json", "yaml"}
	for _, format := range formats {
		err := ValidateFormat(format)
		if err != nil {
			t.Errorf("Expected no error for format '%s', got: %v", format, err)
		}
	}
}

func TestValidateFormat_Invalid(t *testing.T) {
	t.Parallel()
	err := ValidateFormat("xml")
	if err == nil {
		t.Error("Expected error for invalid format 'xml'")
	}
}

func TestPrintText(t *testing.T) {
	t.Parallel()
	// Capture stdout
	oldStdout := os.Stdout
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = writePipe

	status := InitializationStatus{
		SessionID:           "test-session-123",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		CommandFiles: &CommandFilesJSON{
			Implement: "/path/to/implement.md",
			Review:    "/path/to/review.md",
			Commit:    "/path/to/commit.md",
		},
		AgentArgs: []string{"--arg1", "value1"},
	}

	PrintText(status)

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(readPipe); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	expectedStrings := []string{
		"=== fluxid Workflow Initialization ===",
		"Agent: claude",
		"Session ID: test-session-123",
		"Max Review Cycles: 10",
		"Max Implement Retries: 3",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s", expected, output)
		}
	}
}

//nolint:dupl,paralleltest // Test setup duplication is acceptable; cannot parallel due to os.Stdout modification.
func TestPrintJSON(t *testing.T) {
	oldStdout := os.Stdout
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = writePipe

	//nolint:exhaustruct // Optional fields CommandFiles and AgentArgs not needed in test.
	status := InitializationStatus{
		SessionID:           "test-session-json",
		Agent:               "claude",
		MaxReviewCycles:     5,
		MaxImplementRetries: 2,
	}

	err = PrintJSON(status)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(readPipe); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Verify it's valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Errorf("Expected valid JSON, got error: %v\nOutput: %s", err, output)
	}

	// Verify some fields
	if decoded["session_id"] != "test-session-json" {
		t.Errorf("Expected session_id='test-session-json', got: %v", decoded["session_id"])
	}
}

//nolint:dupl,paralleltest // Test setup duplication is acceptable; cannot parallel due to os.Stdout modification.
func TestPrintYAML(t *testing.T) {
	oldStdout := os.Stdout
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = writePipe

	//nolint:exhaustruct // Optional fields CommandFiles and AgentArgs not needed in test.
	status := InitializationStatus{
		SessionID:           "test-session-yaml",
		Agent:               "claude",
		MaxReviewCycles:     5,
		MaxImplementRetries: 2,
	}

	err = PrintYAML(status)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	_ = writePipe.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(readPipe); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Verify it's valid YAML
	var decoded map[string]interface{}
	if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
		t.Errorf("Expected valid YAML, got error: %v\nOutput: %s", err, output)
	}

	// Verify some fields
	if decoded["session_id"] != "test-session-yaml" {
		t.Errorf("Expected session_id='test-session-yaml', got: %v", decoded["session_id"])
	}
}

func TestPrintJSONToWriter_Success(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	status := InitializationStatus{
		SessionID:           "test-json-writer",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	err := PrintJSONToWriter(&buf, status)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	if decoded["session_id"] != "test-json-writer" {
		t.Errorf("Expected session_id='test-json-writer', got: %v", decoded["session_id"])
	}
}

func TestPrintJSONToWriter_Error(t *testing.T) {
	t.Parallel()
	failWriter := &failingWriter{}

	status := InitializationStatus{
		SessionID:           "test",
		Agent:               "claude",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	err := PrintJSONToWriter(failWriter, status)
	if err == nil {
		t.Error("Expected error from failing writer, got nil")
	}
	if !strings.Contains(err.Error(), "failed to encode") {
		t.Errorf("Expected error message to contain 'failed to encode', got: %v", err)
	}
}

func TestPrintYAMLToWriter_Success(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	status := InitializationStatus{
		SessionID:           "test-yaml-writer",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	err := PrintYAMLToWriter(&buf, status)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify valid YAML
	var decoded map[string]interface{}
	if err := yaml.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Errorf("Expected valid YAML, got error: %v", err)
	}

	if decoded["session_id"] != "test-yaml-writer" {
		t.Errorf("Expected session_id='test-yaml-writer', got: %v", decoded["session_id"])
	}
}

func TestPrintYAMLToWriter_Error(t *testing.T) {
	t.Parallel()
	failWriter := &failingWriter{}

	status := InitializationStatus{
		SessionID:           "test",
		Agent:               "claude",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	err := PrintYAMLToWriter(failWriter, status)
	if err == nil {
		t.Error("Expected error from failing writer, got nil")
	}
	if !strings.Contains(err.Error(), "failed to") {
		t.Errorf("Expected error message to contain 'failed to', got: %v", err)
	}
}

func TestPrintTextToWriter_WithCommandFiles(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	status := InitializationStatus{
		SessionID:           "test-text-writer",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		CommandFiles: &CommandFilesJSON{
			Implement: "/path/to/implement.md",
			Review:    "/path/to/review.md",
			Commit:    "/path/to/commit.md",
		},
		AgentArgs: []string{"--arg1", "value1"},
	}

	PrintTextToWriter(&buf, status)

	output := buf.String()

	expectedStrings := []string{
		"=== fluxid Workflow Initialization ===",
		"Agent: claude",
		"Session ID: test-text-writer",
		"Command Files:",
		"Implement: /path/to/implement.md",
		"Review: /path/to/review.md",
		"Commit: /path/to/commit.md",
		"Agent Args: [--arg1 value1]",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s", expected, output)
		}
	}
}

func TestPrintTextToWriter_WithoutCommandFiles(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	status := InitializationStatus{
		SessionID:           "test-text-no-files",
		Agent:               "claude",
		MaxReviewCycles:     10,
		MaxImplementRetries: 3,
		CommandFiles:        nil,
		AgentArgs:           nil,
	}

	PrintTextToWriter(&buf, status)

	output := buf.String()

	// Should NOT contain command files section
	if strings.Contains(output, "Command Files:") {
		t.Error("Expected output to NOT contain 'Command Files:' section")
	}

	// Should NOT contain agent args
	if strings.Contains(output, "Agent Args:") {
		t.Error("Expected output to NOT contain 'Agent Args:' section")
	}

	// Should contain basic fields
	if !strings.Contains(output, "Session ID: test-text-no-files") {
		t.Error("Expected output to contain session ID")
	}
}
