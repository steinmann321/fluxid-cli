//nolint:paralleltest // E2E tests with subprocess execution
package tests

import (
	"errors"
	"fluxid-loop/internal/ipc"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestM03E05GracefulAbortViaSignal tests that a single SIGINT triggers graceful abort.
// NOTE: This test verifies that signal handler sets abort flag. The actual graceful
// exit happens via abort check in the workflow loop (tested in other tests).
func TestM03E05GracefulAbortViaSignal(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createLongRunningStub(t, root, 30)

	sessionID := "test-abort-signal-" + time.Now().Format("20060102150405")
	binPath := filepath.Join(root, "bin", "fluxid")

	ctx, cancel := testContext(5 * time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "--claude", "--fluxid-iterations", "3", "--fluxid-implement-retries", "3")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"FLUXID_SESSION_ID="+sessionID,
	)

	// Capture output
	var outputBuf strings.Builder
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start fluxid: %v", err)
	}

	// Wait for workflow to start first phase
	<-time.After(300 * time.Millisecond)

	// Send SIGINT to trigger graceful abort
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	// Give signal handler time to set abort flag
	<-time.After(100 * time.Millisecond)

	// Verify abort flag was set by signal handler
	aborted, err := ipc.CheckAbortFlag(sessionID)
	if err != nil {
		t.Fatalf("Failed to check abort flag: %v", err)
	}
	if !aborted {
		t.Error("Expected abort flag to be set by signal handler")
	}

	// Wait for command to complete (may be killed by signal or exit gracefully)
	_ = cmd.Wait()

	output := outputBuf.String()

	// Verify signal was received and message displayed
	if !strings.Contains(output, "Received signal") {
		t.Errorf("Expected signal received message in output:\n%s", output)
	}

	// Cleanup
	_ = ipc.ClearAbortFlag(sessionID)
}

// TestM03E05ForcedExitOnSecondSignal tests that two rapid SIGINTs force immediate exit.
func TestM03E05ForcedExitOnSecondSignal(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createLongRunningStub(t, root, 30)

	sessionID := "test-forced-exit-" + time.Now().Format("20060102150405")
	binPath := filepath.Join(root, "bin", "fluxid")

	ctx, cancel := testContext(10 * time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "--claude", "--fluxid-iterations", "1")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"FLUXID_SESSION_ID="+sessionID,
	)

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start fluxid: %v", err)
	}

	// Wait for workflow to start
	<-time.After(500 * time.Millisecond)

	startTime := time.Now()

	// Send first SIGINT
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send first SIGINT: %v", err)
	}

	// Immediately send second SIGINT
	<-time.After(100 * time.Millisecond)
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send second SIGINT: %v", err)
	}

	// Wait for command to complete
	err := cmd.Wait()
	elapsed := time.Since(startTime)

	// Verify immediate exit (should exit quickly, not wait for phase completion)
	if elapsed > 5*time.Second {
		t.Errorf("Expected immediate exit, took %v", elapsed)
	}

	// Verify exit code 130
	if err == nil {
		t.Fatal("Expected fluxid to exit with error")
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Expected ExitError, got: %v", err)
	}

	if exitErr.ExitCode() != 130 {
		t.Errorf("Expected exit code 130, got: %d", exitErr.ExitCode())
	}

	// Cleanup
	_ = ipc.ClearAbortFlag(sessionID)
}

// TestM03E05AbortViaIPCCommand tests abort using 'fluxid ipc abort' command.
func TestM03E05AbortViaIPCCommand(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createLongRunningStub(t, root, 30)

	sessionID := "test-ipc-abort-" + time.Now().Format("20060102150405")
	binPath := filepath.Join(root, "bin", "fluxid")

	ctx, cancel := testContext(5 * time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "--claude", "--fluxid-iterations", "3", "--fluxid-implement-retries", "3")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"FLUXID_SESSION_ID="+sessionID,
	)

	// Capture output
	var outputBuf strings.Builder
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	// Start the workflow command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start fluxid: %v", err)
	}

	// Wait for workflow to start
	<-time.After(300 * time.Millisecond)

	// Issue abort via IPC command
	abortCmd := exec.CommandContext(ctx, binPath, "ipc", "abort", "--session", sessionID)
	abortOutput, err := abortCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run ipc abort: %v\nOutput: %s", err, abortOutput)
	}

	// Verify abort confirmation message
	outputStr := string(abortOutput)
	if !strings.Contains(outputStr, "Abort requested") {
		t.Errorf("Expected abort confirmation message, got: %s", outputStr)
	}

	// Wait for workflow to exit
	err = cmd.Wait()

	// Verify graceful abort
	if err == nil {
		t.Fatal("Expected fluxid to exit with error after abort")
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Expected ExitError, got: %v\nOutput:\n%s", err, outputBuf.String())
	}

	if exitErr.ExitCode() != 130 {
		t.Errorf("Expected exit code 130 (graceful abort), got: %d\nOutput:\n%s", exitErr.ExitCode(), outputBuf.String())
	}

	// Cleanup
	_ = ipc.ClearAbortFlag(sessionID)
}

// TestM03E05AbortMessageContent verifies abort messages are clear and helpful.
func TestM03E05AbortMessageContent(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createLongRunningStub(t, root, 30)

	sessionID := "test-abort-messages-" + time.Now().Format("20060102150405")
	binPath := filepath.Join(root, "bin", "fluxid")

	ctx, cancel := testContext(5 * time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "--claude", "--fluxid-iterations", "3", "--fluxid-implement-retries", "3")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"FLUXID_SESSION_ID="+sessionID,
	)

	// Capture output
	var outputBuf strings.Builder
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start fluxid: %v", err)
	}

	// Wait for workflow to start
	<-time.After(300 * time.Millisecond)

	// Send SIGINT
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	// Wait for completion
	_ = cmd.Wait()

	output := outputBuf.String()

	// Verify abort-related messages
	expectedMessages := []string{
		"graceful abort",
		"Workflow Aborted",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected output to contain '%s', got:\n%s", msg, output)
		}
	}

	// Cleanup
	_ = ipc.ClearAbortFlag(sessionID)
}
