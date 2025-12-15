// Package config provides configuration management for fluxid,
// supporting home, project, and CLI-level configuration with precedence rules.
package config

import (
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

// GetHomeConfigPath returns the path to the home config file.
func GetHomeConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".fluxid", "config.yaml"), nil
}

// GetProjectConfigPath returns the path to the project config file.
func GetProjectConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, ".fluxid", "config.yaml"), nil
}

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

// resolveField resolves a configuration field with precedence: CLI > env > project > home > default.
// For project and home sources, file paths are included in the source string.
func resolveField[T any](
	fieldName string,
	cliValue *T,
	envValue *T,
	projectValue *T,
	homeValue *T,
	defaultValue T,
	sources map[string]string,
	projectConfigPath string,
	homeConfigPath string,
) T {
	switch {
	case cliValue != nil:
		sources[fieldName] = SourceCLI
		return *cliValue
	case envValue != nil:
		sources[fieldName] = SourceEnv
		return *envValue
	case projectValue != nil:
		if projectConfigPath != "" {
			sources[fieldName] = fmt.Sprintf("%s (%s)", SourceProject, projectConfigPath)
		} else {
			sources[fieldName] = SourceProject
		}
		return *projectValue
	case homeValue != nil:
		if homeConfigPath != "" {
			sources[fieldName] = fmt.Sprintf("%s (%s)", SourceHome, homeConfigPath)
		} else {
			sources[fieldName] = SourceHome
		}
		return *homeValue
	default:
		sources[fieldName] = SourceDefault
		return defaultValue
	}
}

// Resolve merges project, home, env config with defaults and tracks sources.
// Precedence: CLI > env > project > home > defaults.
// CLI overrides can be provided as non-nil values and will take precedence.
func Resolve(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
	envConfig *EnvConfig,
	cliAgent *string,
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

	// Get config file paths for display in source strings
	var projectConfigPath, homeConfigPath string
	if projectConfig != nil {
		if path, err := GetProjectConfigPath(); err == nil {
			projectConfigPath = path
		}
	}
	if homeConfig != nil {
		if path, err := GetHomeConfigPath(); err == nil {
			homeConfigPath = path
		}
	}

	// Extract config values or nil
	var envAgent, projectAgent, homeAgent *string
	if envConfig != nil {
		envAgent = envConfig.Agent
	}
	if projectConfig != nil {
		projectAgent = projectConfig.Agent
	}
	if homeConfig != nil {
		homeAgent = homeConfig.Agent
	}

	var envImplementRetries, projectImplementRetries, homeImplementRetries *int
	if envConfig != nil {
		envImplementRetries = envConfig.ImplementRetries
	}
	if projectConfig != nil {
		projectImplementRetries = projectConfig.ImplementRetries
	}
	if homeConfig != nil {
		homeImplementRetries = homeConfig.ImplementRetries
	}

	var envIterations, projectIterations, homeIterations *int
	if envConfig != nil {
		envIterations = envConfig.Iterations
	}
	if projectConfig != nil {
		projectIterations = projectConfig.Iterations
	}
	if homeConfig != nil {
		homeIterations = homeConfig.Iterations
	}

	var envCommitEnabled, projectCommitEnabled, homeCommitEnabled *bool
	if envConfig != nil {
		envCommitEnabled = envConfig.CommitEnabled
	}
	if projectConfig != nil {
		projectCommitEnabled = projectConfig.CommitEnabled
	}
	if homeConfig != nil {
		homeCommitEnabled = homeConfig.CommitEnabled
	}

	// Resolve each field using the helper
	resolved.Agent = resolveField(
		"agent", cliAgent, envAgent, projectAgent, homeAgent, DefaultAgent, resolved.Sources,
		projectConfigPath, homeConfigPath,
	)
	resolved.ImplementRetries = resolveField(
		"implement_retries", cliImplementRetries, envImplementRetries,
		projectImplementRetries, homeImplementRetries, DefaultImplementRetries, resolved.Sources,
		projectConfigPath, homeConfigPath,
	)
	resolved.Iterations = resolveField(
		"iterations", cliIterations, envIterations, projectIterations,
		homeIterations, DefaultIterations, resolved.Sources,
		projectConfigPath, homeConfigPath,
	)
	resolved.CommitEnabled = resolveField(
		"commit_enabled", cliCommitEnabled, envCommitEnabled,
		projectCommitEnabled, homeCommitEnabled, DefaultCommitEnabled, resolved.Sources,
		projectConfigPath, homeConfigPath,
	)

	return resolved
}
