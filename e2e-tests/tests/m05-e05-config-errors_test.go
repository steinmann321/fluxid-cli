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

// TestM05E05MultipleAgentFlagsError validates clear error for conflicting CLI flags.
//
//nolint:funlen // E2E test with multiple error scenarios
func TestM05E05MultipleAgentFlagsError(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	tests := []struct {
		name        string
		args        []string
		wantError   string
		wantExitErr bool
	}{
		{
			name:        "two agent flags",
			args:        []string{"--claude", "--codex"},
			wantError:   "multiple agent flags specified",
			wantExitErr: true,
		},
		{
			name:        "three agent flags",
			args:        []string{"--claude", "--codex", "--opencode"},
			wantError:   "multiple agent flags specified",
			wantExitErr: true,
		},
		{
			name:        "claude and opencode",
			args:        []string{"--claude", "--opencode"},
			wantError:   "multiple agent flags specified",
			wantExitErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			binPath := filepath.Join(root, "bin", "fluxid")
			cmd := exec.CommandContext(t.Context(), binPath, testCase.args...)
			cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			if !testCase.wantExitErr {
				if err != nil {
					t.Fatalf("Expected success with args %v, got error: %v\nStderr: %s", testCase.args, err, stderr.String())
				}
				return
			}

			// Error is expected from here on
			if err == nil {
				t.Fatalf("Expected error with args %v, but succeeded", testCase.args)
			}

			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Errorf("Expected ExitError, got: %v", err)
			}

			stderrStr := stderr.String()
			if !strings.Contains(stderrStr, testCase.wantError) {
				t.Errorf("Expected error text %q, got:\n%s", testCase.wantError, stderrStr)
			}

			// Verify helpful guidance is present
			hasAllFlags := strings.Contains(stderrStr, "--claude") &&
				strings.Contains(stderrStr, "--codex") &&
				strings.Contains(stderrStr, "--opencode")
			if !hasAllFlags {
				t.Errorf("Expected error to mention available agent flags, got:\n%s", stderrStr)
			}
		})
	}
}

func testEnvVarAgent(t *testing.T, root string) {
	t.Helper()

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)
	cmd.Env = append(
		os.Environ(),
		"FLUXID_AGENT=foo",
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err == nil {
		t.Fatal("Expected error for unsupported agent, but succeeded")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "unsupported agent") {
		t.Errorf("Expected error about unsupported agent, got:\n%s", stderrStr)
	}
	if !strings.Contains(stderrStr, "foo") {
		t.Errorf("Expected error to mention invalid agent 'foo', got:\n%s", stderrStr)
	}
	hasAllAgents := strings.Contains(stderrStr, "claude") &&
		strings.Contains(stderrStr, "codex") &&
		strings.Contains(stderrStr, "opencode")
	if !hasAllAgents {
		t.Errorf("Expected error to list supported agents, got:\n%s", stderrStr)
	}
}

func testHomeConfigAgent(t *testing.T, root string) {
	t.Helper()

	tmpHome := t.TempDir()
	fluxidDir := filepath.Join(tmpHome, ".fluxid")
	// #nosec G301 -- Test fixture directory with standard permissions
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create .fluxid directory: %v", err)
	}
	configPath := filepath.Join(fluxidDir, "config.yaml")
	content := "agent: invalid-agent\n"
	// #nosec G306 -- Test fixture file with standard permissions
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)
	cmd.Env = append(
		os.Environ(),
		"HOME="+tmpHome,
		"PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err == nil {
		t.Fatal("Expected error for unsupported agent, but succeeded")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "unsupported agent") {
		t.Errorf("Expected error about unsupported agent, got:\n%s", stderrStr)
	}
	if !strings.Contains(stderrStr, "invalid-agent") {
		t.Errorf("Expected error to mention invalid agent 'invalid-agent', got:\n%s", stderrStr)
	}
}

func testProjectConfigAgent(t *testing.T, root string) {
	t.Helper()

	tmpProject := t.TempDir()
	fluxidDir := filepath.Join(tmpProject, ".fluxid")
	// #nosec G301 -- Test fixture directory with standard permissions
	if err := os.MkdirAll(fluxidDir, 0o755); err != nil {
		t.Fatalf("Failed to create .fluxid directory: %v", err)
	}
	configPath := filepath.Join(fluxidDir, "config.yaml")
	content := "agent: bar\n"
	// #nosec G306 -- Test fixture file with standard permissions
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath)
	cmd.Dir = tmpProject
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err == nil {
		t.Fatal("Expected error for unsupported agent, but succeeded")
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "unsupported agent") {
		t.Errorf("Expected error about unsupported agent, got:\n%s", stderrStr)
	}
	if !strings.Contains(stderrStr, "bar") {
		t.Errorf("Expected error to mention invalid agent 'bar', got:\n%s", stderrStr)
	}
}

// TestM05E05UnsupportedAgentError validates clear error for unsupported agent values.
func TestM05E05UnsupportedAgentError(t *testing.T) {
	t.Parallel()

	root := getProjectRoot(t)
	buildFluxid(t, root)

	t.Run("unsupported agent via environment variable", func(t *testing.T) {
		t.Parallel()
		testEnvVarAgent(t, root)
	})

	t.Run("unsupported agent via home config", func(t *testing.T) {
		t.Parallel()
		testHomeConfigAgent(t, root)
	})

	t.Run("unsupported agent via project config", func(t *testing.T) {
		t.Parallel()
		testProjectConfigAgent(t, root)
	})
}
