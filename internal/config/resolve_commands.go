package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveCommandFiles resolves command file paths with validation.
// Returns error if command files are specified but cannot be resolved or validated.
func ResolveCommandFiles(projectConfig *ProjectConfig, homeConfig *HomeConfig) (*ResolvedCommandFiles, error) {
	return resolveCommandFiles(projectConfig, homeConfig)
}

// resolveCommandFiles resolves command file paths with project-first precedence.
// Returns nil if no command files are configured (not an error).
// Returns error if command files are configured but cannot be resolved or validated.
func resolveCommandFiles(projectConfig *ProjectConfig, homeConfig *HomeConfig) (*ResolvedCommandFiles, error) {
	// Determine which config to use for commands (project takes precedence)
	var cmds *Commands
	var baseDir string

	if projectConfig != nil && projectConfig.Commands != nil {
		// Check if project commands are fully specified
		hasImplement := projectConfig.Commands.Implement != nil
		hasReview := projectConfig.Commands.Review != nil
		hasCommit := projectConfig.Commands.Commit != nil
		if hasImplement || hasReview || hasCommit {
			cmds = projectConfig.Commands
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current directory: %w", err)
			}
			baseDir = filepath.Join(cwd, ".fluxid", "commands")
		}
	}

	// Fallback to home if project doesn't have commands
	if cmds == nil && homeConfig != nil && homeConfig.Commands != nil {
		hasImplement := homeConfig.Commands.Implement != nil
		hasReview := homeConfig.Commands.Review != nil
		hasCommit := homeConfig.Commands.Commit != nil
		if hasImplement || hasReview || hasCommit {
			cmds = homeConfig.Commands
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			baseDir = filepath.Join(homeDir, ".fluxid", "commands")
		}
	}

	// If no commands configured, return nil (not an error)
	if cmds == nil {
		return nil, nil
	}

	// Resolve and validate all three command files
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
		return "", fmt.Errorf("commands.%s is required but not specified", cmdName)
	}

	// Construct absolute path
	absPath := filepath.Join(baseDir, *filename)

	// Validate file exists and is readable
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("command file not found: %s (commands.%s)", absPath, cmdName)
		}
		return "", fmt.Errorf("cannot access command file %s (commands.%s): %w", absPath, cmdName, err)
	}

	// Ensure it's a regular file
	if !fileInfo.Mode().IsRegular() {
		return "", fmt.Errorf("command file %s (commands.%s) is not a regular file", absPath, cmdName)
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
