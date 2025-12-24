//nolint:paralleltest // CLI argument parsing error tests
package command

import (
	"os"
	"strings"
	"testing"
)

func TestParseArgsIterationsNegative(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "-5", "--claude"}
	_, err := ParseArgs()

	if err == nil {
		t.Error("Expected error for negative iterations value")
	}

	if !strings.Contains(err.Error(), "must be a positive integer") {
		t.Errorf("Expected error about positive integer, got: %v", err)
	}
}

func TestParseArgsIterationsZero(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "0", "--claude"}
	_, err := ParseArgs()

	if err == nil {
		t.Error("Expected error for zero iterations value")
	}

	if !strings.Contains(err.Error(), "must be a positive integer") {
		t.Errorf("Expected error about positive integer, got: %v", err)
	}
}

func TestParseArgsImplementRetriesNegative(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-implement-retries", "-3", "--claude"}
	_, err := ParseArgs()

	if err == nil {
		t.Error("Expected error for negative implement retries value")
	}

	if !strings.Contains(err.Error(), "must be a positive integer") {
		t.Errorf("Expected error about positive integer, got: %v", err)
	}
}

func TestParseArgsIterationsInvalid(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "not-a-number", "--claude"}
	_, err := ParseArgs()

	if err == nil {
		t.Error("Expected error for invalid iterations value")
	}
}

func TestParseArgsImplementRetriesInvalid(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-implement-retries", "invalid", "--claude"}
	_, err := ParseArgs()

	if err == nil {
		t.Error("Expected error for invalid implement retries value")
	}
}

func TestParseArgsIterationsMissing(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations"}
	_, err := ParseArgs()

	if err == nil {
		t.Error("Expected error for missing iterations value")
	}
}

func TestParseArgsImplementRetriesMissing(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-implement-retries"}
	_, err := ParseArgs()

	if err == nil {
		t.Error("Expected error for missing implement retries value")
	}
}

func TestParseArgsOutputFormatMissing(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-output"}
	_, err := ParseArgs()

	if err == nil {
		t.Error("Expected error for missing output format value")
	}
}
