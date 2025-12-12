package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// HomeConfig represents the user's ~/.fluxid/config.yaml configuration.
type HomeConfig struct {
	Agent            *string `yaml:"agent"`
	ImplementRetries *int    `yaml:"implement_retries"`
	Iterations       *int    `yaml:"iterations"`
	CommitEnabled    *bool   `yaml:"commit_enabled"`
}

// ResolvedConfig contains the final configuration values with source tracking.
type ResolvedConfig struct {
	Agent            string
	ImplementRetries int
	Iterations       int
	CommitEnabled    bool

	// Source tracking: "default" or "home"
	Sources map[string]string
}

// Defaults for configuration values.
const (
	DefaultAgent            = "claude"
	DefaultImplementRetries = 3
	DefaultIterations       = 20
	DefaultCommitEnabled    = true
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

	return nil
}

// Resolve merges home config with defaults and tracks sources.
// CLI overrides can be provided as non-nil values and will take precedence.
func Resolve(homeConfig *HomeConfig, cliIterations, cliImplementRetries *int) *ResolvedConfig {
	resolved := &ResolvedConfig{
		Sources: make(map[string]string),
	}

	// Agent (no CLI override for now)
	if homeConfig != nil && homeConfig.Agent != nil {
		resolved.Agent = *homeConfig.Agent
		resolved.Sources["agent"] = "home"
	} else {
		resolved.Agent = DefaultAgent
		resolved.Sources["agent"] = "default"
	}

	// ImplementRetries (CLI can override)
	if cliImplementRetries != nil {
		resolved.ImplementRetries = *cliImplementRetries
		resolved.Sources["implement_retries"] = "cli"
	} else if homeConfig != nil && homeConfig.ImplementRetries != nil {
		resolved.ImplementRetries = *homeConfig.ImplementRetries
		resolved.Sources["implement_retries"] = "home"
	} else {
		resolved.ImplementRetries = DefaultImplementRetries
		resolved.Sources["implement_retries"] = "default"
	}

	// Iterations (CLI can override)
	if cliIterations != nil {
		resolved.Iterations = *cliIterations
		resolved.Sources["iterations"] = "cli"
	} else if homeConfig != nil && homeConfig.Iterations != nil {
		resolved.Iterations = *homeConfig.Iterations
		resolved.Sources["iterations"] = "home"
	} else {
		resolved.Iterations = DefaultIterations
		resolved.Sources["iterations"] = "default"
	}

	// CommitEnabled (no CLI override for now)
	if homeConfig != nil && homeConfig.CommitEnabled != nil {
		resolved.CommitEnabled = *homeConfig.CommitEnabled
		resolved.Sources["commit_enabled"] = "home"
	} else {
		resolved.CommitEnabled = DefaultCommitEnabled
		resolved.Sources["commit_enabled"] = "default"
	}

	return resolved
}
