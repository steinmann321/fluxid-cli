package tests

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestM05E05AgentNotInPathError validates clear error when agent binary is missing from PATH.
//
//nolint:funlen // E2E test with runtime error handling
func TestM05E05AgentNotInPathError(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	tests := []struct {
		name       string
		agent      string
		wantError  string
		errorCheck func(t *testing.T, errMsg string)
	}{
		{
			name:      "codex not in PATH",
			agent:     "codex",
			wantError: "not found in PATH",
			errorCheck: func(t *testing.T, errMsg string) { //nolint:thelper // Test table validation func
				if !strings.Contains(errMsg, "codex") {
					t.Errorf("Expected error to mention agent 'codex', got:\n%s", errMsg)
				}
				if !strings.Contains(errMsg, "which") {
					t.Errorf("Expected error to suggest 'which' command, got:\n%s", errMsg)
				}
			},
		},
		{
			name:      "opencode not in PATH",
			agent:     "opencode",
			wantError: "not found in PATH",
			errorCheck: func(t *testing.T, errMsg string) { //nolint:thelper // Test table validation func
				if !strings.Contains(errMsg, "opencode") {
					t.Errorf("Expected error to mention agent 'opencode', got:\n%s", errMsg)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// v2.0: Create temporary home with config and command files
			tmpHome := t.TempDir()
			setupConfigWithCommands(t, tmpHome, "claude")

			binPath := filepath.Join(root, "bin", "fluxid")
			cmd := exec.CommandContext(t.Context(), binPath, "--"+testCase.agent)
			cmd.Env = []string{
				"PATH=/usr/bin:/bin",
				"HOME=" + tmpHome,
			}

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Error is expected
			if err == nil {
				t.Fatalf("Expected error when %s not in PATH, but succeeded", testCase.agent)
			}

			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Errorf("Expected ExitError, got: %v", err)
			}

			stderrStr := stderr.String()
			if !strings.Contains(stderrStr, testCase.wantError) {
				t.Errorf("Expected error text %q, got:\n%s", testCase.wantError, stderrStr)
			}

			// Run additional error checks
			if testCase.errorCheck != nil {
				testCase.errorCheck(t, stderrStr)
			}
		})
	}
}

// TestM05E05NoChildProcessSpawnedOnError validates that no agent process is spawned on error conditions.
// v2.0: environment variable support removed (Phase 7).
//
//nolint:funlen // E2E test with process spawning validation
func TestM05E05NoChildProcessSpawnedOnError(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	tests := []struct {
		name      string
		args      []string
		needsHome bool
	}{
		{
			name:      "multiple agent flags",
			args:      []string{"--claude", "--codex"},
			needsHome: false,
		},
		{
			name:      "agent not in PATH",
			args:      []string{"--codex"},
			needsHome: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var cmdEnv []string
			if testCase.needsHome {
				// v2.0: Create temporary home with config and command files
				tmpHome := t.TempDir()
				setupConfigWithCommands(t, tmpHome, "claude")
				cmdEnv = []string{
					"PATH=/usr/bin:/bin",
					"HOME=" + tmpHome,
				}
			} else {
				cmdEnv = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"))
			}

			binPath := filepath.Join(root, "bin", "fluxid")
			cmd := exec.CommandContext(t.Context(), binPath, testCase.args...)
			cmd.Env = cmdEnv

			var stderr bytes.Buffer
			var stdout bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &stdout

			err := cmd.Run()

			// Error is expected
			if err == nil {
				t.Fatal("Expected error, but succeeded")
			}

			// Verify no agent was invoked - check stdout/stderr for absence of agent output
			combinedOutput := stdout.String() + stderr.String()

			// Agent stub would output "Agent Args:" if invoked
			if strings.Contains(combinedOutput, "Agent Args:") {
				t.Errorf("Agent was invoked despite error condition. Output:\n%s", combinedOutput)
			}

			// Verify no session ID in output (would indicate workflow started)
			if strings.Contains(combinedOutput, "Session ID:") {
				t.Errorf("Workflow started despite error condition. Output:\n%s", combinedOutput)
			}

			// Verify no phase execution (would indicate orchestration started)
			if strings.Contains(combinedOutput, "Starting phase:") {
				t.Errorf("Phase execution started despite error condition. Output:\n%s", combinedOutput)
			}
		})
	}
}

// TestM05E05ErrorMessagesAreConciseAndActionable validates that all error messages include helpful guidance.
// v2.0: environment variable support removed (Phase 7).
//
//nolint:funlen // E2E test with error message validation
func TestM05E05ErrorMessagesAreConciseAndActionable(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	tests := []struct {
		name        string
		args        []string
		needsHome   bool
		wantInError []string
	}{
		{
			name:        "multiple flags error message quality",
			args:        []string{"--claude", "--codex"},
			needsHome:   false,
			wantInError: []string{"multiple", "agent", "--claude", "--codex", "--opencode"},
		},
		{
			name:        "PATH error message quality",
			args:        []string{"--codex"},
			needsHome:   true,
			wantInError: []string{"not found", "PATH", "which", "codex"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var cmdEnv []string
			if testCase.needsHome {
				// v2.0: Create temporary home with config and command files
				tmpHome := t.TempDir()
				setupConfigWithCommands(t, tmpHome, "claude")
				cmdEnv = []string{
					"PATH=/usr/bin:/bin",
					"HOME=" + tmpHome,
				}
			} else {
				cmdEnv = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"))
			}

			binPath := filepath.Join(root, "bin", "fluxid")
			cmd := exec.CommandContext(t.Context(), binPath, testCase.args...)
			cmd.Env = cmdEnv

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Error is expected
			if err == nil {
				t.Fatal("Expected error, but succeeded")
			}

			stderrStr := stderr.String()

			// Verify all expected phrases are present
			for _, want := range testCase.wantInError {
				if !strings.Contains(stderrStr, want) {
					t.Errorf("Expected error to contain %q, got:\n%s", want, stderrStr)
				}
			}

			// Verify error is concise (not excessively long)
			lines := strings.Split(strings.TrimSpace(stderrStr), "\n")
			if len(lines) > 10 {
				t.Errorf("Error message is too verbose (%d lines). Expected concise error. Got:\n%s", len(lines), stderrStr)
			}
		})
	}
}
