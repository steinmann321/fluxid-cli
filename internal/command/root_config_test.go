package command

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecute_UnsupportedAgent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config with unsupported agent
	configDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configContent := "agent: unsupported-agent-xyz\n"
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Run with dry-run, but agent comes from config
	os.Args = []string{"fluxid", "--dry-run"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for unsupported agent")
	}
}

func TestExecute_ProjectConfigLoadError(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)
	t.Setenv("HOME", dataDir)

	// Create invalid project config (malformed YAML)
	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dataDir, ".fluxid", "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: [yaml content"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "echo"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for project config load error")
	}
}

// TestExecute_AgentFromEnv removed - environment variable support removed in v2.0

func TestExecute_ConfigPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create home config with valid supported agent and commands
	configDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configContent := `agent: claude
max_review_cycles: 5
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

	os.Args = []string{"fluxid", "--dry-run"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 with config file, got %d", exitCode)
	}
}

//nolint:dupl // Test functions intentionally similar for different output formats
func TestExecute_OutputFormatJSON(t *testing.T) {
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
	for _, file := range []string{"implement.md", "review.md", "commit.md"} {
		if err := os.WriteFile(filepath.Join(configDir, file), []byte("# Command"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "--output=json", "claude", "test"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --output=json, got %d", exitCode)
	}
}

//nolint:dupl // Test functions intentionally similar for different output formats
func TestExecute_OutputFormatYAML(t *testing.T) {
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
	for _, file := range []string{"implement.md", "review.md", "commit.md"} {
		if err := os.WriteFile(filepath.Join(configDir, file), []byte("# Command"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "--output=yaml", "claude", "test"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --output=yaml, got %d", exitCode)
	}
}

func TestExecute_InvalidOutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "--output=invalid-format", "echo", "test"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid output format")
	}
}

// TestExecute_EnvConfigError removed - environment variable support removed in v2.0

func TestExecute_HomeConfigLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create invalid home config (malformed YAML)
	homeConfigDir := filepath.Join(tmpDir, ".fluxid")
	if err := os.MkdirAll(homeConfigDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(homeConfigDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: {yaml: [content"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--dry-run", "echo"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for home config load error")
	}
}

func TestExecute_CustomSessionID(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	customSessionID := "custom-session-id-123"
	t.Setenv("FLUXID_SESSION_ID", customSessionID)

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

	os.Args = []string{"fluxid", "--dry-run", "echo", "test"}

	exitCode := Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 with custom session ID, got %d", exitCode)
	}
}
