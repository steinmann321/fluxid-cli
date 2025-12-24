package main

import (
	"os"
	"path/filepath"
	"testing"
)

// ERROR PATH TESTS - These tests verify error handling and exit code propagation
// The 2:1 error-to-success ratio ensures comprehensive coverage of failure scenarios

// TestMain_ConfigLoadingError tests that malformed config files return exit code 1.
func TestMain_ConfigLoadingError(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up temporary home directory with invalid YAML config
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create config directory with malformed YAML that will fail to parse
	configDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	malformedYAML := "agent: [unclosed bracket\niterations: not valid yaml"
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(malformedYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set valid arguments
	os.Args = []string{"fluxid", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for config loading error, got %d", exitCode)
	}
}

// TestMain_ArgParsingError_NegativeIterations tests that negative iteration values return exit code 1.
func TestMain_ArgParsingError_NegativeIterations(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up valid environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set invalid arguments (negative iterations)
	os.Args = []string{"fluxid", "--fluxid-iterations", "-5", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for negative iterations argument, got %d", exitCode)
	}
}

// TestMain_ArgParsingError_InvalidIterations tests that non-numeric iteration values return exit code 1.
func TestMain_ArgParsingError_InvalidIterations(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up valid environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set invalid arguments (non-numeric iterations)
	os.Args = []string{"fluxid", "--fluxid-iterations", "not-a-number", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid iterations argument, got %d", exitCode)
	}
}

// TestMain_ArgParsingError_ZeroIterations tests that zero iteration values return exit code 1.
func TestMain_ArgParsingError_ZeroIterations(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up valid environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set invalid arguments (zero iterations)
	os.Args = []string{"fluxid", "--fluxid-iterations", "0", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for zero iterations argument, got %d", exitCode)
	}
}

// TestMain_ArgParsingError_MissingIterationsValue tests that missing iteration values return exit code 1.
func TestMain_ArgParsingError_MissingIterationsValue(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up valid environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set invalid arguments (missing iterations value)
	os.Args = []string{"fluxid", "--fluxid-iterations"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing iterations value, got %d", exitCode)
	}
}

// TestMain_ArgParsingError_NegativeImplementRetries tests that negative implement retry values return exit code 1.
func TestMain_ArgParsingError_NegativeImplementRetries(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up valid environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set invalid arguments (negative implement retries)
	os.Args = []string{"fluxid", "--fluxid-implement-retries", "-3", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for negative implement retries argument, got %d", exitCode)
	}
}

// TestMain_ArgParsingError_InvalidImplementRetries tests that non-numeric implement retry values return exit code 1.
func TestMain_ArgParsingError_InvalidImplementRetries(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up valid environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set invalid arguments (non-numeric implement retries)
	os.Args = []string{"fluxid", "--fluxid-implement-retries", "invalid", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid implement retries argument, got %d", exitCode)
	}
}

// TestMain_ArgParsingError_MissingOutputFormatValue tests that missing output format values return exit code 1.
func TestMain_ArgParsingError_MissingOutputFormatValue(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up valid environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set invalid arguments (missing output format value)
	os.Args = []string{"fluxid", "--fluxid-output"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing output format value, got %d", exitCode)
	}
}

// TestMain_AgentValidationError_UnsupportedAgent tests that unsupported agent names return exit code 1.
func TestMain_AgentValidationError_UnsupportedAgent(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up environment with unsupported agent via env config
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create home config with unsupported agent name
	configDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	invalidConfig := "agent: invalid-agent-xyz\n"
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(invalidConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// No agent flag provided, will use config agent (which is invalid)
	os.Args = []string{"fluxid"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for unsupported agent name, got %d", exitCode)
	}
}

// TestMain_AgentValidationError_AgentNotInPath tests that agents not in PATH return exit code 1.
func TestMain_AgentValidationError_AgentNotInPath(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up valid environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Set PATH to empty to ensure agent is not found
	t.Setenv("PATH", "")

	// Set valid agent name but it won't be in PATH
	os.Args = []string{"fluxid", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 when agent not in PATH, got %d", exitCode)
	}
}

// TestMain_AgentValidationError_AgentNotExecutable tests that non-executable agents return exit code 1.
func TestMain_AgentValidationError_AgentNotExecutable(t *testing.T) {
	// Save original os.Exit and restore after test
	originalExit := osExit
	osExit = mockExit
	defer func() { osExit = originalExit }()

	// Set up valid environment
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create a non-executable agent binary
	agentDir := t.TempDir()
	agentPath := agentDir + "/claude"
	if err := os.WriteFile(agentPath, []byte("#!/bin/sh\necho test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set PATH to include the non-executable agent
	t.Setenv("PATH", agentDir)

	// Set agent
	os.Args = []string{"fluxid", "--claude"}

	// Reset exit code
	exitCode = -1

	// Execute main
	main()

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for non-executable agent, got %d", exitCode)
	}
}
