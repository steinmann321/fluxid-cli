//nolint:paralleltest,goconst // CLI argument parsing tests with repeated test strings
package command

import (
	"os"
	"testing"
)

func TestParseArgs_IterationsEqualsOnly(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations=25"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIIterations == nil || *args.CLIIterations != 25 {
		t.Errorf("Expected iterations=25, got %v", args.CLIIterations)
	}
}

func TestParseArgs_IterationsSpaceSyntaxRejected(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "25"}
	_, err := ParseArgs()
	if err == nil {
		t.Error("Expected error for space syntax, got nil")
	}

	expectedMsg := "requires equals syntax"
	if err != nil && !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

func TestParseArgs_RetriesEqualsOnly(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-implement-retries=5"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIImplementRetries == nil || *args.CLIImplementRetries != 5 {
		t.Errorf("Expected retries=5, got %v", args.CLIImplementRetries)
	}
}

func TestParseArgs_RetriesSpaceSyntaxRejected(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-implement-retries", "5"}
	_, err := ParseArgs()
	if err == nil {
		t.Error("Expected error for space syntax, got nil")
	}

	expectedMsg := "requires equals syntax"
	if err != nil && !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

func TestParseArgs_OutputSpaceSyntaxRejected(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--output", "json"}
	_, err := ParseArgs()
	if err == nil {
		t.Error("Expected error for space syntax, got nil")
	}

	expectedMsg := "requires equals syntax"
	if err != nil && !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

// TestParseArgs_BooleanFlagsNoEquals removed - commit toggle flags removed in v2.0

func TestParseArgs_AllEqualsFormat(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"fluxid",
		"--fluxid-iterations=30",
		"--fluxid-implement-retries=10",
		"--output=json",
	}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIIterations == nil || *args.CLIIterations != 30 {
		t.Errorf("Expected iterations=30, got %v", args.CLIIterations)
	}

	if args.CLIImplementRetries == nil || *args.CLIImplementRetries != 10 {
		t.Errorf("Expected retries=10, got %v", args.CLIImplementRetries)
	}

	if args.CLIOutputFormat == nil || *args.CLIOutputFormat != "json" {
		t.Errorf("Expected output=json, got %v", args.CLIOutputFormat)
	}
}

func TestParseArgs_EmptyValueAfterEqualsAllowed(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--output="}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIOutputFormat == nil || *args.CLIOutputFormat != "" {
		t.Errorf("Expected empty output format, got %v", args.CLIOutputFormat)
	}
}

func TestParseArgs_InvalidIntegerWithEquals(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations=abc"}
	_, err := ParseArgs()
	if err == nil {
		t.Error("Expected error for invalid integer, got nil")
	}

	expectedMsg := "valid integer"
	if err != nil && !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}
