package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	errCommandRequired       = errors.New("command is required but not specified")
	errCommandFileNotFound   = errors.New("command file not found")
	errCommandNotRegularFile = errors.New("command file is not a regular file")
	errCommandsRequired      = errors.New("commands section is required in at least one config file")
)

// ResolveCommandFiles resolves command file paths with validation.
// Returns error if command files are specified but cannot be resolved or validated.
func ResolveCommandFiles(projectConfig *ProjectConfig, homeConfig *HomeConfig) (*ResolvedCommandFiles, error) {
	return resolveCommandFiles(projectConfig, homeConfig)
}

// resolveCommandFiles resolves command file paths with project-first precedence.
// Commands section is REQUIRED in at least one config file.
// Returns error if command files are not configured or cannot be resolved or validated.
func resolveCommandFiles(projectConfig *ProjectConfig, homeConfig *HomeConfig) (*ResolvedCommandFiles, error) {
	// Determine which config to use for commands (project takes precedence)
	cmds, baseDir, err := selectCommandsConfig(projectConfig, homeConfig)
	if err != nil {
		return nil, err
	}

	// Commands section is required - no defaults
	if cmds == nil {
		return nil, errCommandsRequired
	}

	// Resolve and validate all three command files
	return resolveAllCommandFiles(baseDir, cmds)
}

func selectCommandsConfig(projectConfig *ProjectConfig, homeConfig *HomeConfig) (*Commands, string, error) {
	// Try project config first
	if cmds, baseDir, err := tryProjectCommands(projectConfig); cmds != nil || err != nil {
		return cmds, baseDir, err
	}

	// Fallback to home config
	return tryHomeCommands(homeConfig)
}

func tryProjectCommands(projectConfig *ProjectConfig) (*Commands, string, error) {
	if projectConfig != nil && projectConfig.Commands != nil {
		if hasAnyCommand(projectConfig.Commands) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, "", fmt.Errorf("failed to get current directory: %w", err)
			}
			// Paths are relative to .fluxid/ directory (where config.yaml lives)
			baseDir := filepath.Join(cwd, ".fluxid")
			return projectConfig.Commands, baseDir, nil
		}
	}
	return nil, "", nil
}

func tryHomeCommands(homeConfig *HomeConfig) (*Commands, string, error) {
	if homeConfig != nil && homeConfig.Commands != nil {
		if hasAnyCommand(homeConfig.Commands) {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, "", fmt.Errorf("failed to get home directory: %w", err)
			}
			// Paths are relative to ~/.fluxid/ directory (where config.yaml lives)
			baseDir := filepath.Join(homeDir, ".fluxid")
			return homeConfig.Commands, baseDir, nil
		}
	}
	return nil, "", nil
}

func hasAnyCommand(cmds *Commands) bool {
	return (cmds.Implement != nil) || (cmds.Review != nil) || (cmds.Commit != nil)
}

func resolveAllCommandFiles(baseDir string, cmds *Commands) (*ResolvedCommandFiles, error) {
	implementPath, err := resolveAndValidateCommandFile(baseDir, cmds.Implement, "implement")
	if err != nil {
		return nil, err
	}

	reviewPath, err := resolveAndValidateCommandFile(baseDir, cmds.Review, "review")
	if err != nil {
		return nil, err
	}

	commitPath, err := resolveAndValidateCommandFile(baseDir, cmds.Commit, "commit")
	if err != nil {
		return nil, err
	}

	return &ResolvedCommandFiles{
		ImplementPath: implementPath,
		ReviewPath:    reviewPath,
		CommitPath:    commitPath,
	}, nil
}

// resolveAndValidateCommandFile resolves a single command file path and validates its existence.
func resolveAndValidateCommandFile(baseDir string, filename *string, cmdName string) (string, error) {
	if filename == nil || *filename == "" {
		return "", fmt.Errorf("commands.%s: %w", cmdName, errCommandRequired)
	}

	// Construct absolute path
	absPath := filepath.Join(baseDir, *filename)

	// Validate file exists and is readable
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%s (commands.%s): %w", absPath, cmdName, errCommandFileNotFound)
		}
		return "", fmt.Errorf("cannot access command file %s (commands.%s): %w", absPath, cmdName, err)
	}

	// Ensure it's a regular file
	if !fileInfo.Mode().IsRegular() {
		return "", fmt.Errorf("%s (commands.%s): %w", absPath, cmdName, errCommandNotRegularFile)
	}

	// Check read permissions by attempting to open the file
	// #nosec G304 -- absPath is constructed from config values, validated above
	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot read command file %s (commands.%s): %w", absPath, cmdName, err)
	}
	if closeErr := file.Close(); closeErr != nil {
		return "", fmt.Errorf("failed to close command file %s (commands.%s): %w", absPath, cmdName, closeErr)
	}

	return absPath, nil
}
