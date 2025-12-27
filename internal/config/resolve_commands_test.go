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
	// Command file paths must be absolute (no relative paths allowed)
	homeFluxidDir := filepath.Join(homeDir, ".fluxid")
	projectFluxidDir := filepath.Join(projectDir, ".fluxid")

	if err := os.MkdirAll(homeFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create home .fluxid dir: %v", err)
	}
	if err := os.MkdirAll(projectFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create project .fluxid dir: %v", err)
	}

	// Create test command files in commands/ subdirectory
	homeCommandsDir := filepath.Join(homeFluxidDir, "commands")
	projectCommandsDir := filepath.Join(projectFluxidDir, "commands")
	if err := os.MkdirAll(homeCommandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create home commands dir: %v", err)
	}
	if err := os.MkdirAll(projectCommandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create project commands dir: %v", err)
	}

	homeImpl := filepath.Join(homeCommandsDir, "implement.sh")
	homeRev := filepath.Join(homeCommandsDir, "review.sh")
	homeCommit := filepath.Join(homeCommandsDir, "commit.sh")
	projImpl := filepath.Join(projectCommandsDir, "implement.sh")
	projRev := filepath.Join(projectCommandsDir, "review.sh")
	projCommit := filepath.Join(projectCommandsDir, "commit.sh")

	// Create alternate location for testing absolute paths
	altCommandsDir := filepath.Join(projectCommandsDir, "alt")
	if err := os.MkdirAll(altCommandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create alt commands dir: %v", err)
	}
	projAltImpl := filepath.Join(altCommandsDir, "implement.sh")
	projAltRev := filepath.Join(altCommandsDir, "review.sh")
	projAltCommit := filepath.Join(altCommandsDir, "commit.sh")

	for _, f := range []string{
		homeImpl, homeRev, homeCommit, projImpl, projRev, projCommit,
		projAltImpl, projAltRev, projAltCommit,
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
			name: "project commands with absolute paths",
			projectConfig: &ProjectConfig{
				Commands: &Commands{
					Implement: &projImpl,
					Review:    &projRev,
					Commit:    &projCommit,
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   false,
		},
		{
			name: "home commands with absolute paths",
			homeConfig: &HomeConfig{
				Commands: &Commands{
					Implement: &homeImpl,
					Review:    &homeRev,
					Commit:    &homeCommit,
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   false,
		},
		{
			name: "project commands with absolute paths in alternate location",
			projectConfig: &ProjectConfig{
				Commands: &Commands{
					Implement: &projAltImpl,
					Review:    &projAltRev,
					Commit:    &projAltCommit,
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   false,
		},
		{
			name: "missing command file with absolute path",
			projectConfig: &ProjectConfig{
				Commands: &Commands{
					Implement: strPtr(filepath.Join(projectCommandsDir, "missing.sh")),
					Review:    &projRev,
					Commit:    &projCommit,
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
					Implement: &projImpl,
					Commit:    &projCommit,
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   true,
			errMsg:    "commands.review: command is required",
		},
		{
			name: "relative path should fail",
			projectConfig: &ProjectConfig{
				Commands: &Commands{
					Implement: strPtr("commands/implement.sh"),
					Review:    strPtr("commands/review.sh"),
					Commit:    strPtr("commands/commit.sh"),
				},
			},
			setupWd:   projectDir,
			setupHome: homeDir,
			wantNil:   false,
			wantErr:   true,
			errMsg:    "must be absolute path",
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
		filename *string
		cmdName  string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "nil filename",
			filename: nil,
			cmdName:  "implement",
			wantErr:  true,
			errMsg:   "commands.implement: command is required",
		},
		{
			name:     "empty filename",
			filename: strPtr(""),
			cmdName:  "review",
			wantErr:  true,
			errMsg:   "commands.review: command is required",
		},
		{
			name:     "valid absolute path",
			filename: &validFile,
			cmdName:  "commit",
			wantErr:  false,
		},
		{
			name:     "absolute path not found",
			filename: strPtr(filepath.Join(tmpDir, "nonexistent.sh")),
			cmdName:  "implement",
			wantErr:  true,
			errMsg:   "command file not found",
		},
		{
			name:     "absolute path to directory",
			filename: &nonRegularFile,
			cmdName:  "review",
			wantErr:  true,
			errMsg:   "is not a regular file",
		},
		{
			name:     "relative path should fail",
			filename: strPtr("valid.sh"),
			cmdName:  "implement",
			wantErr:  true,
			errMsg:   "must be absolute path",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := validateCommandFile(testCase.filename, testCase.cmdName)

			checkValidateFileError(t, err, result, testCase.wantErr, testCase.errMsg)
		})
	}
}

func checkValidateFileError(t *testing.T, err error, result string, wantErr bool, errMsg string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Errorf("validateCommandFile() expected error, got nil")
			return
		}
		if errMsg != "" && !strings.Contains(err.Error(), errMsg) {
			t.Errorf("validateCommandFile() error = %v, want error containing %q", err, errMsg)
		}
		return
	}
	if err != nil {
		t.Errorf("validateCommandFile() unexpected error: %v", err)
		return
	}
	if result == "" {
		t.Errorf("validateCommandFile() expected non-empty result, got empty string")
	}
}
