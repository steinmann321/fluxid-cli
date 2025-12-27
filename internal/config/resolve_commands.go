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
	errCommandMustBeAbsolute = errors.New("command file path must be absolute path")
)

// ResolveCommandFiles resolves command file paths with validation.
// All command file paths MUST be absolute paths.
// Returns error if command files are specified but cannot be resolved or validated.
func ResolveCommandFiles(projectConfig *ProjectConfig, homeConfig *HomeConfig) (*ResolvedCommandFiles, error) {
	return resolveCommandFiles(projectConfig, homeConfig)
}

// resolveCommandFiles resolves command file paths with project-first precedence.
// Commands section is REQUIRED in at least one config file.
// All command file paths MUST be absolute paths.
// Returns error if command files are not configured or cannot be validated.
func resolveCommandFiles(projectConfig *ProjectConfig, homeConfig *HomeConfig) (*ResolvedCommandFiles, error) {
	// Determine which config to use for commands (project takes precedence)
	cmds, err := selectCommandsConfig(projectConfig, homeConfig)
	if err != nil {
		return nil, err
	}

	// Commands section is required - no defaults
	if cmds == nil {
		return nil, errCommandsRequired
	}

	// Validate all three command files (must be absolute paths)
	return validateAllCommandFiles(cmds)
}

func selectCommandsConfig(projectConfig *ProjectConfig, homeConfig *HomeConfig) (*Commands, error) {
	// Try project config first
	if cmds, err := tryProjectCommands(projectConfig); cmds != nil || err != nil {
		return cmds, err
	}

	// Fallback to home config
	return tryHomeCommands(homeConfig)
}

//nolint:unparam,nilnil // Allows nil returns when no commands configured
func tryProjectCommands(projectConfig *ProjectConfig) (*Commands, error) {
	if projectConfig != nil && projectConfig.Commands != nil {
		if hasAnyCommand(projectConfig.Commands) {
			return projectConfig.Commands, nil
		}
	}
	return nil, nil
}

//nolint:nilnil // Allows nil returns when no commands configured
func tryHomeCommands(homeConfig *HomeConfig) (*Commands, error) {
	if homeConfig != nil && homeConfig.Commands != nil {
		if hasAnyCommand(homeConfig.Commands) {
			return homeConfig.Commands, nil
		}
	}
	return nil, nil
}

func hasAnyCommand(cmds *Commands) bool {
	return (cmds.Implement != nil) || (cmds.Review != nil) || (cmds.Commit != nil)
}

func validateAllCommandFiles(cmds *Commands) (*ResolvedCommandFiles, error) {
	implementPath, err := validateCommandFile(cmds.Implement, "implement")
	if err != nil {
		return nil, err
	}

	reviewPath, err := validateCommandFile(cmds.Review, "review")
	if err != nil {
		return nil, err
	}

	commitPath, err := validateCommandFile(cmds.Commit, "commit")
	if err != nil {
		return nil, err
	}

	return &ResolvedCommandFiles{
		ImplementPath: implementPath,
		ReviewPath:    reviewPath,
		CommitPath:    commitPath,
	}, nil
}

// validateCommandFile validates a command file path (must be absolute) and checks existence.
func validateCommandFile(filename *string, cmdName string) (string, error) {
	if filename == nil || *filename == "" {
		return "", fmt.Errorf("commands.%s: %w", cmdName, errCommandRequired)
	}

	absPath := *filename

	// Validate that path is absolute
	if !filepath.IsAbs(absPath) {
		return "", fmt.Errorf("%s (commands.%s): %w", absPath, cmdName, errCommandMustBeAbsolute)
	}

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
	// #nosec G304 -- absPath is from config and validated above
	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot read command file %s (commands.%s): %w", absPath, cmdName, err)
	}
	if closeErr := file.Close(); closeErr != nil {
		return "", fmt.Errorf("failed to close command file %s (commands.%s): %w", absPath, cmdName, closeErr)
	}

	return absPath, nil
}
