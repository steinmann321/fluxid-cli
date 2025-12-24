package command

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"os"
	"path/filepath"
	"testing"
)

// Test constants.
const (
	testAgentClaude = "claude"
	testFormatJSON  = "json"
)

func TestHandleSpecialCommands_WriteHistory(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataDir)
	t.Setenv("FLUXID_SESSION_ID", "test-write-history")

	if err := os.MkdirAll(filepath.Join(dataDir, ".fluxid"), 0o755); err != nil {
		t.Fatal(err)
	}

	// This would normally be called via main, but we can test the path exists
	// by verifying the IPC write history functionality works
	err := ipc.WriteHistoryEntry("test-write-history", "Test message")
	if err != nil {
		t.Errorf("Expected no error writing history, got: %v", err)
	}
}

//nolint:paralleltest // Uses t.Setenv indirectly
func TestBuildFinalConfig_DefaultValues(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            testAgentClaude,
		Iterations:       3,
		ImplementRetries: 2,
		CommitEnabled:    true,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "",
			ReviewPath:    "",
			CommitPath:    "",
		},
		Sources: map[string]string{},
	}

	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           []string{"test", "args"},
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if cfg.Agent != testAgentClaude {
		t.Errorf("Expected agent '%s', got '%s'", testAgentClaude, cfg.Agent)
	}
	if cfg.MaxReviewCycles != 3 {
		t.Errorf("Expected 3 review cycles, got %d", cfg.MaxReviewCycles)
	}
}

func TestBuildFinalConfig_WithSessionID(t *testing.T) {
	t.Setenv("FLUXID_SESSION_ID", "test-uuid-12345")

	resolved := &config.ResolvedConfig{
		Agent:            testAgentClaude,
		Iterations:       1,
		ImplementRetries: 1,
		CommitEnabled:    false,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "",
			ReviewPath:    "",
			CommitPath:    "",
		},
		Sources: map[string]string{},
	}

	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           []string{},
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if cfg.SessionID != "test-uuid-12345" {
		t.Errorf("Expected SessionID 'test-uuid-12345', got '%s'", cfg.SessionID)
	}
}

//nolint:paralleltest // Uses t.Setenv indirectly
func TestBuildFinalConfig_WithOutputFormat(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            testAgentClaude,
		Iterations:       1,
		ImplementRetries: 1,
		CommitEnabled:    false,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "",
			ReviewPath:    "",
			CommitPath:    "",
		},
		Sources: map[string]string{},
	}

	outputFormat := testFormatJSON
	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           []string{},
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     &outputFormat,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if cfg.OutputFormat != output.FormatJSON {
		t.Errorf("Expected OutputFormat JSON, got %v", cfg.OutputFormat)
	}
}

//nolint:paralleltest // Uses t.Setenv indirectly
func TestBuildFinalConfig_DryRunOverride(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            testAgentClaude,
		Iterations:       3,
		ImplementRetries: 2,
		CommitEnabled:    true,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "",
			ReviewPath:    "",
			CommitPath:    "",
		},
		Sources: map[string]string{},
	}

	dryRun := true
	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           []string{},
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           &dryRun,
		CLIOutputFormat:     nil,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !cfg.DryRun {
		t.Error("Expected DryRun to be true")
	}
	if cfg.MaxReviewCycles != 3 {
		t.Errorf("Expected MaxReviewCycles to be 3 (from resolved), got %d", cfg.MaxReviewCycles)
	}
}

func TestParseFluxidFlag_CommitEnabled(t *testing.T) {
	t.Parallel()
	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           nil,
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
	}
	skip, handled, err := parseFluxidFlag("--fluxid-commit-enabled", 0, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !handled {
		t.Error("Expected flag to be handled")
	}
	if skip != 0 {
		t.Errorf("Expected skip=0, got %d", skip)
	}
	if args.CLICommitEnabled == nil || !*args.CLICommitEnabled {
		t.Error("Expected commit enabled to be true")
	}
}

func TestParseFluxidFlag_NoCommit(t *testing.T) {
	t.Parallel()
	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           nil,
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
	}
	skip, handled, err := parseFluxidFlag("--fluxid-no-commit", 0, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !handled {
		t.Error("Expected flag to be handled")
	}
	if skip != 0 {
		t.Errorf("Expected skip=0, got %d", skip)
	}
	if args.CLICommitEnabled == nil || *args.CLICommitEnabled {
		t.Error("Expected commit enabled to be false")
	}
}

func TestParseFluxidFlag_DryRun(t *testing.T) {
	t.Parallel()
	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           nil,
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
	}
	skip, handled, err := parseFluxidFlag("--fluxid-dry-run", 0, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !handled {
		t.Error("Expected flag to be handled")
	}
	if skip != 0 {
		t.Errorf("Expected skip=0, got %d", skip)
	}
	if args.CLIDryRun == nil || !*args.CLIDryRun {
		t.Error("Expected dry run to be true")
	}
}

func TestParseFluxidFlag_UnknownFlag(t *testing.T) {
	t.Parallel()
	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           nil,
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
	}
	skip, handled, err := parseFluxidFlag("--fluxid-unknown", 0, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if handled {
		t.Error("Expected flag not to be handled")
	}
	if skip != 0 {
		t.Errorf("Expected skip=0, got %d", skip)
	}
}

func TestHandleSpecialCommands_HelpFlag(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Set os.Args to include --help
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"fluxid", "--help"}

	exitCode, handled := handleSpecialCommands()
	if !handled {
		t.Error("Expected help command to be handled")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help, got %d", exitCode)
	}
}

//nolint:paralleltest // Uses t.Setenv indirectly
func TestBuildFinalConfig_UsesResolvedValues(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            testAgentClaude,
		Iterations:       3,
		ImplementRetries: 2,
		CommitEnabled:    true,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "",
			ReviewPath:    "",
			CommitPath:    "",
		},
		Sources: map[string]string{},
	}

	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           []string{},
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if cfg.MaxReviewCycles != 3 {
		t.Errorf("Expected 3 review cycles from resolved, got %d", cfg.MaxReviewCycles)
	}
	if cfg.MaxImplementRetries != 2 {
		t.Errorf("Expected 2 implement retries from resolved, got %d", cfg.MaxImplementRetries)
	}
	if !cfg.CommitEnabled {
		t.Error("Expected CommitEnabled to be true from resolved")
	}
}

func TestValidateOutputFormat_Valid(t *testing.T) {
	t.Parallel()
	formats := []string{"text", "json", "yaml"}
	for _, format := range formats {
		err := output.ValidateFormat(format)
		if err != nil {
			t.Errorf("Expected no error for format '%s', got: %v", format, err)
		}
	}
}

func TestValidateOutputFormat_Invalid(t *testing.T) {
	t.Parallel()
	err := output.ValidateFormat("xml")
	if err == nil {
		t.Error("Expected error for invalid format 'xml'")
	}
}

//nolint:paralleltest // Uses t.Setenv indirectly
func TestBuildFinalConfig_InvalidOutputFormat(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            testAgentClaude,
		Iterations:       1,
		ImplementRetries: 1,
		CommitEnabled:    false,
		CommandFiles: &config.ResolvedCommandFiles{
			ImplementPath: "",
			ReviewPath:    "",
			CommitPath:    "",
		},
		Sources: map[string]string{},
	}

	invalidFormat := "xml"
	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           []string{},
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     &invalidFormat,
	}

	_, err := buildFinalConfig(resolved, args)
	if err == nil {
		t.Error("Expected error for invalid output format")
	}
}
