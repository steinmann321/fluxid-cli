//nolint:gocritic // Test helper with string concatenation
package tests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Common test configuration constants.
// v2.0: commit_enabled removed (Phase 10 - commits always enabled).
const (
	basicHomeConfig = `iterations: 10
implement_retries: 5
`
	fullHomeConfig = `agent: claude
implement_retries: 5
iterations: 10
`
)

// createHomeConfigDir creates a ~/.fluxid directory in the given home directory.
func createHomeConfigDir(t *testing.T, homeDir string) string {
	t.Helper()

	fluxidDir := filepath.Join(homeDir, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create .fluxid dir: %v", err)
	}
	return fluxidDir
}

// writeHomeConfig writes a config.yaml file to ~/.fluxid/config.yaml.
// v2.0: Automatically adds commands section and creates command files if not present.
func writeHomeConfig(t *testing.T, fluxidDir, content string) {
	t.Helper()

	// v2.0: Ensure commands section is present
	if !strings.Contains(content, "commands:") {
		// Ensure there's a newline before the commands section
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content = content + fmt.Sprintf(`commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, fluxidDir, fluxidDir, fluxidDir)
	}

	configPath := filepath.Join(fluxidDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// v2.0: Create command files
	createCommandFiles(t, fluxidDir)
}

// runFluxidWithHome runs fluxid with a custom HOME directory and returns stdout.
func runFluxidWithHome(t *testing.T, root, homeDir string) string {
	t.Helper()

	// Create a dummy task file in home
	taskPath := filepath.Join(homeDir, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, stdout.String())
	}

	return stdout.String()
}

// runFluxidWithHomeAndArgs runs fluxid with custom HOME directory and additional arguments.
func runFluxidWithHomeAndArgs(t *testing.T, root, homeDir string, args ...string) string {
	t.Helper()

	// Ensure a dummy task file exists
	taskPath := filepath.Join(homeDir, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}
	args = append(args, "--file="+taskPath)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"PATH"+":"+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, stdout.String())
	}

	return stdout.String()
}

// runFluxidExpectError runs fluxid expecting it to fail and returns the stderr output.
func runFluxidExpectError(t *testing.T, root, homeDir string) (string, int) {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected fluxid to fail, but it succeeded")
	}

	var exitCode int
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}

	return stderr.String(), exitCode
}

// setupHomeWithConfig creates a temporary home directory with the given config content.
func setupHomeWithConfig(t *testing.T, configContent string) string {
	t.Helper()

	tmpHome := t.TempDir()
	fluxidDir := createHomeConfigDir(t, tmpHome)
	writeHomeConfig(t, fluxidDir, configContent)
	return tmpHome
}

// runFluxidInDir runs fluxid in a specific working directory.
func runFluxidInDir(t *testing.T, root, homeDir, workDir string) {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, stdout.String())
	}
}

// runFluxidInDirWithOutput runs fluxid in a specific working directory and returns output.

// createProjectWithConfig creates a temporary project dir with .fluxid/config.yaml content.
// v2.0: Automatically adds commands section and creates command files if not present.
func createProjectWithConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	fluxidDir := filepath.Join(dir, ".fluxid")
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create project .fluxid dir: %v", err)
	}

	// v2.0: Ensure commands section is present
	if !strings.Contains(content, "commands:") {
		// Ensure there's a newline before the commands section
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content = content + fmt.Sprintf(`commands:
  implement: %s/implement.md
  review: %s/review.md
  commit: %s/commit.md
`, fluxidDir, fluxidDir, fluxidDir)
	}

	cfgPath := filepath.Join(fluxidDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write project config: %v", err)
	}

	// v2.0: Create command files
	createCommandFiles(t, fluxidDir)

	return dir
}

// runFluxidInDirWithArgs runs fluxid in a specific working directory with additional arguments.
func runFluxidInDirWithArgs(t *testing.T, root, homeDir, workDir string, args ...string) string {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, stdout.String())
	}

	return stdout.String()
}

// runFluxidInDirExpectError runs fluxid in a directory expecting it to fail.
func runFluxidInDirExpectError(t *testing.T, root, homeDir, workDir string) (string, int) {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected fluxid to fail, but it succeeded")
	}

	var exitCode int
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}

	return stderr.String(), exitCode
}
