package config

import (
	"errors"
	"fmt"
)

// SupportedAgents lists the supported agent values.
var SupportedAgents = []string{"claude", "codex", "opencode"} //nolint:gochecknoglobals // Shared config constant

// validateAgent validates that the agent value is not empty and is supported.
func validateAgent(agent string) error {
	if agent == "" {
		return errors.New("agent cannot be empty")
	}

	// Check if agent is in the supported list
	for _, supported := range SupportedAgents {
		if agent == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported agent %q. Supported agents: %v", agent, SupportedAgents)
}

// ValidateAgent is a public wrapper for validateAgent that can be called from main.
func ValidateAgent(agent string) error {
	return validateAgent(agent)
}

// validateHomeConfig validates the home config values.
func validateHomeConfig(cfg *HomeConfig) error {
	if cfg.ImplementRetries != nil && *cfg.ImplementRetries < 1 {
		return fmt.Errorf("implement_retries must be a positive integer (≥1), got: %d", *cfg.ImplementRetries)
	}

	if cfg.Iterations != nil && *cfg.Iterations < 1 {
		return fmt.Errorf("iterations must be a positive integer (≥1), got: %d", *cfg.Iterations)
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
		return fmt.Errorf("implement_retries must be a positive integer (≥1), got: %d", *cfg.ImplementRetries)
	}

	if cfg.Iterations != nil && *cfg.Iterations < 1 {
		return fmt.Errorf("iterations must be a positive integer (≥1), got: %d", *cfg.Iterations)
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
		if !hasImplement {
			return errors.New("commands.implement is required when commands are specified")
		}
		if !hasReview {
			return errors.New("commands.review is required when commands are specified")
		}
		if !hasCommit {
			return errors.New("commands.commit is required when commands are specified")
		}
	}

	return nil
}
