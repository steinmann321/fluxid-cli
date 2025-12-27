//nolint:exhaustruct // Test file with test data structures
package config

import (
	"strings"
	"testing"
)

//nolint:funlen // Table-driven test with comprehensive test cases
func TestValidateCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmds    *Commands
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil commands is valid",
			cmds:    nil,
			wantErr: false,
		},
		{
			name:    "all nil pointers is valid",
			cmds:    &Commands{},
			wantErr: false,
		},
		{
			name: "all three specified is valid",
			cmds: &Commands{
				Implement: strPtr("implement.sh"),
				Review:    strPtr("review.sh"),
				Commit:    strPtr("commit.sh"),
			},
			wantErr: false,
		},
		{
			name: "missing implement",
			cmds: &Commands{
				Review: strPtr("review.sh"),
				Commit: strPtr("commit.sh"),
			},
			wantErr: true,
			errMsg:  "commands.implement is required",
		},
		{
			name: "missing review",
			cmds: &Commands{
				Implement: strPtr("implement.sh"),
				Commit:    strPtr("commit.sh"),
			},
			wantErr: true,
			errMsg:  "commands.review is required",
		},
		{
			name: "missing commit",
			cmds: &Commands{
				Implement: strPtr("implement.sh"),
				Review:    strPtr("review.sh"),
			},
			wantErr: true,
			errMsg:  "commands.commit is required",
		},
		{
			name: "only implement specified",
			cmds: &Commands{
				Implement: strPtr("implement.sh"),
			},
			wantErr: true,
			errMsg:  "commands.review is required",
		},
		{
			name: "empty string values",
			cmds: &Commands{
				Implement: strPtr(""),
				Review:    strPtr(""),
				Commit:    strPtr(""),
			},
			wantErr: false, // Empty strings are treated as not specified
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
