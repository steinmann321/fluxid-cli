//nolint:exhaustruct,paralleltest,revive,thelper,usetesting // Test file with test data structures
package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

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
					Agent:            strPtr("test-agent"),
					Iterations:       intPtr(10),
					ImplementRetries: intPtr(5),
					CommitEnabled:    boolPtr(false),
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantConfig: &HomeConfig{
				Agent:            strPtr("test-agent"),
				Iterations:       intPtr(10),
				ImplementRetries: intPtr(5),
				CommitEnabled:    boolPtr(false),
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary home directory
			tmpHome := t.TempDir()
			originalHome := os.Getenv("HOME")
			t.Setenv("HOME", tmpHome)
			defer func() {
				_ = os.Setenv("HOME", originalHome)
			}()

			tt.setupHome(t, tmpHome)

			got, err := LoadHomeConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadHomeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !equalHomeConfig(got, tt.wantConfig) {
				t.Errorf("LoadHomeConfig() = %+v, want %+v", got, tt.wantConfig)
			}
		})
	}
}

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
					Agent:            strPtr("project-agent"),
					Iterations:       intPtr(15),
					ImplementRetries: intPtr(7),
					CommitEnabled:    boolPtr(true),
				}

				data, _ := yaml.Marshal(config)
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantConfig: &ProjectConfig{
				Agent:            strPtr("project-agent"),
				Iterations:       intPtr(15),
				ImplementRetries: intPtr(7),
				CommitEnabled:    boolPtr(true),
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary project directory and change to it
			tmpProject := t.TempDir()
			originalWd, _ := os.Getwd()
			if err := os.Chdir(tmpProject); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Chdir(originalWd)
			}()

			tt.setupProject(t, tmpProject)

			got, err := LoadProjectConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadProjectConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !equalProjectConfig(got, tt.wantConfig) {
				t.Errorf("LoadProjectConfig() = %+v, want %+v", got, tt.wantConfig)
			}
		})
	}
}
