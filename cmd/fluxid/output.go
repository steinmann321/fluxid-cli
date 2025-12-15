// Package main implements the fluxid CLI workflow controller for coding agents.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// OutputFormat represents the supported output formats.
type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
	OutputFormatYAML OutputFormat = "yaml"
)

// ValidateOutputFormat validates the output format string.
func ValidateOutputFormat(format string) error {
	switch format {
	case "text", "json", "yaml":
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s (supported: text, json, yaml)", format)
	}
}

// InitializationStatus represents the initialization status output.
type InitializationStatus struct {
	SessionID              string            `json:"session_id" yaml:"session_id"`
	Agent                  string            `json:"agent" yaml:"agent"`
	AgentSource            string            `json:"agent_source" yaml:"agent_source"`
	MaxReviewCycles        int               `json:"max_review_cycles" yaml:"max_review_cycles"`
	ReviewCyclesSource     string            `json:"review_cycles_source" yaml:"review_cycles_source"`
	MaxImplementRetries    int               `json:"max_implement_retries" yaml:"max_implement_retries"`
	ImplementRetriesSource string            `json:"implement_retries_source" yaml:"implement_retries_source"`
	CommitEnabled          bool              `json:"commit_enabled" yaml:"commit_enabled"`
	CommitEnabledSource    string            `json:"commit_enabled_source" yaml:"commit_enabled_source"`
	CommandFiles           *CommandFilesJSON `json:"command_files,omitempty" yaml:"command_files,omitempty"`
	AgentArgs              []string          `json:"agent_args,omitempty" yaml:"agent_args,omitempty"`
}

// CommandFilesJSON represents command file paths in JSON/YAML output.
type CommandFilesJSON struct {
	Implement string `json:"implement" yaml:"implement"`
	Review    string `json:"review" yaml:"review"`
	Commit    string `json:"commit" yaml:"commit"`
}

// PrintInitializationStatusJSON outputs initialization status as JSON to stdout.
func PrintInitializationStatusJSON(cfg Config) error {
	status := InitializationStatus{
		SessionID:              cfg.SessionID,
		Agent:                  cfg.Agent,
		AgentSource:            cfg.Sources["agent"],
		MaxReviewCycles:        cfg.MaxReviewCycles,
		ReviewCyclesSource:     cfg.Sources["iterations"],
		MaxImplementRetries:    cfg.MaxImplementRetries,
		ImplementRetriesSource: cfg.Sources["implement_retries"],
		CommitEnabled:          cfg.CommitEnabled,
		CommitEnabledSource:    cfg.Sources["commit_enabled"],
		CommandFiles:           nil,
		AgentArgs:              nil,
	}

	if cfg.CommandFiles != nil {
		status.CommandFiles = &CommandFilesJSON{
			Implement: cfg.CommandFiles.ImplementPath,
			Review:    cfg.CommandFiles.ReviewPath,
			Commit:    cfg.CommandFiles.CommitPath,
		}
	}

	if len(cfg.AgentArgs) > 0 {
		status.AgentArgs = cfg.AgentArgs
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(status); err != nil {
		return fmt.Errorf("failed to encode initialization status: %w", err)
	}
	return nil
}

// PrintInitializationStatusYAML outputs initialization status as YAML to stdout.
func PrintInitializationStatusYAML(cfg Config) error {
	status := InitializationStatus{
		SessionID:              cfg.SessionID,
		Agent:                  cfg.Agent,
		AgentSource:            cfg.Sources["agent"],
		MaxReviewCycles:        cfg.MaxReviewCycles,
		ReviewCyclesSource:     cfg.Sources["iterations"],
		MaxImplementRetries:    cfg.MaxImplementRetries,
		ImplementRetriesSource: cfg.Sources["implement_retries"],
		CommitEnabled:          cfg.CommitEnabled,
		CommitEnabledSource:    cfg.Sources["commit_enabled"],
		CommandFiles:           nil,
		AgentArgs:              nil,
	}

	if cfg.CommandFiles != nil {
		status.CommandFiles = &CommandFilesJSON{
			Implement: cfg.CommandFiles.ImplementPath,
			Review:    cfg.CommandFiles.ReviewPath,
			Commit:    cfg.CommandFiles.CommitPath,
		}
	}

	if len(cfg.AgentArgs) > 0 {
		status.AgentArgs = cfg.AgentArgs
	}

	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	if err := encoder.Encode(status); err != nil {
		return fmt.Errorf("failed to encode initialization status: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close YAML encoder: %w", err)
	}
	return nil
}

// PrintInitializationStatusText outputs initialization status as human-readable text to stdout.
func PrintInitializationStatusText(cfg Config) {
	_, _ = fmt.Fprintf(os.Stdout, "=== fluxid Workflow Initialization ===\n")
	_, _ = fmt.Fprintf(os.Stdout, "Agent: %s (source: %s)\n", cfg.Agent, cfg.Sources["agent"])
	_, _ = fmt.Fprintf(os.Stdout, "Session ID: %s\n", cfg.SessionID)
	_, _ = fmt.Fprintf(
		os.Stdout,
		"Max Review Cycles: %d (source: %s)\n",
		cfg.MaxReviewCycles,
		cfg.Sources["iterations"],
	)
	_, _ = fmt.Fprintf(
		os.Stdout,
		"Max Implement Retries: %d (source: %s)\n",
		cfg.MaxImplementRetries,
		cfg.Sources["implement_retries"],
	)
	_, _ = fmt.Fprintf(
		os.Stdout,
		"Commit Enabled: %v (source: %s)\n",
		cfg.CommitEnabled,
		cfg.Sources["commit_enabled"],
	)

	if cfg.CommandFiles != nil {
		_, _ = fmt.Fprintf(os.Stdout, "\n")
		_, _ = fmt.Fprintf(os.Stdout, "Command Files:\n")
		_, _ = fmt.Fprintf(os.Stdout, "  Implement: %s\n", cfg.CommandFiles.ImplementPath)
		_, _ = fmt.Fprintf(os.Stdout, "  Review: %s\n", cfg.CommandFiles.ReviewPath)
		_, _ = fmt.Fprintf(os.Stdout, "  Commit: %s\n", cfg.CommandFiles.CommitPath)
	}

	if len(cfg.AgentArgs) > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "Agent Args: %v\n", cfg.AgentArgs)
	}
	_, _ = fmt.Fprintf(os.Stdout, "======================================\n")
	_, _ = fmt.Fprintf(os.Stdout, "\n")
}
