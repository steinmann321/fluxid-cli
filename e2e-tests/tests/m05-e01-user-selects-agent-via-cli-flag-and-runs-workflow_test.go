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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			binPath := filepath.Join(root, "bin", "fluxid")
			cmd := exec.CommandContext(t.Context(), binPath, tt.flag)
			cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stdout

			if err := cmd.Run(); err != nil {
				t.Fatalf("fluxid %s failed: %v\nOutput:\n%s", tt.flag, err, stdout.String())
			}

			output := stdout.String()

			// Verify agent is shown in initialization
			expectedAgent := fmt.Sprintf("Agent: %s", tt.wantAgent)
			if !strings.Contains(output, expectedAgent) {
				t.Errorf("Expected agent %s in output, got:\n%s", tt.wantAgent, output)
			}

			// Verify source is CLI
			if !strings.Contains(output, "source: cli") {
				t.Errorf("Expected source: cli in output, got:\n%s", output)
			}

			// Verify workflow completion
			if !strings.Contains(output, "Status: SUCCESS") {
				t.Errorf("Expected successful completion, got:\n%s", output)
			}
		})
	}
}

// TestM05E01ExactlyOneAgentFlagRequired validates mutual exclusion of agent flags.
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			binPath := filepath.Join(root, "bin", "fluxid")
			cmd := exec.CommandContext(t.Context(), binPath, tt.args...)
			cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			if !tt.expectError {
				if err != nil {
					t.Fatalf("Expected success with args %v, got error: %v", tt.args, err)
				}
				return
			}

			// Error is expected from here on
			if err == nil {
				t.Fatalf("Expected error with args %v, but succeeded", tt.args)
			}

			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
				t.Errorf("Expected exit code 1, got: %v", err)
			}

			if !strings.Contains(stderr.String(), tt.errorText) {
				t.Errorf("Expected error text %q, got: %s", tt.errorText, stderr.String())
			}
		})
	}
}

// TestM05E01AgentBinaryPathResolution validates PATH resolution and executability.
func TestM05E01AgentBinaryPathResolution(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	t.Run("agent found in PATH", func(t *testing.T) {
		t.Parallel()

		binPath := filepath.Join(root, "bin", "fluxid")
		cmd := exec.CommandContext(t.Context(), binPath, "--codex")
		cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

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

		binPath := filepath.Join(root, "bin", "fluxid")
		cmd := exec.CommandContext(t.Context(), binPath, "--codex")

		// Set minimal PATH that doesn't include our stubs, but preserve HOME
		homeDir, _ := os.UserHomeDir()
		cmd.Env = []string{
			"PATH=/usr/bin:/bin",
			fmt.Sprintf("HOME=%s", homeDir),
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
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	testAgentArgsPassthrough(t, root, "--codex", "Agent Args:", []string{"--custom-arg", "test-value", "--another-flag"})
}

// TestM05E01OrchestrationMatchesBaseline validates workflow orchestration is unchanged.
func TestM05E01OrchestrationMatchesBaseline(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--opencode")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid --opencode failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()

	// Verify phase execution order matches baseline
	if !strings.Contains(output, "Review Cycle 1/20") {
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

	expectedEnv := fmt.Sprintf("FLUXID_SESSION_ID=%s", sessionID)
	if !strings.Contains(output, expectedEnv) {
		t.Errorf("FLUXID_SESSION_ID not propagated to agent process")
	}

	// Verify completion
	if !strings.Contains(output, "Status: SUCCESS") {
		t.Errorf("Missing success status in completion summary")
	}
}
