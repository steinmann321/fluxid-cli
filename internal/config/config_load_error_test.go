package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAgentUnsupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		agent   string
		wantErr bool
	}{
		{
			name:    "unsupported agent gpt4",
			agent:   "gpt4",
			wantErr: true,
		},
		{
			name:    "unsupported agent gemini",
			agent:   "gemini",
			wantErr: true,
		},
		{
			name:    "unsupported agent random",
			agent:   "random-ai",
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateAgent(testCase.agent)
			if testCase.wantErr {
				if err == nil {
					t.Error("ValidateAgent() expected error, got nil")
				}
			} else if err != nil {
				t.Errorf("ValidateAgent() unexpected error: %v", err)
			}
		})
	}
}

func TestLoadCustomConfigInvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: [unclosed"), 0o600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, _, err = LoadCustomConfig(configPath)
	if err == nil {
		t.Error("LoadCustomConfig() expected error for invalid YAML, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("LoadCustomConfig() error = %v, want error containing 'failed to parse'", err)
	}
}

func TestLoadCustomConfigValidationError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-values.yaml")

	// Write config with invalid values
	content := `
implement_retries: -5
commit_retries: 0
`
	err := os.WriteFile(configPath, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, _, err = LoadCustomConfig(configPath)
	if err == nil {
		t.Error("LoadCustomConfig() expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid config") {
		t.Errorf("LoadCustomConfig() error = %v, want error containing 'invalid config'", err)
	}
}

func TestLoadCustomConfigNotFound(t *testing.T) {
	t.Parallel()

	_, _, err := LoadCustomConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadCustomConfig() expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "config file not found") {
		t.Errorf("LoadCustomConfig() error = %v, want error containing 'config file not found'", err)
	}
}

func TestGetHomeConfigPathError(t *testing.T) {
	t.Parallel()

	// This test is tricky because os.UserHomeDir() is hard to make fail
	// We can at least call the function to increase coverage
	path, err := GetHomeConfigPath()
	if err != nil {
		// If it fails, that's actually what we want to test
		if !strings.Contains(err.Error(), "failed to get home directory") {
			t.Errorf("GetHomeConfigPath() error = %v, want error containing 'failed to get home directory'", err)
		}
	} else {
		// If it succeeds, verify the path is reasonable
		if path == "" {
			t.Error("GetHomeConfigPath() returned empty path")
		}
		if !strings.HasSuffix(path, filepath.Join(".fluxid", "config.yaml")) {
			t.Errorf("GetHomeConfigPath() = %v, want path ending with .fluxid/config.yaml", path)
		}
	}
}

func TestGetProjectConfigPathError(t *testing.T) {
	t.Parallel()

	// Similar to GetHomeConfigPath, os.Getwd() is hard to make fail
	// We can at least call the function to increase coverage
	path, err := GetProjectConfigPath()
	if err != nil {
		// If it fails, that's what we want to test
		if !strings.Contains(err.Error(), "failed to get current directory") {
			t.Errorf("GetProjectConfigPath() error = %v, want error containing 'failed to get current directory'", err)
		}
	} else {
		// If it succeeds, verify the path is reasonable
		if path == "" {
			t.Error("GetProjectConfigPath() returned empty path")
		}
		if !strings.HasSuffix(path, filepath.Join(".fluxid", "config.yaml")) {
			t.Errorf("GetProjectConfigPath() = %v, want path ending with .fluxid/config.yaml", path)
		}
	}
}
