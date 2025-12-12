// Package config provides configuration management for fluxid,
// supporting home, project, and CLI-level configuration with precedence rules.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Configuration source constants.
const (
	SourceDefault = "default"
	SourceHome    = "home"
	SourceProject = "project"
	SourceEnv     = "env"
	SourceCLI     = "cli"
)

// Commands represents command file paths configuration.
type Commands struct {
	Implement *string `yaml:"implement"`
	Review    *string `yaml:"review"`
	Commit    *string `yaml:"commit"`
}

// HomeConfig represents the user's ~/.fluxid/config.yaml configuration.
type HomeConfig struct {
	Agent            *string   `yaml:"agent"`
	ImplementRetries *int      `yaml:"implement_retries"`
	Iterations       *int      `yaml:"iterations"`
	CommitEnabled    *bool     `yaml:"commit_enabled"`
	Commands         *Commands `yaml:"commands"`
}

// ProjectConfig represents the project's ./.fluxid/config.yaml configuration.
// Structurally identical to HomeConfig but semantically distinct.
type ProjectConfig struct {
	Agent            *string   `yaml:"agent"`
	ImplementRetries *int      `yaml:"implement_retries"`
	Iterations       *int      `yaml:"iterations"`
	CommitEnabled    *bool     `yaml:"commit_enabled"`
	Commands         *Commands `yaml:"commands"`
}

// ResolvedCommandFiles contains the resolved absolute paths to command files.
type ResolvedCommandFiles struct {
	ImplementPath string
	ReviewPath    string
	CommitPath    string
}

// ResolvedConfig contains the final configuration values with source tracking.
type ResolvedConfig struct {
	Agent            string
	ImplementRetries int
	Iterations       int
	CommitEnabled    bool
	CommandFiles     *ResolvedCommandFiles

	// Source tracking: "default" or "home"
	Sources map[string]string
}

// Defaults for configuration values.
const (
	DefaultAgent            = "claude"
	DefaultImplementRetries = 3
	DefaultIterations       = 20
	DefaultCommitEnabled    = false
)

// LoadHomeConfig reads and parses ~/.fluxid/config.yaml if it exists.
// Returns nil if the file doesn't exist (not an error).
func LoadHomeConfig() (*HomeConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".fluxid", "config.yaml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // Not an error - file simply doesn't exist
	}

	// #nosec G304 -- configPath is constructed from user's home directory, not user input
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var homeConfig HomeConfig
	if err := yaml.Unmarshal(data, &homeConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Validate the loaded config
	if err := validateHomeConfig(&homeConfig); err != nil {
		return nil, fmt.Errorf("invalid config in %s: %w", configPath, err)
	}

	return &homeConfig, nil
}

// LoadProjectConfig reads and parses ./.fluxid/config.yaml if it exists.
// Returns nil if the file doesn't exist (not an error).
func LoadProjectConfig() (*ProjectConfig, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(cwd, ".fluxid", "config.yaml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // Not an error - file simply doesn't exist
	}

	// #nosec G304 -- configPath is constructed from current working directory, not user input
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config file %s: %w", configPath, err)
	}

	var projectConfig ProjectConfig
	if err := yaml.Unmarshal(data, &projectConfig); err != nil {
		return nil, fmt.Errorf("failed to parse project config file %s: %w", configPath, err)
	}

	// Validate the loaded config
	if err := validateProjectConfig(&projectConfig); err != nil {
		return nil, fmt.Errorf("invalid config in %s: %w", configPath, err)
	}

	return &projectConfig, nil
}

