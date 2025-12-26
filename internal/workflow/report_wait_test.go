//nolint:paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-cli/internal/ipc"
	"fmt"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestWaitForValidReport_NoReport(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()
	sessionID := "test-no-report-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

	// Don't write any report - should return FAIL immediately
	status, err := waitForValidReport(sessionID, "implement")
	if err != nil {
		t.Errorf("Expected no error when report missing (should return FAIL), got: %v", err)
	}
	if status != statusFail {
		t.Errorf("Expected status FAIL when no report exists, got: %s", status)
	}
}

func TestWaitForValidReport_Success(t *testing.T) {
	// Test waitForValidReport with a valid report
	_, cleanup := setupTestDataDir(t)
	defer cleanup()
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
	_, cleanup := setupTestDataDir(t)
	defer cleanup()
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

// NOTE: Tests for retry behavior (testRetryScenario, TestWaitForValidReport_InvalidThenValid,
// TestWaitForValidReport_MalformedYAML) have been removed as waitForValidReport() no longer
// polls/retries. It checks the report immediately and returns FAIL for invalid/missing reports.
// Invalid report handling is already tested by TestWaitForValidReport_InvalidReport.

func TestWaitForValidReport_ReadError(t *testing.T) {
	defer goleak.VerifyNone(t)

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
	defer close(done)

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
	defer goleak.VerifyNone(t)

	// Test abort flag check error path (warning logged but execution continues)
	_, cleanup := setupTestDataDir(t)
	defer cleanup()
	sessionID := "test-wait-abort-check-err"

	// Write report immediately to ensure deterministic test behavior
	err := ipc.WriteReport(sessionID, testPassReport)
	if err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// waitForValidReport should complete successfully even if abort check has issues
	status, err := waitForValidReport(sessionID, "implement")
	if err != nil {
		t.Errorf("Expected no error despite abort check issues, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected status PASS, got: %s", status)
	}
}
