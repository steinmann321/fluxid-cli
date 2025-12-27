// Package assets provides embedded default configuration and command files.
package assets

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Embed all assets using embed.FS
//
//go:embed commands/*.md
var commandsFS embed.FS

//go:embed templates/*.yaml
var templatesFS embed.FS

//go:embed default_config.yaml
var defaultConfigYAML string

// GetDefaultConfig returns the default configuration YAML content.
func GetDefaultConfig() string {
	return defaultConfigYAML
}

// AssetCounts tracks the number of files copied during initialization.
type AssetCounts struct {
	Commands  int
	Templates int
}

// CopyAssetsToDir copies all embedded assets to the specified directory.
// Creates: <targetDir>/.fluxid/{commands/,templates/,config.yaml}
//
// Returns file counts and error.
// Error if:
//   - targetDir/.fluxid already exists
//   - Failed to create directories
//   - Failed to write files
//
//nolint:err113,mnd // Init errors are descriptive, file permissions are standard
func CopyAssetsToDir(targetDir string) (AssetCounts, error) {
	var counts AssetCounts
	fluxidDir := filepath.Join(targetDir, ".fluxid")

	// Check if .fluxid already exists
	if _, err := os.Stat(fluxidDir); err == nil {
		return counts, fmt.Errorf(".fluxid directory already exists at %s", fluxidDir)
	}

	// Create base directory
	// #nosec G301 -- 0o755 is standard user directory permission
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		return counts, fmt.Errorf("failed to create .fluxid directory: %w", err)
	}

	// Copy commands
	cmdCount, err := copyEmbeddedDir(commandsFS, "commands", filepath.Join(fluxidDir, "commands"))
	if err != nil {
		return counts, fmt.Errorf("failed to copy commands: %w", err)
	}
	counts.Commands = cmdCount

	// Copy templates
	tplCount, err := copyEmbeddedDir(templatesFS, "templates", filepath.Join(fluxidDir, "templates"))
	if err != nil {
		return counts, fmt.Errorf("failed to copy templates: %w", err)
	}
	counts.Templates = tplCount

	// Write config.yaml with placeholder replacement
	configPath := filepath.Join(fluxidDir, "config.yaml")
	configContent := replacePlaceholders(defaultConfigYAML, fluxidDir)
	// #nosec G306 -- 0o644 is standard config file permission
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		return counts, fmt.Errorf("failed to write config.yaml: %w", err)
	}

	return counts, nil
}

// replacePlaceholders replaces {{FLUXID_DIR}} with the actual absolute path.
func replacePlaceholders(content string, fluxidDir string) string {
	// Get absolute path
	absPath, err := filepath.Abs(fluxidDir)
	if err != nil {
		// Fallback to fluxidDir if abs fails
		absPath = fluxidDir
	}
	return strings.ReplaceAll(content, "{{FLUXID_DIR}}", absPath)
}

// copyEmbeddedDir recursively copies files from embedded FS to destination.
// Returns the count of files copied.
//
//nolint:mnd,wrapcheck // File permissions are standard, errors already contextual
func copyEmbeddedDir(fsys embed.FS, srcDir, dstDir string) (int, error) {
	fileCount := 0

	// Create destination directory
	// #nosec G301 -- 0o755 is standard user directory permission
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return 0, err
	}

	// Walk embedded filesystem
	//nolint:varnamelen // 'd' is standard for fs.DirEntry in WalkDir callback
	err := fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == srcDir {
			return nil
		}

		// Calculate relative path and destination
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755) // #nosec G301
		}

		// Copy file
		if err := copyEmbeddedFile(fsys, path, dstPath); err != nil {
			return err
		}
		fileCount++
		return nil
	})

	return fileCount, err
}

// copyEmbeddedFile copies a single file from embedded FS to destination.
//
//nolint:wrapcheck // Errors already contextual in caller
func copyEmbeddedFile(fsys embed.FS, src, dst string) error {
	srcFile, err := fsys.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst) // #nosec G304
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}
