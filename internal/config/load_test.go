//nolint:exhaustruct,paralleltest,revive,thelper,usetesting // Test file with test data structures
package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

//nolint:cyclop,funlen // Unit test with table-driven tests for home config scenarios
func TestLoadHomeConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupHome  func(t *testing.T, homeDir string)
		wantConfig *HomeConfig
		wantErr    bool
	}{
		{
			name: "no config file",
			setupHome: func(t *testing.T, homeDir string) {
				// Don't create config file
			},
			wantConfig: nil,
			wantErr:    false,
		},
		{
			name: "valid config file",
			setupHome: func(t *testing.T, homeDir string) {
				configDir := filepath.Join(homeDir, ".fluxid")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}

				config := HomeConfig{
					Agent:            strPtr("claude"),
					Iterations:       intPtr(10),
					ImplementRetries: intPtr(5),
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantConfig: &HomeConfig{
				Agent:            strPtr("claude"),
				Iterations:       intPtr(10),
				ImplementRetries: intPtr(5),
			},
			wantErr: false,
		},
		{
			name: "invalid YAML",
			setupHome: func(t *testing.T, homeDir string) {
				configDir := filepath.Join(homeDir, ".fluxid")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}

				invalidYAML := []byte("agent: [invalid yaml structure")
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), invalidYAML, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid config values",
			setupHome: func(t *testing.T, homeDir string) {
				configDir := filepath.Join(homeDir, ".fluxid")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}

				config := HomeConfig{
					ImplementRetries: intPtr(0), // Invalid: must be >= 1
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "partial config",
			setupHome: func(t *testing.T, homeDir string) {
				configDir := filepath.Join(homeDir, ".fluxid")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}

				config := HomeConfig{
					Iterations: intPtr(25),
					// Other fields unset
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantConfig: &HomeConfig{
				Iterations: intPtr(25),
			},
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary home directory
			tmpHome := t.TempDir()
			t.Setenv("HOME", tmpHome)

			testCase.setupHome(t, tmpHome)

			got, err := LoadHomeConfig()
			if (err != nil) != testCase.wantErr {
				t.Errorf("LoadHomeConfig() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}

			if !testCase.wantErr && !equalHomeConfig(got, testCase.wantConfig) {
				t.Errorf("LoadHomeConfig() = %+v, want %+v", got, testCase.wantConfig)
			}
		})
	}
}

//nolint:cyclop,funlen // Unit test with table-driven tests for default config scenarios
func TestLoadProjectConfig(t *testing.T) {
	tests := []struct {
		name         string
		setupProject func(t *testing.T, projectDir string)
		wantConfig   *ProjectConfig
		wantErr      bool
	}{
		{
			name: "no config file",
			setupProject: func(t *testing.T, projectDir string) {
				// Don't create config file
			},
			wantConfig: nil,
			wantErr:    false,
		},
		{
			name: "valid config file",
			setupProject: func(t *testing.T, projectDir string) {
				configDir := filepath.Join(projectDir, ".fluxid")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}

				config := ProjectConfig{
					Agent:            strPtr("codex"),
					Iterations:       intPtr(15),
					ImplementRetries: intPtr(7),
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantConfig: &ProjectConfig{
				Agent:            strPtr("codex"),
				Iterations:       intPtr(15),
				ImplementRetries: intPtr(7),
			},
			wantErr: false,
		},
		{
			name: "invalid YAML",
			setupProject: func(t *testing.T, projectDir string) {
				configDir := filepath.Join(projectDir, ".fluxid")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}

				invalidYAML := []byte("agent: [invalid")
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), invalidYAML, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantConfig: nil,
			wantErr:    true,
		},
		{
			name: "invalid project config",
			setupProject: func(t *testing.T, projectDir string) {
				configDir := filepath.Join(projectDir, ".fluxid")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}

				config := ProjectConfig{
					Iterations: intPtr(0), // Invalid
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantConfig: nil,
			wantErr:    true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary project directory and change to it
			tmpProject := t.TempDir()
			originalWd, _ := os.Getwd()
			if err := os.Chdir(tmpProject); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Chdir(originalWd)
			}()

			testCase.setupProject(t, tmpProject)

			got, err := LoadProjectConfig()
			if (err != nil) != testCase.wantErr {
				t.Errorf("LoadProjectConfig() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}

			if !testCase.wantErr && !equalProjectConfig(got, testCase.wantConfig) {
				t.Errorf("LoadProjectConfig() = %+v, want %+v", got, testCase.wantConfig)
			}
		})
	}
}

func TestGetHomeConfigPath(t *testing.T) {
	t.Parallel()
	path, err := GetHomeConfigPath()
	if err != nil {
		t.Errorf("GetHomeConfigPath() unexpected error: %v", err)
	}
	if path == "" {
		t.Error("GetHomeConfigPath() returned empty path")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("GetHomeConfigPath() returned non-absolute path: %s", path)
	}
}

func TestGetProjectConfigPath(t *testing.T) {
	t.Parallel()
	path, err := GetProjectConfigPath()
	if err != nil {
		t.Errorf("GetProjectConfigPath() unexpected error: %v", err)
	}
	if path == "" {
		t.Error("GetProjectConfigPath() returned empty path")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("GetProjectConfigPath() returned non-absolute path: %s", path)
	}
}