// validateHomeConfig validates the home config values.
func validateHomeConfig(cfg *HomeConfig) error {
	if cfg.ImplementRetries != nil && *cfg.ImplementRetries < 1 {
		return fmt.Errorf("implement_retries must be a positive integer (≥1), got: %d", *cfg.ImplementRetries)
	}

	if cfg.Iterations != nil && *cfg.Iterations < 1 {
		return fmt.Errorf("iterations must be a positive integer (≥1), got: %d", *cfg.Iterations)
	}

	if cfg.Agent != nil && *cfg.Agent == "" {
		return errors.New("agent cannot be empty")
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

	if cfg.Agent != nil && *cfg.Agent == "" {
		return errors.New("agent cannot be empty")
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

// Resolve merges project, home, env config with defaults and tracks sources.
// Precedence: CLI > env > project > home > defaults.
// CLI overrides can be provided as non-nil values and will take precedence.
func Resolve(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
	envConfig *EnvConfig,
	cliIterations, cliImplementRetries *int,
	cliCommitEnabled *bool,
) *ResolvedConfig {
	resolved := &ResolvedConfig{
		Agent:            DefaultAgent,
		ImplementRetries: DefaultImplementRetries,
		Iterations:       DefaultIterations,
		CommitEnabled:    DefaultCommitEnabled,
		CommandFiles:     nil,
		Sources:          make(map[string]string),
	}

	// Agent
	// Precedence: CLI > env > project > home > default
	// Note: no CLI flag for agent yet
	switch {
	case envConfig != nil && envConfig.Agent != nil:
		resolved.Agent = *envConfig.Agent
		resolved.Sources["agent"] = SourceEnv
	case projectConfig != nil && projectConfig.Agent != nil:
		resolved.Agent = *projectConfig.Agent
		resolved.Sources["agent"] = SourceProject
	case homeConfig != nil && homeConfig.Agent != nil:
		resolved.Agent = *homeConfig.Agent
		resolved.Sources["agent"] = SourceHome
	default:
		resolved.Agent = DefaultAgent
		resolved.Sources["agent"] = SourceDefault
	}

	// ImplementRetries
	// Precedence: CLI > env > project > home > default
	switch {
	case cliImplementRetries != nil:
		resolved.ImplementRetries = *cliImplementRetries
		resolved.Sources["implement_retries"] = SourceCLI
	case envConfig != nil && envConfig.ImplementRetries != nil:
		resolved.ImplementRetries = *envConfig.ImplementRetries
		resolved.Sources["implement_retries"] = SourceEnv
	case projectConfig != nil && projectConfig.ImplementRetries != nil:
		resolved.ImplementRetries = *projectConfig.ImplementRetries
		resolved.Sources["implement_retries"] = SourceProject
	case homeConfig != nil && homeConfig.ImplementRetries != nil:
		resolved.ImplementRetries = *homeConfig.ImplementRetries
		resolved.Sources["implement_retries"] = SourceHome
	default:
		resolved.ImplementRetries = DefaultImplementRetries
		resolved.Sources["implement_retries"] = SourceDefault
	}

	// Iterations
	// Precedence: CLI > env > project > home > default
	switch {
	case cliIterations != nil:
		resolved.Iterations = *cliIterations
		resolved.Sources["iterations"] = SourceCLI
	case envConfig != nil && envConfig.Iterations != nil:
		resolved.Iterations = *envConfig.Iterations
		resolved.Sources["iterations"] = SourceEnv
	case projectConfig != nil && projectConfig.Iterations != nil:
		resolved.Iterations = *projectConfig.Iterations
		resolved.Sources["iterations"] = SourceProject
	case homeConfig != nil && homeConfig.Iterations != nil:
		resolved.Iterations = *homeConfig.Iterations
		resolved.Sources["iterations"] = SourceHome
	default:
		resolved.Iterations = DefaultIterations
		resolved.Sources["iterations"] = SourceDefault
	}

	// CommitEnabled
	// Precedence: CLI > env > project > home > default
	switch {
	case cliCommitEnabled != nil:
		resolved.CommitEnabled = *cliCommitEnabled
		resolved.Sources["commit_enabled"] = SourceCLI
	case envConfig != nil && envConfig.CommitEnabled != nil:
		resolved.CommitEnabled = *envConfig.CommitEnabled
		resolved.Sources["commit_enabled"] = SourceEnv
	case projectConfig != nil && projectConfig.CommitEnabled != nil:
		resolved.CommitEnabled = *projectConfig.CommitEnabled
		resolved.Sources["commit_enabled"] = SourceProject
	case homeConfig != nil && homeConfig.CommitEnabled != nil:
		resolved.CommitEnabled = *homeConfig.CommitEnabled
		resolved.Sources["commit_enabled"] = SourceHome
	default:
		resolved.CommitEnabled = DefaultCommitEnabled
		resolved.Sources["commit_enabled"] = SourceDefault
	}

	return resolved
}
