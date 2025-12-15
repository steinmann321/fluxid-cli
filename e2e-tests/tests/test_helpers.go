package tests

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Common test constants.
const (
	standardCommandFilesConfig = `commands:
  implement: implement.md
  review: review.md
  commit: commit.md
`
)

// readCombinedOutput reads from stdout and stderr pipes concurrently
// and combines them into a single buffer. Optionally handles stdin interaction
// when a specific prompt is detected.
func readCombinedOutput(
	stdout, stderr io.Reader,
	stdin io.WriteCloser,
	promptMarker, stdinResponse string,
	timeout time.Duration,
) (string, error) {
	var output bytes.Buffer
	var outputMutex sync.Mutex
	done := make(chan error, 1)
	promptSeen := false

	var wg sync.WaitGroup
	wg.Add(2)

	// Read stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputMutex.Lock()
			output.WriteString(line + "\n")
			outputMutex.Unlock()

			// Handle interactive prompt if configured
			if promptMarker != "" && strings.Contains(line, promptMarker) && !promptSeen {
				promptSeen = true
				time.Sleep(50 * time.Millisecond)
				if stdin != nil && stdinResponse != "" {
					if _, err := stdin.Write([]byte(stdinResponse + "\n")); err != nil {
						done <- err
						return
					}
				}
			}
		}
		if scanner.Err() != nil {
			done <- scanner.Err()
		}
	}()

	// Read stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			outputMutex.Lock()
			output.WriteString(line + "\n")
			outputMutex.Unlock()
		}
		if scanner.Err() != nil {
			done <- scanner.Err()
		}
	}()

	// Wait for both readers to finish
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, io.EOF) {
			return output.String(), fmt.Errorf("error reading output: %w", err)
		}
	case <-time.After(timeout):
		return output.String(), fmt.Errorf("timeout after %v", timeout)
	}

	return output.String(), nil
}

// findProjectRoot walks up from the starting directory to find the project root.
func findProjectRoot(start string) (string, error) {
	cur := start
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return "", errors.New("project root with go.mod not found")
}

// getProjectRoot returns the absolute path to the project root directory.
func getProjectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	root, err := findProjectRoot(wd)
	if err != nil {
		t.Fatalf("find project root failed: %v", err)
	}

	return root
}

// buildFluxid builds the fluxid binary for testing.
func buildFluxid(t *testing.T, root string) {
	t.Helper()

	// Build fluxid binary
	build := exec.CommandContext(context.Background(), "go", "build", "-o", "bin/fluxid", "./cmd/fluxid")
	build.Dir = root

	var stderr bytes.Buffer
	build.Stderr = &stderr

	if err := build.Run(); err != nil {
		t.Fatalf("build failed: %v\nStderr: %s", err, stderr.String())
	}
}

// testAgentArgsPassthrough is a helper to test that agent-specific arguments are passed through correctly.
// It runs fluxid with the given agent flag and custom args, then verifies the args appear in output.
func testAgentArgsPassthrough(t *testing.T, root, agentFlag, argsLabel string, customArgs []string) {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	args := append([]string{agentFlag}, customArgs...)
	// #nosec G204 -- Test helper with controlled test inputs
	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

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

// setupConfigDir creates a .fluxid config directory and returns its path.
func setupConfigDir(t *testing.T, baseDir string) string {
	t.Helper()
	fluxidDir := filepath.Join(baseDir, ".fluxid")
	// #nosec G301 -- Test fixture directory with standard permissions
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create .fluxid directory: %v", err)
	}
	return fluxidDir
}

// writeConfigFile writes a config file with the given agent name.
func writeConfigFile(t *testing.T, configPath, agent string) {
	t.Helper()
	config := fmt.Sprintf("agent: %s\n", agent)
	// #nosec G306 -- Test fixture file with standard permissions
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
}

// writeRawConfigFile writes raw config content to the specified path.
func writeRawConfigFile(t *testing.T, configPath, content string) {
	t.Helper()
	// #nosec G306 -- Test fixture file with standard permissions
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
}

// runFluxidWithConfig runs fluxid with custom home, working dir, env, and CLI args.
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
		cmdEnv = append(cmdEnv, fmt.Sprintf("HOME=%s", homeDir))
	}
	cmdEnv = append(cmdEnv, env...)
	cmd.Env = cmdEnv

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	return stdout.String(), err
}

