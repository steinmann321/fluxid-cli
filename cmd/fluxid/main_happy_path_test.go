package main

import (
	"fmt"
	"os"
	"testing"
)

// TestMain_SuccessfulExecution tests the happy path where command.Execute() returns 0
// and main() exits with code 0.
//
//nolint:dupl // Test functions intentionally similar for different test scenarios
func TestMain_SuccessfulExecution(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up minimal environment for successful execution
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create config with commands section using absolute paths
	configDir := tmpDir + "/.fluxid"
	commandsDir := configDir + "/commands"
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create command files
	for _, file := range []string{"implement.md", "review.md", "commit.md"} {
		if err := os.WriteFile(commandsDir+"/"+file, []byte("# Command"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Write config with absolute paths
	configContent := fmt.Sprintf(`agent: claude
commands:
  implement: %s/commands/implement.md
  review: %s/commands/review.md
  commit: %s/commands/commit.md
`, configDir, configDir, configDir)
	if err := os.WriteFile(configDir+"/config.yaml", []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Simulate dry-run mode to avoid actual agent execution
	os.Args = []string{"fluxid", "--fluxid-dry-run", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 0
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for successful execution, got %d", exitCode)
	}
}

// TestMain_HelpFlag tests that --help flag returns exit code 0.
func TestMain_HelpFlag(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set --help flag
	os.Args = []string{"fluxid", "--help"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 0
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help flag, got %d", exitCode)
	}
}

// TestMain_HelpShortFlag tests that -h flag returns exit code 0.
func TestMain_HelpShortFlag(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set -h flag
	os.Args = []string{"fluxid", "-h"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 0
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h flag, got %d", exitCode)
	}
}

// TestMain_DryRunMode tests that dry-run mode executes successfully.
//
//nolint:dupl // Test functions intentionally similar for different test scenarios
func TestMain_DryRunMode(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create config with commands section using absolute paths
	configDir := tmpDir + "/.fluxid"
	commandsDir := configDir + "/commands"
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create command files
	for _, file := range []string{"implement.md", "review.md", "commit.md"} {
		if err := os.WriteFile(commandsDir+"/"+file, []byte("# Command"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Write config with absolute paths
	configContent := fmt.Sprintf(`agent: claude
commands:
  implement: %s/commands/implement.md
  review: %s/commands/review.md
  commit: %s/commands/commit.md
`, configDir, configDir, configDir)
	if err := os.WriteFile(configDir+"/config.yaml", []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set dry-run mode with valid agent
	os.Args = []string{"fluxid", "--fluxid-dry-run", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 0
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run mode, got %d", exitCode)
	}
}

// TestMain_SuccessfulExecutionWithIterations tests successful execution with custom iterations.
func TestMain_SuccessfulExecutionWithIterations(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create config with commands section using absolute paths
	configDir := tmpDir + "/.fluxid"
	commandsDir := configDir + "/commands"
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create command files
	for _, file := range []string{"implement.md", "review.md", "commit.md"} {
		if err := os.WriteFile(commandsDir+"/"+file, []byte("# Command"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Write config with absolute paths
	configContent := fmt.Sprintf(`agent: claude
commands:
  implement: %s/commands/implement.md
  review: %s/commands/review.md
  commit: %s/commands/commit.md
`, configDir, configDir, configDir)
	if err := os.WriteFile(configDir+"/config.yaml", []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set dry-run mode with custom iterations using equals syntax
	os.Args = []string{"fluxid", "--fluxid-dry-run", "--claude", "--fluxid-iterations=3"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 0
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for successful execution with iterations, got %d", exitCode)
	}
}
