//nolint:paralleltest // CLI argument parsing tests
package main

import (
	"os"
	"testing"
)

func TestParseArgsWithNoAgentFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "20"}
	args, err := parseArgs()
	if err != nil {
		t.Errorf("Expected no error when agent flag is optional, got: %v", err)
	}

	if args.cliAgent != nil {
		t.Errorf("Expected cliAgent=nil when no agent flag provided, got: %v", args.cliAgent)
	}
}

func containsString(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	return s == substr || s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr || findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseArgsWithClaudeOnly(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--claude"}
	args, err := parseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.cliAgent == nil || *args.cliAgent != "claude" {
		t.Errorf("Expected cliAgent=claude, got %v", args.cliAgent)
	}

	if args.cliIterations != nil {
		t.Errorf("Expected cliIterations=nil, got %v", args.cliIterations)
	}

	if args.cliImplementRetries != nil {
		t.Errorf("Expected cliImplementRetries=nil, got %v", args.cliImplementRetries)
	}

	if args.cliCommitEnabled != nil {
		t.Errorf("Expected cliCommitEnabled=nil, got %v", args.cliCommitEnabled)
	}

	if len(args.agentArgs) != 0 {
		t.Errorf("Expected empty agentArgs, got %v", args.agentArgs)
	}
}

func TestParseArgsWithAllFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"fluxid", "--claude", "--model", "gpt-4",
		"--fluxid-iterations", "25",
		"--fluxid-implement-retries", "8",
		"--fluxid-no-commit",
	}
	args, err := parseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.cliAgent == nil || *args.cliAgent != "claude" {
		t.Errorf("Expected cliAgent=claude, got %v", args.cliAgent)
	}

	if args.cliIterations == nil || *args.cliIterations != 25 {
		t.Errorf("Expected cliIterations=25, got %v", args.cliIterations)
	}

	if args.cliImplementRetries == nil || *args.cliImplementRetries != 8 {
		t.Errorf("Expected cliImplementRetries=8, got %v", args.cliImplementRetries)
	}

	if args.cliCommitEnabled == nil || *args.cliCommitEnabled != false {
		t.Errorf("Expected cliCommitEnabled=false, got %v", args.cliCommitEnabled)
	}

	if len(args.agentArgs) != 2 || args.agentArgs[0] != "--model" || args.agentArgs[1] != "gpt-4" {
		t.Errorf("Expected agentArgs=[--model, gpt-4], got %v", args.agentArgs)
	}
}

func TestParseArgsMultipleAgentFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--claude", "--codex"}
	_, err := parseArgs()

	if err == nil {
		t.Error("Expected error for multiple agent flags")
	}

	if !containsString(err.Error(), "multiple agent flags specified") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestParseArgsCodexAgent(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--codex", "--api-key", "test"}
	args, err := parseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.cliAgent == nil || *args.cliAgent != "codex" {
		t.Errorf("Expected cliAgent=codex, got %v", args.cliAgent)
	}

	if len(args.agentArgs) != 2 || args.agentArgs[0] != "--api-key" || args.agentArgs[1] != "test" {
		t.Errorf("Expected agentArgs=[--api-key, test], got %v", args.agentArgs)
	}
}

func TestParseArgsOpencodeAgent(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--opencode"}
	args, err := parseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.cliAgent == nil || *args.cliAgent != "opencode" {
		t.Errorf("Expected cliAgent=opencode, got %v", args.cliAgent)
	}
}

func TestOsEnvGetenvNonexistent(t *testing.T) {
	env := osEnv{}
	got := env.Getenv("NONEXISTENT_FLUXID_TEST_VAR_12345")
	if got != "" {
		t.Errorf("Expected empty string for nonexistent var, got %v", got)
	}
}

func TestParseArgsWithFluxidOutputFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-output", "json", "--claude"}
	args, err := parseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	//nolint:goconst // String comparison for test clarity
	if args.cliOutputFormat == nil || *args.cliOutputFormat != "json" {
		t.Errorf("Expected cliOutputFormat=json, got %v", args.cliOutputFormat)
	}
}

func TestParseArgsWithDryRunFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-dry-run", "--claude"}
	args, err := parseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.cliDryRun == nil || *args.cliDryRun != true {
		t.Errorf("Expected cliDryRun=true, got %v", args.cliDryRun)
	}
}
