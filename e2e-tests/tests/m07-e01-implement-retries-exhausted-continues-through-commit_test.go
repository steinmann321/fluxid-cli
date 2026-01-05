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

// TestImplementRetriesExhaustedContinuesThroughCommit verifies that when all implement retries fail,
// the workflow continues through commit phase to review instead of terminating.
// Flow: implement (FAIL) → implement (FAIL) → commit → review (PASS) → workflow succeeds.
//
//nolint:cyclop,funlen,gocognit // E2E test with multiple conditional checks for output validation
func TestImplementRetriesExhaustedContinuesThroughCommit(t *testing.T) {
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

	// Create config with commands section (using claude as default, will be overridden by --agent flag)
	setupConfigWithCommands(t, homeDir, "claude")

	t.Cleanup(func() {
		_ = os.RemoveAll(reportsDir)
	})

	sessionID := fmt.Sprintf("test-impl-exhaust-%d", time.Now().UnixNano())

	// Build fluxid binary
	root := getProjectRoot(t)
	buildFluxid(t, root)
	fluxidBin := filepath.Join(root, "bin", "fluxid")

	// v2.0: Create mock agent in bin directory with valid agent name
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}
	agentScript := filepath.Join(binDir, "codex")
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
		commitWritten := false
		waitCycles := 0

		for {
			select {
			case <-stopReportWriter:
				return
			case <-time.After(100 * time.Millisecond):
				currentReport, _ := ipc.ReadReport(sessionID)

				// State machine: write reports based on current phase
				//nolint:gocritic // Complex test state management, if-else more readable than switch
				if implementAttempts < 2 && (currentReport == "" || strings.Contains(currentReport, "fluxid.implement")) {
					// First 2 implement attempts: write FAIL reports
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
				} else if implementAttempts >= 2 && !commitWritten {
					// Wait a few cycles after implement exhausts, then write commit PASS
					waitCycles++
					if waitCycles >= 2 {
						commitReport := fmt.Sprintf(`command: fluxid.commit
artifact: test-commit
timestamp: %s
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`, time.Now().Format(time.RFC3339))
						_ = ipc.WriteReport(sessionID, commitReport)
						commitWritten = true
						reportWritten <- "commit-pass"
					}
				} else if commitWritten && strings.Contains(currentReport, "fluxid.review") {
					// After commit succeeds, write review PASS
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

	// Create task file
	taskPath := filepath.Join(homeDir, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}

	// v2.0: Run fluxid workflow with --codex flag
	cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin,
		"--fluxid-iterations=1",
		"--fluxid-implement-retries=2",
		"--codex",
		"--file="+taskPath,
	)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+binDir+":"+os.Getenv("PATH"),
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
	if !strings.Contains(outputStr, "Status: SUCCESS") {
		t.Error("Expected workflow to complete successfully")
	}

	waitGroup.Wait()
}

// TestImplementRetriesExhaustedWithCommitEnabled verifies that when implement retries fail,
// the workflow continues through commit phase and then review.
// Flow: implement (FAIL) → implement (FAIL) → commit → review (PASS) → workflow succeeds.
//
//nolint:cyclop,funlen,gocognit // E2E test with multiple conditional checks for output validation
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

	// Create config with commands section (using claude as default, will be overridden by --agent flag)
	setupConfigWithCommands(t, homeDir, "claude")

	t.Cleanup(func() {
		_ = os.RemoveAll(reportsDir)
	})

	sessionID := fmt.Sprintf("test-impl-exhaust-commit-%d", time.Now().UnixNano())

	// Build fluxid binary
	root := getProjectRoot(t)
	buildFluxid(t, root)
	fluxidBin := filepath.Join(root, "bin", "fluxid")

	// v2.0: Create mock agent in bin directory with valid agent name
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}
	agentScript := filepath.Join(binDir, "codex")
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
		commitWritten := false
		waitCycles := 0
		for {
			select {
			case <-stopReportWriter:
				return
			case <-time.After(100 * time.Millisecond):
				currentReport, _ := ipc.ReadReport(sessionID)

				//nolint:gocritic // Complex test state management, if-else more readable than switch
				if implementAttempts < 2 && (currentReport == "" || strings.Contains(currentReport, "fluxid.implement")) {
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
				} else if implementAttempts >= 2 && !commitWritten {
					// Wait a few cycles after implement exhausts, then write commit PASS
					waitCycles++
					if waitCycles >= 2 {
						commitReport := fmt.Sprintf(`command: fluxid.commit
artifact: test-commit
timestamp: %s
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`, time.Now().Format(time.RFC3339))
						_ = ipc.WriteReport(sessionID, commitReport)
						commitWritten = true
					}
				} else if commitWritten && strings.Contains(currentReport, "fluxid.review") {
					// Review phase is running, write review PASS
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

	// Create task file
	taskPath := filepath.Join(homeDir, "task.txt")
	if err := os.WriteFile(taskPath, []byte("task"), 0o644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}

	// v2.0: Run fluxid workflow with --codex flag
	cmd := exec.CommandContext(testCtx(30*time.Second), fluxidBin,
		"--fluxid-iterations=1",
		"--fluxid-implement-retries=2",
		"--codex",
		"--file="+taskPath,
	)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+binDir+":"+os.Getenv("PATH"),
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
	if !strings.Contains(outputStr, "Commit attempt") {
		t.Error("Expected commit phase to execute after implement retries exhausted")
	}

	// Verify review phase executed
	if !strings.Contains(outputStr, "Running review phase") {
		t.Error("Expected review phase to execute")
	}

	// Verify workflow completed
	if !strings.Contains(outputStr, "Status: SUCCESS") {
		t.Error("Expected workflow to complete successfully")
	}

	waitGroup.Wait()
}
