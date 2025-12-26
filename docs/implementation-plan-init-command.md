# Implementation Plan: `fluxid init` Command

**Date:** 2025-12-26
**Status:** Ready for Implementation
**Scope:** Add initialization command to bootstrap fluxid configuration

## Table of Contents
1. [Overview](#overview)
2. [Requirements](#requirements)
3. [Architecture](#architecture)
4. [Implementation Details](#implementation-details)
5. [Testing Strategy](#testing-strategy)
6. [Implementation Checklist](#implementation-checklist)

## Overview

Implement `fluxid init` command to bootstrap new fluxid projects by copying default configuration and command files to either:
- Global location: `~/.fluxid/` (when run without arguments)
- Project location: `<path>/.fluxid/` (when run with path argument)

Additionally, update path resolution logic to use `.fluxid/commands/` subdirectory instead of `.fluxid/` root.

### User Requirements
- `fluxid init` → Initialize in `~/.fluxid/`
- `fluxid init <project-path>` → Initialize in `<project-path>/.fluxid/`
- Only copy report schema and example to templates (implementation workflow only)
- Update command path resolution to `.fluxid/commands/` (breaking change, no backward compatibility)

## Requirements

### Functional Requirements
1. Initialize configuration in global or project-specific location
2. Copy 30 command markdown files to `.fluxid/commands/`
3. Copy report schema and example to `.fluxid/templates/`
4. Create default `config.yaml` with sensible defaults
5. Prevent overwriting existing `.fluxid/` directories
6. Provide clear success/error messages

### Non-Functional Requirements
- Embed assets in binary (no external file dependencies)
- Fast initialization (<100ms)
- Clear error messages for common failure cases
- Follow existing fluxid command patterns

## Architecture

### Directory Structure After Init

```
~/.fluxid/                           # or <project>/.fluxid/
├── config.yaml                      # Default configuration
├── commands/                        # Command markdown files
│   ├── fluxid.implement.md         # TDD implementation specialist
│   ├── fluxid.review-Implementation.md  # Implementation reviewer
│   ├── fluxid.commit.md            # Commit gatekeeper
│   ├── fluxid.create-product.md
│   ├── fluxid.clarify-product.md
│   ├── fluxid.create-milestones.md
│   ├── fluxid.create-epics.md
│   ├── fluxid.validate-epic.md
│   ├── fluxid.validate-milestone.md
│   └── ... (21 more command files)
└── templates/                       # Report templates
    ├── report-schema.yaml          # YAML report schema
    └── report-example.yaml         # Example report
```

### Component Architecture

```
┌─────────────────────────────────────────┐
│  fluxid init [path]                     │
│  (User Command)                         │
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│  internal/command/init.go               │
│  - handleInit(args)                     │
│  - getInitTargetDir(args)               │
│  - printInitHelp()                      │
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│  internal/assets/assets.go              │
│  - CopyAssetsToDir(targetDir)           │
│  - Embedded commands/ (30 files)        │
│  - Embedded templates/ (2 files)        │
│  - Embedded default_config.yaml         │
└─────────────────────────────────────────┘
```

## Implementation Details

### 1. Embedded Assets Package

**Location:** `internal/assets/`

#### 1.1 Directory Structure

```
internal/assets/
├── assets.go                        # Embed logic and copy functions
├── default_config.yaml              # Default configuration
├── commands/                        # 30 command files (copied from .fluxid/commands/)
│   ├── fluxid.implement.md
│   ├── fluxid.review-Implementation.md
│   ├── fluxid.commit.md
│   ├── fluxid.create-product.md
│   ├── fluxid.clarify-product.md
│   ├── fluxid.create-milestones.md
│   ├── fluxid.create-epics.md
│   ├── fluxid.create-tasks.md
│   ├── fluxid.create-e2e-tasks.md
│   ├── fluxid.create-progress.md
│   ├── fluxid.validate-epic.md
│   ├── fluxid.validate-milestone.md
│   ├── fluxid.validate-task.md
│   ├── fluxid.validate-e2e.md
│   ├── fluxid.validate-structure.md
│   ├── fluxid.validate-layers.md
│   ├── fluxid.validate-hooks.md
│   ├── fluxid.review-e2e.md
│   ├── fluxid.review-architecture.md
│   ├── fluxid.review-epics.md
│   ├── fluxid.review-implementation-e2e.md
│   ├── fluxid.create-layers.md
│   ├── fluxid.implement-delegate.md
│   ├── fluxid.implement-e2e.md
│   ├── fluxid.implement-cli.md
│   ├── fluxid.review-cli.md
│   ├── fluxid.create-tasks-deprecated.md
│   └── ... (verify exact count and names)
└── templates/
    ├── report-schema.yaml          # Copy from internal/ipc/schema.yaml
    └── report-example.yaml         # Copy from .fluxid/templates/report-example.yaml
```

#### 1.2 assets.go Implementation

```go
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
func CopyAssetsToDir(targetDir string) error {
    fluxidDir := filepath.Join(targetDir, ".fluxid")

    // Check if .fluxid already exists
    if _, err := os.Stat(fluxidDir); err == nil {
        return fmt.Errorf(".fluxid directory already exists at %s", fluxidDir)
    }

    // Create base directory
    if err := os.MkdirAll(fluxidDir, 0755); err != nil {
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
    if err := os.WriteFile(configPath, []byte(defaultConfigYAML), 0644); err != nil {
        return fmt.Errorf("failed to write config.yaml: %w", err)
    }

    return nil
}

// copyEmbeddedDir recursively copies files from embedded FS to destination
func copyEmbeddedDir(fsys embed.FS, srcDir, dstDir string) error {
    // Create destination directory
    if err := os.MkdirAll(dstDir, 0755); err != nil {
        return err
    }

    // Walk embedded filesystem
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
            return os.MkdirAll(dstPath, 0755)
        }

        // Copy file
        return copyEmbeddedFile(fsys, path, dstPath)
    })
}

// copyEmbeddedFile copies a single file from embedded FS to destination
func copyEmbeddedFile(fsys embed.FS, src, dst string) error {
    srcFile, err := fsys.Open(src)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    dstFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    if _, err := io.Copy(dstFile, srcFile); err != nil {
        return err
    }

    return dstFile.Sync()
}
```

#### 1.3 default_config.yaml Content

```yaml
# Fluxid Configuration
# Workflow automation for coding agents

# Command file paths (relative to .fluxid/commands/)
commands:
  implement: fluxid.implement.md
  review: fluxid.review-Implementation.md
  commit: fluxid.commit.md

# Optional: Override defaults
# agent: claude                  # Default: claude
# implement_retries: 3           # Default: 3
# iterations: 20                 # Default: 20
```

**Key Notes:**
- Paths are simple filenames (no `commands/` prefix)
- baseDir will be `.fluxid/commands/` so these resolve correctly
- Optional fields commented out to show defaults without forcing values

### 2. Init Command Implementation

**Location:** `internal/command/init.go`

```go
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
//   fluxid init              -> Initialize in ~/.fluxid/
//   fluxid init <path>       -> Initialize in <path>/.fluxid/
//   fluxid init --help       -> Show help
//
// Returns:
//   0 on success
//   1 on error
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
    fmt.Fprintf(os.Stderr, "  %s/commands/      (30 command files)\n", absPath)
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
            if err := os.MkdirAll(absPath, 0755); err != nil {
                return "", fmt.Errorf("failed to create directory %s: %w", absPath, err)
            }
        }

        return absPath, nil
    }

    // Too many arguments
    return "", fmt.Errorf("too many arguments: expected 0 or 1, got %d", len(args))
}

// printInitHelp prints help text for the init command
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
  - Copies 30 command files to .fluxid/commands/
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
```

### 3. Command Dispatcher Integration

**Location:** `internal/command/root.go`

**Update:** Add init command to `handleSpecialCommands()` (around line 34)

```go
// handleSpecialCommands checks for and handles commands that bypass config loading
func handleSpecialCommands() (int, bool) {
    // Check for init command first (before loading config)
    // Init creates the config, so it must run before config loading
    if len(os.Args) > 1 && os.Args[1] == "init" {
        return handleInit(os.Args[2:]), true
    }

    // Check for IPC command (before loading config)
    if len(os.Args) > 1 && os.Args[1] == "ipc" {
        return handleIPCCommand(os.Args[2:]), true
    }

    // Check for --write-history flag (standalone operation)
    for i := 1; i < len(os.Args); i++ {
        if os.Args[i] == "--write-history" {
            return handleWriteHistory(os.Args[i+1:]), true
        }
    }

    // Check for --help flag
    for _, arg := range os.Args[1:] {
        if arg == flagHelp || arg == "-h" {
            printUsage()
            return 0, true
        }
    }

    return 0, false
}
```

**Location:** `internal/command/ipc.go`

**Update:** Add init to usage text in `printUsage()`

```go
func printUsage() {
    fmt.Fprintf(os.Stderr, "Usage:\n")
    fmt.Fprintf(os.Stderr, "  fluxid init [path]\n")
    fmt.Fprintf(os.Stderr, "  fluxid --claude [--fluxid-iterations N] [--fluxid-implement-retries R] [claude-args]\n")
    fmt.Fprintf(os.Stderr, "  fluxid --write-history <message> [--help]\n")
    // ... rest of usage

    fmt.Fprintf(os.Stderr, "\nCommands:\n")
    fmt.Fprintf(os.Stderr, "  init [path]              Initialize fluxid configuration (global or project)\n")
    fmt.Fprintf(os.Stderr, "  (default)                Run workflow controller (requires --claude)\n")
    // ... rest of commands
}
```

### 4. Path Resolution Updates

**Location:** `internal/config/resolve_commands.go`

**Update Line 60** - `tryProjectCommands()`:

```go
func tryProjectCommands(projectConfig *ProjectConfig) (*Commands, string, error) {
    if projectConfig != nil && projectConfig.Commands != nil {
        if hasAnyCommand(projectConfig.Commands) {
            cwd, err := os.Getwd()
            if err != nil {
                return nil, "", fmt.Errorf("failed to get current directory: %w", err)
            }
            // Paths are relative to .fluxid/commands/ directory
            baseDir := filepath.Join(cwd, ".fluxid", "commands")
            return projectConfig.Commands, baseDir, nil
        }
    }
    return nil, "", nil
}
```

**Update Line 75** - `tryHomeCommands()`:

```go
func tryHomeCommands(homeConfig *HomeConfig) (*Commands, string, error) {
    if homeConfig != nil && homeConfig.Commands != nil {
        if hasAnyCommand(homeConfig.Commands) {
            homeDir, err := os.UserHomeDir()
            if err != nil {
                return nil, "", fmt.Errorf("failed to get home directory: %w", err)
            }
            // Paths are relative to ~/.fluxid/commands/ directory
            baseDir := filepath.Join(homeDir, ".fluxid", "commands")
            return homeConfig.Commands, baseDir, nil
        }
    }
    return nil, "", nil
}
```

**Impact:** This is a BREAKING CHANGE. Existing configs must move command files into `.fluxid/commands/` subdirectory.

**Migration for existing users:**
```bash
mkdir .fluxid/commands
mv .fluxid/*.md .fluxid/commands/
```

## Testing Strategy

### Unit Tests

#### 1. `internal/assets/assets_test.go`

```go
package assets

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCopyAssetsToDir_Success(t *testing.T) {
    tmpDir := t.TempDir()

    err := CopyAssetsToDir(tmpDir)
    if err != nil {
        t.Fatalf("CopyAssetsToDir failed: %v", err)
    }

    // Verify structure created
    fluxidDir := filepath.Join(tmpDir, ".fluxid")
    if _, err := os.Stat(fluxidDir); os.IsNotExist(err) {
        t.Error(".fluxid directory not created")
    }

    // Verify config.yaml
    configPath := filepath.Join(fluxidDir, "config.yaml")
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        t.Error("config.yaml not created")
    }

    // Verify commands directory with files
    commandsDir := filepath.Join(fluxidDir, "commands")
    entries, err := os.ReadDir(commandsDir)
    if err != nil {
        t.Fatalf("Failed to read commands dir: %v", err)
    }
    if len(entries) != 30 {
        t.Errorf("Expected 30 command files, got %d", len(entries))
    }

    // Verify templates directory with files
    templatesDir := filepath.Join(fluxidDir, "templates")
    entries, err = os.ReadDir(templatesDir)
    if err != nil {
        t.Fatalf("Failed to read templates dir: %v", err)
    }
    if len(entries) != 2 {
        t.Errorf("Expected 2 template files, got %d", len(entries))
    }
}

func TestCopyAssetsToDir_AlreadyExists(t *testing.T) {
    tmpDir := t.TempDir()

    // Create .fluxid directory
    fluxidDir := filepath.Join(tmpDir, ".fluxid")
    if err := os.MkdirAll(fluxidDir, 0755); err != nil {
        t.Fatalf("Failed to create test dir: %v", err)
    }

    // Should fail
    err := CopyAssetsToDir(tmpDir)
    if err == nil {
        t.Error("Expected error when .fluxid already exists")
    }
}
```

#### 2. `internal/command/init_test.go`

```go
package command

import (
    "os"
    "path/filepath"
    "testing"
)

func TestHandleInit_NoArgs(t *testing.T) {
    // This would initialize in home dir - complex to test
    // Instead test getInitTargetDir directly
    targetDir, err := getInitTargetDir([]string{})
    if err != nil {
        t.Fatalf("getInitTargetDir failed: %v", err)
    }

    homeDir, _ := os.UserHomeDir()
    if targetDir != homeDir {
        t.Errorf("Expected home dir %s, got %s", homeDir, targetDir)
    }
}

func TestGetInitTargetDir_WithPath(t *testing.T) {
    tmpDir := t.TempDir()
    testPath := filepath.Join(tmpDir, "project")

    targetDir, err := getInitTargetDir([]string{testPath})
    if err != nil {
        t.Fatalf("getInitTargetDir failed: %v", err)
    }

    // Should create the directory
    if _, err := os.Stat(targetDir); os.IsNotExist(err) {
        t.Error("Target directory not created")
    }
}

func TestGetInitTargetDir_TooManyArgs(t *testing.T) {
    _, err := getInitTargetDir([]string{"path1", "path2"})
    if err == nil {
        t.Error("Expected error for too many args")
    }
}
```

### Integration Tests

#### Update: `internal/config/resolve_commands_test.go`

Key changes:
- Create test files in `.fluxid/commands/` subdirectory instead of `.fluxid/`
- Update all path expectations

Example:
```go
// Setup
homeCommandsDir := filepath.Join(homeDir, ".fluxid", "commands")
if err := os.MkdirAll(homeCommandsDir, 0755); err != nil {
    t.Fatalf("Failed to create commands dir: %v", err)
}

homeImpl := filepath.Join(homeCommandsDir, "implement.sh")
// ... create files in commands/ subdirectory
```

#### Update: `e2e-tests/tests/m02-e03-user-provides-command-files-and-system-resolves-paths_test.go`

Update all test cases to:
- Create `.fluxid/commands/` directory
- Place command files in `commands/` subdirectory
- Update config to reference files without `commands/` prefix

### E2E Tests

#### New: `e2e-tests/tests/m02-eXX-user-initializes-fluxid-configuration_test.go`

```go
package tests

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

func TestInitGlobalDirectory(t *testing.T) {
    // Would need to mock home directory or use test env
    // For now, test with explicit path
}

func TestInitProjectDirectory(t *testing.T) {
    tmpDir := t.TempDir()
    projectDir := filepath.Join(tmpDir, "myproject")

    // Run fluxid init
    cmd := exec.Command("fluxid", "init", projectDir)
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("fluxid init failed: %v\nOutput: %s", err, output)
    }

    // Verify structure
    fluxidDir := filepath.Join(projectDir, ".fluxid")
    if _, err := os.Stat(fluxidDir); os.IsNotExist(err) {
        t.Error(".fluxid not created")
    }

    // Verify config
    configPath := filepath.Join(fluxidDir, "config.yaml")
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        t.Error("config.yaml not created")
    }

    // Verify commands
    commandsDir := filepath.Join(fluxidDir, "commands")
    entries, _ := os.ReadDir(commandsDir)
    if len(entries) == 0 {
        t.Error("No command files created")
    }
}

func TestInitAlreadyExists(t *testing.T) {
    tmpDir := t.TempDir()

    // Create .fluxid first
    fluxidDir := filepath.Join(tmpDir, ".fluxid")
    os.MkdirAll(fluxidDir, 0755)

    // Should fail
    cmd := exec.Command("fluxid", "init", tmpDir)
    err := cmd.Run()
    if err == nil {
        t.Error("Expected error when .fluxid exists")
    }
}
```

## Implementation Checklist

### Phase 1: Asset Preparation
- [ ] Create `internal/assets/` directory
- [ ] Copy 30 command files from `.fluxid/commands/` to `internal/assets/commands/`
- [ ] Copy `internal/ipc/schema.yaml` to `internal/assets/templates/report-schema.yaml`
- [ ] Copy `.fluxid/templates/report-example.yaml` to `internal/assets/templates/report-example.yaml`
- [ ] Create `internal/assets/default_config.yaml` with default configuration
- [ ] Implement `internal/assets/assets.go` with embed directives and copy functions
- [ ] Verify all files embed correctly (check binary size)

### Phase 2: Command Implementation
- [ ] Create `internal/command/init.go` with `handleInit()`, `getInitTargetDir()`, `printInitHelp()`
- [ ] Update `internal/command/root.go` - Add init check in `handleSpecialCommands()` (line ~34)
- [ ] Update `internal/command/ipc.go` - Add init to `printUsage()` help text
- [ ] Test basic init command manually

### Phase 3: Path Resolution
- [ ] Update `internal/config/resolve_commands.go` line 60 - Add "commands" to baseDir in `tryProjectCommands()`
- [ ] Update `internal/config/resolve_commands.go` line 75 - Add "commands" to baseDir in `tryHomeCommands()`
- [ ] Test path resolution with new structure

### Phase 4: Unit Tests
- [ ] Create `internal/assets/assets_test.go`
  - [ ] Test `CopyAssetsToDir()` success
  - [ ] Test error when directory exists
  - [ ] Test file count and structure
- [ ] Create `internal/command/init_test.go`
  - [ ] Test `getInitTargetDir()` with no args
  - [ ] Test `getInitTargetDir()` with path arg
  - [ ] Test error on too many args
  - [ ] Test help flag

### Phase 5: Integration Tests
- [ ] Update `internal/config/resolve_commands_test.go`
  - [ ] Change file creation to use `.fluxid/commands/` subdirectory
  - [ ] Update all test expectations
  - [ ] Run tests and verify pass
- [ ] Update `e2e-tests/tests/m02-e03-user-provides-command-files-and-system-resolves-paths_test.go`
  - [ ] Update all test cases for new directory structure
  - [ ] Run E2E tests and verify pass

### Phase 6: E2E Tests
- [ ] Create `e2e-tests/tests/m02-eXX-user-initializes-fluxid-configuration_test.go`
  - [ ] Test init in project directory
  - [ ] Test error when .fluxid exists
  - [ ] Test help flag
  - [ ] Verify created structure

### Phase 7: Validation
- [ ] Build binary: `make build`
- [ ] Manual test: `bin/fluxid init /tmp/test-project`
- [ ] Verify all files created correctly
- [ ] Manual test: `bin/fluxid init` (global location)
- [ ] Run full test suite: `make test`
- [ ] Test error cases manually

### Phase 8: Documentation
- [ ] Update README.md with init command usage
- [ ] Add migration notes for existing users
- [ ] Update any relevant documentation

## Error Scenarios and Handling

| Scenario | Behavior | Exit Code |
|----------|----------|-----------|
| `.fluxid/` already exists | Error message, suggest manual removal | 1 |
| Invalid path argument | Error message with path details | 1 |
| Permission denied | Propagate OS error with context | 1 |
| Too many arguments | Print usage, error message | 1 |
| Help flag | Print help text | 0 |
| Success | Print success message with paths | 0 |

## Usage Examples

```bash
# Initialize global configuration
$ fluxid init
✓ Successfully initialized fluxid configuration

Location: /Users/username/.fluxid

Created:
  /Users/username/.fluxid/config.yaml
  /Users/username/.fluxid/commands/      (30 command files)
  /Users/username/.fluxid/templates/     (2 template files)

Next steps:
  1. Review and customize config.yaml
  2. Run 'fluxid --claude' to start a workflow

# Initialize in current project
$ fluxid init .
✓ Successfully initialized fluxid configuration

Location: /Users/username/myproject/.fluxid

Created:
  /Users/username/myproject/.fluxid/config.yaml
  /Users/username/myproject/.fluxid/commands/      (30 command files)
  /Users/username/myproject/.fluxid/templates/     (2 template files)

Next steps:
  1. Review and customize config.yaml
  2. Run 'fluxid --claude' to start a workflow

# Initialize in specific project
$ fluxid init /path/to/project

# Error: already exists
$ fluxid init .
Error: failed to initialize: .fluxid directory already exists at /Users/username/myproject/.fluxid
If you want to reinitialize, please remove the existing directory first.

# Get help
$ fluxid init --help
[help text displayed]
```

## Migration Guide for Existing Users

If you have an existing `.fluxid/` configuration, follow these steps to migrate to the new structure:

### Manual Migration

```bash
# Navigate to your project or home directory
cd ~/myproject  # or cd ~

# Create commands subdirectory
mkdir .fluxid/commands

# Move all command markdown files
mv .fluxid/*.md .fluxid/commands/

# Your config.yaml should already reference just the filenames
# Example: implement: fluxid.implement.md (NOT commands/fluxid.implement.md)
```

### Migration Script

```bash
#!/bin/bash
# migrate-fluxid.sh - Migrate .fluxid structure to v2.0

if [ ! -d ".fluxid" ]; then
    echo "Error: No .fluxid directory found"
    exit 1
fi

if [ -d ".fluxid/commands" ]; then
    echo "Already migrated - .fluxid/commands exists"
    exit 0
fi

# Create commands directory
mkdir -p .fluxid/commands

# Move all .md files
mv .fluxid/*.md .fluxid/commands/ 2>/dev/null || true

echo "✓ Migration complete"
echo "  Moved command files to .fluxid/commands/"
echo "  Verify your config.yaml references filenames without 'commands/' prefix"
```

## Implementation Notes

### Design Decisions

1. **Embed vs External Files**: Use embed for zero external dependencies
2. **Minimal Default Config**: Commented optional fields to show possibilities without forcing choices
3. **Breaking Change**: Clean break on path resolution (no backward compatibility) since not yet public
4. **Error Messages**: Clear, actionable error messages for common failures
5. **Directory Safety**: Fail if `.fluxid/` exists to prevent accidental overwrites

### Performance Considerations

- Embedded asset size: ~100KB (negligible)
- File copy operations: <10ms for 32 files
- No performance impact on normal workflow execution

### Security Considerations

- File permissions: 0755 for directories, 0644 for files
- Path validation via `filepath.Abs()` and `filepath.Join()`
- No force overwrite - explicit user action required to replace existing config

## Success Criteria

- [ ] `fluxid init` successfully creates global config
- [ ] `fluxid init <path>` successfully creates project config
- [ ] All 30 command files copied correctly
- [ ] Report schema and example copied to templates
- [ ] Config file created with correct defaults
- [ ] Path resolution uses `.fluxid/commands/` correctly
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] All E2E tests pass
- [ ] Binary builds successfully
- [ ] Manual testing confirms correct behavior

## References

- Command structure: `internal/command/root.go`, `internal/command/ipc.go`
- Existing embed pattern: `internal/ipc/schema.go`
- Path resolution: `internal/config/resolve_commands.go`
- Config structure: `internal/types/config.go`
- Example config: `.fluxid/scripts/loop/config.yaml`
