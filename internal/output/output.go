// Package output implements output formatting for fluxid initialization status.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	yamlIndentSpaces = 2 // YAML indentation level in spaces
)

var errUnsupportedOutputFormat = errors.New("unsupported output format (supported: text, json, yaml)")

// Format represents the supported output formats.
type Format string

// Supported output formats.
const (
	FormatText Format = "text"
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// ValidateFormat validates the output format string.
func ValidateFormat(format string) error {
	switch format {
	case "text", "json", "yaml":
		return nil
	default:
		return fmt.Errorf("%s: %w", format, errUnsupportedOutputFormat)
	}
}

// InitializationStatus represents the initialization status output.
// This is used by the Config type in the command package.
type InitializationStatus struct {
	SessionID           string            `json:"session_id" yaml:"session_id"`
	Agent               string            `json:"agent" yaml:"agent"`
	MaxReviewCycles     int               `json:"max_review_cycles" yaml:"max_review_cycles"`
	MaxImplementRetries int               `json:"max_implement_retries" yaml:"max_implement_retries"`
	MaxCommitRetries    int               `json:"max_commit_retries" yaml:"max_commit_retries"`
	TaskFile            string            `json:"task_file" yaml:"task_file"`
	CommandFiles        *CommandFilesJSON `json:"command_files,omitempty" yaml:"command_files,omitempty"`
	AgentArgs           []string          `json:"agent_args,omitempty" yaml:"agent_args,omitempty"`
}

// CommandFilesJSON represents command file paths in JSON/YAML output.
type CommandFilesJSON struct {
	Implement string `json:"implement" yaml:"implement"`
	Review    string `json:"review" yaml:"review"`
	Commit    string `json:"commit" yaml:"commit"`
}

// PrintJSON outputs initialization status as JSON to stdout.
func PrintJSON(status InitializationStatus) error {
	return PrintJSONToWriter(os.Stdout, status)
}

// PrintJSONToWriter outputs initialization status as JSON to the provided writer.
func PrintJSONToWriter(w io.Writer, status InitializationStatus) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(status); err != nil {
		return fmt.Errorf("failed to encode initialization status: %w", err)
	}
	return nil
}

// PrintYAML outputs initialization status as YAML to stdout.
func PrintYAML(status InitializationStatus) error {
	return PrintYAMLToWriter(os.Stdout, status)
}

// PrintYAMLToWriter outputs initialization status as YAML to the provided writer.
func PrintYAMLToWriter(w io.Writer, status InitializationStatus) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(yamlIndentSpaces)
	if err := encoder.Encode(status); err != nil {
		return fmt.Errorf("failed to encode initialization status: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close YAML encoder: %w", err)
	}
	return nil
}

// PrintText outputs initialization status as human-readable text to stdout.
func PrintText(status InitializationStatus) {
	PrintTextToWriter(os.Stdout, status)
}

// PrintTextToWriter outputs initialization status as human-readable text to the provided writer.
func PrintTextToWriter(writer io.Writer, status InitializationStatus) {
	_, _ = fmt.Fprintf(writer, "=== fluxid Workflow Initialization ===\n")
	_, _ = fmt.Fprintf(writer, "Agent: %s\n", status.Agent)
	_, _ = fmt.Fprintf(writer, "Session ID: %s\n", status.SessionID)
	_, _ = fmt.Fprintf(writer, "Max Review Cycles: %d\n", status.MaxReviewCycles)
	_, _ = fmt.Fprintf(writer, "Max Implement Retries: %d\n", status.MaxImplementRetries)
	_, _ = fmt.Fprintf(writer, "Max Commit Retries: %d\n", status.MaxCommitRetries)

	if status.TaskFile != "" {
		_, _ = fmt.Fprintf(writer, "Task File: %s\n", status.TaskFile)
	}

	if status.CommandFiles != nil {
		_, _ = fmt.Fprintf(writer, "\n")
		_, _ = fmt.Fprintf(writer, "Command Files:\n")
		_, _ = fmt.Fprintf(writer, "  Implement: %s\n", status.CommandFiles.Implement)
		_, _ = fmt.Fprintf(writer, "  Review: %s\n", status.CommandFiles.Review)
		_, _ = fmt.Fprintf(writer, "  Commit: %s\n", status.CommandFiles.Commit)
	}

	if len(status.AgentArgs) > 0 {
		_, _ = fmt.Fprintf(writer, "Agent Args: %v\n", status.AgentArgs)
	}
	_, _ = fmt.Fprintf(writer, "======================================\n")
	_, _ = fmt.Fprintf(writer, "\n")
}
