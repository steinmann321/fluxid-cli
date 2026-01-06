//nolint:cyclop // E2E tests: comprehensive scenarios justify complexity
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

// TestReportGetSchema tests retrieval of the report schema.
// Per User Story 3 (FR-003): Agent retrieves report schema to understand validation rules.
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
//nolint:cyclop // E2E test complexity justified by comprehensive validation
//nolint:cyclop // E2E test: comprehensive validation scenarios
func TestReportGetSchema(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--get-schema")

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

	// Verify schema contains expected properties section
	properties, hasProperties := schema["properties"]
	if !hasProperties {
		t.Fatal("Schema should have 'properties' section")
	}

	propertiesMap, ok := properties.(map[string]interface{})
	if !ok {
		t.Fatal("Schema 'properties' should be an object")
	}

	// Verify schema documents expected fields
	expectedFields := []string{"command", "artifact", "timestamp", "status", "issues"}
	for _, field := range expectedFields {
		if _, exists := propertiesMap[field]; !exists {
			t.Errorf("Schema should document field '%s', but it's missing from properties", field)
		}
	}

	// Verify schema documents the enum constraint for status
	// We're checking that the schema output contains information about valid enum values
	if !strings.Contains(schemaOutput, "PASS") || !strings.Contains(schemaOutput, "FAIL") {
		t.Errorf("Schema should document status enum values (PASS, FAIL), but output doesn't contain them")
	}

	// Verify schema documents required fields
	if !strings.Contains(schemaOutput, "required") {
		t.Errorf("Schema should indicate which fields are required")
	}
}

// TestReportGetSchemaNoSessionRequired tests that --get-schema works without session.
// Per FR-003: Schema retrieval doesn't require session context.
//
//nolint:cyclop // Complexity inherent to validation/workflow logic
//nolint:cyclop // E2E test complexity justified by comprehensive validation
//nolint:cyclop // E2E test: comprehensive validation scenarios
func TestReportGetSchemaNoSessionRequired(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), binPath, "report", "--get-schema")

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
