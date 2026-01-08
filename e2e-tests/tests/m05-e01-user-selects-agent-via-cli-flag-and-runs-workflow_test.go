//nolint:paralleltest // E2E tests use shared infrastructure
package tests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const uuidV4Pattern = `([0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})`

// TestM05E01UserSelectsAgentViaCLIFlag validates agent selection via CLI flags.
//
//nolint:funlen // E2E test with agent selection validation
func TestM05E01UserSelectsAgentViaCLIFlag(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tests := []struct {
		name      string
		agent     string
		flag      string
		wantAgent string
	}{
		{
			name:      "claude agent via --claude",
			agent:     "claude",
			flag:      "--claude",
			wantAgent: "claude",
		},
		{
			name:      "codex agent via --codex",
			agent:     "codex",
			flag:      "--codex",
			wantAgent: "codex",
		},
		{
			name:      "opencode agent via --opencode",
			agent:     "opencode",
			flag:      "--opencode",
			wantAgent: "opencode",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// v2.0: Create temporary home with config and command files
			tmpHome := t.TempDir()
			setupConfigWithCommands(t, tmpHome, "claude")

			binPath := filepath.Join(root, "bin", "fluxid")
			// Create a dummy task file in home
			taskPath := filepath.Join(tmpHome, "task.txt")
			if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(t.Context(), binPath, testCase.flag, "--fluxid-iterations=1", "--file="+taskPath)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
				"HOME="+tmpHome,
			)

			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stdout

			if err := cmd.Run(); err != nil {
				t.Fatalf("fluxid %s failed: %v\nOutput:\n%s", testCase.flag, err, stdout.String())
			}

			output := stdout.String()

			// v2.0: source tracking removed (Phase 9)
			// Verify agent is shown in initialization
			expectedAgent := "Agent: " + testCase.wantAgent
			if !strings.Contains(output, expectedAgent) {
				t.Errorf("Expected agent %s in output, got:\n%s", testCase.wantAgent, output)
			}

			// Verify workflow completion
			if !strings.Contains(output, "Status: SUCCESS") {
				t.Errorf("Expected successful completion, got:\n%s", output)
			}
		})
	}
}

// TestM05E01ExactlyOneAgentFlagRequired validates mutual exclusion of agent flags.
//
//nolint:funlen // E2E test with flag validation checks
func TestM05E01ExactlyOneAgentFlagRequired(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorText   string
	}{
		{
			name:        "no agent flag uses config/default",
			args:        []string{},
			expectError: false,
			errorText:   "",
		},
		{
			name:        "two agent flags",
			args:        []string{"--claude", "--codex"},
			expectError: true,
			errorText:   "multiple agent flags specified",
		},
		{
			name:        "three agent flags",
			args:        []string{"--claude", "--codex", "--opencode"},
			expectError: true,
			errorText:   "multiple agent flags specified",
		},
		{
			name:        "one agent flag succeeds",
			args:        []string{"--claude"},
			expectError: false,
			errorText:   "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// v2.0: Create temporary home with config and command files
			tmpHome := t.TempDir()
			setupConfigWithCommands(t, tmpHome, "claude")

			binPath := filepath.Join(root, "bin", "fluxid")
			args := testCase.args
			// Add minimal iterations for tests that run workflows (success cases).
			if !testCase.expectError {
				args = append(args, "--fluxid-iterations=1")
			}
			// Create a dummy task file in home
			taskPath := filepath.Join(tmpHome, "task.txt")
			if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
				t.Fatal(err)
			}
			args = append(args, "--file="+taskPath)
			cmd := exec.CommandContext(t.Context(), binPath, args...)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
				"HOME="+tmpHome,
			)

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			if !testCase.expectError {
				if err != nil {
					t.Fatalf("Expected success with args %v, got error: %v", testCase.args, err)
				}
				return
			}

			// Error is expected from here on
			if err == nil {
				t.Fatalf("Expected error with args %v, but succeeded", testCase.args)
			}

			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
				t.Errorf("Expected exit code 1, got: %v", err)
			}

			if !strings.Contains(stderr.String(), testCase.errorText) {
				t.Errorf("Expected error text %q, got: %s", testCase.errorText, stderr.String())
			}
		})
	}
}

