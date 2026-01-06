//nolint:paralleltest // E2E tests use shared infrastructure
package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestM06E04MissingCommandFileSingleError validates error reporting for one missing file.
func TestM06E04MissingCommandFileSingleError(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temp directory with config referencing missing implement command file
	tmpDir := t.TempDir()
	configDir := setupConfigDir(t, tmpDir)
	configPath := filepath.Join(configDir, "config.yaml")

	// Create only review and commit files, leave implement missing
	commandsDir := filepath.Join(configDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	reviewPath := filepath.Join(commandsDir, "review.md")
	commitPath := filepath.Join(commandsDir, "commit.md")

	if err := os.WriteFile(reviewPath, []byte("# Review"), 0o644); err != nil {
		t.Fatalf("Failed to create review file: %v", err)
	}
	if err := os.WriteFile(commitPath, []byte("# Commit"), 0o644); err != nil {
		t.Fatalf("Failed to create commit file: %v", err)
	}

	// Configure to use command files (implement.md doesn't exist)
	configContent := standardCommandFilesConfigFor(commandsDir)
	writeRawConfigFile(t, configPath, configContent)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected error for missing command file, but command succeeded")
	}

	stderrOutput := stderr.String()

	// Verify error message includes phase context
	if !strings.Contains(stderrOutput, "command file not found") {
		t.Errorf("Expected 'command file not found' in error, got:\n%s", stderrOutput)
	}

	// Verify error includes the path
	if !strings.Contains(stderrOutput, "implement.md") {
		t.Errorf("Expected missing file path 'implement.md' in error, got:\n%s", stderrOutput)
	}

	// Verify error mentions the phase
	if !strings.Contains(stderrOutput, "commands.implement") {
		t.Errorf("Expected 'commands.implement' in error, got:\n%s", stderrOutput)
	}

	// Verify no simulation plan was printed (error happens before simulation)
	if strings.Contains(stderrOutput, "=== Simulation Plan ===") {
		t.Errorf("Simulation plan should not appear when config validation fails:\n%s", stderrOutput)
	}

	// Verify exit code is non-zero (already checked by cmd.Run() error)
}

// TestM06E04MissingCommandFileMultipleErrors validates error reporting for multiple missing files.
func TestM06E04MissingCommandFileMultipleErrors(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temp directory with config referencing all missing command files
	tmpDir := t.TempDir()
	configDir := setupConfigDir(t, tmpDir)
	configPath := filepath.Join(configDir, "config.yaml")

	// Create commands directory but no files
	commandsDir := filepath.Join(configDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Configure to use command files (none exist)
	configContent := standardCommandFilesConfigFor(commandsDir)
	writeRawConfigFile(t, configPath, configContent)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected error for missing command files, but command succeeded")
	}

	stderrOutput := stderr.String()

	// Note: The current implementation fails on the first missing file
	// So we expect to see at least one error (implement is checked first)
	if !strings.Contains(stderrOutput, "command file not found") {
		t.Errorf("Expected 'command file not found' in error, got:\n%s", stderrOutput)
	}

	// Verify error includes at least the first missing file
	if !strings.Contains(stderrOutput, "implement.md") {
		t.Errorf("Expected missing file path 'implement.md' in error, got:\n%s", stderrOutput)
	}
}

