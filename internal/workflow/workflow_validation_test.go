//nolint:paralleltest // Tests use global mutex, cannot run in parallel
//nolint:paralleltest // Workflow tests with subprocess execution
package workflow

import (
	"fluxid-cli/internal/ipc"
	"fmt"
	"testing"
	"time"
)

func TestWaitForValidReport_WithValidReport(t *testing.T) {
	// Test waitForValidReport when a valid report already exists
	sessionID := "test-wait-valid-report-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

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
	sessionID := "test-wait-fail-report-" + fmt.Sprintf("%d", time.Now().UnixNano()) //nolint:perfsprint

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

	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	if err := ipc.WriteReport(sessionID, reportYAML); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	resultChan := make(chan string, 1)
	defer close(resultChan)
	errorChan := make(chan error, 1)
	defer close(errorChan)

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
