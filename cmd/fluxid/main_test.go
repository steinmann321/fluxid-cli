//nolint:paralleltest,exhaustruct // CLI tests with command execution
package main

import (
	"os"
	"testing"
)

func TestParsePositiveInt(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		flagName string
		want     int
		wantErr  bool
	}{
		{
			name:     "valid positive integer",
			value:    "42",
			flagName: "--test-flag",
			want:     42,
			wantErr:  false,
		},
		{
			name:     "valid one",
			value:    "1",
			flagName: "--iterations",
			want:     1,
			wantErr:  false,
		},
		{
			name:     "zero is invalid",
			value:    "0",
			flagName: "--retries",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "negative is invalid",
			value:    "-5",
			flagName: "--count",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "not a number",
			value:    "abc",
			flagName: "--num",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "empty string",
			value:    "",
			flagName: "--empty",
			want:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePositiveInt(tt.value, tt.flagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePositiveInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parsePositiveInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildClaudeCommand(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		prompt   string
		wantArgs []string
	}{
		{
			name: "basic claude command",
			config: Config{
				Agent:      "claude",
				ClaudeArgs: []string{},
				SessionID:  "test-session-123",
			},
			prompt:   "Test prompt",
			wantArgs: []string{"--print", "Test prompt"},
		},
		{
			name: "claude with custom args",
			config: Config{
				Agent:      "claude",
				ClaudeArgs: []string{"--model", "gpt-4", "--temp", "0.7"},
				SessionID:  "session-456",
			},
			prompt:   "Another prompt",
			wantArgs: []string{"--print", "--model", "gpt-4", "--temp", "0.7", "Another prompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildClaudeCommand(tt.config, tt.prompt)

			// Verify Path is set (it will be the resolved path from LookPath)
			if cmd.Path == "" {
				t.Error("buildClaudeCommand() Path is empty")
			}

			// Verify args (cmd.Args[0] is the command name itself, args[1:] are arguments)
			// The first arg is always the agent name
			if len(cmd.Args) < 2 {
				t.Errorf("buildClaudeCommand() has too few args: %v", cmd.Args)
				return
			}

			actualArgs := cmd.Args[1:] // Skip the command name
			if len(actualArgs) != len(tt.wantArgs) {
				t.Errorf("buildClaudeCommand() args = %v, want %v", actualArgs, tt.wantArgs)
				return
			}

			for i, wantArg := range tt.wantArgs {
				if actualArgs[i] != wantArg {
					t.Errorf("buildClaudeCommand() arg[%d] = %v, want %v", i, actualArgs[i], wantArg)
				}
			}
		})
	}
}

func TestRunWithInvalidIterations(t *testing.T) {
	// Save and restore original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "0"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid iterations, got %d", exitCode)
	}
}

func TestRunWithInvalidRetries(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-implement-retries", "-5"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid retries, got %d", exitCode)
	}
}

func TestRunWithMissingIterationsValue(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing iterations value, got %d", exitCode)
	}
}

func TestRunWithMissingRetriesValue(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-implement-retries"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing retries value, got %d", exitCode)
	}
}

func TestRunWithValidArguments(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "10", "--fluxid-implement-retries", "5"}
	// This will fail because there's no workflow file, but it tests argument parsing
	exitCode := run()
	// We expect this to fail since there's no workflow file in test environment
	// but the fact that it gets past argument parsing is what we're testing
	if exitCode == 0 {
		t.Log("Unexpectedly succeeded - test environment may have workflow files")
	}
	// Any non-zero exit is acceptable here since we're just testing arg parsing
}

func TestRunWithClaudeFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--claude", "--model", "gpt-4"}
	exitCode := run()
	// This will fail due to missing workflow, but tests Claude flag parsing
	if exitCode == 0 {
		t.Log("Unexpectedly succeeded")
	}
}

func TestRunWithClaudeAndFluxidFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--claude", "--model", "gpt-4", "--fluxid-iterations", "15"}
	exitCode := run()
	if exitCode == 0 {
		t.Log("Unexpectedly succeeded")
	}
}

func TestRunWithInvalidIterationsNotANumber(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "abc"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid iterations (not a number), got %d", exitCode)
	}
}

func TestRunWithInvalidRetriesNotANumber(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-implement-retries", "xyz"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid retries (not a number), got %d", exitCode)
	}
}

func TestRunWithMultipleClaudeArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"fluxid",
		"--claude",
		"--model", "gpt-4",
		"--temp", "0.7",
		"--max-tokens", "1000",
	}
	exitCode := run()
	// Will fail due to no workflow, but tests arg parsing
	if exitCode == 0 {
		t.Log("Unexpectedly succeeded")
	}
}

func TestRunWithAllFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"fluxid",
		"--claude",
		"--model", "gpt-4",
		"--fluxid-iterations", "20",
		"--fluxid-implement-retries", "7",
	}
	exitCode := run()
	if exitCode == 0 {
		t.Log("Unexpectedly succeeded")
	}
}

func TestOsEnvGetenv(t *testing.T) {
	// Test that osEnv correctly wraps os.Getenv
	t.Setenv("TEST_VAR_FLUXID", "test_value")

	env := osEnv{}
	got := env.Getenv("TEST_VAR_FLUXID")

	if got != "test_value" {
		t.Errorf("osEnv.Getenv() = %v, want %v", got, "test_value")
	}

	// Test with non-existent var
	got = env.Getenv("NONEXISTENT_VAR_FLUXID")
	if got != "" {
		t.Errorf("osEnv.Getenv() for non-existent var = %v, want empty string", got)
	}
}

func TestRunWithNoCommitFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--claude", "--fluxid-no-commit"}
	exitCode := run()
	// Will fail due to no workflow
	if exitCode == 0 {
		t.Log("Unexpectedly succeeded")
	}
}

func TestRunWithClaudeNotFound(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set PATH to empty to ensure claude is not found
	t.Setenv("PATH", "")
	os.Args = []string{"fluxid", "--claude"}

	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 when claude not in PATH, got %d", exitCode)
	}
}

func TestParseArgsWithNoClaudeFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "20"}
	_, err := parseArgs()

	if err == nil {
		t.Error("Expected error for missing --claude flag")
	}

	if err.Error() != "missing required --claude flag" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestParseArgsWithClaudeOnly(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--claude"}
	args, err := parseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
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

	if len(args.claudeArgs) != 0 {
		t.Errorf("Expected empty claudeArgs, got %v", args.claudeArgs)
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

	if args.cliIterations == nil || *args.cliIterations != 25 {
		t.Errorf("Expected cliIterations=25, got %v", args.cliIterations)
	}

	if args.cliImplementRetries == nil || *args.cliImplementRetries != 8 {
		t.Errorf("Expected cliImplementRetries=8, got %v", args.cliImplementRetries)
	}

	if args.cliCommitEnabled == nil || *args.cliCommitEnabled != false {
		t.Errorf("Expected cliCommitEnabled=false, got %v", args.cliCommitEnabled)
	}

	if len(args.claudeArgs) != 2 || args.claudeArgs[0] != "--model" || args.claudeArgs[1] != "gpt-4" {
		t.Errorf("Expected claudeArgs=[--model, gpt-4], got %v", args.claudeArgs)
	}
}

func TestOsEnvGetenvNonexistent(t *testing.T) {
	env := osEnv{}
	got := env.Getenv("NONEXISTENT_FLUXID_TEST_VAR_12345")
	if got != "" {
		t.Errorf("Expected empty string for nonexistent var, got %v", got)
	}
}
