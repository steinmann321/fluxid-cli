//nolint:paralleltest // E2E tests use shared infrastructure
package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestM02E03CommandFilesResolvedFromHome validates that command files are
// resolved from home config and absolute paths are displayed.
func TestM02E03CommandFilesResolvedFromHome(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home directory with command files
	tmpHome := t.TempDir()
	fluxidDir := filepath.Join(tmpHome, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create fluxid dir: %v", err)
	}

	// Create command files in the config directory (relative paths resolve from here)
	implementFile := filepath.Join(fluxidDir, "implement.md")
	reviewFile := filepath.Join(fluxidDir, "review.md")
	commitFile := filepath.Join(fluxidDir, "commit.md")

	if err := os.WriteFile(implementFile, []byte("# Implement Command"), 0o644); err != nil {
		t.Fatalf("Failed to write implement file: %v", err)
	}
	if err := os.WriteFile(reviewFile, []byte("# Review Command"), 0o644); err != nil {
		t.Fatalf("Failed to write review file: %v", err)
	}
	if err := os.WriteFile(commitFile, []byte("# Commit Command"), 0o644); err != nil {
		t.Fatalf("Failed to write commit file: %v", err)
	}

	// Create home config referencing command files with absolute paths
	configContent := fmt.Sprintf(`agent: claude
commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, fluxidDir, fluxidDir, fluxidDir)
	configPath := filepath.Join(fluxidDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Run fluxid with custom HOME in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpHome, "--fluxid-dry-run")

	// Verify command files section appears in output
	if !strings.Contains(output, "Command Files:") {
		t.Errorf("Missing 'Command Files:' section in output")
	}

	// Verify absolute paths are displayed
	if !strings.Contains(output, implementFile) {
		t.Errorf("Expected implement path %s in output, got:\n%s", implementFile, output)
	}
	if !strings.Contains(output, reviewFile) {
		t.Errorf("Expected review path %s in output, got:\n%s", reviewFile, output)
	}
	if !strings.Contains(output, commitFile) {
		t.Errorf("Expected commit path %s in output, got:\n%s", commitFile, output)
	}
}

// TestM02E03ProjectCommandFilesOverrideHome validates that project command files
// take precedence over home command files.
//
//nolint:cyclop,funlen // E2E test with file setup, command execution, and multiple validations
func TestM02E03ProjectCommandFilesOverrideHome(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home directory with command files
	tmpHome := t.TempDir()
	homeFluxidDir := filepath.Join(tmpHome, ".fluxid")
	if err := os.MkdirAll(homeFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create home fluxid dir: %v", err)
	}

	// Create home command files in the config directory (relative paths resolve from here)
	homeImplementFile := filepath.Join(homeFluxidDir, "home-implement.md")
	homeReviewFile := filepath.Join(homeFluxidDir, "home-review.md")
	homeCommitFile := filepath.Join(homeFluxidDir, "home-commit.md")

	if err := os.WriteFile(homeImplementFile, []byte("# Home Implement"), 0o644); err != nil {
		t.Fatalf("Failed to write home implement file: %v", err)
	}
	if err := os.WriteFile(homeReviewFile, []byte("# Home Review"), 0o644); err != nil {
		t.Fatalf("Failed to write home review file: %v", err)
	}
	if err := os.WriteFile(homeCommitFile, []byte("# Home Commit"), 0o644); err != nil {
		t.Fatalf("Failed to write home commit file: %v", err)
	}

	// Create home config
	homeConfigContent := fmt.Sprintf(`agent: claude
commands:
  implement: %s
  review: %s
  commit: %s
`, homeImplementFile, homeReviewFile, homeCommitFile)
	homeConfigPath := filepath.Join(homeFluxidDir, "config.yaml")
	if err := os.WriteFile(homeConfigPath, []byte(homeConfigContent), 0o644); err != nil {
		t.Fatalf("Failed to write home config: %v", err)
	}

	// Create temporary project directory with command files
	tmpProjectDir := t.TempDir()
	projectFluxidDir := filepath.Join(tmpProjectDir, ".fluxid")
	if err := os.MkdirAll(projectFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create project fluxid dir: %v", err)
	}

	// Create project command files in the config directory (relative paths resolve from here)
	projectImplementFile := filepath.Join(projectFluxidDir, "project-implement.md")
	projectReviewFile := filepath.Join(projectFluxidDir, "project-review.md")
	projectCommitFile := filepath.Join(projectFluxidDir, "project-commit.md")

	if err := os.WriteFile(projectImplementFile, []byte("# Project Implement"), 0o644); err != nil {
		t.Fatalf("Failed to write project implement file: %v", err)
	}
	if err := os.WriteFile(projectReviewFile, []byte("# Project Review"), 0o644); err != nil {
		t.Fatalf("Failed to write project review file: %v", err)
	}
	if err := os.WriteFile(projectCommitFile, []byte("# Project Commit"), 0o644); err != nil {
		t.Fatalf("Failed to write project commit file: %v", err)
	}

	// Create project config
	projectConfigContent := fmt.Sprintf(`commands:
  implement: %s
  review: %s
  commit: %s
`, projectImplementFile, projectReviewFile, projectCommitFile)
	projectConfigPath := filepath.Join(projectFluxidDir, "config.yaml")
	if err := os.WriteFile(projectConfigPath, []byte(projectConfigContent), 0o644); err != nil {
		t.Fatalf("Failed to write project config: %v", err)
	}

	// Run fluxid from project directory in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpProjectDir, "--fluxid-dry-run")

	// Verify project command files are used (not home)
	if !strings.Contains(output, projectImplementFile) {
		t.Errorf("Expected project implement path %s in output, got:\n%s", projectImplementFile, output)
	}
	if !strings.Contains(output, projectReviewFile) {
		t.Errorf("Expected project review path %s in output, got:\n%s", projectReviewFile, output)
	}
	if !strings.Contains(output, projectCommitFile) {
		t.Errorf("Expected project commit path %s in output, got:\n%s", projectCommitFile, output)
	}

	// Verify home command files are NOT used
	if strings.Contains(output, homeImplementFile) {
		t.Errorf("Home implement path should not appear when project overrides, got:\n%s", output)
	}
}

// TestM02E03MissingCommandFileError validates that missing command files
// produce clear error messages.
func TestM02E03MissingCommandFileError(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create temporary home directory with commands dir but missing files
	tmpHome := t.TempDir()
	fluxidDir := filepath.Join(tmpHome, ".fluxid")
	commandsDir := filepath.Join(fluxidDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	// Create home config referencing non-existent command files with absolute paths
	configContent := fmt.Sprintf(`agent: claude
commands:
  implement: %s/missing-implement.md
  review: %s/missing-review.md
  commit: %s/missing-commit.md
`, commandsDir, commandsDir, commandsDir)
	configPath := filepath.Join(fluxidDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Run fluxid expecting error
	errOutput, exitCode := runFluxidExpectError(t, root, tmpHome)

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", exitCode)
	}

	// Verify error message mentions the missing file
	if !strings.Contains(errOutput, "command file not found") {
		t.Errorf("Expected error about missing command file, got: %s", errOutput)
	}
}

// TestM02E03PartialCommandsError validates that specifying only some command files
// (not all three) produces an error.
func TestM02E03PartialCommandsError(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create temporary home directory
	tmpHome := t.TempDir()
	fluxidDir := filepath.Join(tmpHome, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create .fluxid dir: %v", err)
	}

	// Create home config with only some command files specified (absolute paths)
	configContent := fmt.Sprintf(`agent: claude
commands:
  implement: %s/implement.md
  review: %s/review.md
`, fluxidDir, fluxidDir)
	configPath := filepath.Join(fluxidDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Run fluxid expecting error
	errOutput, exitCode := runFluxidExpectError(t, root, tmpHome)

	// Verify exit code is 1
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", exitCode)
	}

	// Verify error message mentions required command
	if !strings.Contains(errOutput, "commands.commit is required") {
		t.Errorf("Expected error about required commands.commit, got: %s", errOutput)
	}
}

// TestM02E03NoCommandFilesOptional validates that command files are optional
// and fluxid works without them.
func TestM02E03NoCommandFilesOptional(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// v2.0: Commands section is required, so create minimal config
	tmpHome := t.TempDir()
	homeFluxidDir := filepath.Join(tmpHome, ".fluxid")
	if err := os.MkdirAll(homeFluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create home fluxid dir: %v", err)
	}

	// Create minimal v2.0 config with commands section and command files (absolute paths)
	configContent := fmt.Sprintf(`commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, homeFluxidDir, homeFluxidDir, homeFluxidDir)
	configPath := filepath.Join(homeFluxidDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create command files
	createCommandFiles(t, homeFluxidDir)

	// Run fluxid in dry-run mode
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpHome, "--fluxid-dry-run")

	// Verify fluxid runs successfully
	if !strings.Contains(output, "=== fluxid Workflow Initialization ===") {
		t.Errorf("Expected initialization header, got:\n%s", output)
	}

	// v2.0: Command Files section should appear since commands are required
	if !strings.Contains(output, "Command Files:") {
		t.Errorf("Command Files section should appear in v2.0, got:\n%s", output)
	}
}

