//nolint:exhaustruct // Test file with test data structures
package config

import (
	"strings"
	"testing"
)

//nolint:funlen // Unit test with table-driven home config validation
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
			errMsg:  "got 0: implement_retries must be a positive integer",
		},
		{
			name: "invalid implement retries negative",
			config: &HomeConfig{
				ImplementRetries: intPtr(-1),
			},
			wantErr: true,
			errMsg:  "got -1: implement_retries must be a positive integer",
		},
		{
			name: "invalid iterations zero",
			config: &HomeConfig{
				Iterations: intPtr(0),
			},
			wantErr: true,
			errMsg:  "got 0: iterations must be a positive integer",
		},
		{
			name: "invalid iterations negative",
			config: &HomeConfig{
				Iterations: intPtr(-5),
			},
			wantErr: true,
			errMsg:  "got -5: iterations must be a positive integer",
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateHomeConfig(testCase.config)
			if testCase.wantErr {
				if err == nil {
					t.Errorf("validateHomeConfig() expected error containing %q, got nil", testCase.errMsg)
				} else if testCase.errMsg != "" && err.Error()[:len(testCase.errMsg)] != testCase.errMsg {
					t.Errorf("validateHomeConfig() error = %v, want error containing %q", err, testCase.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateHomeConfig() unexpected error: %v", err)
			}
		})
	}
}

//nolint:funlen // Unit test with table-driven project config validation
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
			errMsg:  "got 0: implement_retries must be a positive integer",
		},
		{
			name: "invalid iterations",
			config: &ProjectConfig{
				Iterations: intPtr(-1),
			},
			wantErr: true,
			errMsg:  "got -1: iterations must be a positive integer",
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateProjectConfig(testCase.config)
			if testCase.wantErr {
				if err == nil {
					t.Errorf("validateProjectConfig() expected error, got nil")
				} else if testCase.errMsg != "" && err.Error()[:len(testCase.errMsg)] != testCase.errMsg {
					t.Errorf("validateProjectConfig() error = %v, want error containing %q", err, testCase.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateProjectConfig() unexpected error: %v", err)
			}
		})
	}
}

//nolint:funlen // Unit test with comprehensive config resolution tests
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := Resolve(
				testCase.projectConfig,
				testCase.homeConfig,
				nil,
				nil,
				testCase.cliIterations,
				testCase.cliImplementRetries,
				nil,
			)

			if result.Agent != testCase.wantAgent {
				t.Errorf("Agent = %v, want %v", result.Agent, testCase.wantAgent)
			}
			if result.Iterations != testCase.wantIterations {
				t.Errorf("Iterations = %v, want %v", result.Iterations, testCase.wantIterations)
			}
			if result.ImplementRetries != testCase.wantImplementRetries {
				t.Errorf("ImplementRetries = %v, want %v", result.ImplementRetries, testCase.wantImplementRetries)
			}
			if result.CommitEnabled != testCase.wantCommitEnabled {
				t.Errorf("CommitEnabled = %v, want %v", result.CommitEnabled, testCase.wantCommitEnabled)
			}

			// Check sources (for home/project sources, check prefix since they include file paths)
			if !strings.HasPrefix(result.Sources["agent"], testCase.wantAgentSource) {
				t.Errorf("Agent source = %v, want prefix %v", result.Sources["agent"], testCase.wantAgentSource)
			}
			if !strings.HasPrefix(result.Sources["iterations"], testCase.wantIterationsSource) {
				t.Errorf("Iterations source = %v, want prefix %v", result.Sources["iterations"], testCase.wantIterationsSource)
			}
			if !strings.HasPrefix(result.Sources["implement_retries"], testCase.wantRetriesSource) {
				t.Errorf(
					"ImplementRetries source = %v, want prefix %v",
					result.Sources["implement_retries"],
					testCase.wantRetriesSource,
				)
			}
			if !strings.HasPrefix(result.Sources["commit_enabled"], testCase.wantCommitEnabledSource) {
				t.Errorf(
					"CommitEnabled source = %v, want prefix %v",
					result.Sources["commit_enabled"],
					testCase.wantCommitEnabledSource,
				)
			}
		})
	}
}
