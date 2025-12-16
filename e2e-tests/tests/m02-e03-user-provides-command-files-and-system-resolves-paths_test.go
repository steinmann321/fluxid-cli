package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestM02E03CommandFilesResolvedFromHome validates that command files are
// resolved from home config and absolute paths are displayed.
func TestM02E03CommandFilesResolvedFromHome(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home directory with command files
	tmpHome := t.TempDir()
	fluxidDir := filepath.Join(tmpHome, ".fluxid")
	commandsDir := filepath.Join(fluxidDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	// Create command files
	implementFile := filepath.Join(commandsDir, "implement.md")
	reviewFile := filepath.Join(commandsDir, "review.md")
	commitFile := filepath.Join(commandsDir, "commit.md")

	if err := os.WriteFile(implementFile, []byte("# Implement Command"), 0o644); err != nil {
		t.Fatalf("Failed to write implement file: %v", err)
	}
	if err := os.WriteFile(reviewFile, []byte("# Review Command"), 0o644); err != nil {
		t.Fatalf("Failed to write review file: %v", err)
	}
	if err := os.WriteFile(commitFile, []byte("# Commit Command"), 0o644); err != nil {
		t.Fatalf("Failed to write commit file: %v", err)
	}

	// Create home config referencing command files
	configContent := `agent: claude
commands:
  implement: implement.md
  review: review.md
  commit: commit.md
`
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
func TestM02E03ProjectCommandFilesOverrideHome(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home directory with command files
	tmpHome := t.TempDir()
	homeFluxidDir := filepath.Join(tmpHome, ".fluxid")
	homeCommandsDir := filepath.Join(homeFluxidDir, "commands")
	if err := os.MkdirAll(homeCommandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create home commands dir: %v", err)
	}

	// Create home command files
	homeImplementFile := filepath.Join(homeCommandsDir, "home-implement.md")
	homeReviewFile := filepath.Join(homeCommandsDir, "home-review.md")
	homeCommitFile := filepath.Join(homeCommandsDir, "home-commit.md")

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
	homeConfigContent := `agent: claude
commands:
  implement: home-implement.md
  review: home-review.md
  commit: home-commit.md
`
	homeConfigPath := filepath.Join(homeFluxidDir, "config.yaml")
	if err := os.WriteFile(homeConfigPath, []byte(homeConfigContent), 0o644); err != nil {
		t.Fatalf("Failed to write home config: %v", err)
	}

	// Create temporary project directory with command files
	tmpProjectDir := t.TempDir()
	projectFluxidDir := filepath.Join(tmpProjectDir, ".fluxid")
	projectCommandsDir := filepath.Join(projectFluxidDir, "commands")
	if err := os.MkdirAll(projectCommandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create project commands dir: %v", err)
	}

	// Create project command files
	projectImplementFile := filepath.Join(projectCommandsDir, "project-implement.md")
	projectReviewFile := filepath.Join(projectCommandsDir, "project-review.md")
	projectCommitFile := filepath.Join(projectCommandsDir, "project-commit.md")

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
	projectConfigContent := `commands:
  implement: project-implement.md
  review: project-review.md
  commit: project-commit.md
`
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
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create temporary home directory with commands dir but missing files
	tmpHome := t.TempDir()
	fluxidDir := filepath.Join(tmpHome, ".fluxid")
	commandsDir := filepath.Join(fluxidDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	// Create home config referencing non-existent command files
	configContent := `agent: claude
commands:
  implement: missing-implement.md
  review: missing-review.md
  commit: missing-commit.md
`
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
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create temporary home directory
	tmpHome := t.TempDir()
	fluxidDir := filepath.Join(tmpHome, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create .fluxid dir: %v", err)
	}

	// Create home config with only some command files specified
	configContent := `agent: claude
commands:
  implement: implement.md
  review: review.md
`
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
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home directory without command files
	tmpHome := t.TempDir()

	// Run fluxid without any command files in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpHome, "--fluxid-dry-run")

	// Verify fluxid runs successfully
	if !strings.Contains(output, "=== fluxid Workflow Initialization ===") {
		t.Errorf("Expected initialization header, got:\n%s", output)
	}

	// Verify "Command Files:" section does NOT appear
	if strings.Contains(output, "Command Files:") {
		t.Errorf("Command Files section should not appear when no commands configured, got:\n%s", output)
	}
}

// TestM02E03AbsolutePathsDisplayed validates that displayed paths are absolute.
func TestM02E03AbsolutePathsDisplayed(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temporary home directory with command files
	tmpHome := t.TempDir()
	fluxidDir := filepath.Join(tmpHome, ".fluxid")
	commandsDir := filepath.Join(fluxidDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	// Create command files
	if err := os.WriteFile(filepath.Join(commandsDir, "impl.md"), []byte("# Impl"), 0o644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "rev.md"), []byte("# Rev"), 0o644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "com.md"), []byte("# Com"), 0o644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Create home config with relative paths
	configContent := `commands:
  implement: impl.md
  review: rev.md
  commit: com.md
`
	configPath := filepath.Join(fluxidDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Run fluxid in dry-run mode (only need init status)
	output := runFluxidInDirWithArgs(t, root, tmpHome, tmpHome, "--fluxid-dry-run")

	// Verify absolute paths are displayed (should contain tmpHome path)
	expectedPath := filepath.Join(commandsDir, "impl.md")
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
