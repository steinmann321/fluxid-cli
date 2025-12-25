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

	// Set args for dry-run mode
	os.Args = []string{"fluxid", "--dry-run", "echo", "test"}

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
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write valid YAML with agent setting and commands
	configFile := filepath.Join(configDir, "config.yaml")
	validYAML := `agent: claude
max_review_cycles: 5
commands:
  implement: implement.md
  review: review.md
  commit: commit.md
`
	if err := os.WriteFile(configFile, []byte(validYAML), 0o644); err != nil {
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

	os.Args = []string{"fluxid", "--dry-run", "claude"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for valid config, got %d", exitCode)
	}
}

// TestExecute_BuildInitStatusWithEnvConfig removed - environment variable support removed in v2.0
