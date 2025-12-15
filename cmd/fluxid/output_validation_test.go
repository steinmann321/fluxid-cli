//nolint:paralleltest // Output tests with log capture
package main

import (
	"strings"
	"testing"
)

func TestValidateOutputFormatValidText(t *testing.T) {
	err := ValidateOutputFormat("text")
	if err != nil {
		t.Errorf("ValidateOutputFormat('text') should not return error, got: %v", err)
	}
}

func TestValidateOutputFormatValidJSON(t *testing.T) {
	err := ValidateOutputFormat("json")
	if err != nil {
		t.Errorf("ValidateOutputFormat('json') should not return error, got: %v", err)
	}
}

func TestValidateOutputFormatInvalidFormat(t *testing.T) {
	err := ValidateOutputFormat("invalid")
	if err == nil {
		t.Error("ValidateOutputFormat('invalid') should return error")
	}
}

func TestValidateOutputFormatInvalidFormatXML(t *testing.T) {
	err := ValidateOutputFormat("xml")
	if err == nil {
		t.Error("ValidateOutputFormat('xml') should return error")
	}

	// Verify error message contains helpful information
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Errorf("Expected error message to contain 'unsupported output format', got: %v", err)
	}
}

func TestValidateOutputFormatValidYAML(t *testing.T) {
	err := ValidateOutputFormat("yaml")
	if err != nil {
		t.Errorf("ValidateOutputFormat('yaml') should not return error, got: %v", err)
	}
}

func TestValidateOutputFormatEmptyString(t *testing.T) {
	err := ValidateOutputFormat("")
	if err == nil {
		t.Error("ValidateOutputFormat('') should return error")
	}
}

func TestValidateOutputFormatCaseInsensitivity(t *testing.T) {
	// The current implementation is case-sensitive, so uppercase should fail
	err := ValidateOutputFormat("TEXT")
	if err == nil {
		t.Error("ValidateOutputFormat('TEXT') should return error (case-sensitive)")
	}

	err = ValidateOutputFormat("JSON")
	if err == nil {
		t.Error("ValidateOutputFormat('JSON') should return error (case-sensitive)")
	}
}
