package config

import (
	"errors"
	"fmt"
)

var (
	errAgentEmpty                  = errors.New("FLUXID_AGENT cannot be empty")
	errIterationsInvalidInt        = errors.New("FLUXID_ITERATIONS must be a valid integer")
	errIterationsNotPositive       = errors.New("FLUXID_ITERATIONS must be a positive integer (≥1)")
	errImplementRetriesInvalidInt  = errors.New("FLUXID_IMPLEMENT_RETRIES must be a valid integer")
	errImplementRetriesNotPositive = errors.New("FLUXID_IMPLEMENT_RETRIES must be a positive integer (≥1)")
	errCommitEnabledInvalid        = errors.New("FLUXID_COMMIT_ENABLED must be true/false/1/0/yes/no")
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
	if err := parseEnvAgent(envGetter, env, &hasAny); err != nil {
		return nil, err
	}

	// Parse FLUXID_ITERATIONS
	if err := parseEnvIterations(envGetter, env, &hasAny); err != nil {
		return nil, err
	}

	// Parse FLUXID_IMPLEMENT_RETRIES
	if err := parseEnvImplementRetries(envGetter, env, &hasAny); err != nil {
		return nil, err
	}

	// Parse FLUXID_COMMIT_ENABLED
	if err := parseEnvCommitEnabled(envGetter, env, &hasAny); err != nil {
		return nil, err
	}

	if !hasAny {
		//nolint:nilnil // Valid: no environment variables set is not an error, return nil to indicate "no env config"
		return nil, nil
	}

	return env, nil
}

func parseEnvAgent(envGetter EnvGetter, env *EnvConfig, hasAny *bool) error {
	if val := envGetter.Getenv("FLUXID_AGENT"); val != "" {
		if val == "" {
			return errAgentEmpty
		}
		env.Agent = &val
		*hasAny = true
	}
	return nil
}

func parseEnvIterations(envGetter EnvGetter, env *EnvConfig, hasAny *bool) error {
	if val := envGetter.Getenv("FLUXID_ITERATIONS"); val != "" {
		var number int
		_, err := fmt.Sscanf(val, "%d", &number)
		if err != nil {
			return fmt.Errorf("%w, got: %s", errIterationsInvalidInt, val)
		}
		if number < 1 {
			return fmt.Errorf("%w, got: %d", errIterationsNotPositive, number)
		}
		env.Iterations = &number
		*hasAny = true
	}
	return nil
}

func parseEnvImplementRetries(envGetter EnvGetter, env *EnvConfig, hasAny *bool) error {
	if val := envGetter.Getenv("FLUXID_IMPLEMENT_RETRIES"); val != "" {
		var number int
		_, err := fmt.Sscanf(val, "%d", &number)
		if err != nil {
			return fmt.Errorf("%w, got: %s", errImplementRetriesInvalidInt, val)
		}
		if number < 1 {
			return fmt.Errorf("%w, got: %d", errImplementRetriesNotPositive, number)
		}
		env.ImplementRetries = &number
		*hasAny = true
	}
	return nil
}

func parseEnvCommitEnabled(envGetter EnvGetter, env *EnvConfig, hasAny *bool) error {
	if val := envGetter.Getenv("FLUXID_COMMIT_ENABLED"); val != "" {
		switch val {
		case "true", "1", "yes":
			b := true
			env.CommitEnabled = &b
		case "false", "0", "no":
			b := false
			env.CommitEnabled = &b
		default:
			return fmt.Errorf("%w, got: %s", errCommitEnabledInvalid, val)
		}
		*hasAny = true
	}
	return nil
}
