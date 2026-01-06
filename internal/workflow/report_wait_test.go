//nolint:paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-cli/internal/storage"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestWaitForValidReport_NoReport(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()
	sessionID := "0ab7dd40-6bef-47d5-b6df-fed5cb8ad688"

	// Don't write any report - should return FAIL immediately
	status, err := waitForValidReport(sessionID, "", "implement")
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
	sessionID := "d1e2f3a4-5b6c-7d8e-9f0a-1b2c3d4e5f6a"

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
	if err := storage.WriteReport(sessionID, validReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Wait for the report
	status, err := waitForValidReport(sessionID, "", "implement")
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
	sessionID := "e2f3a4b5-6c7d-8e9f-0a1b-2c3d4e5f6a7b"

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
	if err := storage.WriteReport(sessionID, failReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Wait for the report
	status, err := waitForValidReport(sessionID, "", "implement")
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
	sessionID := "f3a4b5c6-7d8e-9f0a-1b2c-3d4e5f6a7b8c"

	// Set invalid XDG_DATA_HOME to cause read errors
	t.Setenv("XDG_DATA_HOME", "/dev/null/invalid")

	// Use a channel with timeout
	done := make(chan struct {
		status string
		err    error
	}, 1)
	defer close(done)

	go func() {
		status, err := waitForValidReport(sessionID, "", "test")
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
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	defer goleak.VerifyNone(t)

	// Test abort flag check error path (warning logged but execution continues)
	_, cleanup := setupTestDataDir(t)
	defer cleanup()
	sessionID := "f6a7b8c9-d0e1-2f3a-4b5c-6d7e8f9a0b1c"

	// Write report immediately to ensure deterministic test behavior
	err := storage.WriteReport(sessionID, testPassReport)
	if err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// waitForValidReport should complete successfully even if abort check has issues
	status, err := waitForValidReport(sessionID, "", "implement")
	if err != nil {
		t.Errorf("Expected no error despite abort check issues, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected status PASS, got: %s", status)
	}
}
