package config

import (
	"testing"
)

type mockEnv struct {
	vars map[string]string
}

func (m *mockEnv) Getenv(key string) string {
	return m.vars[key]
}

func TestLoadEnvConfig_NoVars(t *testing.T) {
	t.Parallel()

	env := &mockEnv{vars: map[string]string{}}
	cfg, err := LoadEnvConfig(env)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg != nil {
		t.Errorf("Expected nil config when no env vars set, got: %+v", cfg)
	}
}

func TestLoadEnvConfig_Agent(t *testing.T) {
	t.Parallel()

	env := &mockEnv{vars: map[string]string{
		"FLUXID_AGENT": "opencode",
	}}
	cfg, err := LoadEnvConfig(env)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	if cfg.Agent == nil || *cfg.Agent != "opencode" {
		t.Errorf("Expected Agent=opencode, got: %v", cfg.Agent)
	}
}

func TestLoadEnvConfig_Iterations(t *testing.T) {
	t.Parallel()

	env := &mockEnv{vars: map[string]string{
		"FLUXID_ITERATIONS": "30",
	}}
	cfg, err := LoadEnvConfig(env)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	if cfg.Iterations == nil || *cfg.Iterations != 30 {
		t.Errorf("Expected Iterations=30, got: %v", cfg.Iterations)
	}
}

func TestLoadEnvConfig_ImplementRetries(t *testing.T) {
	t.Parallel()

	env := &mockEnv{vars: map[string]string{
		"FLUXID_IMPLEMENT_RETRIES": "7",
	}}
	cfg, err := LoadEnvConfig(env)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	if cfg.ImplementRetries == nil || *cfg.ImplementRetries != 7 {
		t.Errorf("Expected ImplementRetries=7, got: %v", cfg.ImplementRetries)
	}
}

func TestLoadEnvConfig_CommitEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := &mockEnv{vars: map[string]string{
				"FLUXID_COMMIT_ENABLED": testCase.value,
			}}
			cfg, err := LoadEnvConfig(env)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if cfg == nil {
				t.Fatal("Expected non-nil config")
			}

			if cfg.CommitEnabled == nil || *cfg.CommitEnabled != testCase.expected {
				t.Errorf("Expected CommitEnabled=%v, got: %v", testCase.expected, cfg.CommitEnabled)
			}
		})
	}
}

func TestLoadEnvConfig_InvalidIterations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{"negative", "-5"},
		{"zero", "0"},
		{"non-numeric", "abc"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := &mockEnv{vars: map[string]string{
				"FLUXID_ITERATIONS": testCase.value,
			}}
			_, err := LoadEnvConfig(env)

			if err == nil {
				t.Errorf("Expected error for invalid FLUXID_ITERATIONS=%s", testCase.value)
			}
		})
	}
}

func TestLoadEnvConfig_InvalidImplementRetries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{"negative", "-3"},
		{"zero", "0"},
		{"non-numeric", "xyz"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			env := &mockEnv{vars: map[string]string{
				"FLUXID_IMPLEMENT_RETRIES": testCase.value,
			}}
			_, err := LoadEnvConfig(env)

			if err == nil {
				t.Errorf("Expected error for invalid FLUXID_IMPLEMENT_RETRIES=%s", testCase.value)
			}
		})
	}
}

func TestLoadEnvConfig_InvalidCommitEnabled(t *testing.T) {
	t.Parallel()

	env := &mockEnv{vars: map[string]string{
		"FLUXID_COMMIT_ENABLED": "maybe",
	}}
	_, err := LoadEnvConfig(env)

	if err == nil {
		t.Error("Expected error for invalid FLUXID_COMMIT_ENABLED")
	}
}

//nolint:cyclop // Unit test validating all environment variable configurations
func TestLoadEnvConfig_AllFields(t *testing.T) {
	t.Parallel()

	env := &mockEnv{vars: map[string]string{
		"FLUXID_AGENT":             "codex",
		"FLUXID_ITERATIONS":        "25",
		"FLUXID_IMPLEMENT_RETRIES": "8",
		"FLUXID_COMMIT_ENABLED":    "true",
	}}
	cfg, err := LoadEnvConfig(env)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	if cfg.Agent == nil || *cfg.Agent != "codex" {
		t.Errorf("Expected Agent=codex, got: %v", cfg.Agent)
	}

	if cfg.Iterations == nil || *cfg.Iterations != 25 {
		t.Errorf("Expected Iterations=25, got: %v", cfg.Iterations)
	}

	if cfg.ImplementRetries == nil || *cfg.ImplementRetries != 8 {
		t.Errorf("Expected ImplementRetries=8, got: %v", cfg.ImplementRetries)
	}

	if cfg.CommitEnabled == nil || !*cfg.CommitEnabled {
		t.Errorf("Expected CommitEnabled=true, got: %v", cfg.CommitEnabled)
	}
}
