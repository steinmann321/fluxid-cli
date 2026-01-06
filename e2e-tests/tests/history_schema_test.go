//nolint:cyclop,funlen // E2E tests: comprehensive scenarios justify complexity
package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestHistoryGetSchema tests retrieval of the history schema.
// Per User Story 5 (FR-006): Agent retrieves history schema to understand validation rules.
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
//nolint:cyclop // E2E test complexity justified by comprehensive validation
//nolint:cyclop,funlen // E2E test: comprehensive validation scenarios
func TestHistoryGetSchema(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "history", "--get-schema")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Should succeed with exit 0
	if err := cmd.Run(); err != nil {
		t.Fatalf("Getting schema should succeed, got error: %v\nStderr: %s", err, stderr.String())
	}

	// Verify silent success (no stderr output per FR-042)
	if stderr.Len() > 0 {
		t.Errorf("Expected no stderr output, got: %s", stderr.String())
	}

	// Verify stdout contains valid YAML schema
	schemaOutput := stdout.String()
	if schemaOutput == "" {
		t.Fatal("Expected schema output on stdout, got empty string")
	}

	// Parse YAML to verify it's valid
	var schema map[string]interface{}
	if err := yaml.Unmarshal([]byte(schemaOutput), &schema); err != nil {
		t.Fatalf("Schema output should be valid YAML, got parse error: %v\nOutput: %s", err, schemaOutput)
	}

	// Verify schema type is array (history is a YAML array)
	schemaType, hasType := schema["type"]
	if !hasType {
		t.Fatal("Schema should have 'type' field")
	}

	if schemaType != "array" {
		t.Errorf("Schema type should be 'array' for history, got: %v", schemaType)
	}

	// Verify schema has items definition
	_, hasItems := schema["items"]
	if !hasItems {
		t.Fatal("Schema should have 'items' definition for array elements")
	}

	// Verify schema documents required fields for events
	if !strings.Contains(schemaOutput, "timestamp") {
		t.Error("Schema should document 'timestamp' field")
	}
	if !strings.Contains(schemaOutput, "step") {
		t.Error("Schema should document 'step' field")
	}
	if !strings.Contains(schemaOutput, "status") {
		t.Error("Schema should document 'status' field")
	}
	if !strings.Contains(schemaOutput, "summary") {
		t.Error("Schema should document 'summary' field")
	}

	// Verify schema documents the enum constraint for status
	if !strings.Contains(schemaOutput, "SUCCESS") || !strings.Contains(schemaOutput, "FAIL") {
		t.Error("Schema should document status enum values (SUCCESS, FAIL)")
	}

	// Verify schema documents required fields
	if !strings.Contains(schemaOutput, "required") {
		t.Error("Schema should indicate which fields are required")
	}
}

// TestHistoryGetSchemaNoSessionRequired tests that --get-schema works without session.
// Per FR-006: Schema retrieval doesn't require session context.
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
//nolint:cyclop // E2E test complexity justified by comprehensive validation
//nolint:cyclop,funlen // E2E test: comprehensive validation scenarios
func TestHistoryGetSchemaNoSessionRequired(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "history", "--get-schema")

	// Explicitly clear FLUXID_SESSION_ID to verify it's not required
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Should succeed even without session
	if err := cmd.Run(); err != nil {
		t.Fatalf("Getting schema should work without session, got error: %v\nStderr: %s", err, stderr.String())
	}

	// Verify we got schema output
	if stdout.Len() == 0 {
		t.Error("Expected schema output even without session context")
	}

	// Verify no error about missing session
	stderrOutput := stderr.String()
	if strings.Contains(stderrOutput, "FLUXID_SESSION_ID") || strings.Contains(stderrOutput, "session") {
		t.Errorf("Should not require session for schema retrieval, but got session-related error: %s", stderrOutput)
	}
}
