// Package config provides configuration management for fluxid.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateWorkflowConfig validates the workflow configuration at startup.
// This implements fail-fast validation to catch config errors before workflow execution begins.
func ValidateWorkflowConfig(cfg *WorkflowConfig, configDir string) error {
	// V001: Workflow section exists
	if cfg == nil {
		//nolint:err113 // Clear configuration error message
		return errors.New("workflow section is required in config.yaml")
	}

	// V002: Review section exists
	if cfg.Review.Command == "" {
		//nolint:err113 // Clear configuration error message
		return errors.New("workflow.review section is required")
	}

	// V003: Review command specified
	if cfg.Review.Command == "" {
		//nolint:err113 // Clear configuration error message
		return errors.New("workflow.review.command is required")
	}

	// V004: At least 1 custom step
	if len(cfg.Steps) == 0 {
		//nolint:err113 // Clear configuration error message
		return errors.New("at least one custom workflow step is required before review")
	}

	// V005-V009: Validate all custom steps
	if err := validateWorkflowSteps(cfg.Steps, configDir); err != nil {
		return err
	}

	// Validate review command path
	if err := ValidateCommandPath(cfg.Review.Command, configDir); err != nil {
		return fmt.Errorf("review step: %w", err)
	}

	// Validate review retries
	if cfg.Review.Retries < 0 {
		//nolint:err113 // Clear configuration error message
		return errors.New("review step: retries cannot be negative")
	}

	return nil
}

// validateWorkflowSteps validates all custom workflow steps.
// This checks for empty names, duplicate names, valid command paths, and non-negative retries.
func validateWorkflowSteps(steps []WorkflowStepConfig, configDir string) error {
	stepNames := make(map[string]bool)
	for _, step := range steps {
		// V005: Step names non-empty
		if strings.TrimSpace(step.Name) == "" {
			//nolint:err113 // Clear configuration error message
			return errors.New("step name cannot be empty or whitespace-only")
		}

		// V006: Step names unique
		if stepNames[step.Name] {
			//nolint:err113 // Clear configuration error message with context
			return fmt.Errorf("duplicate step name: %s", step.Name)
		}
		stepNames[step.Name] = true

		// V007, V008: Command file exists and readable
		if err := ValidateCommandPath(step.Command, configDir); err != nil {
			return fmt.Errorf("step %s: %w", step.Name, err)
		}

		// V009: Retries non-negative
		if step.Retries < 0 {
			//nolint:err113 // Clear configuration error message with context
			return fmt.Errorf("step %s: retries cannot be negative", step.Name)
		}
	}
	return nil
}

// ValidateCommandPath validates a command file path.
// Supports both absolute and relative paths (relative paths resolved from configDir).
func ValidateCommandPath(path string, configDir string) error {
	resolvedPath := path
	if !filepath.IsAbs(path) {
		resolvedPath = filepath.Join(configDir, path)
	}

	// #nosec G304 -- configDir is validated, path comes from config file
	fileInfo, err := os.Stat(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			//nolint:err113 // Clear configuration error message with context
			return fmt.Errorf("command file not found: %s", path)
		}
		//nolint:err113 // Clear configuration error message with context
		return fmt.Errorf("command file not readable: %s", path)
	}

	if fileInfo.IsDir() {
		//nolint:err113 // Clear configuration error message with context
		return fmt.Errorf("command path is a directory, not a file: %s", path)
	}

	return nil
}
