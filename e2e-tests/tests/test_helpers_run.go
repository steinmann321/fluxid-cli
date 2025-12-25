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

// testAgentArgsPassthrough is a helper to test that agent-specific arguments are passed through correctly.
// It runs fluxid with the given agent flag and custom args, then verifies the args appear in output.
func testAgentArgsPassthrough(t *testing.T, root, agentFlag, argsLabel string, customArgs []string) {
	t.Helper()

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	args := append([]string{agentFlag}, customArgs...)
	args = append(args, "--fluxid-iterations=1")
	// #nosec G204 -- Test helper with controlled test inputs
	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid with custom args failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify args label is shown in initialization
	if !strings.Contains(output, argsLabel) {
		t.Errorf("%s not displayed in output", argsLabel)
	}

	// Verify all custom args appear in output (from stub echo)
	for _, arg := range customArgs {
		if !strings.Contains(output, arg) {
			t.Errorf("Custom argument %q not passed through to agent", arg)
		}
	}
}

// runFluxidWithConfig runs fluxid with custom home, working dir, env, and CLI args.
//
//nolint:unparam // env parameter maintained for API consistency
func runFluxidWithConfig(t *testing.T, root, workDir, homeDir string, env []string, args []string) (string, error) {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test helper with controlled test inputs
	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Dir = workDir

	cmdEnv := append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
	)
	if homeDir != "" {
		cmdEnv = append(cmdEnv, "HOME="+homeDir)
	}
	cmdEnv = append(cmdEnv, env...)
	cmd.Env = cmdEnv

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	return stdout.String(), err
}

// runFluxidWithOutputFormat runs fluxid with the specified output format and returns stdout.
func runFluxidWithOutputFormat(
	t *testing.T,
	root, format string,
	extraArgs ...string,
) string {
	t.Helper()

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	// v2.0: equals syntax only (Phase 8)
	args := []string{"--fluxid-output=" + format, "--fluxid-dry-run", "--claude"}
	args = append(args, extraArgs...)

	// #nosec G204 -- Test helper with controlled test inputs
	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf(
			"fluxid --fluxid-output %s --fluxid-dry-run failed: %v\nStdout:\n%s\nStderr:\n%s",
			format, err, stdout.String(), stderr.String(),
		)
	}

	return stdout.String()
}

// runFluxidDefaultFormat runs fluxid without output format flag and returns stdout.
func runFluxidDefaultFormat(t *testing.T, root string) string {
	t.Helper()

	// Create temporary home with v2.0 config
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test helper with controlled test inputs
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nStdout:\n%s\nStderr:\n%s",
			err, stdout.String(), stderr.String())
	}

	return stdout.String()
}
