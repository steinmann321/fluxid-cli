//nolint:exhaustruct // Test file with test data structures
package config

import (
	"strings"
	"testing"
)

func TestValidateHomeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *HomeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: &HomeConfig{
				Agent:            strPtr("claude"),
				ImplementRetries: intPtr(5),
				Iterations:       intPtr(10),
				CommitEnabled:    boolPtr(true),
			},
			wantErr: false,
		},
		{
			name:    "valid empty config",
			config:  &HomeConfig{},
			wantErr: false,
		},
		{
			name: "invalid implement retries zero",
			config: &HomeConfig{
				ImplementRetries: intPtr(0),
			},
			wantErr: true,
			errMsg:  "implement_retries must be a positive integer",
		},
		{
			name: "invalid implement retries negative",
			config: &HomeConfig{
				ImplementRetries: intPtr(-1),
			},
			wantErr: true,
			errMsg:  "implement_retries must be a positive integer",
		},
		{
			name: "invalid iterations zero",
			config: &HomeConfig{
				Iterations: intPtr(0),
			},
			wantErr: true,
			errMsg:  "iterations must be a positive integer",
		},
		{
			name: "invalid iterations negative",
			config: &HomeConfig{
				Iterations: intPtr(-5),
			},
			wantErr: true,
			errMsg:  "iterations must be a positive integer",
		},
		{
			name: "invalid empty agent",
			config: &HomeConfig{
				Agent: strPtr(""),
			},
			wantErr: true,
			errMsg:  "agent cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateHomeConfig(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateHomeConfig() expected error containing %q, got nil", tt.errMsg)
				} else if tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("validateHomeConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateHomeConfig() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateProjectConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *ProjectConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &ProjectConfig{
				Agent:            strPtr("opencode"),
				ImplementRetries: intPtr(3),
				Iterations:       intPtr(15),
				CommitEnabled:    boolPtr(false),
			},
			wantErr: false,
		},
		{
			name:    "valid empty config",
			config:  &ProjectConfig{},
			wantErr: false,
		},
		{
			name: "invalid implement retries",
			config: &ProjectConfig{
				ImplementRetries: intPtr(0),
			},
			wantErr: true,
			errMsg:  "implement_retries must be a positive integer",
		},
		{
			name: "invalid iterations",
			config: &ProjectConfig{
				Iterations: intPtr(-1),
			},
			wantErr: true,
			errMsg:  "iterations must be a positive integer",
		},
		{
			name: "invalid empty agent",
			config: &ProjectConfig{
				Agent: strPtr(""),
			},
			wantErr: true,
			errMsg:  "agent cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateProjectConfig(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateProjectConfig() expected error, got nil")
				} else if tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("validateProjectConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateProjectConfig() unexpected error: %v", err)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		projectConfig           *ProjectConfig
		homeConfig              *HomeConfig
		cliIterations           *int
		cliImplementRetries     *int
		wantAgent               string
		wantIterations          int
		wantImplementRetries    int
		wantCommitEnabled       bool
		wantAgentSource         string
		wantIterationsSource    string
		wantRetriesSource       string
		wantCommitEnabledSource string
	}{
		{
			name:                    "all defaults",
			projectConfig:           nil,
			homeConfig:              nil,
			cliIterations:           nil,
			cliImplementRetries:     nil,
			wantAgent:               DefaultAgent,
			wantIterations:          DefaultIterations,
			wantImplementRetries:    DefaultImplementRetries,
			wantCommitEnabled:       DefaultCommitEnabled,
			wantAgentSource:         SourceDefault,
			wantIterationsSource:    SourceDefault,
			wantRetriesSource:       SourceDefault,
			wantCommitEnabledSource: SourceDefault,
		},
		{
			name:          "home config only",
			projectConfig: nil,
			homeConfig: &HomeConfig{
				Agent:            strPtr("codex"),
				Iterations:       intPtr(25),
				ImplementRetries: intPtr(5),
				CommitEnabled:    boolPtr(false),
			},
			cliIterations:           nil,
			cliImplementRetries:     nil,
			wantAgent:               "codex",
			wantIterations:          25,
			wantImplementRetries:    5,
			wantCommitEnabled:       false,
			wantAgentSource:         SourceHome,
			wantIterationsSource:    SourceHome,
			wantRetriesSource:       SourceHome,
			wantCommitEnabledSource: SourceHome,
		},
		{
			name: "project overrides home",
			projectConfig: &ProjectConfig{
				Agent:            strPtr("opencode"),
				Iterations:       intPtr(30),
				ImplementRetries: intPtr(7),
				CommitEnabled:    boolPtr(true),
			},
			homeConfig: &HomeConfig{
				Agent:            strPtr("claude"),
				Iterations:       intPtr(15),
				ImplementRetries: intPtr(3),
				CommitEnabled:    boolPtr(false),
			},
			cliIterations:           nil,
			cliImplementRetries:     nil,
			wantAgent:               "opencode",
			wantIterations:          30,
			wantImplementRetries:    7,
			wantCommitEnabled:       true,
			wantAgentSource:         SourceProject,
			wantIterationsSource:    SourceProject,
			wantRetriesSource:       SourceProject,
			wantCommitEnabledSource: SourceProject,
		},
		{
			name: "CLI overrides all for iterations and retries",
			projectConfig: &ProjectConfig{
				Agent:            strPtr("codex"),
				Iterations:       intPtr(99),
				ImplementRetries: intPtr(99),
				CommitEnabled:    boolPtr(true),
			},
			homeConfig: &HomeConfig{
				Agent:            strPtr("claude"),
				Iterations:       intPtr(88),
				ImplementRetries: intPtr(88),
				CommitEnabled:    boolPtr(false),
			},
			cliIterations:           intPtr(5),
			cliImplementRetries:     intPtr(2),
			wantAgent:               "codex", // CLI doesn't override agent yet
			wantIterations:          5,
			wantImplementRetries:    2,
			wantCommitEnabled:       true, // CLI doesn't override commit_enabled yet
			wantAgentSource:         SourceProject,
			wantIterationsSource:    SourceCLI,
			wantRetriesSource:       SourceCLI,
			wantCommitEnabledSource: SourceProject,
		},
		{
			name: "partial configs merge correctly",
			projectConfig: &ProjectConfig{
				Iterations: intPtr(50),
				// Agent, ImplementRetries, CommitEnabled unset
			},
			homeConfig: &HomeConfig{
				Agent:            strPtr("fallback-agent"),
				ImplementRetries: intPtr(8),
				// Iterations and CommitEnabled unset
			},
			cliIterations:           nil,
			cliImplementRetries:     nil,
			wantAgent:               "fallback-agent",     // from home
			wantIterations:          50,                   // from project
			wantImplementRetries:    8,                    // from home
			wantCommitEnabled:       DefaultCommitEnabled, // default
			wantAgentSource:         SourceHome,
			wantIterationsSource:    SourceProject,
			wantRetriesSource:       SourceHome,
			wantCommitEnabledSource: SourceDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Resolve(tt.projectConfig, tt.homeConfig, nil, nil, tt.cliIterations, tt.cliImplementRetries, nil)

			if result.Agent != tt.wantAgent {
				t.Errorf("Agent = %v, want %v", result.Agent, tt.wantAgent)
			}
			if result.Iterations != tt.wantIterations {
				t.Errorf("Iterations = %v, want %v", result.Iterations, tt.wantIterations)
			}
			if result.ImplementRetries != tt.wantImplementRetries {
				t.Errorf("ImplementRetries = %v, want %v", result.ImplementRetries, tt.wantImplementRetries)
			}
			if result.CommitEnabled != tt.wantCommitEnabled {
				t.Errorf("CommitEnabled = %v, want %v", result.CommitEnabled, tt.wantCommitEnabled)
			}

			// Check sources (for home/project sources, check prefix since they include file paths)
			if !strings.HasPrefix(result.Sources["agent"], tt.wantAgentSource) {
				t.Errorf("Agent source = %v, want prefix %v", result.Sources["agent"], tt.wantAgentSource)
			}
			if !strings.HasPrefix(result.Sources["iterations"], tt.wantIterationsSource) {
				t.Errorf("Iterations source = %v, want prefix %v", result.Sources["iterations"], tt.wantIterationsSource)
			}
			if !strings.HasPrefix(result.Sources["implement_retries"], tt.wantRetriesSource) {
				t.Errorf("ImplementRetries source = %v, want prefix %v", result.Sources["implement_retries"], tt.wantRetriesSource)
			}
			if !strings.HasPrefix(result.Sources["commit_enabled"], tt.wantCommitEnabledSource) {
				t.Errorf("CommitEnabled source = %v, want prefix %v", result.Sources["commit_enabled"], tt.wantCommitEnabledSource)
			}
		})
	}
}
