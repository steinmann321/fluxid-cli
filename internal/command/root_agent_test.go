package command

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/goleak"
)

func TestExecute_InvalidAgent(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("PATH", "") // Empty PATH to ensure agent not found

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
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("PATH", "") // Empty PATH to ensure no agent is found

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
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
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
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Reset signal count for clean test state
	signalCount.Store(0)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--fluxid-iterations", "not-a-number"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for parse args error")
	}
}

func TestExecute_WorkflowExecutionWithEchoAgent(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create config with commands section
	configDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configContent := `agent: claude
commands:
  implement: implement.md
  review: review.md
  commit: commit.md
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create command files
	if err := os.WriteFile(filepath.Join(configDir, "implement.md"), []byte("# Implement"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "review.md"), []byte("# Review"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "commit.md"), []byte("# Commit"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "echo"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run with echo, got %d", exitCode)
	}
}

func TestExecute_ValidateAgentNotExecutable(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
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
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "--output", "invalid-format"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for build final config error")
	}
}

func TestExecute_AgentStatError(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create config with commands section
	configDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configContent := `agent: claude
commands:
  implement: implement.md
  review: review.md
  commit: commit.md
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create command files
	if err := os.WriteFile(filepath.Join(configDir, "implement.md"), []byte("# Implement"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "review.md"), []byte("# Review"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "commit.md"), []byte("# Commit"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create an executable agent in PATH
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	agentPath := filepath.Join(binDir, "claude")
	if err := os.WriteFile(agentPath, []byte("#!/bin/sh\necho test"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", binDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// First verify agent is executable
	os.Args = []string{"fluxid", "--dry-run", "claude"}
	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for executable agent in dry-run, got %d", exitCode)
	}
}

func TestExecute_NoAgentSpecified(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
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
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "--output", "invalid-format", "echo"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid output format")
	}
}

func TestExecute_AgentNotInPathError(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
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
