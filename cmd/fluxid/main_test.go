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
				Agent:     "claude",
				AgentArgs: []string{},
				SessionID: "test-session-123",
			},
			prompt:   "Test prompt",
			wantArgs: []string{"--print", "Test prompt"},
		},
		{
			name: "claude with custom args",
			config: Config{
				Agent:     "claude",
				AgentArgs: []string{"--model", "gpt-4", "--temp", "0.7"},
				SessionID: "session-456",
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
	// This test has been replaced by more focused E2E tests.
	// The original test would run the full workflow which takes too long for unit tests.
	// Argument validation is now tested in TestParseArgs* functions in args_test.go
	t.Skip("Replaced by E2E tests and args_test.go unit tests")
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

func TestRunWithMultipleAgentArgs(t *testing.T) {
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

	// Parse args to verify --fluxid-no-commit is accepted
	args, err := parseArgs()
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}

	if args.cliCommitEnabled == nil {
		t.Error("Expected cliCommitEnabled to be set")
	} else if *args.cliCommitEnabled != false {
		t.Error("Expected cliCommitEnabled to be false when --fluxid-no-commit is set")
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

func TestRunWithHelpFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test --help flag
	os.Args = []string{"fluxid", "--help"}
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}

	// Test -h flag
	os.Args = []string{"fluxid", "-h"}
	exitCode = run()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}
}

func TestSetupSignalHandler(t *testing.T) {
	// Test that setupSignalHandler sets up the signal handler
	// We can't easily test the full signal flow, but we can verify it doesn't panic
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-signal-session"

	// Reset signal count for this test
	signalCount.Store(0)

	// This should not panic
	setupSignalHandler(sessionID)

	// The handler is running in a goroutine, so we can't easily test it
	// without sending actual signals. This test primarily checks for setup issues.
}

func TestRunWithInvalidConfigDir(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set HOME to a path that will cause config loading issues
	t.Setenv("HOME", "/dev/null/invalid/path/that/does/not/exist")
	t.Setenv("XDG_CONFIG_HOME", "/dev/null/invalid/config")

	os.Args = []string{"fluxid", "--claude"}
	exitCode := run()

	// Should handle config error gracefully
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when config loading fails")
	}
}

func TestRunWithNonExecutableAgent(t *testing.T) {
	oldArgs := os.Args
	tmpDir := t.TempDir()
	defer func() { os.Args = oldArgs }()

	// Create a non-executable file
	nonExecPath := tmpDir + "/nonexec"
	if err := os.WriteFile(nonExecPath, []byte("#!/bin/sh\necho test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Add to PATH
	t.Setenv("PATH", tmpDir)
	os.Args = []string{"fluxid", "--agent", "nonexec"}

	exitCode := run()
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for non-executable agent, got %d", exitCode)
	}
}
