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

// Configuration source constants removed in v2.0 - source tracking no longer supported

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
	CommitRetries    *int      `yaml:"commit_retries"`
	Iterations       *int      `yaml:"iterations"`
	Commands         *Commands `yaml:"commands"`
}

// ProjectConfig represents the project's ./.fluxid/config.yaml configuration.
// Structurally identical to HomeConfig but semantically distinct.
type ProjectConfig struct {
	Agent            *string   `yaml:"agent"`
	ImplementRetries *int      `yaml:"implement_retries"`
	CommitRetries    *int      `yaml:"commit_retries"`
	Iterations       *int      `yaml:"iterations"`
	Commands         *Commands `yaml:"commands"`
}

// ResolvedCommandFiles contains the resolved absolute paths to command files.
type ResolvedCommandFiles struct {
	ImplementPath string
	ReviewPath    string
	CommitPath    string
}

// ResolvedConfig contains the final configuration values.
type ResolvedConfig struct {
	Agent            string
	ImplementRetries int
	CommitRetries    int
	Iterations       int
	CommandFiles     *ResolvedCommandFiles
}

// Defaults for configuration values.
const (
	DefaultAgent            = "claude"
	DefaultImplementRetries = 3
	DefaultCommitRetries    = 100
	DefaultIterations       = 20
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
		//nolint:nilnil // Valid: no config file found is not an error, return nil to indicate "no home config"
		return nil, nil
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
		//nolint:nilnil // Valid: no config file found is not an error, return nil to indicate "no project config"
		return nil, nil
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

// LoadDefaultConfig loads default configuration from project and/or home config files.
// At least one config file must exist. Returns error if neither exists.
// Load order: Project config (.fluxid/config.yaml) → User config (~/.fluxid/config.yaml)
// Fails fast on invalid YAML - does not fall back to user config if project config is invalid.
func LoadDefaultConfig() (*ProjectConfig, *HomeConfig, error) {
	// Try to load project config first
	projectConfig, err := LoadProjectConfig()
	if err != nil {
		// Fail fast on project config error (invalid YAML or validation failure)
		return nil, nil, err
	}

	// Try to load home config
	homeConfig, err := LoadHomeConfig()
	if err != nil {
		// Fail fast on home config error (invalid YAML or validation failure)
		return nil, nil, err
	}

	// At least one config must exist
	if projectConfig == nil && homeConfig == nil {
		//nolint:err113 // Configuration error with clear message, sentinel error not needed
		return nil, nil, errors.New("at least one default config must exist: ~/.fluxid/config.yaml or .fluxid/config.yaml")
	}

	return projectConfig, homeConfig, nil
}

// CustomConfig represents a custom config file loaded via --config flag.
// Structurally identical to HomeConfig/ProjectConfig but semantically distinct.
type CustomConfig struct {
	Agent            *string   `yaml:"agent"`
	ImplementRetries *int      `yaml:"implement_retries"`
	CommitRetries    *int      `yaml:"commit_retries"`
	Iterations       *int      `yaml:"iterations"`
	Commands         *Commands `yaml:"commands"`
}

// LoadCustomConfig reads and parses a custom config file from the given path.
// The path can be relative or absolute.
func LoadCustomConfig(configPath string) (*CustomConfig, string, error) {
	// Convert to absolute path if needed
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve config path %s: %w", configPath, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		//nolint:err113 // Dynamic error message includes file path for better diagnostics
		return nil, "", fmt.Errorf("config file not found: %s", absPath)
	}

	// #nosec G304 -- configPath comes from CLI flag, user-controlled
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read config file %s: %w", absPath, err)
	}

	var customConfig CustomConfig
	if err := yaml.Unmarshal(data, &customConfig); err != nil {
		return nil, "", fmt.Errorf("failed to parse config file %s: %w", absPath, err)
	}

	// Validate the loaded config
	if err := validateCustomConfig(&customConfig); err != nil {
		return nil, "", fmt.Errorf("invalid config in %s: %w", absPath, err)
	}

	// Return config and its directory for path resolution
	configDir := filepath.Dir(absPath)
	return &customConfig, configDir, nil
}

// validateCustomConfig validates the custom config values.
func validateCustomConfig(cfg *CustomConfig) error {
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

//nolint:ireturn,nolintlint // Generic function returns type parameter T
func resolveField[T any](
	cliValue *T,
	projectValue *T,
	homeValue *T,
	defaultValue T,
) T {
	switch {
	case cliValue != nil:
		return *cliValue
	case projectValue != nil:
		return *projectValue
	case homeValue != nil:
		return *homeValue
	default:
		return defaultValue
	}
}

// Resolve merges project and home config with defaults.
// Precedence: CLI > project > home > defaults.
// CLI overrides can be provided as non-nil values and will take precedence.
func Resolve(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
	cliAgent *string,
	cliIterations, cliImplementRetries, cliCommitRetries *int,
) *ResolvedConfig {
	resolved := &ResolvedConfig{
		Agent:            DefaultAgent,
		ImplementRetries: DefaultImplementRetries,
		CommitRetries:    DefaultCommitRetries,
		Iterations:       DefaultIterations,
		CommandFiles:     nil,
	}

	// Extract all config values
	agentValues := extractAgentValues(projectConfig, homeConfig)
	implementRetriesValues := extractImplementRetriesValues(projectConfig, homeConfig)
	commitRetriesValues := extractCommitRetriesValues(projectConfig, homeConfig)
	iterationsValues := extractIterationsValues(projectConfig, homeConfig)

	// Resolve each field using the helper
	resolved.Agent = resolveField(
		cliAgent, agentValues.project, agentValues.home, DefaultAgent,
	)
	resolved.ImplementRetries = resolveField(
		cliImplementRetries, implementRetriesValues.project, implementRetriesValues.home, DefaultImplementRetries,
	)
	resolved.CommitRetries = resolveField(
		cliCommitRetries, commitRetriesValues.project, commitRetriesValues.home, DefaultCommitRetries,
	)
	resolved.Iterations = resolveField(
		cliIterations, iterationsValues.project, iterationsValues.home, DefaultIterations,
	)

	return resolved
}

type configValues[T any] struct {
	project *T
	home    *T
}

func extractAgentValues(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
) configValues[string] {
	var values configValues[string]
	if projectConfig != nil {
		values.project = projectConfig.Agent
	}
	if homeConfig != nil {
		values.home = homeConfig.Agent
	}
	return values
}

func extractImplementRetriesValues(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
) configValues[int] {
	var values configValues[int]
	if projectConfig != nil {
		values.project = projectConfig.ImplementRetries
	}
	if homeConfig != nil {
		values.home = homeConfig.ImplementRetries
	}
	return values
}

func extractCommitRetriesValues(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
) configValues[int] {
	var values configValues[int]
	if projectConfig != nil {
		values.project = projectConfig.CommitRetries
	}
	if homeConfig != nil {
		values.home = homeConfig.CommitRetries
	}
	return values
}

func extractIterationsValues(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
) configValues[int] {
	var values configValues[int]
	if projectConfig != nil {
		values.project = projectConfig.Iterations
	}
	if homeConfig != nil {
		values.home = homeConfig.Iterations
	}
	return values
}
