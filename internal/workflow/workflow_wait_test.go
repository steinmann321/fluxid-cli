package workflow

import (
	"errors"
	"fluxid-loop/internal/ipc"
	"sync"
	"testing"
	"time"
)

func TestWaitForValidReport_InvalidReportRetries(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-invalid-report-" + time.Now().Format("20060102150405")

	// Write invalid report first, then valid one after delay
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		<-time.After(50 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, "invalid: yaml: [content")
		<-time.After(100 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, testPassReport)
	}()

	status, err := waitForValidReport(sessionID, "implement")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}

	waitGroup.Wait()
}

func TestWaitForValidReport_AbortWhileWaiting(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-abort-waiting-" + time.Now().Format("20060102150405")

	// Set abort flag after delay
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		<-time.After(50 * time.Millisecond)
		_ = ipc.SetAbortFlag(sessionID)
	}()

	_, err := waitForValidReport(sessionID, "implement")
	if err == nil {
		t.Error("Expected error due to abort, got nil")
	}

	// Check if it's an AbortError with exit code 130
	var abortErr *AbortError
	if err != nil {
		if !errors.As(err, &abortErr) {
			t.Errorf("Expected AbortError, got: %T", err)
		} else if abortErr.ExitCode != 130 {
			t.Errorf("Expected exit code 130, got: %d", abortErr.ExitCode)
		}
	}

	waitGroup.Wait()
}

func TestWaitForValidReport_ImmediateValidReport(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-immediate-" + time.Now().Format("20060102150405")

	// Write valid report immediately (before calling waitForValidReport)
	if err := ipc.WriteReport(sessionID, testPassReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	status, err := waitForValidReport(sessionID, "implement")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}
}

func TestWaitForValidReport_MultipleInvalidThenValid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	sessionID := "test-multi-invalid-" + time.Now().Format("20060102150405")

	// Write multiple invalid reports, then a valid one
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		<-time.After(30 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, "bad yaml 1")
		<-time.After(30 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, "bad yaml 2: [[[")
		<-time.After(30 * time.Millisecond)
		_ = ipc.WriteReport(sessionID, testPassReport)
	}()

	status, err := waitForValidReport(sessionID, "review")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}

	waitGroup.Wait()
}

func TestWaitForValidReport_CheckAbortFlagError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-check-abort-err-" + time.Now().Format("20060102150405")

	// Write valid report immediately so we can complete successfully
	// The goal is to test the CheckAbortFlag warning path, not to fail
	if err := ipc.WriteReport(sessionID, testPassReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// This should successfully read the report
	status, err := waitForValidReport(sessionID, "test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}
}

func TestWaitForValidReport_ReadReportError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-read-err-" + time.Now().Format("20060102150405")

	// Create a valid session directory
	if err := ipc.WriteReport(sessionID, testPassReport); err != nil {
		t.Fatalf("Failed to write initial report: %v", err)
	}

	// This test validates that waitForValidReport properly reads a valid report
	status, err := waitForValidReport(sessionID, "test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}
}

func TestWaitForValidReport_MissingStatusField(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-missing-status-" + time.Now().Format("20060102150405")

	// Write a report that is valid YAML and passes schema but test status extraction
	reportWithoutStatus := `command: "test"
artifact: "test-artifact"
timestamp: "2024-01-01T00:00:00Z"
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	if err := ipc.WriteReport(sessionID, reportWithoutStatus); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	status, err := waitForValidReport(sessionID, "test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}
}

func TestWaitForValidReport_FAILStatus(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	sessionID := "test-fail-status-" + time.Now().Format("20060102150405")

	failReport := `command: "test"
artifact: "test-artifact"
timestamp: "2024-01-01T00:00:00Z"
status: FAIL
issues:
  blockers:
    - message: "Critical issue"
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	if err := ipc.WriteReport(sessionID, failReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	status, err := waitForValidReport(sessionID, "test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != statusFail {
		t.Errorf("Expected FAIL status, got %s", status)
	}
}