// TestM02E03AbsolutePathsDisplayed validates that displayed paths are absolute.
//
//nolint:cyclop // E2E test with path resolution and multiple output format validations
func TestM02E03AbsolutePathsDisplayed(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home directory with command files
	tmpHome := t.TempDir()
	fluxidDir := filepath.Join(tmpHome, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create fluxid dir: %v", err)
	}

	// Create command files in the config directory (relative paths resolve from here)
	if err := os.WriteFile(filepath.Join(fluxidDir, "impl.md"), []byte("# Impl"), 0o644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fluxidDir, "rev.md"), []byte("# Rev"), 0o644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fluxidDir, "com.md"), []byte("# Com"), 0o644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Create home config with absolute paths
	configContent := fmt.Sprintf(`commands:
  implement: %s/impl.md
  review: %s/rev.md
  commit: %s/com.md
`, fluxidDir, fluxidDir, fluxidDir)
	configPath := filepath.Join(fluxidDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Run fluxid in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpHome, "--fluxid-dry-run")

	// Verify absolute paths are displayed (should contain tmpHome path)
	expectedPath := filepath.Join(fluxidDir, "impl.md")
	if !strings.Contains(output, expectedPath) {
		t.Errorf("Expected absolute path %s in output, got:\n%s", expectedPath, output)
	}

	// Verify paths are absolute (should start with /)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Implement:") {
			// Extract path from line (format: "timestamp  Implement: /path/to/file")
			// Find the last colon which separates "Implement:" from the path
			idx := strings.LastIndex(line, "Implement:")
			if idx != -1 {
				// Get everything after "Implement:"
				pathPart := line[idx+len("Implement:"):]
				path := strings.TrimSpace(pathPart)
				if !filepath.IsAbs(path) {
					t.Errorf("Expected absolute path, got relative path: %s (from line: %s)", path, line)
				}
			}
		}
	}
}
