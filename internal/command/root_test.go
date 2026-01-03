package command

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/goleak"
)

//nolint:paralleltest // Cannot run in parallel - modifies global os.Args
func TestExecute_Help(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--help"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global os.Args
func TestExecute_HelpShort(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "-h"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for -h, got %d", exitCode)
	}
}

func TestExecute_DryRunSuccess(t *testing.T) {
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
	commandsDir := filepath.Join(configDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configContent := fmt.Sprintf(`agent: claude
commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, commandsDir, commandsDir, commandsDir)
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create command files
	if err := os.WriteFile(filepath.Join(commandsDir, "implement.md"), []byte("# Implement"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "review.md"), []byte("# Review"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "commit.md"), []byte("# Commit"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args for dry-run mode
	// Create dummy task file
	taskPath := filepath.Join(tmpDir, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatal(err)
	}
	os.Args = []string{"fluxid", "--dry-run", "echo", "test", "--file=" + taskPath}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run, got %d", exitCode)
	}
}

func TestExecute_LoadAllConfigsWithMultipleSources(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer func() {
		cleanupAllSignalHandlers()
		signalCount.Store(0)
	}()
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create a valid home config
	configDir := filepath.Join(tmpDir, ".fluxid")
	commandsDir := filepath.Join(configDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create command files
	if err := os.WriteFile(filepath.Join(commandsDir, "implement.md"), []byte("# Implement"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "review.md"), []byte("# Review"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "commit.md"), []byte("# Commit"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write valid YAML with agent setting and commands (absolute paths)
	configFile := filepath.Join(configDir, "config.yaml")
	validYAML := fmt.Sprintf(`agent: claude
max_review_cycles: 5
commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, commandsDir, commandsDir, commandsDir)
	if err := os.WriteFile(configFile, []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create dummy task file
	taskPath := filepath.Join(tmpDir, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatal(err)
	}
	os.Args = []string{"fluxid", "--dry-run", "claude", "--file=" + taskPath}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for valid config, got %d", exitCode)
	}
}

// TestExecute_BuildInitStatusWithEnvConfig removed - environment variable support removed in v2.0
