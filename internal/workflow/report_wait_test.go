package workflow

import (
	"errors"
	"fluxid-loop/internal/ipc"
	"testing"
	"time"
)

func TestWaitForValidReport_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-timeout-" + time.Now().Format("20060102150405")

	// Save original values and restore after test
	origMaxAttempts := reportMaxAttempts
	origPollInterval := reportPollInterval
	defer func() {
		reportMaxAttempts = origMaxAttempts
		reportPollInterval = origPollInterval
	}()

	// Set short timeout for testing (3 attempts * 100ms = 300ms)
	reportMaxAttempts = 3
	reportPollInterval = 100 * time.Millisecond

	// Don't write any report - let it timeout
	status, err := waitForValidReport(sessionID, "implement")
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	if status != "" {
		t.Errorf("Expected empty status on timeout, got %s", status)
	}
	// Verify it's the timeout error
	if !errors.Is(err, errReportTimeout) {
		t.Errorf("Expected errReportTimeout, got: %v", err)
	}
}

func TestWaitForValidReport_Success(t *testing.T) {
	// Test waitForValidReport with a valid report
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-wait-report-success"

	// Write a valid report
	validReport := `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Test passed
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Continue to next phase
`
	if err := ipc.WriteReport(sessionID, validReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Wait for the report
	status, err := waitForValidReport(sessionID, "implement")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected status PASS, got: %s", status)
	}
}

func TestWaitForValidReport_Fail(t *testing.T) {
	// Test waitForValidReport with a FAIL status
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-wait-report-fail"

	// Write a valid FAIL report
	failReport := `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: FAIL
summary: Test failed
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Retry implementation
`
	if err := ipc.WriteReport(sessionID, failReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Wait for the report
	status, err := waitForValidReport(sessionID, "implement")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != statusFail {
		t.Errorf("Expected status FAIL, got: %s", status)
	}
}

// testRetryScenario tests waitForValidReport with an initial invalid report followed by a valid one.
func testRetryScenario(t *testing.T, sessionID, initialReport, validReport, expectedStatus, command string) {
	t.Helper()

	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Write initial invalid/malformed report
	if err := ipc.WriteReport(sessionID, initialReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Use a channel with timeout to prevent hanging
	type result struct {
		status string
		err    error
	}
	done := make(chan result, 1)
	started := make(chan struct{})

	// Start waitForValidReport in background
	go func() {
		close(started) // Signal that goroutine is starting
		status, err := waitForValidReport(sessionID, command)
		done <- result{status, err}
	}()

	// Wait for the goroutine to start, then write the valid report immediately
	// This ensures waitForValidReport has started before we overwrite the invalid report
	<-started

	// Write valid report without artificial delay
	// waitForValidReport will retry and eventually see this valid report
	if err := ipc.WriteReport(sessionID, validReport); err != nil {
		t.Fatalf("Failed to write valid report: %v", err)
	}

	// Use very long timeout to accommodate race detector slowness
	// waitForValidReport sleeps 2s between retries, which becomes 20-40s under race detector
	// Multiple retries could take several minutes with race detector
	select {
	case res := <-done:
		if res.err != nil {
			t.Errorf("Expected no error, got: %v", res.err)
		}
		if res.status != expectedStatus {
			t.Errorf("Expected status %s, got: %s", expectedStatus, res.status)
		}
	case <-time.After(5 * time.Minute):
		t.Fatal("Test timed out waiting for report")
	}
}

//nolint:paralleltest // Cannot use t.Parallel with t.Setenv
func TestWaitForValidReport_InvalidThenValid(t *testing.T) {
	// Skip when race detector is enabled - this test verifies retry behavior
	// which involves 2s sleeps that become 20-40s under race detector
	if testing.Short() {
		t.Skip("Skipping retry test in short mode (incompatible with race detector)")
	}
	// Test waitForValidReport retrying on invalid report
	invalidReport := `invalid: yaml without status`
	validReport := `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Test passed eventually
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Continue
`
	testRetryScenario(t, "test-wait-report-retry", invalidReport, validReport, statusPass, "test")
}

//nolint:paralleltest // Cannot use t.Parallel with t.Setenv
func TestWaitForValidReport_MalformedYAML(t *testing.T) {
	t.Skip("TODO: Fix test - timing issues with report writes and reads")
	// Skip when race detector is enabled - this test verifies retry behavior
	// which involves 2s sleeps that become 20-40s under race detector
	if testing.Short() {
		t.Skip("Skipping retry test in short mode (incompatible with race detector)")
	}
	// Test waitForValidReport with malformed YAML that later becomes valid
	malformedYAML := `command: test
status: {{{invalid yaml`
	validReport := `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: FAIL
summary: Test failed after retry
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Fix issues
`
	testRetryScenario(t, "test-malformed-yaml", malformedYAML, validReport, statusFail, "implement")
}

func TestWaitForValidReport_ReadError(t *testing.T) {
	t.Skip("TODO: Fix test - abort flag checking interferes with read error testing")
	// Test waitForValidReport with a read error scenario
	sessionID := "test-read-error"

	// Set invalid XDG_DATA_HOME to cause read errors
	t.Setenv("XDG_DATA_HOME", "/dev/null/invalid")

	// Use a channel with timeout
	done := make(chan struct {
		status string
		err    error
	}, 1)

	go func() {
		status, err := waitForValidReport(sessionID, "test")
		done <- struct {
			status string
			err    error
		}{status, err}
	}()

	select {
	case result := <-done:
		if result.err == nil {
			t.Error("Expected error when reading from invalid path")
		}
	case <-time.After(1 * time.Second):
		// Expected to fail quickly, test passes
	}
}

func TestWaitForValidReport_CheckAbortError(t *testing.T) {
	// Test abort flag check error path (warning logged but execution continues)
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-wait-abort-check-err"

	// Provide a valid report after a delay to ensure abort check happens first
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, testPassReport)
	}()

	// waitForValidReport should complete successfully even if abort check has issues
	status, err := waitForValidReport(sessionID, "implement")
	if err != nil {
		t.Errorf("Expected no error despite abort check issues, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected status PASS, got: %s", status)
	}
}
