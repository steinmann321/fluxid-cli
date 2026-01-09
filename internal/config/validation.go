package config

import (
	"errors"
	"fmt"
)

// SupportedAgents lists the supported agent values.
//
//nolint:gochecknoglobals // Shared config constant
var SupportedAgents = []string{"claude", "codex", "opencode", "gemini"}

var (
	errValidationAgentEmpty       = errors.New("agent cannot be empty")
	errAgentUnsupported           = errors.New("unsupported agent")
	errValidationImplementRetries = errors.New("implement_retries must be a positive integer (≥1)")
	errValidationCommitRetries    = errors.New("commit_retries must be a positive integer (≥1)")
	errValidationIterations       = errors.New("iterations must be a positive integer (≥1)")
	errCommandImplementRequired   = errors.New("commands.implement is required when commands are specified")
	errCommandReviewRequired      = errors.New("commands.review is required when commands are specified")
	errCommandCommitRequired      = errors.New("commands.commit is required when commands are specified")
)

// validateAgent validates that the agent value is not empty and is supported.
func validateAgent(agent string) error {
	if agent == "" {
		return errValidationAgentEmpty
	}

	// Check if agent is in the supported list
	for _, supported := range SupportedAgents {
		if agent == supported {
			return nil
		}
	}

	return fmt.Errorf("%q (supported agents: %v): %w", agent, SupportedAgents, errAgentUnsupported)
}

// ValidateAgent is a public wrapper for validateAgent that can be called from main.
func ValidateAgent(agent string) error {
	return validateAgent(agent)
}

// validateHomeConfig validates the home config values.
func validateHomeConfig(cfg *HomeConfig) error {
	if cfg.ImplementRetries != nil && *cfg.ImplementRetries < 1 {
		return fmt.Errorf("got %d: %w", *cfg.ImplementRetries, errValidationImplementRetries)
	}

	if cfg.CommitRetries != nil && *cfg.CommitRetries < 1 {
		return fmt.Errorf("got %d: %w", *cfg.CommitRetries, errValidationCommitRetries)
	}

	if cfg.Iterations != nil && *cfg.Iterations < 1 {
		return fmt.Errorf("got %d: %w", *cfg.Iterations, errValidationIterations)
	}

	if cfg.Agent != nil {
		if err := validateAgent(*cfg.Agent); err != nil {
			return err
		}
	}

	// Validate commands structure if provided
	if err := validateCommands(cfg.Commands); err != nil {
		return err
	}

	return nil
}

// validateProjectConfig validates the project config values.
func validateProjectConfig(cfg *ProjectConfig) error {
	if cfg.ImplementRetries != nil && *cfg.ImplementRetries < 1 {
		return fmt.Errorf("got %d: %w", *cfg.ImplementRetries, errValidationImplementRetries)
	}

	if cfg.CommitRetries != nil && *cfg.CommitRetries < 1 {
		return fmt.Errorf("got %d: %w", *cfg.CommitRetries, errValidationCommitRetries)
	}

	if cfg.Iterations != nil && *cfg.Iterations < 1 {
		return fmt.Errorf("got %d: %w", *cfg.Iterations, errValidationIterations)
	}

	if cfg.Agent != nil {
		if err := validateAgent(*cfg.Agent); err != nil {
			return err
		}
	}

	// Validate commands structure if provided
	if err := validateCommands(cfg.Commands); err != nil {
		return err
	}

	return nil
}

// validateCommands validates the commands configuration.
// Either all three command files must be specified or none.
func validateCommands(cmds *Commands) error {
	if cmds == nil {
		return nil
	}

	// Check if any command is specified
	hasImplement := cmds.Implement != nil && *cmds.Implement != ""
	hasReview := cmds.Review != nil && *cmds.Review != ""
	hasCommit := cmds.Commit != nil && *cmds.Commit != ""

	// If any command is specified, all must be specified
	if hasImplement || hasReview || hasCommit {
		return validateAllCommandsPresent(hasImplement, hasReview, hasCommit)
	}

	return nil
}

func validateAllCommandsPresent(hasImplement, hasReview, hasCommit bool) error {
	if !hasImplement {
		return errCommandImplementRequired
	}
	if !hasReview {
		return errCommandReviewRequired
	}
	if !hasCommit {
		return errCommandCommitRequired
	}
	return nil
}
