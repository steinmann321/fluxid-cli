//nolint:paralleltest // E2E tests with subprocess execution
package tests

import (
	"fluxid-loop/internal/ipc"
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

// TestImplementRetriesExhaustedContinuesToReview verifies that when all implement retries fail,
// the workflow continues to the review phase instead of terminating.
// Flow: implement (FAIL) → implement (FAIL) → review (PASS) → workflow succeeds.
//
//nolint:cyclop,funlen // E2E test with multiple conditional checks for output validation
func TestImplementRetriesExhaustedContinuesToReview(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Setup
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	reportsDir := filepath.Join(os.TempDir(), "fluxid-reports")

	if err := os.MkdirAll(homeDir, 0o750); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o750); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}
	if err := os.MkdirAll(reportsDir, 0o750); err != nil {
		t.Fatalf("Failed to create reports dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(reportsDir)
	})

	sessionID := fmt.Sprintf("test-impl-exhaust-%d", time.Now().UnixNano())

	// Build fluxid binary
	fluxidBin := buildFluxidForE2E(t)

	// Create mock agent script that succeeds (we'll control via reports)
	agentScript := filepath.Join(tmpDir, "mock-agent.sh")
	agentContent := `#!/bin/bash
echo "Mock agent executing..."
exit 0
`
	if err := os.WriteFile(agentScript, []byte(agentContent), 0o755); err != nil {
		t.Fatalf("Failed to write agent script: %v", err)
	}

	// Start report writer in background
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	reportWritten := make(chan string, 10)
	stopReportWriter := make(chan bool)

	go func() {
		defer waitGroup.Done()
		implementAttempts := 0
		for {
			select {
			case <-stopReportWriter:
				return
			case <-time.After(100 * time.Millisecond):
				// Read current report to determine phase
				currentReport, _ := ipc.ReadReport(sessionID)

				// Count implement attempts by checking if we've written FAIL reports
				if currentReport == "" || strings.Contains(currentReport, "fluxid.implement") {
					if implementAttempts < 2 {
						// First 2 attempts: FAIL
						failReport := fmt.Sprintf(`command: fluxid.implement
artifact: test-artifact-attempt-%d
timestamp: %s
status: FAIL
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`, implementAttempts+1, time.Now().Format(time.RFC3339))
						_ = ipc.WriteReport(sessionID, failReport)
						implementAttempts++
						reportWritten <- "implement-fail"
					}
				} else if strings.Contains(currentReport, "fluxid.implement") && implementAttempts >= 2 {
					// After implement retries exhausted, write review PASS
					reviewReport := fmt.Sprintf(`command: fluxid.review
artifact: test-review
timestamp: %s
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`, time.Now().Format(time.RFC3339))
					_ = ipc.WriteReport(sessionID, reviewReport)
					reportWritten <- "review-pass"
					return
				}
			}
		}
	}()

	// Run fluxid workflow
	cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin,
		"--fluxid-iterations", "1",
		"--fluxid-implement-retries", "2",
		"--fluxid-no-commit",
		"--agent", agentScript,
	)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"FLUXID_SESSION_ID="+sessionID,
	)
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	close(stopReportWriter)

	// Verify workflow succeeded
	if err != nil {
		t.Errorf("Workflow should succeed even with implement failures, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Verify implement attempts were logged
	if !strings.Contains(outputStr, "Implement attempt 1/2") {
		t.Error("Expected to see first implement attempt in output")
	}
	if !strings.Contains(outputStr, "Implement attempt 2/2") {
		t.Error("Expected to see second implement attempt in output")
	}

	// Verify review phase was executed
	if !strings.Contains(outputStr, "Running review phase") {
		t.Error("Expected review phase to execute after implement retries exhausted")
	}

	// Verify workflow completed successfully
	if !strings.Contains(outputStr, "Workflow completed successfully") {
		t.Error("Expected workflow to complete successfully")
	}

	waitGroup.Wait()
}

// TestImplementRetriesExhaustedWithCommitEnabled verifies that when implement retries fail,
// the workflow continues to commit phase (if enabled) and then review.
// Flow: implement (FAIL) → implement (FAIL) → commit → review (PASS) → workflow succeeds.
//
//nolint:cyclop,funlen // E2E test with multiple conditional checks for output validation
func TestImplementRetriesExhaustedWithCommitEnabled(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Setup
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	reportsDir := filepath.Join(os.TempDir(), "fluxid-reports")

	if err := os.MkdirAll(homeDir, 0o750); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o750); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}
	if err := os.MkdirAll(reportsDir, 0o750); err != nil {
		t.Fatalf("Failed to create reports dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(reportsDir)
	})

	sessionID := fmt.Sprintf("test-impl-exhaust-commit-%d", time.Now().UnixNano())

	// Build fluxid binary
	fluxidBin := buildFluxidForE2E(t)

	// Create mock agent script
	agentScript := filepath.Join(tmpDir, "mock-agent.sh")
	agentContent := `#!/bin/bash
echo "Mock agent executing..."
exit 0
`
	if err := os.WriteFile(agentScript, []byte(agentContent), 0o755); err != nil {
		t.Fatalf("Failed to write agent script: %v", err)
	}

	// Start report writer
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	stopReportWriter := make(chan bool)

	go func() {
		defer waitGroup.Done()
		implementAttempts := 0
		for {
			select {
			case <-stopReportWriter:
				return
			case <-time.After(100 * time.Millisecond):
				currentReport, _ := ipc.ReadReport(sessionID)

				if currentReport == "" || strings.Contains(currentReport, "fluxid.implement") {
					if implementAttempts < 2 {
						// FAIL reports for implement attempts
						failReport := fmt.Sprintf(`command: fluxid.implement
artifact: test-artifact-attempt-%d
timestamp: %s
status: FAIL
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`, implementAttempts+1, time.Now().Format(time.RFC3339))
						_ = ipc.WriteReport(sessionID, failReport)
						implementAttempts++
					}
				} else if implementAttempts >= 2 {
					// After retries exhausted, write review PASS
					reviewReport := fmt.Sprintf(`command: fluxid.review
artifact: test-review
timestamp: %s
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`, time.Now().Format(time.RFC3339))
					_ = ipc.WriteReport(sessionID, reviewReport)
					return
				}
			}
		}
	}()

	// Run with commit enabled
	cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin,
		"--fluxid-iterations", "1",
		"--fluxid-implement-retries", "2",
		"--fluxid-commit", // Enable commit phase
		"--agent", agentScript,
	)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"FLUXID_SESSION_ID="+sessionID,
	)
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	close(stopReportWriter)

	// Verify workflow succeeded
	if err != nil {
		t.Errorf("Workflow should succeed with commit enabled, got error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Verify implement attempts
	if !strings.Contains(outputStr, "Implement attempt 1/2") {
		t.Error("Expected first implement attempt")
	}
	if !strings.Contains(outputStr, "Implement attempt 2/2") {
		t.Error("Expected second implement attempt")
	}

	// Verify commit phase executed
	if !strings.Contains(outputStr, "Running commit phase") {
		t.Error("Expected commit phase to execute after implement retries exhausted")
	}

	// Verify review phase executed
	if !strings.Contains(outputStr, "Running review phase") {
		t.Error("Expected review phase to execute")
	}

	// Verify workflow completed
	if !strings.Contains(outputStr, "Workflow completed successfully") {
		t.Error("Expected workflow to complete successfully")
	}

	waitGroup.Wait()
}

// Helper to build fluxid binary for e2e tests.
func buildFluxidForE2E(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "fluxid")

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
