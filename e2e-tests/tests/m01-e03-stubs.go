package tests

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// writeExecutableStub writes a bash script with executable permissions.
// This is a test helper that requires 0755 permissions for shell scripts.
//
// #nosec G306 -- Test stubs require executable permissions
//
//nolint:wrapcheck // Test helper, error context clear from usage
func writeExecutableStub(path string, content []byte) error {
	return os.WriteFile(path, content, 0o755)
}

// createStreamingStubClaude creates a stub that generates streaming output over time.
func createStreamingStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Streaming output stub - generates output over time

echo "Claude stub: Starting streaming output test"

# Generate burst output with small delays
for i in {1..10}; do
  echo "BURST_LINE $i: $(date +%s%N)"
  sleep 0.05
done

echo "FLUXID_SESSION_ID=$FLUXID_SESSION_ID"
exit 0
`

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

// createInteractiveStubClaude creates a stub that prompts for input and echoes it back.
func createInteractiveStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Interactive stub - prompts for input and echoes it back

echo "Claude stub: Interactive test"

# Only prompt during implement phase
if echo "$@" | grep -q "Implement the required"; then
  echo "PROMPT: Enter your name:"
  read -r response
  echo "RECEIVED: $response"
fi

exit 0
`

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

// createLargeOutputStubClaude creates a stub that generates many lines to test buffer handling.
func createLargeOutputStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Large output stub - generates many lines to test buffer handling

echo "Claude stub: Large output test"

# Generate 1500 lines of output
for i in {1..1500}; do
  echo "LARGE_OUTPUT_LINE $i: Lorem ipsum dolor sit amet, consectetur adipiscing elit."
done

exit 0
`

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

// createMixedStreamStubClaude creates a stub that outputs to both stdout and stderr.
func createMixedStreamStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Mixed stream stub - outputs to both stdout and stderr

echo "Claude stub: Mixed stream test"

for i in {1..5}; do
  echo "STDOUT: MSG_$i message on stdout"
  echo "STDERR: MSG_$i message on stderr" >&2
  sleep 0.02
done

exit 0
`

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

// createWorkflowContinuationStubClaude creates a stub that is interactive during
// implement phase but silent for other phases.
func createWorkflowContinuationStubClaude(t *testing.T, root string) {
	t.Helper()

	stubPath := filepath.Join(root, "bin", "claude")
	stubScript := `#!/bin/bash
# Workflow continuation stub - interactive during implement, silent for others

# Check which phase we're in
if echo "$@" | grep -q "Implement the required"; then
  echo "IMPLEMENT_PROMPT: Ready to implement? (type anything to continue)"
  read -r response
  echo "IMPLEMENT_RESPONSE: Got '$response', continuing..."
elif echo "$@" | grep -q "Create a git commit"; then
  echo "Commit phase executing..."
elif echo "$@" | grep -q "Review the implementation"; then
  echo "Review phase executing..."
fi

exit 0
`

	_ = writeExecutableStub(stubPath, []byte(stubScript)) // Ignore error - test will fail if stub missing
}

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
			output.WriteString(line + "\n")

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
			output.WriteString(line + "\n")
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