// verifyConfigLine checks that a line containing fieldPrefix also contains sourcePattern.
// This handles file paths in source strings like "source: home (/path/to/config.yaml)".
func verifyConfigLine(t *testing.T, output, fieldPrefix, sourcePattern string) {
	t.Helper()
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, fieldPrefix) && strings.Contains(line, sourcePattern) {
			return
		}
	}
	t.Errorf("Expected line containing %q and %q, got:\n%s", fieldPrefix, sourcePattern, output)
}

// verifyAgentArgsAndSource is a helper to verify agent arguments and source in test output.
// This reduces duplication in M06 tests for JSON and YAML validation.
func verifyAgentArgsAndSource(
	t *testing.T,
	agentArgs []string,
	agent, agentSource string,
	expectedAgent, expectedSource string,
) {
	t.Helper()

	// Verify agent args are captured
	if len(agentArgs) != 2 {
		t.Errorf("Expected 2 agent args, got: %d (%v)", len(agentArgs), agentArgs)
	}
	if len(agentArgs) >= 2 {
		if agentArgs[0] != "arg1" || agentArgs[1] != "arg2" {
			t.Errorf("Expected agent args [arg1, arg2], got: %v", agentArgs)
		}
	}

	// Verify agent is as expected
	if agent != expectedAgent {
		t.Errorf("Expected agent %q, got: %q", expectedAgent, agent)
	}

	// Verify source is as expected
	if agentSource != expectedSource {
		t.Errorf("Expected agent_source %q, got: %q", expectedSource, agentSource)
	}
}

// runFluxidWithOutputFormat runs fluxid with the specified output format and returns stdout.
func runFluxidWithOutputFormat(
	t *testing.T,
	root, format string,
	extraArgs ...string,
) string {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	args := []string{"--fluxid-output", format, "--fluxid-dry-run", "--claude"}
	args = append(args, extraArgs...)

	// #nosec G204 -- Test helper with controlled test inputs
	cmd := exec.CommandContext(t.Context(), binPath, args...)

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

// verifyConfigValues checks that max review cycles and implement retries match expected values.
func verifyConfigValues(
	t *testing.T,
	maxReviewCycles, maxImplementRetries int,
	reviewCyclesSource, implementRetriesSource string,
	expectedReviewCycles, expectedImplementRetries int,
	expectedSource string,
) {
	t.Helper()

	if maxReviewCycles != expectedReviewCycles {
		t.Errorf("Expected max_review_cycles to be %d, got: %d", expectedReviewCycles, maxReviewCycles)
	}
	if maxImplementRetries != expectedImplementRetries {
		t.Errorf("Expected max_implement_retries to be %d, got: %d", expectedImplementRetries, maxImplementRetries)
	}

	if reviewCyclesSource != expectedSource {
		t.Errorf("Expected review_cycles_source to be %q, got: %q", expectedSource, reviewCyclesSource)
	}
	if implementRetriesSource != expectedSource {
		t.Errorf("Expected implement_retries_source to be %q, got: %q", expectedSource, implementRetriesSource)
	}
}

// runFluxidDefaultFormat runs fluxid without output format flag and returns stdout.
func runFluxidDefaultFormat(t *testing.T, root string) string {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	// #nosec G204 -- Test helper with controlled test inputs
	cmd := exec.CommandContext(t.Context(), binPath, "--fluxid-dry-run", "--claude")

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

// verifyDefaultTextFormat verifies output is in text format and not structured format.
// unmarshalFunc should attempt to unmarshal the output as the structured format
// and return an error if it's not valid.
func verifyDefaultTextFormat(
	t *testing.T,
	output string,
	unmarshalFunc func(string) error,
	formatName string,
) {
	t.Helper()

	// Verify text format markers are present
	if !strings.Contains(output, "=== fluxid Workflow Initialization ===") {
		t.Errorf("Expected text format header in output, got:\n%s", output)
	}

	// Verify it's NOT the structured format (should fail to parse)
	if err := unmarshalFunc(output); err == nil {
		t.Errorf("Expected text output, but got valid %s:\n%s", formatName, output)
	}
}
