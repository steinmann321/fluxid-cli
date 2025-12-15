//nolint:paralleltest // Tests modify os.Args
package main

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

	args := &cliArgs{
		cliAgent:            nil,
		agentArgs:           nil,
		cliIterations:       nil,
		cliImplementRetries: nil,
		cliCommitEnabled:    nil,
		cliDryRun:           nil,
		cliOutputFormat:     nil,
	}
	skip, handled, err := parseFluxidFlag("--fluxid-output", 1, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !handled {
		t.Error("Expected flag to be handled")
	}
	if skip != 1 {
		t.Errorf("Expected skip=1, got %d", skip)
	}
	if args.cliOutputFormat == nil || *args.cliOutputFormat != "json" {
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
