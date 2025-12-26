//nolint:paralleltest // E2E tests with subprocess execution
package tests

import (
	"fluxid-cli/internal/ipc"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

const (
	statusPass = "PASS"
	statusFail = "FAIL"
)

// TestLoopProgressionPassAfterImplementAndReview tests the basic PASS flow:
// implement (PASS) → review (PASS) → workflow complete.
func TestLoopProgressionPassAfterImplementAndReview(t *testing.T) {
	sessionID := "test-session-pass-flow"
	setupReportDir(t)

	// Simulate fluxid workflow with mock phases
	implementPhase := func() {
		writeReport(t, sessionID, statusPass, "implement")
	}

	reviewPhase := func() {
		writeReport(t, sessionID, statusPass, "review")
	}

	// Run phases
	implementPhase()
	status := readReportStatus(t, sessionID)
	if status != statusPass {
		t.Errorf("Expected implement status PASS, got %s", status)
	}

	reviewPhase()
	status = readReportStatus(t, sessionID)
	if status != statusPass {
		t.Errorf("Expected review status PASS, got %s", status)
	}
}

// TestLoopProgressionFailThenPass tests FAIL triggering retry:
// implement (FAIL) → implement (PASS) → review (PASS) → complete.
func TestLoopProgressionFailThenPass(t *testing.T) {
	sessionID := "test-session-fail-then-pass"
	setupReportDir(t)

	// First attempt: FAIL
	writeReport(t, sessionID, statusFail, "implement-attempt-1")
	status := readReportStatus(t, sessionID)
	if status != statusFail {
		t.Errorf("Expected FAIL status, got %s", status)
	}

	// Second attempt: PASS
	writeReport(t, sessionID, statusPass, "implement-attempt-2")
	status = readReportStatus(t, sessionID)
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}

	// Review: PASS
	writeReport(t, sessionID, statusPass, "review")
	status = readReportStatus(t, sessionID)
	if status != statusPass {
		t.Errorf("Expected PASS status, got %s", status)
	}
}

// TestInvalidReportValidation tests that invalid reports are rejected.
func TestInvalidReportValidation(t *testing.T) {
	sessionID := "test-session-invalid"
	setupReportDir(t)

	// Write invalid report (missing required fields)
	invalidReport := `command: test
status: PASS
`
	err := ipc.WriteReport(sessionID, invalidReport)
	if err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	// Validate should fail
	reportYAML, err := ipc.ReadReport(sessionID)
	if err != nil {
		t.Fatalf("Failed to read report: %v", err)
	}

	err = ipc.ValidateReport(reportYAML)
	if err == nil {
		t.Error("Expected validation error for invalid report, got nil")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Expected validation error message, got: %v", err)
	}
}

// TestMissingReportHandling tests behavior when no report exists.
func TestMissingReportHandling(t *testing.T) {
	sessionID := "test-session-missing"
	setupReportDir(t)

	// Read non-existent report
	reportYAML, err := ipc.ReadReport(sessionID)
	if err != nil {
		t.Fatalf("Expected no error for missing report, got: %v", err)
	}

	if reportYAML != "" {
		t.Errorf("Expected empty string for missing report, got: %s", reportYAML)
	}
}

// TestIPCWriteReportCommand tests the write-report IPC command.
func TestIPCWriteReportCommand(t *testing.T) {
	sessionID := "test-session-ipc-write"
	setupReportDir(t)

	validReport := `command: fluxid.implement
artifact: m03-e04
timestamp: 2025-12-12T10:00:00Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Write via IPC command
	cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin, "ipc", "write-report")
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)
	cmd.Stdin = strings.NewReader(validReport)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("write-report failed: %v\nOutput: %s", err, output)
	}

	// Verify report was written
	reportYAML, err := ipc.ReadReport(sessionID)
	if err != nil {
		t.Fatalf("Failed to read report: %v", err)
	}

	if !strings.Contains(reportYAML, "status: PASS") {
		t.Errorf("Report missing expected content, got: %s", reportYAML)
	}
}

// TestIPCReadReportCommand tests the read-report IPC command.
func TestIPCReadReportCommand(t *testing.T) {
	sessionID := "test-session-ipc-read"
	setupReportDir(t)

	// Write a report first
	writeReport(t, sessionID, statusPass, "test-artifact")

	// Build fluxid binary
	fluxidBin := buildFluxidBinary(t)

	// Read via IPC command
	cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin, "ipc", "read-report")
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+sessionID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("read-report failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "status: PASS") {
		t.Errorf("Output missing expected status, got: %s", outputStr)
	}
}

// TestWaitForValidReportTimeout tests that waitForValidReport retries on missing reports.
// This test simulates the retry behavior by writing a report after a delay.
func TestWaitForValidReportRetry(t *testing.T) {
	defer goleak.VerifyNone(t)

	sessionID := "test-session-wait-retry"
	setupReportDir(t)

	// Start goroutine that writes report after delay
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		<-time.After(1 * time.Second)
		writeReport(t, sessionID, statusPass, "delayed-write")
	}()

	// Simulate waiting behavior (in real workflow, this is in waitForValidReport)
	start := time.Now()
	var reportYAML string
	var err error

	for i := 0; i < 5; i++ {
		reportYAML, err = ipc.ReadReport(sessionID)
		if err != nil {
			t.Fatalf("Failed to read report: %v", err)
		}
		if reportYAML != "" {
			break
		}
		<-time.After(500 * time.Millisecond)
	}

	elapsed := time.Since(start)

	if reportYAML == "" {
		t.Error("Report was never written")
	}

	if elapsed < 1*time.Second {
		t.Errorf("Expected at least 1 second delay, got %v", elapsed)
	}

	waitGroup.Wait()
}

// Helper functions

func setupReportDir(t *testing.T) {
	t.Helper()
	reportsDir := filepath.Join(os.TempDir(), "fluxid-reports")
	if err := os.MkdirAll(reportsDir, 0o750); err != nil {
		t.Fatalf("Failed to create reports directory: %v", err)
	}
	t.Cleanup(func() {
		// Clean up test reports after test
		_ = os.RemoveAll(reportsDir)
	})
}

func writeReport(t *testing.T, sessionID string, status string, artifact string) {
	t.Helper()
	report := fmt.Sprintf(`command: fluxid.test
artifact: %s
timestamp: %s
status: %s
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`, artifact, time.Now().Format(time.RFC3339), status)

	if err := ipc.WriteReport(sessionID, report); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}
}

func readReportStatus(t *testing.T, sessionID string) string {
	t.Helper()
	reportYAML, err := ipc.ReadReport(sessionID)
	if err != nil {
		t.Fatalf("Failed to read report: %v", err)
	}

	if reportYAML == "" {
		return ""
	}

	// Parse status from YAML
	lines := strings.Split(reportYAML, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "status:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func buildFluxidBinary(t *testing.T) string {
	t.Helper()

	// Build in temp directory
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "fluxid")

	// Find project root (go up from e2e-tests/tests/)
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	fluxidPath := filepath.Join(projectRoot, "cmd", "fluxid")
	cmd := exec.CommandContext(testCtx(30*time.Second), "go", "build", "-o", binPath, fluxidPath)
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build fluxid: %v\nOutput: %s", err, output)
	}

	return binPath
}
