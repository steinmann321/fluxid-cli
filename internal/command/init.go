package command

import (
	"fluxid-cli/internal/assets"
	"fmt"
	"os"
	"path/filepath"
)

// handleInit processes the init command to bootstrap fluxid configuration.
//
// Usage:
//
//	fluxid init              -> Initialize in ~/.fluxid/
//	fluxid init <path>       -> Initialize in <path>/.fluxid/
//	fluxid init --help       -> Show help
//
// Returns:
//
//	0 on success
//	1 on error
func handleInit(args []string) int {
	// Check for --help flag
	for _, arg := range args {
		if arg == flagHelp || arg == "-h" {
			printInitHelp()
			return 0
		}
	}

	// Determine target directory
	targetDir, err := getInitTargetDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Copy assets to target directory
	if err := assets.CopyAssetsToDir(targetDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize: %v\n", err)
		return 1
	}

	// Print success message
	fluxidDir := filepath.Join(targetDir, ".fluxid")
	absPath, _ := filepath.Abs(fluxidDir)

	fmt.Fprintf(os.Stderr, "✓ Successfully initialized fluxid configuration\n\n")
	fmt.Fprintf(os.Stderr, "Location: %s\n\n", absPath)
	fmt.Fprintf(os.Stderr, "Created:\n")
	fmt.Fprintf(os.Stderr, "  %s/config.yaml\n", absPath)
	fmt.Fprintf(os.Stderr, "  %s/commands/      (28 command files)\n", absPath)
	fmt.Fprintf(os.Stderr, "  %s/templates/     (2 template files)\n", absPath)
	fmt.Fprintf(os.Stderr, "\nNext steps:\n")
	fmt.Fprintf(os.Stderr, "  1. Review and customize config.yaml\n")
	fmt.Fprintf(os.Stderr, "  2. Run 'fluxid --claude' to start a workflow\n")

	return 0
}

// getInitTargetDir determines the target directory for initialization.
//
// Args:
//   - Empty: Use home directory
//   - One arg: Use specified path (create if doesn't exist)
//   - Multiple args: Error
func getInitTargetDir(args []string) (string, error) {
	if len(args) == 0 {
		// No args: use home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return homeDir, nil
	}

	if len(args) == 1 {
		// One arg: use specified path
		targetPath := args[0]

		// Convert to absolute path
		absPath, err := filepath.Abs(targetPath)
		if err != nil {
			return "", fmt.Errorf("invalid path %s: %w", targetPath, err)
		}

		// Create directory if it doesn't exist
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			// #nosec G301 -- 0o755 is standard user directory permission
			//nolint:mnd // 0o755 is standard file permission constant
			if err := os.MkdirAll(absPath, 0o755); err != nil {
				return "", fmt.Errorf("failed to create directory %s: %w", absPath, err)
			}
		}

		return absPath, nil
	}

	// Too many arguments
	//nolint:err113 // Init command error messages are user-facing and descriptive
	return "", fmt.Errorf("too many arguments: expected 0 or 1, got %d", len(args))
}

// printInitHelp prints help text for the init command.
func printInitHelp() {
	helpText := `fluxid init - Initialize fluxid configuration

USAGE:
  fluxid init [path]

DESCRIPTION:
  Initialize a new fluxid configuration by copying default command files,
  templates, and configuration to the specified directory.

ARGUMENTS:
  path      Optional target directory (creates .fluxid/ subdirectory)
            If omitted, initializes in ~/.fluxid/ (global configuration)

BEHAVIOR:
  - Creates .fluxid/ directory in target location
  - Copies 28 command files to .fluxid/commands/
  - Copies 2 template files to .fluxid/templates/
  - Creates default config.yaml
  - Fails if .fluxid/ already exists (safety check)

EXAMPLES:
  fluxid init                    Initialize global config in ~/.fluxid/
  fluxid init .                  Initialize in current directory
  fluxid init /path/to/project   Initialize in specific project
  fluxid init --help             Show this help

SEE ALSO:
  Use --config flag to specify custom config location
  Edit .fluxid/config.yaml to customize command file paths
`
	fmt.Fprint(os.Stderr, helpText)
}
