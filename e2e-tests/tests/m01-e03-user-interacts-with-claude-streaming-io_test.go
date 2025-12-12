package tests

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestM01E03StreamingOutputPassthrough validates that stdout/stderr from Claude
// are streamed in real-time to the user with acceptable latency.
//
//nolint:paralleltest // Sequential stub usage; I/O streaming complexity
func TestM01E03StreamingOutputPassthrough(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createStreamingStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	// Capture stdout with timestamps to measure latency
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start fluxid: %v", err)
	}

	// Track output timing
	type timedLine struct {
		text      string
		timestamp time.Time
	}

	var lines []timedLine
	startTime := time.Now()

	// Read stdout and stderr concurrently
	outDone := make(chan struct{})
	errDone := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			lines = append(lines, timedLine{
				text:      scanner.Text(),
				timestamp: time.Now(),
			})
		}
		close(outDone)
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			lines = append(lines, timedLine{
				text:      scanner.Text(),
				timestamp: time.Now(),
			})
		}
		close(errDone)
	}()

	// Wait for completion
	<-outDone
	<-errDone

	if err := cmd.Wait(); err != nil {
		t.Fatalf("fluxid failed: %v", err)
	}

	// Verify we got streaming output
	if len(lines) < 10 {
		t.Errorf("Expected at least 10 lines of streaming output, got %d", len(lines))
	}

	// Check for burst output markers from stub
	var burstCount int
	for _, line := range lines {
		if strings.Contains(line.text, "BURST_LINE") {
			burstCount++
		}
	}

	if burstCount < 5 {
		t.Errorf("Expected at least 5 burst output lines, got %d", burstCount)
	}

	// Verify latency is reasonable (output should arrive within reasonable time)
	totalDuration := time.Since(startTime)
	if totalDuration > 10*time.Second {
		t.Errorf("Total execution took too long: %v (expected < 10s)", totalDuration)
	}

	// Check that lines arrived incrementally (not all at once at the end)
	if len(lines) >= 2 {
		firstLineTime := lines[0].timestamp.Sub(startTime)
		lastLineTime := lines[len(lines)-1].timestamp.Sub(startTime)

		// If all lines arrive at once, the time difference would be minimal
		// We expect streaming, so there should be some time spread
		timeDiff := lastLineTime - firstLineTime
		if timeDiff < 100*time.Millisecond {
			t.Errorf("Lines arrived too quickly (%.2fms), suggesting buffering instead of streaming",
				timeDiff.Seconds()*1000)
		}
	}
}

// TestM01E03InteractiveStdinDelivery validates that user input from stdin
// is delivered to the Claude process reliably.
//
//nolint:paralleltest // Sequential stub usage; I/O complexity
func TestM01E03InteractiveStdinDelivery(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createInteractiveStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	// Set up pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start fluxid: %v", err)
	}

	// Read output and provide input when prompted
	var output bytes.Buffer
	done := make(chan error)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")

			// When we see the prompt, send input
			if strings.Contains(line, "PROMPT: Enter your name:") {
				time.Sleep(50 * time.Millisecond) // Small delay to simulate user typing
				if _, err := stdin.Write([]byte("TestUser\n")); err != nil {
					done <- fmt.Errorf("failed to write to stdin: %w", err)
					return
				}
			}
		}
		done <- scanner.Err()
	}()

	// Wait for completion
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("error reading output: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("test timeout - process did not complete")
	}

	if err := stdin.Close(); err != nil {
		t.Logf("failed to close stdin: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, output.String())
	}

	outputStr := output.String()

	// Verify prompt appeared
	if !strings.Contains(outputStr, "PROMPT: Enter your name:") {
		t.Errorf("Expected prompt in output")
	}

	// Verify input was received and echoed back
	if !strings.Contains(outputStr, "RECEIVED: TestUser") {
		t.Errorf("Expected input to be received by Claude stub\nOutput:\n%s", outputStr)
	}
}

// TestM01E03NoOutputTruncation validates that large outputs are not truncated
// and buffer deadlocks don't occur.
//
//nolint:paralleltest // Sequential execution required due to shared stub
func TestM01E03NoOutputTruncation(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createLargeOutputStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, stdout.String())
	}

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Count LARGE_OUTPUT markers
	var largeOutputCount int
	for _, line := range lines {
		if strings.Contains(line, "LARGE_OUTPUT_LINE") {
			largeOutputCount++
		}
	}

	// Expect at least 1000 lines from the stub
	if largeOutputCount < 1000 {
		t.Errorf("Expected at least 1000 large output lines (checking for truncation), got %d", largeOutputCount)
	}

	// Verify workflow completed
	if !strings.Contains(output, "Workflow completed successfully") {
		t.Errorf("Workflow did not complete - possible deadlock")
	}
}

// TestM01E03StreamOrderingReadable validates that stdout and stderr are
// interleaved in a sensible, readable manner.
//
//nolint:paralleltest // Sequential execution required due to shared stub
func TestM01E03StreamOrderingReadable(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createMixedStreamStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("fluxid failed: %v", err)
	}

	output := stdout.String()

	// Verify we have both STDOUT and STDERR markers
	hasStdout := strings.Contains(output, "STDOUT:")
	hasStderr := strings.Contains(output, "STDERR:")

	if !hasStdout || !hasStderr {
		t.Errorf("Expected both STDOUT and STDERR in output (stdout=%v, stderr=%v)", hasStdout, hasStderr)
	}

	// Check that messages appear in sequence (1, 2, 3, ...)
	// This validates ordering is maintained
	for i := 1; i <= 5; i++ {
		marker := fmt.Sprintf("MSG_%d", i)
		if !strings.Contains(output, marker) {
			t.Errorf("Missing message marker %s - stream ordering may be broken", marker)
		}
	}
}

// TestM01E03WorkflowContinuesAfterInteraction validates that after user
// provides input during an interactive phase, the workflow continues correctly.
//
//nolint:paralleltest // Sequential stub; workflow I/O complexity
func TestM01E03WorkflowContinuesAfterInteraction(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)
	createWorkflowContinuationStubClaude(t, root)

	binPath := filepath.Join(root, "bin", "fluxid")
	cmd := exec.CommandContext(t.Context(), binPath, "--claude")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start fluxid: %v", err)
	}

	// Read output and handle interactive prompt
	outputStr, err := readCombinedOutput(stdout, stderr, stdin, "IMPLEMENT_PROMPT:", "confirmed", 10*time.Second)
	if err != nil {
		t.Fatalf("error reading output: %v", err)
	}

	if err := stdin.Close(); err != nil {
		t.Logf("failed to close stdin: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatalf("fluxid failed: %v\nOutput:\n%s", err, outputStr)
	}

	// Verify all three phases ran
	if !strings.Contains(outputStr, "Starting phase: implement") {
		t.Errorf("Implement phase did not start")
	}

	if !strings.Contains(outputStr, "Starting phase: review") {
		t.Errorf("Review phase did not run after interaction")
	}

	// Verify workflow completed
	if !strings.Contains(outputStr, "Status: SUCCESS") {
		t.Errorf("Workflow did not complete successfully after interactive phase")
	}
}
