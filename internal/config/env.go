package config

import (
	"errors"
	"fmt"
)

// EnvConfig represents configuration from environment variables.
type EnvConfig struct {
	Agent            *string
	ImplementRetries *int
	Iterations       *int
	CommitEnabled    *bool
}

// EnvGetter is an interface for retrieving environment variables.
type EnvGetter interface {
	Getenv(key string) string
}

// LoadEnvConfig reads configuration from environment variables.
// Returns nil if no environment variables are set.
func LoadEnvConfig(envGetter EnvGetter) (*EnvConfig, error) {
	env := &EnvConfig{
		Agent:            nil,
		ImplementRetries: nil,
		Iterations:       nil,
		CommitEnabled:    nil,
	}
	hasAny := false

	// Parse FLUXID_AGENT
	if val := envGetter.Getenv("FLUXID_AGENT"); val != "" {
		if val == "" {
			return nil, errors.New("FLUXID_AGENT cannot be empty")
		}
		env.Agent = &val
		hasAny = true
	}

	// Parse FLUXID_ITERATIONS
	if val := envGetter.Getenv("FLUXID_ITERATIONS"); val != "" {
		var n int
		_, err := fmt.Sscanf(val, "%d", &n)
		if err != nil {
			return nil, fmt.Errorf("FLUXID_ITERATIONS must be a valid integer, got: %s", val)
		}
		if n < 1 {
			return nil, fmt.Errorf("FLUXID_ITERATIONS must be a positive integer (≥1), got: %d", n)
		}
		env.Iterations = &n
		hasAny = true
	}

	// Parse FLUXID_IMPLEMENT_RETRIES
	if val := envGetter.Getenv("FLUXID_IMPLEMENT_RETRIES"); val != "" {
		var n int
		_, err := fmt.Sscanf(val, "%d", &n)
		if err != nil {
			return nil, fmt.Errorf("FLUXID_IMPLEMENT_RETRIES must be a valid integer, got: %s", val)
		}
		if n < 1 {
			return nil, fmt.Errorf("FLUXID_IMPLEMENT_RETRIES must be a positive integer (≥1), got: %d", n)
		}
		env.ImplementRetries = &n
		hasAny = true
	}

	// Parse FLUXID_COMMIT_ENABLED
	if val := envGetter.Getenv("FLUXID_COMMIT_ENABLED"); val != "" {
		switch val {
		case "true", "1", "yes":
			b := true
			env.CommitEnabled = &b
		case "false", "0", "no":
			b := false
			env.CommitEnabled = &b
		default:
			return nil, fmt.Errorf("FLUXID_COMMIT_ENABLED must be true/false/1/0/yes/no, got: %s", val)
		}
		hasAny = true
	}

	if !hasAny {
		return nil, nil
	}

	return env, nil
}
