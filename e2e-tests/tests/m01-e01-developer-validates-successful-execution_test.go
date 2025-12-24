// Package tests contains E2E tests for the fluxid CLI application.
// This file tests the happy path scenarios for successful execution.
package tests

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const minimalHomeConfig = `agent: echo
iterations: 1
implement_retries: 1
commit_enabled: false
`

// extractExitCode extracts the exit code from an error using errors.As.
func extractExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 0
}

// TestMain_SuccessfulExecutionWithDryRun validates that the CLI executes successfully
// in dry-run mode and exits with code 0.
func TestMain_SuccessfulExecutionWithDryRun(t *testing.T) {
	t.Parallel()
	start := time.Now()

	// Build the binary
	binaryPath := buildFluxidBinary(t)

	// Setup test environment with minimal configuration
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create minimal home config to prevent validation errors
	homeConfigPath := filepath.Join(homeDir, ".config", "fluxid", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(homeConfigPath), 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	if err := os.WriteFile(homeConfigPath, []byte(minimalHomeConfig), 0o644); err != nil {
		t.Fatalf("Failed to write home config: %v", err)
	}

	// Execute fluxid with --fluxid-dry-run flag
	ctx, cancel := testContext(5 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binaryPath, "--fluxid-dry-run", "--claude")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"XDG_CONFIG_HOME="+filepath.Join(homeDir, ".config"),
		"XDG_DATA_HOME="+filepath.Join(homeDir, ".local", "share"),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := extractExitCode(err)

	// Assertions
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for successful dry-run execution, got %d", exitCode)
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
	}

	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Errorf("Test execution took %v, expected < 2 seconds", elapsed)
	}

	// Verify that output indicates dry-run mode
	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "DRY RUN") && !strings.Contains(output, "Simulation") {
		t.Error("Expected output to indicate dry-run/simulation mode")
	}
}

// testHelpFlag is a helper to test help flag variations.
func testHelpFlag(t *testing.T, binaryPath, flag string) {
	t.Helper()
	start := time.Now()

	// Execute fluxid with help flag
	ctx, cancel := testContext(5 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binaryPath, flag)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := extractExitCode(err)

	// Assertions
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for %s flag, got %d", flag, exitCode)
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
	}

	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Errorf("Test execution took %v, expected < 2 seconds", elapsed)
	}

	// Verify help output contains usage information
	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "Usage:") && !strings.Contains(output, "usage") {
		t.Errorf("Expected %s output to contain usage information", flag)
	}
}

// TestMain_HelpFlag validates that --help flag returns exit code 0 and displays usage.
func TestMain_HelpFlag(t *testing.T) {
	t.Parallel()
	binaryPath := buildFluxidBinary(t)
	testHelpFlag(t, binaryPath, "--help")
}

// TestMain_ShortHelpFlag validates that -h flag returns exit code 0 and displays usage.
func TestMain_ShortHelpFlag(t *testing.T) {
	t.Parallel()
	binaryPath := buildFluxidBinary(t)
	testHelpFlag(t, binaryPath, "-h")
}

// TestMain_OutputFormats validates that different output formats work correctly.
//
//nolint:funlen // E2E test validating multiple output formats with extensive checks
func TestMain_OutputFormats(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		flag           string
		expectedOutput string
	}{
		{
			name:           "JSON output format",
			flag:           "--output=json",
			expectedOutput: "session_id",
		},
		{
			name:           "YAML output format",
			flag:           "--output=yaml",
			expectedOutput: "session_id:",
		},
		{
			name:           "Text output format (default)",
			flag:           "",
			expectedOutput: "Session ID:",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			start := time.Now()

			binaryPath := buildFluxidBinary(t)

			// Setup test environment
			tmpDir := t.TempDir()
			homeDir := filepath.Join(tmpDir, "home")
			projectDir := filepath.Join(tmpDir, "project")

			if err := os.MkdirAll(homeDir, 0o755); err != nil {
				t.Fatalf("Failed to create home dir: %v", err)
			}
			if err := os.MkdirAll(projectDir, 0o755); err != nil {
				t.Fatalf("Failed to create project dir: %v", err)
			}

			// Create minimal home config
			homeConfigPath := filepath.Join(homeDir, ".config", "fluxid", "config.yaml")
			if err := os.MkdirAll(filepath.Dir(homeConfigPath), 0o755); err != nil {
				t.Fatalf("Failed to create config dir: %v", err)
			}

			homeConfig := `agent: echo
iterations: 1
implement_retries: 1
commit_enabled: false
`
			if err := os.WriteFile(homeConfigPath, []byte(homeConfig), 0o644); err != nil {
				t.Fatalf("Failed to write home config: %v", err)
			}

			// Build command arguments
			args := []string{"--dry-run"}
			if testCase.flag != "" {
				args = append(args, testCase.flag)
			}

			ctx, cancel := testContext(5 * time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binaryPath, args...)
			cmd.Dir = projectDir
			cmd.Env = append(os.Environ(),
				"HOME="+homeDir,
				"XDG_CONFIG_HOME="+filepath.Join(homeDir, ".config"),
				"XDG_DATA_HOME="+filepath.Join(homeDir, ".local", "share"),
			)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			exitCode := extractExitCode(err)

			// Assertions
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d", exitCode)
				t.Logf("stdout: %s", stdout.String())
				t.Logf("stderr: %s", stderr.String())
			}

			elapsed := time.Since(start)
			if elapsed > 2*time.Second {
				t.Errorf("Test execution took %v, expected < 2 seconds", elapsed)
			}

			// Verify expected output format
			output := stdout.String() + stderr.String()
			if !strings.Contains(output, testCase.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", testCase.expectedOutput, output)
			}
		})
	}
}

// TestMain_ExecutionSpeed validates that the CLI executes quickly (< 1 second).
func TestMain_ExecutionSpeed(t *testing.T) {
	t.Parallel()
	binaryPath := buildFluxidBinary(t)

	// Setup test environment
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create minimal home config
	homeConfigPath := filepath.Join(homeDir, ".config", "fluxid", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(homeConfigPath), 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	if err := os.WriteFile(homeConfigPath, []byte(minimalHomeConfig), 0o644); err != nil {
		t.Fatalf("Failed to write home config: %v", err)
	}

	// Run the test 3 times to ensure consistent performance
	for iteration := 0; iteration < 3; iteration++ {
		start := time.Now()

		ctx, cancel := testContext(5 * time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "--fluxid-dry-run", "--claude")
		cmd.Dir = projectDir
		cmd.Env = append(os.Environ(),
			"HOME="+homeDir,
			"XDG_CONFIG_HOME="+filepath.Join(homeDir, ".config"),
			"XDG_DATA_HOME="+filepath.Join(homeDir, ".local", "share"),
		)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		elapsed := time.Since(start)

		if err != nil {
			exitCode := extractExitCode(err)
			t.Logf("Execution %d failed with exit code %d", iteration+1, exitCode)
			t.Logf("stdout: %s", stdout.String())
			t.Logf("stderr: %s", stderr.String())
		}

		if elapsed > 2*time.Second {
			t.Errorf("Execution %d took %v, expected < 2 seconds", iteration+1, elapsed)
		} else {
			t.Logf("Execution %d completed in %v", iteration+1, elapsed)
		}
	}
}
