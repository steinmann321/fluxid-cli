package command

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/goleak"
)

//nolint:paralleltest // Cannot run in parallel - modifies global os.Args
func TestExecute_Help(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to request help
	os.Args = []string{"fluxid", "--help"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global os.Args
func TestExecute_HelpShort(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to request help with -h
	os.Args = []string{"fluxid", "-h"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}
}

func TestExecute_DryRunSuccess(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args for dry-run mode
	os.Args = []string{"fluxid", "--dry-run", "echo", "test"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run, got %d", exitCode)
	}
}

func TestExecute_InvalidAgent(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("PATH", "") // Empty PATH to ensure agent not found

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args with non-existent agent (not in dry-run mode)
	os.Args = []string{"fluxid", "nonexistent-agent-xyz"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid agent")
	}
}

func TestExecute_MissingAgentArgument(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("PATH", "") // Empty PATH to ensure no agent is found

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args with --claude but claude won't be in PATH (should fail fast)
	os.Args = []string{"fluxid", "--claude"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when agent is missing from PATH")
	}
}

func TestExecute_InvalidConfigHomeError(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	// Test that invalid home config path returns error
	// Create a file where a directory is expected
	tmpDir := t.TempDir()
	invalidPath := tmpDir + "/file-not-dir"
	if err := os.WriteFile(invalidPath, []byte("test"), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Override HOME to point to the file
	t.Setenv("HOME", invalidPath)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--claude"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for config load error")
	}
}

func TestExecute_ParseArgsError(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Use invalid flag to trigger parse error
	os.Args = []string{"fluxid", "--invalid-flag-that-does-not-exist"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for parse args error")
	}
}

func TestExecute_WorkflowExecutionWithEchoAgent(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Use echo which exists in PATH
	os.Args = []string{"fluxid", "--dry-run", "echo"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run with echo, got %d", exitCode)
	}
}

func TestExecute_ValidateAgentNotExecutable(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create a non-executable file in PATH
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	agentPath := filepath.Join(binDir, "fake-agent")
	if err := os.WriteFile(agentPath, []byte("#!/bin/sh"), 0o644); err != nil { // Not executable
		t.Fatal(err)
	}

	t.Setenv("PATH", binDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "fake-agent"} // NOT in dry-run mode

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for non-executable agent")
	}
}

func TestExecute_BuildFinalConfigError(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Provide agent args but no agent - this should trigger build config error
	os.Args = []string{"fluxid", "--"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for build final config error")
	}
}

func TestExecute_AgentStatError(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create an executable agent in PATH
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	agentPath := filepath.Join(binDir, "fake-agent")
	if err := os.WriteFile(agentPath, []byte("#!/bin/sh\necho test"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", binDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// First verify agent is executable
	os.Args = []string{"fluxid", "--dry-run", "fake-agent"}
	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for executable agent in dry-run, got %d", exitCode)
	}
}

func TestExecute_NoAgentSpecified(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("PATH", "") // Empty PATH

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// No agent specified, no config, should fail
	os.Args = []string{"fluxid"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code when no agent specified")
	}
}

func TestExecute_InvalidOutputFormatBuildConfig(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Use invalid output format to trigger buildFinalConfig error
	os.Args = []string{"fluxid", "--dry-run", "--output", "invalid-format", "echo"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid output format")
	}
}

func TestExecute_AgentNotInPathError(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("PATH", tmpDir) // Empty directory with no binaries

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test with an agent that doesn't exist in PATH
	os.Args = []string{"fluxid", "nonexistent-agent-xyz"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for agent not in PATH")
	}
}

func TestExecute_LoadAllConfigsWithMultipleSources(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create a valid project config
	projectConfigDir := filepath.Join(tmpDir, ".config", "fluxid")
	if err := os.MkdirAll(projectConfigDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write valid YAML with agent setting
	configFile := filepath.Join(projectConfigDir, "config.yaml")
	validYAML := "agent: echo\nmax_review_cycles: 5\n"
	if err := os.WriteFile(configFile, []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "echo"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for valid config, got %d", exitCode)
	}
}

func TestExecute_BuildInitStatusWithEnvConfig(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Cleanup(cleanupAllSignalHandlers)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("FLUXID_AGENT", "echo")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "echo"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for valid env config, got %d", exitCode)
	}
}
