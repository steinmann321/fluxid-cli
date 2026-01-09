package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/types"
	"path/filepath"
	"testing"
)

// TestBuildAgentCommandClaude tests Claude agent command construction.
func TestBuildAgentCommandClaude(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		Agent:     "claude",
		AgentArgs: []string{"--extra-arg"},
	}

	cmd := buildAgentCommand(cfg, "test prompt")

	// Verify command name
	if filepath.Base(cmd.Path) != "claude" {
		t.Errorf("Expected command 'claude', got %q", filepath.Base(cmd.Path))
	}

	// Verify args structure
	args := cmd.Args[1:] // Skip command name
	expectedArgs := []string{
		"--dangerously-skip-permissions",
		"--output-format",
		"stream-json",
		"--verbose",
		"-p",
		"test prompt",
		"--extra-arg",
	}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, args[i])
		}
	}
}

// TestBuildAgentCommandCodex tests Codex agent command construction.
func TestBuildAgentCommandCodex(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		Agent:     "codex",
		AgentArgs: []string{"--extra-arg"},
	}

	cmd := buildAgentCommand(cfg, "test prompt")

	// Verify command name
	if filepath.Base(cmd.Path) != "codex" {
		t.Errorf("Expected command 'codex', got %q", filepath.Base(cmd.Path))
	}

	// Verify args structure
	args := cmd.Args[1:] // Skip command name
	expectedArgs := []string{
		"exec",
		"--json",
		"test prompt",
		"--extra-arg",
	}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, args[i])
		}
	}
}

// TestBuildAgentCommandOpencode tests Opencode agent command construction.
func TestBuildAgentCommandOpencode(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		Agent:     "opencode",
		AgentArgs: []string{"--extra-arg"},
	}

	cmd := buildAgentCommand(cfg, "test prompt")

	// Verify command name
	if filepath.Base(cmd.Path) != "opencode" {
		t.Errorf("Expected command 'opencode', got %q", filepath.Base(cmd.Path))
	}

	// Verify args structure
	args := cmd.Args[1:] // Skip command name
	expectedArgs := []string{
		"run",
		"--format",
		"json",
		"test prompt",
		"--extra-arg",
	}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, args[i])
		}
	}
}

// TestBuildAgentCommandGemini tests Gemini agent command construction.
func TestBuildAgentCommandGemini(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		Agent:     "gemini",
		AgentArgs: []string{"--extra-arg"},
	}

	cmd := buildAgentCommand(cfg, "test prompt")

	// Verify command name
	if filepath.Base(cmd.Path) != "gemini" {
		t.Errorf("Expected command 'gemini', got %q", filepath.Base(cmd.Path))
	}

	// Verify args structure
	args := cmd.Args[1:] // Skip command name
	expectedArgs := []string{
		"--output-format",
		"stream-json",
		"test prompt",
		"--extra-arg",
	}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, args[i])
		}
	}
}

// TestBuildAgentCommandUnknown tests default (Claude-style) command for unknown agent.
func TestBuildAgentCommandUnknown(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		Agent:     "unknown-agent",
		AgentArgs: []string{},
	}

	cmd := buildAgentCommand(cfg, "test prompt")

	// Verify command name matches agent
	if filepath.Base(cmd.Path) != "unknown-agent" {
		t.Errorf("Expected command 'unknown-agent', got %q", filepath.Base(cmd.Path))
	}

	// Verify args use Claude-style defaults
	args := cmd.Args[1:] // Skip command name
	expectedArgs := []string{
		"--dangerously-skip-permissions",
		"--output-format",
		"stream-json",
		"--verbose",
		"-p",
		"test prompt",
	}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, args[i])
		}
	}
}

// TestComposePromptWithBuiltIn tests prompt composition with built-in command.
func TestComposePromptWithBuiltIn(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		TaskFilePath: "/path/to/task.yaml",
		CommandFiles: nil, // Built-in command
	}

	result := composePrompt(cfg, "implement", "base prompt")

	expected := "base prompt\nCommand file: <built-in>\nTask file: /path/to/task.yaml"
	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestGetCommandFilePathBuiltIn tests command file resolution with built-in.
func TestGetCommandFilePathBuiltIn(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		CommandFiles: nil,
	}

	result := getCommandFilePath(cfg, "implement")
	if result != builtInPrompt {
		t.Errorf("Expected %q, got %q", builtInPrompt, result)
	}
}

// TestGetCommandFilePathImplement tests command file resolution for implement phase.
func TestGetCommandFilePathImplement(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/path/to/implement.md",
		},
	}

	result := getCommandFilePath(cfg, "implement")
	if result != "/path/to/implement.md" {
		t.Errorf("Expected '/path/to/implement.md', got %q", result)
	}
}

// TestGetCommandFilePathReview tests command file resolution for review phase.
func TestGetCommandFilePathReview(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		CommandFiles: &config.ResolvedCommandFiles{
			ReviewPath: "/path/to/review.md",
		},
	}

	result := getCommandFilePath(cfg, "review")
	if result != "/path/to/review.md" {
		t.Errorf("Expected '/path/to/review.md', got %q", result)
	}
}

// TestGetCommandFilePathCommit tests command file resolution for commit phase.
func TestGetCommandFilePathCommit(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		CommandFiles: &config.ResolvedCommandFiles{
			CommitPath: "/path/to/commit.md",
		},
	}

	result := getCommandFilePath(cfg, "commit")
	if result != "/path/to/commit.md" {
		t.Errorf("Expected '/path/to/commit.md', got %q", result)
	}
}

// TestGetCommandFilePathUnknownPhase tests fallback for unknown phase.
func TestGetCommandFilePathUnknownPhase(t *testing.T) {
	t.Parallel()

	//nolint:exhaustruct // Test config with only relevant fields
	cfg := types.Config{
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "/path/to/implement.md",
		},
	}

	result := getCommandFilePath(cfg, "unknown-phase")
	if result != builtInPrompt {
		t.Errorf("Expected %q, got %q", builtInPrompt, result)
	}
}
