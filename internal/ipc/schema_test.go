package ipc

import (
	"bytes"
	"strings"
	"testing"
)

func TestGetReportSchema(t *testing.T) {
	t.Parallel()

	schema := GetReportSchema()

	// Verify schema is not empty
	if schema == "" {
		t.Error("GetReportSchema() returned empty string")
	}

	// Verify schema contains expected YAML structure
	expectedFields := []string{
		"command:",
		"artifact:",
		"timestamp:",
		"status:",
		"issues:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(schema, field) {
			t.Errorf("Schema missing expected field: %s", field)
		}
	}

	// Verify schema contains status enum values
	statusValues := []string{"PASS", "FAIL"}
	for _, value := range statusValues {
		if !strings.Contains(schema, value) {
			t.Errorf("Schema missing expected status value: %s", value)
		}
	}
}

func TestWriteReportSchema(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := WriteReportSchema(&buf)
	if err != nil {
		t.Fatalf("WriteReportSchema() error = %v", err)
	}

	output := buf.String()

	// Verify output matches GetReportSchema
	expected := GetReportSchema()
	if output != expected {
		t.Errorf("WriteReportSchema() output doesn't match GetReportSchema()\nGot:\n%s\nWant:\n%s", output, expected)
	}

	// Verify output is valid YAML structure
	if !strings.HasPrefix(output, "# ") && !strings.Contains(output, "command:") {
		t.Error("WriteReportSchema() output doesn't appear to be valid YAML")
	}
}

func TestWriteReportSchemaErrorHandling(t *testing.T) {
	t.Parallel()

	// Test with a writer that always fails
	errWriter := &failingWriter{}
	err := WriteReportSchema(errWriter)

	if err == nil {
		t.Error("WriteReportSchema() with failing writer should return error")
	}

	// Verify error is wrapped correctly
	if !strings.Contains(err.Error(), "failed to write schema") {
		t.Errorf("Error should be wrapped with context, got: %v", err)
	}
}

// failingWriter is a writer that always returns an error.
type failingWriter struct{}

func (f *failingWriter) Write(_ []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}
