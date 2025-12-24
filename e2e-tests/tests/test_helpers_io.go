package tests

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

const (
	promptDelayMillis = 50 // Delay in milliseconds before sending stdin response
)

var errTimeout = errors.New("timeout waiting for output")

// readCombinedOutput reads from stdout and stderr pipes concurrently
// and combines them into a single buffer. Optionally handles stdin interaction
// when a specific prompt is detected.
//
//nolint:cyclop,funlen // Test helper with multiple error paths and concurrent I/O handling
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

	var waitGroup sync.WaitGroup
	waitGroup.Add(minExpectedArgCount)

	// Read stdout
	go func() {
		defer waitGroup.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputMutex.Lock()
			output.WriteString(line + "\n")
			outputMutex.Unlock()

			// Handle interactive prompt if configured
			if promptMarker != "" && strings.Contains(line, promptMarker) && !promptSeen {
				promptSeen = true
				time.Sleep(promptDelayMillis * time.Millisecond)
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
		defer waitGroup.Done()
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
		waitGroup.Wait()
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, io.EOF) {
			return output.String(), fmt.Errorf("error reading output: %w", err)
		}
	case <-time.After(timeout):
		return output.String(), fmt.Errorf("%w after %v", errTimeout, timeout)
	}

	return output.String(), nil
}
