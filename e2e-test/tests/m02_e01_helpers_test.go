package tests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
func writeHomeConfig(t *testing.T, fluxidDir, content string) {
	t.Helper()

	configPath := filepath.Join(fluxidDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
}

// runFluxidWithHome runs fluxid with a custom HOME directory and returns stdout.
func runFluxidWithHome(t *testing.T, root, homeDir string) string {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", homeDir),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
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

	binPath := filepath.Join(root, "bin", "fluxid")
	cmdArgs := append([]string{"--claude"}, args...)
	cmd := exec.CommandContext(t.Context(), binPath, cmdArgs...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", homeDir),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
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
	cmd := exec.CommandContext(t.Context(), binPath, "--claude")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", homeDir),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
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
	cmd := exec.CommandContext(t.Context(), binPath, "--claude")
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", homeDir),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, stdout.String())
	}
}
