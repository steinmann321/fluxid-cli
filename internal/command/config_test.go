//nolint:paralleltest,exhaustruct,godot // Test file with structs and literals
package command

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/output"
	"testing"
)

func TestBuildFinalConfigWithInvalidOutputFormat(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommitEnabled:    true,
		CommandFiles:     &config.ResolvedCommandFiles{},
		Sources:          map[string]string{},
	}

	invalidFormat := "invalid-format"
	args := &CLIArgs{
		AgentArgs:           []string{},
		CLIOutputFormat:     &invalidFormat,
		CLIDryRun:           boolPtr(false),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
	}

	_, err := buildFinalConfig(resolved, args)
	if err == nil {
		t.Error("Expected error for invalid output format, got nil")
	}
}

func TestBuildFinalConfigWithDryRun(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommitEnabled:    true,
		CommandFiles:     &config.ResolvedCommandFiles{},
		Sources:          map[string]string{},
	}

	args := &CLIArgs{
		AgentArgs:           []string{"--model", "test"},
		CLIOutputFormat:     nil,
		CLIDryRun:           boolPtr(true),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if !cfg.DryRun {
		t.Error("Expected DryRun to be true")
	}
}

func TestBuildFinalConfigWithSessionID(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommitEnabled:    true,
		CommandFiles:     &config.ResolvedCommandFiles{},
		Sources:          map[string]string{},
	}

	args := &CLIArgs{
		AgentArgs:           []string{},
		CLIOutputFormat:     nil,
		CLIDryRun:           boolPtr(false),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
	}

	// Test with FLUXID_SESSION_ID set
	t.Setenv("FLUXID_SESSION_ID", "test-session-123")

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if cfg.SessionID != "test-session-123" {
		t.Errorf("Expected SessionID to be test-session-123, got %s", cfg.SessionID)
	}
}

func TestBuildFinalConfigWithJSONFormat(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommitEnabled:    true,
		CommandFiles:     &config.ResolvedCommandFiles{},
		Sources:          map[string]string{},
	}

	jsonFormat := testFormatJSON
	args := &CLIArgs{
		AgentArgs:           []string{},
		CLIOutputFormat:     &jsonFormat,
		CLIDryRun:           boolPtr(false),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if cfg.OutputFormat != output.FormatJSON {
		t.Errorf("Expected OutputFormat to be JSON, got %v", cfg.OutputFormat)
	}
}

func TestBuildFinalConfigWithYAMLFormat(t *testing.T) {
	resolved := &config.ResolvedConfig{
		Agent:            "claude",
		Iterations:       20,
		ImplementRetries: 3,
		CommitEnabled:    true,
		CommandFiles:     &config.ResolvedCommandFiles{},
		Sources:          map[string]string{},
	}

	yamlFormat := "yaml"
	args := &CLIArgs{
		AgentArgs:           []string{},
		CLIOutputFormat:     &yamlFormat,
		CLIDryRun:           boolPtr(false),
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if cfg.OutputFormat != output.FormatYAML {
		t.Errorf("Expected OutputFormat to be YAML, got %v", cfg.OutputFormat)
	}
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
