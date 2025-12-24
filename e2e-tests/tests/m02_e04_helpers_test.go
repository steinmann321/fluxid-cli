package tests

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// runFluxidInDirWithEnv runs fluxid with environment variables set.
func runFluxidInDirWithEnv(t *testing.T, root, homeDir, workDir string, envVars map[string]string) string {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)
	cmd.Dir = workDir

	// Build environment with custom vars
	env := append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)
	for key, val := range envVars {
		env = append(env, key+"="+val)
	}
	cmd.Env = env

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, stdout.String())
	}

	return stdout.String()
}

// runFluxidInDirWithEnvAndArgs runs fluxid with env vars and CLI args.
func runFluxidInDirWithEnvAndArgs(
	t *testing.T, root, homeDir, workDir string, envVars map[string]string, args ...string,
) string {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Dir = workDir

	// Build environment with custom vars
	env := append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)
	for key, val := range envVars {
		env = append(env, key+"="+val)
	}
	cmd.Env = env

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, stdout.String())
	}

	return stdout.String()
}

// runFluxidInDirWithEnvExpectError runs fluxid with env vars expecting failure.
func runFluxidInDirWithEnvExpectError(
	t *testing.T, root, homeDir, workDir string, envVars map[string]string,
) (string, int) {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)
	cmd.Dir = workDir

	// Build environment with custom vars
	env := append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)
	for key, val := range envVars {
		env = append(env, key+"="+val)
	}
	cmd.Env = env

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
