// Package assets provides embedded default configuration and command files.
package assets

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

// CopyAssetsToDir copies all embedded assets to the specified directory.
// Creates: <targetDir>/.fluxid/{commands/,templates/,config.yaml}
//
// Returns error if:
//   - targetDir/.fluxid already exists
//   - Failed to create directories
//   - Failed to write files
//
//nolint:err113,mnd // Init errors are descriptive, file permissions are standard
func CopyAssetsToDir(targetDir string) error {
	fluxidDir := filepath.Join(targetDir, ".fluxid")

	// Check if .fluxid already exists
	if _, err := os.Stat(fluxidDir); err == nil {
		return fmt.Errorf(".fluxid directory already exists at %s", fluxidDir)
	}

	// Create base directory
	// #nosec G301 -- 0o755 is standard user directory permission
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .fluxid directory: %w", err)
	}

	// Copy commands
	if err := copyEmbeddedDir(commandsFS, "commands", filepath.Join(fluxidDir, "commands")); err != nil {
		return fmt.Errorf("failed to copy commands: %w", err)
	}

	// Copy templates
	if err := copyEmbeddedDir(templatesFS, "templates", filepath.Join(fluxidDir, "templates")); err != nil {
		return fmt.Errorf("failed to copy templates: %w", err)
	}

	// Write config.yaml
	configPath := filepath.Join(fluxidDir, "config.yaml")
	// #nosec G306 -- 0o644 is standard config file permission
	if err := os.WriteFile(configPath, []byte(defaultConfigYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %w", err)
	}

	return nil
}

// copyEmbeddedDir recursively copies files from embedded FS to destination.
//
//nolint:mnd,wrapcheck // File permissions are standard, errors already contextual
func copyEmbeddedDir(fsys embed.FS, srcDir, dstDir string) error {
	// Create destination directory
	// #nosec G301 -- 0o755 is standard user directory permission
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}

	// Walk embedded filesystem
	//nolint:varnamelen // 'd' is standard for fs.DirEntry in WalkDir callback
	return fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, err error) error {
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
		return copyEmbeddedFile(fsys, path, dstPath)
	})
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