// TestM05E01AgentBinaryPathResolution validates PATH resolution and executability.
//
//nolint:funlen // E2E test with PATH resolution validation
func TestM05E01AgentBinaryPathResolution(t *testing.T) {
	t.Parallel()
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	t.Run("agent found in PATH", func(t *testing.T) {
		t.Parallel()

		// v2.0: Create temporary home with config and command files
		tmpHome := t.TempDir()
		setupConfigWithCommands(t, tmpHome, "claude")

		// Create task file
		taskPath := filepath.Join(tmpHome, "task.txt")
		if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
			t.Fatalf("Failed to write task file: %v", err)
		}

		binPath := filepath.Join(root, "bin", "fluxid")
		cmd := exec.CommandContext(t.Context(), binPath, "--codex", "--fluxid-iterations=1", "--file="+taskPath)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
			"HOME="+tmpHome,
		)

		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stdout

		if err := cmd.Run(); err != nil {
			t.Fatalf("Expected success when codex is in PATH, got: %v\nOutput:\n%s", err, stdout.String())
		}

		// Verify agent was actually resolved and used
		if !strings.Contains(stdout.String(), "Agent: codex") {
			t.Errorf("Expected codex agent in output")
		}
	})

	t.Run("agent not in PATH", func(t *testing.T) {
		t.Parallel()

		// v2.0: Create temporary home with config and command files
		tmpHome := t.TempDir()
		setupConfigWithCommands(t, tmpHome, "claude")

		binPath := filepath.Join(root, "bin", "fluxid")
		cmd := exec.CommandContext(t.Context(), binPath, "--codex")

		// Set minimal PATH that doesn't include our stubs
		cmd.Env = []string{
			"PATH=/usr/bin:/bin",
			"HOME=" + tmpHome,
		}

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("Expected error when agent not in PATH, but succeeded")
		}

		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got: %v", err)
		}

		output := stderr.String()
		if !strings.Contains(output, "command not found") && !strings.Contains(output, "not found in PATH") {
			t.Errorf("Expected clear error message about missing agent, got: %s", output)
		}

		// Verify helpful suggestion is present
		if !strings.Contains(output, "which") {
			t.Errorf("Expected helpful 'which' suggestion in error, got: %s", output)
		}
	})
}

// TestM05E01AgentArgsPassthrough validates agent-specific args are forwarded.
func TestM05E01AgentArgsPassthrough(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	testAgentArgsPassthrough(t, root, "--codex", "Agent Args:", []string{"--custom-arg", "test-value", "--another-flag"})
}

// TestM05E01OrchestrationMatchesBaseline validates workflow orchestration is unchanged.
func TestM05E01OrchestrationMatchesBaseline(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	// v2.0: Create temporary home with config and command files
	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	// Create a dummy task file in home
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), binPath, "--opencode", "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --opencode failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify phase execution order matches baseline
	if !strings.Contains(output, "REVIEW CYCLE 1/1") {
		t.Errorf("Missing review cycle indicator")
	}

	if !strings.Contains(output, "Starting phase: implement") {
		t.Errorf("Missing implement phase")
	}

	if !strings.Contains(output, "Starting phase: review") {
		t.Errorf("Missing review phase")
	}

	// Verify session ID propagation
	sessionIDPattern := regexp.MustCompile(`Session ID: ` + uuidV4Pattern)
	matches := sessionIDPattern.FindStringSubmatch(output)
	if len(matches) < 2 {
		t.Fatal("Could not find session ID in output")
	}
	sessionID := matches[1]

	expectedEnv := "FLUXID_SESSION_ID=" + sessionID
	if !strings.Contains(output, expectedEnv) {
		t.Errorf("FLUXID_SESSION_ID not propagated to agent process")
	}

	// Verify completion
	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Missing success status in completion summary")
	}
}
