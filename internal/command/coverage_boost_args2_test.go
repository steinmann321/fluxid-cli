//nolint:paralleltest,exhaustruct // Tests modify os.Args, partial struct initialization
package command

import (
	"os"
	"testing"
)

func TestParseFluxidFlag_Output(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set os.Args to simulate flag parsing
	os.Args = []string{"fluxid", "--fluxid-output", "json"}

	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           nil,
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
	}
	// Test equals syntax
	skip, handled, err := parseFluxidFlag("--fluxid-output=json", 1, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !handled {
		t.Error("Expected flag to be handled")
	}
	if skip != 0 {
		t.Errorf("Expected skip=0 for equals syntax, got %d", skip)
	}
	if args.CLIOutputFormat == nil || *args.CLIOutputFormat != "json" {
		t.Error("Expected output format to be 'json'")
	}
}

func TestHandleSpecialCommands_NoSpecialCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "run"}

	_, handled := handleSpecialCommands()
	if handled {
		t.Error("Expected command not to be handled")
	}
}
