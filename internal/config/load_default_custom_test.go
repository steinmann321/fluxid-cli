//nolint:exhaustruct,paralleltest,thelper,usetesting // Test file with test data structures
package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

//nolint:cyclop,funlen,gocognit // Unit test with table-driven tests for home config scenarios
func TestLoadDefaultConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func(t *testing.T) (homeDir, projectDir string)
		wantProject *ProjectConfig
		wantHome    *HomeConfig
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name: "neither config exists - error",
			setupEnv: func(t *testing.T) (string, string) {
				tmpHome := t.TempDir()
				tmpProject := t.TempDir()
				return tmpHome, tmpProject
			},
			wantProject: nil,
			wantHome:    nil,
			wantErr:     true,
			wantErrMsg:  "at least one default config must exist",
		},
		{
			name: "only home config exists",
			setupEnv: func(t *testing.T) (string, string) {
				tmpHome := t.TempDir()
				tmpProject := t.TempDir()

				configDir := filepath.Join(tmpHome, ".fluxid")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}

				config := HomeConfig{
					Agent:      strPtr("claude"),
					Iterations: intPtr(10),
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0o600); err != nil {
					t.Fatal(err)
				}

				return tmpHome, tmpProject
			},
			wantProject: nil,
			wantHome: &HomeConfig{
				Agent:      strPtr("claude"),
				Iterations: intPtr(10),
			},
			wantErr: false,
		},
		{
			name: "only project config exists",
			setupEnv: func(t *testing.T) (string, string) {
				tmpHome := t.TempDir()
				tmpProject := t.TempDir()

				configDir := filepath.Join(tmpProject, ".fluxid")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}

				config := ProjectConfig{
					Agent:      strPtr("codex"),
					Iterations: intPtr(15),
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0o600); err != nil {
					t.Fatal(err)
				}

				return tmpHome, tmpProject
			},
			wantProject: &ProjectConfig{
				Agent:      strPtr("codex"),
				Iterations: intPtr(15),
			},
			wantHome: nil,
			wantErr:  false,
		},
		{
			name: "both configs exist",
			setupEnv: func(t *testing.T) (string, string) {
				tmpHome := t.TempDir()
				tmpProject := t.TempDir()

				// Create home config
				homeConfigDir := filepath.Join(tmpHome, ".fluxid")
				if err := os.MkdirAll(homeConfigDir, 0o755); err != nil {
					t.Fatal(err)
				}

				homeConfig := HomeConfig{
					Agent: strPtr("claude"),
				}
				homeData, _ := yaml.Marshal(homeConfig)
				if err := os.WriteFile(filepath.Join(homeConfigDir, "config.yaml"), homeData, 0o600); err != nil {
					t.Fatal(err)
				}

				// Create project config
				projectConfigDir := filepath.Join(tmpProject, ".fluxid")
				if err := os.MkdirAll(projectConfigDir, 0o755); err != nil {
					t.Fatal(err)
				}

				projectConfig := ProjectConfig{
					Iterations: intPtr(25),
				}
				projectData, _ := yaml.Marshal(projectConfig)
				if err := os.WriteFile(filepath.Join(projectConfigDir, "config.yaml"), projectData, 0o600); err != nil {
					t.Fatal(err)
				}

				return tmpHome, tmpProject
			},
			wantProject: &ProjectConfig{
				Iterations: intPtr(25),
			},
			wantHome: &HomeConfig{
				Agent: strPtr("claude"),
			},
			wantErr: false,
		},
		{
			name: "project config invalid YAML - fail fast",
			setupEnv: func(t *testing.T) (string, string) {
				tmpHome := t.TempDir()
				tmpProject := t.TempDir()

				// Create valid home config
				homeConfigDir := filepath.Join(tmpHome, ".fluxid")
				if err := os.MkdirAll(homeConfigDir, 0o755); err != nil {
					t.Fatal(err)
				}

				homeConfig := HomeConfig{
					Agent: strPtr("claude"),
				}
				homeData, _ := yaml.Marshal(homeConfig)
				if err := os.WriteFile(filepath.Join(homeConfigDir, "config.yaml"), homeData, 0o600); err != nil {
					t.Fatal(err)
				}

				// Create invalid project config
				projectConfigDir := filepath.Join(tmpProject, ".fluxid")
				if err := os.MkdirAll(projectConfigDir, 0o755); err != nil {
					t.Fatal(err)
				}

				invalidYAML := []byte("agent: [invalid")
				if err := os.WriteFile(filepath.Join(projectConfigDir, "config.yaml"), invalidYAML, 0o600); err != nil {
					t.Fatal(err)
				}

				return tmpHome, tmpProject
			},
			wantProject: nil,
			wantHome:    nil,
			wantErr:     true,
			wantErrMsg:  "failed to parse project config",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			homeDir, projectDir := testCase.setupEnv(t)

			// Set HOME and change to project directory
			t.Setenv("HOME", homeDir)
			originalWd, _ := os.Getwd()
			if err := os.Chdir(projectDir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Chdir(originalWd)
			}()

			gotProject, gotHome, err := LoadDefaultConfig()

			if (err != nil) != testCase.wantErr {
				t.Errorf("LoadDefaultConfig() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}

			if testCase.wantErr && testCase.wantErrMsg != "" {
				if err == nil || !contains(err.Error(), testCase.wantErrMsg) {
					t.Errorf("LoadDefaultConfig() error = %v, want error containing %q", err, testCase.wantErrMsg)
				}
				return
			}

			if !testCase.wantErr {
				if !equalProjectConfig(gotProject, testCase.wantProject) {
					t.Errorf("LoadDefaultConfig() project = %+v, want %+v", gotProject, testCase.wantProject)
				}
				if !equalHomeConfig(gotHome, testCase.wantHome) {
					t.Errorf("LoadDefaultConfig() home = %+v, want %+v", gotHome, testCase.wantHome)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

//nolint:cyclop,funlen,gocognit // Unit test with table-driven tests for custom config scenarios
func TestLoadCustomConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupFile  func(t *testing.T) string
		wantConfig *CustomConfig
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "valid custom config",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "custom.yaml")

				config := CustomConfig{
					Agent:      strPtr("codex"),
					Iterations: intPtr(15),
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(configPath, data, 0o600); err != nil {
					t.Fatal(err)
				}

				return configPath
			},
			wantConfig: &CustomConfig{
				Agent:      strPtr("codex"),
				Iterations: intPtr(15),
			},
			wantErr: false,
		},
		{
			name: "config file not found",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent.yaml")
			},
			wantConfig: nil,
			wantErr:    true,
			wantErrMsg: "config file not found",
		},
		{
			name: "invalid YAML",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "invalid.yaml")

				invalidYAML := []byte("agent: [invalid")
				if err := os.WriteFile(configPath, invalidYAML, 0o600); err != nil {
					t.Fatal(err)
				}

				return configPath
			},
			wantConfig: nil,
			wantErr:    true,
			wantErrMsg: "failed to parse",
		},
		{
			name: "partial config",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "partial.yaml")

				config := CustomConfig{
					Agent: strPtr("opencode"),
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(configPath, data, 0o600); err != nil {
					t.Fatal(err)
				}

				return configPath
			},
			wantConfig: &CustomConfig{
				Agent: strPtr("opencode"),
			},
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			configPath := testCase.setupFile(t)

			got, configDir, err := LoadCustomConfig(configPath)

			if (err != nil) != testCase.wantErr {
				t.Errorf("LoadCustomConfig() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}

			if testCase.wantErr && testCase.wantErrMsg != "" {
				if err == nil || !contains(err.Error(), testCase.wantErrMsg) {
					t.Errorf("LoadCustomConfig() error = %v, want error containing %q", err, testCase.wantErrMsg)
				}
				return
			}

			if !testCase.wantErr {
				if got == nil {
					t.Error("LoadCustomConfig() returned nil config")
					return
				}

				if !equalCustomConfig(got, testCase.wantConfig) {
					t.Errorf("LoadCustomConfig() = %+v, want %+v", got, testCase.wantConfig)
				}

				if configDir == "" {
					t.Error("LoadCustomConfig() returned empty config directory")
				}
			}
		})
	}
}

func equalCustomConfig(actual, expected *CustomConfig) bool {
	if actual == nil && expected == nil {
		return true
	}
	if actual == nil || expected == nil {
		return false
	}

	return equalStrPtr(actual.Agent, expected.Agent) &&
		equalIntPtr(actual.Iterations, expected.Iterations) &&
		equalIntPtr(actual.ImplementRetries, expected.ImplementRetries)
}
