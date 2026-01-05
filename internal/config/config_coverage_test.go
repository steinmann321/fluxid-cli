//nolint:exhaustruct // Test file with test data structures
package config

import (
	"strings"
	"testing"
)

//nolint:funlen // Comprehensive validation test with multiple cases
func TestValidateCustomConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *CustomConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: &CustomConfig{
				Agent:            strPtr("claude"),
				ImplementRetries: intPtr(5),
				CommitRetries:    intPtr(10),
				Iterations:       intPtr(20),
			},
			wantErr: false,
		},
		{
			name:    "valid empty config",
			config:  &CustomConfig{},
			wantErr: false,
		},
		{
			name: "invalid implement retries zero",
			config: &CustomConfig{
				ImplementRetries: intPtr(0),
			},
			wantErr: true,
			errMsg:  "got 0: implement_retries must be a positive integer",
		},
		{
			name: "invalid implement retries negative",
			config: &CustomConfig{
				ImplementRetries: intPtr(-5),
			},
			wantErr: true,
			errMsg:  "got -5: implement_retries must be a positive integer",
		},
		{
			name: "invalid commit retries zero",
			config: &CustomConfig{
				CommitRetries: intPtr(0),
			},
			wantErr: true,
			errMsg:  "got 0: commit_retries must be a positive integer",
		},
		{
			name: "invalid commit retries negative",
			config: &CustomConfig{
				CommitRetries: intPtr(-10),
			},
			wantErr: true,
			errMsg:  "got -10: commit_retries must be a positive integer",
		},
		{
			name: "invalid iterations zero",
			config: &CustomConfig{
				Iterations: intPtr(0),
			},
			wantErr: true,
			errMsg:  "got 0: iterations must be a positive integer",
		},
		{
			name: "invalid iterations negative",
			config: &CustomConfig{
				Iterations: intPtr(-3),
			},
			wantErr: true,
			errMsg:  "got -3: iterations must be a positive integer",
		},
		{
			name: "invalid empty agent",
			config: &CustomConfig{
				Agent: strPtr(""),
			},
			wantErr: true,
			errMsg:  "agent cannot be empty",
		},
		{
			name: "invalid unsupported agent",
			config: &CustomConfig{
				Agent: strPtr("gpt4"),
			},
			wantErr: true,
			errMsg:  "\"gpt4\"",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateCustomConfig(testCase.config)
			if testCase.wantErr {
				if err == nil {
					t.Errorf("validateCustomConfig() expected error containing %q, got nil", testCase.errMsg)
				} else if testCase.errMsg != "" && !strings.Contains(err.Error(), testCase.errMsg) {
					t.Errorf("validateCustomConfig() error = %v, want error containing %q", err, testCase.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateCustomConfig() unexpected error: %v", err)
			}
		})
	}
}

//nolint:funlen // Comprehensive validation test for commands
func TestValidateCommandsPartialSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmds    *Commands
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil commands",
			cmds:    nil,
			wantErr: false,
		},
		{
			name:    "empty commands struct",
			cmds:    &Commands{},
			wantErr: false,
		},
		{
			name: "all commands specified",
			cmds: &Commands{
				Implement: strPtr("impl.md"),
				Review:    strPtr("review.md"),
				Commit:    strPtr("commit.md"),
			},
			wantErr: false,
		},
		{
			name: "only implement specified",
			cmds: &Commands{
				Implement: strPtr("impl.md"),
			},
			wantErr: true,
			errMsg:  "commands.review is required",
		},
		{
			name: "only review specified",
			cmds: &Commands{
				Review: strPtr("review.md"),
			},
			wantErr: true,
			errMsg:  "commands.implement is required",
		},
		{
			name: "only commit specified",
			cmds: &Commands{
				Commit: strPtr("commit.md"),
			},
			wantErr: true,
			errMsg:  "commands.implement is required",
		},
		{
			name: "implement and review but no commit",
			cmds: &Commands{
				Implement: strPtr("impl.md"),
				Review:    strPtr("review.md"),
			},
			wantErr: true,
			errMsg:  "commands.commit is required",
		},
		{
			name: "implement and commit but no review",
			cmds: &Commands{
				Implement: strPtr("impl.md"),
				Commit:    strPtr("commit.md"),
			},
			wantErr: true,
			errMsg:  "commands.review is required",
		},
		{
			name: "review and commit but no implement",
			cmds: &Commands{
				Review: strPtr("review.md"),
				Commit: strPtr("commit.md"),
			},
			wantErr: true,
			errMsg:  "commands.implement is required",
		},
		{
			name: "empty string values treated as not specified",
			cmds: &Commands{
				Implement: strPtr(""),
				Review:    strPtr(""),
				Commit:    strPtr(""),
			},
			wantErr: false,
		},
		{
			name: "one empty string requires all",
			cmds: &Commands{
				Implement: strPtr("impl.md"),
				Review:    strPtr(""),
				Commit:    strPtr("commit.md"),
			},
			wantErr: true,
			errMsg:  "commands.review is required",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateCommands(testCase.cmds)
			if testCase.wantErr {
				if err == nil {
					t.Errorf("validateCommands() expected error containing %q, got nil", testCase.errMsg)
				} else if testCase.errMsg != "" && !strings.Contains(err.Error(), testCase.errMsg) {
					t.Errorf("validateCommands() error = %v, want error containing %q", err, testCase.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateCommands() unexpected error: %v", err)
			}
		})
	}
}
