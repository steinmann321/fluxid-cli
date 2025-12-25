//nolint:exhaustruct // Test file with test data structures
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// os.Getwd/Chdir; long table-driven test
//
//nolint:paralleltest,usetesting,funlen // Cannot run in parallel or use t.Chdir - test uses
func TestResolveCommandFiles(t *testing.T) {
	// Cannot run in parallel because we need to change working directory
	// Create temporary directories and files for testing
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	// Paths are relative to .fluxid/ directory (not .fluxid/commands/)
	homeFluxidDir := filepath.Join(homeDir, ".fluxid")
	projectFluxidDir := filepath.Join(projectDir, ".fluxid")

	if err := os.MkdirAll(homeFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create home .fluxid dir: %v", err)
	}
	if err := os.MkdirAll(projectFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create project .fluxid dir: %v", err)
	}

	// Create test command files directly in .fluxid/ directory
	homeImpl := filepath.Join(homeFluxidDir, "implement.sh")
	homeRev := filepath.Join(homeFluxidDir, "review.sh")
	homeCommit := filepath.Join(homeFluxidDir, "commit.sh")
	projImpl := filepath.Join(projectFluxidDir, "implement.sh")
	projRev := filepath.Join(projectFluxidDir, "review.sh")
	projCommit := filepath.Join(projectFluxidDir, "commit.sh")

	// Create subdirectory for testing subdirectory path resolution
	projectScriptsDir := filepath.Join(projectFluxidDir, "scripts")
	if err := os.MkdirAll(projectScriptsDir, 0o755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}
	projScriptImpl := filepath.Join(projectScriptsDir, "implement.sh")
	projScriptRev := filepath.Join(projectScriptsDir, "review.sh")
	projScriptCommit := filepath.Join(projectScriptsDir, "commit.sh")

	for _, f := range []string{
		homeImpl, homeRev, homeCommit, projImpl, projRev, projCommit,
		projScriptImpl, projScriptRev, projScriptCommit,
	} {
		if err := os.WriteFile(f, []byte("#!/bin/bash\necho test"), 0o644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", f, err)
		}
	}

	// Note: Can't use t.Chdir/t.Setenv here as they don't work well with subtests
	origWd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(origWd)
	})
	tests := []struct {
		name          string
		projectConfig *ProjectConfig
		homeConfig    *HomeConfig
		setupWd       string // working directory to set
		setupHome     string // HOME env var to set
		wantNil       bool
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "nil configs - error required",
			projectConfig: nil,
			homeConfig:    nil,
			wantNil:       false,
			wantErr:       true,
			errMsg:        "commands section is required",
		},
		{
			name:          "configs with no commands - error required",
			projectConfig: &ProjectConfig{},
			homeConfig:    &HomeConfig{},
			wantNil:       false,
			wantErr:       true,
			errMsg:        "commands section is required",
		},
		{
			name: "project commands resolved successfully",
			projectConfig: &ProjectConfig{
				Commands: &Commands{
					Implement: strPtr("implement.sh"),
					Review:    strPtr("review.sh"),
					Commit:    strPtr("commit.sh"),
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   false,
		},
		{
			name: "home commands resolved successfully",
			homeConfig: &HomeConfig{
				Commands: &Commands{
					Implement: strPtr("implement.sh"),
					Review:    strPtr("review.sh"),
					Commit:    strPtr("commit.sh"),
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   false,
		},
		{
			name: "project commands with subdirectory paths",
			projectConfig: &ProjectConfig{
				Commands: &Commands{
					Implement: strPtr("scripts/implement.sh"),
					Review:    strPtr("scripts/review.sh"),
					Commit:    strPtr("scripts/commit.sh"),
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   false,
		},
		{
			name: "missing command file",
			projectConfig: &ProjectConfig{
				Commands: &Commands{
					Implement: strPtr("missing.sh"),
					Review:    strPtr("review.sh"),
					Commit:    strPtr("commit.sh"),
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   true,
			errMsg:    "command file not found",
		},
		{
			name: "incomplete commands - missing review",
			projectConfig: &ProjectConfig{
				Commands: &Commands{
					Implement: strPtr("implement.sh"),
					Commit:    strPtr("commit.sh"),
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   true,
			errMsg:    "commands.review: command is required",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Cannot run in parallel because resolveCommandFiles uses os.Getwd()
			// Set up environment if needed
			setupTestEnv(t, testCase.setupWd, testCase.setupHome)

			result, err := ResolveCommandFiles(testCase.projectConfig, testCase.homeConfig)
			checkResolveError(t, err, testCase.wantErr, testCase.errMsg)
			checkResolveResult(t, result, testCase.wantNil, testCase.wantErr)
		})
	}
}

//nolint:usetesting // Cannot use t.Chdir - must use os.Chdir for test setup
func setupTestEnv(t *testing.T, wd, home string) {
	t.Helper()
	if wd != "" {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}
	}
	if home != "" {
		t.Setenv("HOME", home)
	}
}

func checkResolveError(t *testing.T, err error, wantErr bool, errMsg string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Errorf("ResolveCommandFiles() expected error, got nil")
		} else if errMsg != "" && !strings.Contains(err.Error(), errMsg) {
			t.Errorf("ResolveCommandFiles() error = %v, want error containing %q", err, errMsg)
		}
	} else if err != nil {
		t.Errorf("ResolveCommandFiles() unexpected error: %v", err)
	}
}

func checkResolveResult(t *testing.T, result *ResolvedCommandFiles, wantNil, wantErr bool) {
	t.Helper()
	if wantNil && result != nil {
		t.Errorf("ResolveCommandFiles() expected nil result, got %+v", result)
	}
	if !wantNil && !wantErr && result == nil {
		t.Errorf("ResolveCommandFiles() expected non-nil result, got nil")
	}
}

//nolint:funlen // Unit test with file resolution validation
func TestResolveAndValidateCommandFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "valid.sh")
	nonRegularFile := filepath.Join(tmpDir, "dir")

	// Create a valid file
	if err := os.WriteFile(validFile, []byte("#!/bin/bash\necho test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a directory (non-regular file)
	if err := os.Mkdir(nonRegularFile, 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	tests := []struct {
		name     string
		baseDir  string
		filename *string
		cmdName  string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "nil filename",
			baseDir:  tmpDir,
			filename: nil,
			cmdName:  "implement",
			wantErr:  true,
			errMsg:   "commands.implement: command is required",
		},
		{
			name:     "empty filename",
			baseDir:  tmpDir,
			filename: strPtr(""),
			cmdName:  "review",
			wantErr:  true,
			errMsg:   "commands.review: command is required",
		},
		{
			name:     "valid file",
			baseDir:  tmpDir,
			filename: strPtr("valid.sh"),
			cmdName:  "commit",
			wantErr:  false,
		},
		{
			name:     "file not found",
			baseDir:  tmpDir,
			filename: strPtr("nonexistent.sh"),
			cmdName:  "implement",
			wantErr:  true,
			errMsg:   "command file not found",
		},
		{
			name:     "not a regular file",
			baseDir:  tmpDir,
			filename: strPtr("dir"),
			cmdName:  "review",
			wantErr:  true,
			errMsg:   "is not a regular file",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := resolveAndValidateCommandFile(testCase.baseDir, testCase.filename, testCase.cmdName)

			checkValidateFileError(t, err, result, testCase.wantErr, testCase.errMsg)
		})
	}
}

func checkValidateFileError(t *testing.T, err error, result string, wantErr bool, errMsg string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Errorf("resolveAndValidateCommandFile() expected error, got nil")
			return
		}
		if errMsg != "" && !strings.Contains(err.Error(), errMsg) {
			t.Errorf("resolveAndValidateCommandFile() error = %v, want error containing %q", err, errMsg)
		}
		return
	}
	if err != nil {
		t.Errorf("resolveAndValidateCommandFile() unexpected error: %v", err)
		return
	}
	if result == "" {
		t.Errorf("resolveAndValidateCommandFile() expected non-empty result, got empty string")
	}
}

//nolint:funlen // Unit test with command validation scenarios
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
