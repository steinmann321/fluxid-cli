//nolint:paralleltest,goconst // CLI argument parsing tests with repeated test strings
package command

import (
	"os"
	"testing"
)

func TestParseArgsWithNoAgentFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations=20"}
	args, err := ParseArgs()
	if err != nil {
		t.Errorf("Expected no error when agent flag is optional, got: %v", err)
	}

	if args.CLIAgent != nil {
		t.Errorf("Expected cliAgent=nil when no agent flag provided, got: %v", args.CLIAgent)
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
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIAgent == nil || *args.CLIAgent != "claude" {
		t.Errorf("Expected cliAgent=claude, got %v", args.CLIAgent)
	}

	if args.CLIIterations != nil {
		t.Errorf("Expected cliIterations=nil, got %v", args.CLIIterations)
	}

	if args.CLIImplementRetries != nil {
		t.Errorf("Expected cliImplementRetries=nil, got %v", args.CLIImplementRetries)
	}

	if len(args.AgentArgs) != 0 {
		t.Errorf("Expected empty agentArgs, got %v", args.AgentArgs)
	}
}

//nolint:cyclop // Unit test validating all CLI flags and combinations
func TestParseArgsWithAllFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"fluxid", "--claude", "--model", "gpt-4",
		"--fluxid-iterations=25",
		"--fluxid-implement-retries=8",
	}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIAgent == nil || *args.CLIAgent != "claude" {
		t.Errorf("Expected cliAgent=claude, got %v", args.CLIAgent)
	}

	if args.CLIIterations == nil || *args.CLIIterations != 25 {
		t.Errorf("Expected cliIterations=25, got %v", args.CLIIterations)
	}

	if args.CLIImplementRetries == nil || *args.CLIImplementRetries != 8 {
		t.Errorf("Expected cliImplementRetries=8, got %v", args.CLIImplementRetries)
	}

	if len(args.AgentArgs) != 2 || args.AgentArgs[0] != "--model" || args.AgentArgs[1] != "gpt-4" {
		t.Errorf("Expected agentArgs=[--model, gpt-4], got %v", args.AgentArgs)
	}
}

func TestParseArgsMultipleAgentFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--claude", "--codex"}
	_, err := ParseArgs()

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
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIAgent == nil || *args.CLIAgent != "codex" {
		t.Errorf("Expected cliAgent=codex, got %v", args.CLIAgent)
	}

	if len(args.AgentArgs) != 2 || args.AgentArgs[0] != "--api-key" || args.AgentArgs[1] != "test" {
		t.Errorf("Expected agentArgs=[--api-key, test], got %v", args.AgentArgs)
	}
}

func TestParseArgsOpencodeAgent(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--opencode"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIAgent == nil || *args.CLIAgent != "opencode" {
		t.Errorf("Expected cliAgent=opencode, got %v", args.CLIAgent)
	}
}

// TestOsEnvGetenvNonexistent removed - environment variable support removed in v2.0

func TestParseArgsWithFluxidOutputFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-output=json", "--claude"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIOutputFormat == nil || *args.CLIOutputFormat != testFormatJSON {
		t.Errorf("Expected cliOutputFormat=json, got %v", args.CLIOutputFormat)
	}
}

func TestParseArgsWithDryRunFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-dry-run", "--claude"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIDryRun == nil || *args.CLIDryRun != true {
		t.Errorf("Expected cliDryRun=true, got %v", args.CLIDryRun)
	}
}

func TestParseArgsWithOutputEquals(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--output=json", "claude"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIOutputFormat == nil || *args.CLIOutputFormat != "json" {
		t.Errorf("Expected output format 'json', got %v", args.CLIOutputFormat)
	}
}

func TestParseArgsWithFluxidIterationsFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations=5", "claude"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIIterations == nil || *args.CLIIterations != 5 {
		t.Errorf("Expected iterations 5, got %v", args.CLIIterations)
	}
}

func TestParseArgsWithFluxidImplementRetriesFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-implement-retries=3", "claude"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIImplementRetries == nil || *args.CLIImplementRetries != 3 {
		t.Errorf("Expected implement retries 3, got %v", args.CLIImplementRetries)
	}
}

// TestParseArgsWithFluxidCommitEnabledFlag removed - commit toggle flags removed in v2.0
// TestParseArgsWithFluxidNoCommitFlag removed - commit toggle flags removed in v2.0

func TestParseArgsWithFluxidDryRunFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-dry-run", "claude"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIDryRun == nil || *args.CLIDryRun != true {
		t.Errorf("Expected dry run true, got %v", args.CLIDryRun)
	}
}
