//nolint:paralleltest // Workflow tests with subprocess execution
package main

import (
	"fluxid-loop/internal/ipc"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWaitForValidReport_WithValidReport(t *testing.T) {
	// Test waitForValidReport when a valid report already exists
	sessionID := "test-wait-valid-report-" + time.Now().Format("20060102150405")

	// Write a valid PASS report before calling waitForValidReport
	validReport := `command: "test"
artifact: "test-artifact"
timestamp: "2024-01-01T00:00:00Z"
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
summary: "All tests passed"
`

	testWaitForValidReportHelper(t, sessionID, validReport, statusPass)
}

func TestWaitForValidReport_WithFailReport(t *testing.T) {
	// Test waitForValidReport when report has FAIL status
	sessionID := "test-wait-fail-report-" + time.Now().Format("20060102150405")

	// Write a valid FAIL report before calling waitForValidReport
	failReport := `command: "test"
artifact: "test-artifact"
timestamp: "2024-01-01T00:00:00Z"
status: FAIL
issues:
  blockers:
    - message: "Test blocker"
  defects: []
  concerns: []
  observations: []
  enhancements: []
summary: "Tests failed"
`

	testWaitForValidReportHelper(t, sessionID, failReport, statusFail)
}

func testWaitForValidReportHelper(t *testing.T, sessionID, reportYAML, expectedStatus string) {
	t.Helper()

	if err := ipc.WriteReport(sessionID, reportYAML); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}
	defer func() {
		// Clean up
		reportPath := filepath.Join(os.TempDir(), "fluxid-reports", sessionID+".yaml")
		_ = os.Remove(reportPath)
	}()

	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		status, err := waitForValidReport(sessionID, "implement")
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- status
		}
	}()

	// Wait for result with timeout
	select {
	case status := <-resultChan:
		if status != expectedStatus {
			t.Errorf("Expected %s status, got: %s", expectedStatus, status)
		}
	case err := <-errorChan:
		t.Fatalf("waitForValidReport returned error: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("waitForValidReport timed out")
	}
}
