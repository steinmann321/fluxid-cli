//nolint:paralleltest,exhaustruct,godot // Test file with structs and literals
package main

import (
	"fluxid-loop/internal/config"
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
	args := &cliArgs{
		agentArgs:           []string{},
		cliOutputFormat:     &invalidFormat,
		cliDryRun:           boolPtr(false),
		cliIterations:       nil,
		cliImplementRetries: nil,
		cliCommitEnabled:    nil,
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

	args := &cliArgs{
		agentArgs:           []string{"--model", "test"},
		cliOutputFormat:     nil,
		cliDryRun:           boolPtr(true),
		cliIterations:       nil,
		cliImplementRetries: nil,
		cliCommitEnabled:    nil,
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

	args := &cliArgs{
		agentArgs:           []string{},
		cliOutputFormat:     nil,
		cliDryRun:           boolPtr(false),
		cliIterations:       nil,
		cliImplementRetries: nil,
		cliCommitEnabled:    nil,
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

	jsonFormat := "json"
	args := &cliArgs{
		agentArgs:           []string{},
		cliOutputFormat:     &jsonFormat,
		cliDryRun:           boolPtr(false),
		cliIterations:       nil,
		cliImplementRetries: nil,
		cliCommitEnabled:    nil,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if cfg.OutputFormat != OutputFormatJSON {
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
	args := &cliArgs{
		agentArgs:           []string{},
		cliOutputFormat:     &yamlFormat,
		cliDryRun:           boolPtr(false),
		cliIterations:       nil,
		cliImplementRetries: nil,
		cliCommitEnabled:    nil,
	}

	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		t.Fatalf("buildFinalConfig failed: %v", err)
	}

	if cfg.OutputFormat != OutputFormatYAML {
		t.Errorf("Expected OutputFormat to be YAML, got %v", cfg.OutputFormat)
	}
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