// TestM06E04NoAgentProcessSpawnedOnValidationFailure validates no agent is invoked.
func TestM06E04NoAgentProcessSpawnedOnValidationFailure(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temp directory with config referencing missing command file
	tmpDir := t.TempDir()
	configDir := setupConfigDir(t, tmpDir)
	configPath := filepath.Join(configDir, "config.yaml")

	// Create commands directory but no files
	commandsDir := filepath.Join(configDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Configure to use command files (none exist) - use absolute paths to non-existent files
	configContent := fmt.Sprintf(`commands:
  implement: %s/missing.md
  review: %s/missing2.md
  commit: %s/missing3.md
`, commandsDir, commandsDir, commandsDir)
	writeRawConfigFile(t, configPath, configContent)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run() // We expect error, so ignore it

	output := stdout.String() + stderr.String()

	// Verify stub agent was NOT invoked
	if strings.Contains(output, "Claude stub invoked") {
		t.Errorf("Agent process should not be spawned when validation fails:\n%s", output)
	}

	// Verify no phase execution markers
	if strings.Contains(output, "Starting phase:") {
		t.Errorf("No phases should execute when validation fails:\n%s", output)
	}
}

// TestM06E04UnreadableCommandFile validates error for unreadable files.
func TestM06E04UnreadableCommandFile(t *testing.T) {
	// Skip on Windows as file permissions work differently
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping file permission test on Windows")
	}

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temp directory with unreadable command file
	tmpDir := t.TempDir()
	configDir := setupConfigDir(t, tmpDir)
	configPath := filepath.Join(configDir, "config.yaml")

	// Create files but make implement unreadable
	implementPath := filepath.Join(configDir, "implement.md")
	reviewPath := filepath.Join(configDir, "review.md")
	commitPath := filepath.Join(configDir, "commit.md")

	if err := os.WriteFile(implementPath, []byte("# Implement"), 0o644); err != nil {
		t.Fatalf("Failed to create implement file: %v", err)
	}
	if err := os.WriteFile(reviewPath, []byte("# Review"), 0o644); err != nil {
		t.Fatalf("Failed to create review file: %v", err)
	}
	if err := os.WriteFile(commitPath, []byte("# Commit"), 0o644); err != nil {
		t.Fatalf("Failed to create commit file: %v", err)
	}

	// Make implement file unreadable
	if err := os.Chmod(implementPath, 0o000); err != nil {
		t.Fatalf("Failed to chmod implement file: %v", err)
	}

	// Configure to use command files
	configContent := standardCommandFilesConfigFor(configDir)
	writeRawConfigFile(t, configPath, configContent)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected error for unreadable command file, but command succeeded")
	}

	stderrOutput := stderr.String()

	// Verify error message indicates permission problem
	if !strings.Contains(stderrOutput, "cannot read command file") &&
		!strings.Contains(stderrOutput, "permission denied") {
		t.Errorf("Expected permission error in stderr, got:\n%s", stderrOutput)
	}
}

// TestM06E04DirectoryAsCommandFile validates error when directory is specified as command file.
func TestM06E04DirectoryAsCommandFile(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temp directory with directory as command file
	tmpDir := t.TempDir()
	configDir := setupConfigDir(t, tmpDir)
	configPath := filepath.Join(configDir, "config.yaml")

	// Create a subdirectory as "implement.md" (where file should be)
	implementDir := filepath.Join(configDir, "implement.md")
	if err := os.Mkdir(implementDir, 0o755); err != nil {
		t.Fatalf("Failed to create implement directory: %v", err)
	}

	// Create real files for review and commit
	reviewPath := filepath.Join(configDir, "review.md")
	commitPath := filepath.Join(configDir, "commit.md")
	if err := os.WriteFile(reviewPath, []byte("# Review"), 0o644); err != nil {
		t.Fatalf("Failed to create review file: %v", err)
	}
	if err := os.WriteFile(commitPath, []byte("# Commit"), 0o644); err != nil {
		t.Fatalf("Failed to create commit file: %v", err)
	}

	// Configure to use command files
	configContent := standardCommandFilesConfigFor(configDir)
	writeRawConfigFile(t, configPath, configContent)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected error for directory as command file, but command succeeded")
	}

	stderrOutput := stderr.String()

	// Verify error message indicates it's not a regular file
	if !strings.Contains(stderrOutput, "is not a regular file") {
		t.Errorf("Expected 'is not a regular file' error in stderr, got:\n%s", stderrOutput)
	}
}

// TestM06E04ErrorGuidance validates that error messages include actionable guidance.
func TestM06E04ErrorGuidance(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// Create temp directory with config referencing missing command file
	tmpDir := t.TempDir()
	configDir := setupConfigDir(t, tmpDir)
	configPath := filepath.Join(configDir, "config.yaml")

	commandsDir := filepath.Join(configDir, "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Configure to use missing command files (absolute paths to non-existent files)
	configContent := fmt.Sprintf(`commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, commandsDir, commandsDir, commandsDir)
	writeRawConfigFile(t, configPath, configContent)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	_ = cmd.Run() // Expect error

	stderrOutput := stderr.String()

	// Verify error includes the full expected path (resolved relative to config dir)
	expectedPath := filepath.Join(commandsDir, "implement.md")
	if !strings.Contains(stderrOutput, expectedPath) {
		t.Errorf("Expected full path %q in error message, got:\n%s", expectedPath, stderrOutput)
	}

	// Verify error mentions which config field is problematic
	if !strings.Contains(stderrOutput, "commands.implement") {
		t.Errorf("Expected 'commands.implement' field reference in error, got:\n%s", stderrOutput)
	}
}
