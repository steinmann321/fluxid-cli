package tests

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestM03E01GetReportSchemaOutputsValidYAML verifies that the command
// outputs valid YAML to stdout without errors.
func TestM03E01GetReportSchemaOutputsValidYAML(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "ipc", "get-report-schema")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()

	// Verify output is not empty
	if len(output) == 0 {
		t.Fatal("Output is empty")
	}

	// Parse as YAML to verify it's valid
	var schema map[string]interface{}
	if err := yaml.Unmarshal([]byte(output), &schema); err != nil {
		t.Fatalf("Output is not valid YAML: %v\nOutput:\n%s", err, output)
	}

	// Verify schema is non-empty
	if len(schema) == 0 {
		t.Fatal("Parsed schema is empty")
	}
}

// TestM03E01GetReportSchemaContainsRequiredFields verifies that the schema
// contains all required fields and the status enum with PASS/FAIL values.
func TestM03E01GetReportSchemaContainsRequiredFields(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "ipc", "get-report-schema")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	var schema map[string]interface{}
	if err := yaml.Unmarshal([]byte(output), &schema); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// Verify 'required' field exists and contains expected keys
	required, ok := schema["required"].([]interface{})
	if !ok {
		t.Fatal("Schema does not have a 'required' array")
	}

	expectedRequired := map[string]bool{
		"command":   false,
		"artifact":  false,
		"timestamp": false,
		"status":    false,
		"issues":    false,
	}

	for _, field := range required {
		fieldName, ok := field.(string)
		if !ok {
			continue
		}
		if _, exists := expectedRequired[fieldName]; exists {
			expectedRequired[fieldName] = true
		}
	}

	for field, found := range expectedRequired {
		if !found {
			t.Errorf("Required field '%s' not found in schema", field)
		}
	}

	// Verify properties exist
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema does not have a 'properties' object")
	}

	// Verify 'status' property has enum with PASS and FAIL
	status, ok := properties["status"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema 'properties' does not contain 'status'")
	}

	statusEnum, ok := status["enum"].([]interface{})
	if !ok {
		t.Fatal("Status property does not have an 'enum' array")
	}

	hasPass := false
	hasFail := false
	for _, value := range statusEnum {
		strValue, ok := value.(string)
		if !ok {
			continue
		}
		if strValue == "PASS" {
			hasPass = true
		}
		if strValue == "FAIL" {
			hasFail = true
		}
	}

	if !hasPass {
		t.Error("Status enum does not contain 'PASS'")
	}
	if !hasFail {
		t.Error("Status enum does not contain 'FAIL'")
	}

	// Verify 'issues' property exists and has required categories
	issues, ok := properties["issues"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema 'properties' does not contain 'issues'")
	}

	issuesProperties, ok := issues["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Issues property does not have 'properties'")
	}

	expectedCategories := []string{"blockers", "defects", "concerns", "observations", "enhancements"}
	for _, category := range expectedCategories {
		if _, exists := issuesProperties[category]; !exists {
			t.Errorf("Issues properties does not contain category '%s'", category)
		}
	}
}

// TestM03E01GetReportSchemaHelp verifies that --help flag works
// and shows usage information.
func TestM03E01GetReportSchemaHelp(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "ipc", "get-report-schema", "--help")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()

	// Verify help text contains key information
	expectedStrings := []string{
		"Usage:",
		"Description:",
		"Example:",
		"fluxid ipc get-report-schema",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output missing expected text: %q\nOutput:\n%s", expected, output)
		}
	}
}

// TestM03E01GetReportSchemaExitsCleanly verifies that the command
// exits with status code 0 on success.
func TestM03E01GetReportSchemaExitsCleanly(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "ipc", "get-report-schema")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command should exit cleanly but failed: %v\nStderr: %s", err, stderr.String())
	}

	// Verify output went to stdout, not stderr
	if stderr.Len() > 0 {
		t.Errorf("Expected no stderr output, got: %s", stderr.String())
	}

	if stdout.Len() == 0 {
		t.Error("Expected stdout output, got none")
	}
}
