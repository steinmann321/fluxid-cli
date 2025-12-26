package workflow

import (
	"context"
	"sync"
	"testing"
	"time"
)

// Test helper constants for workflow tests.
const (
	testAgentEcho  = "echo"
	testAgentTrue  = "true"
	testAgentFalse = "false"
)

// testDataDirMutex prevents concurrent tests from interfering with each other's XDG_DATA_HOME.
//
// FLAKINESS PREVENTION:
// Multiple tests running in parallel (especially across packages) can interfere when they
// all modify the XDG_DATA_HOME environment variable. Since environment variables are
// process-global, this causes race conditions where:
//  1. Test A sets XDG_DATA_HOME and starts a goroutine that reads it
//  2. Test B sets XDG_DATA_HOME to a different value
//  3. Test A's main code reads XDG_DATA_HOME and gets B's value
//  4. Test A's goroutine writes to directory A, but main code reads from directory B
//
// This mutex ensures only one test at a time can set up and use XDG_DATA_HOME.
//
//nolint:gochecknoglobals // Global mutex required for test isolation across parallel test execution
var testDataDirMutex sync.Mutex

// setupTestDataDir sets up an isolated data directory for the test.
// It locks testDataDirMutex to prevent concurrent tests from interfering.
//
// IMPORTANT: Always call the returned cleanup function via defer:
//
//	tmpDir, cleanup := setupTestDataDir(t)
//	defer cleanup()
//
// This pattern ensures proper test isolation and prevents flakiness.
//
//nolint:nonamedreturns // Named returns improve clarity for this test helper API
func setupTestDataDir(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()

	testDataDirMutex.Lock()

	tmpDir = t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	cleanup = func() {
		testDataDirMutex.Unlock()
	}

	return tmpDir, cleanup
}

// testContext creates a context with timeout for testing.
// This helper avoids direct context.Background() calls in test files.
//
//nolint:unused // Reserved for future test refactoring
func testContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// Phase command constants for test validation.
const (
	phaseCommandImplement = "fluxid.implement"
	phaseCommandCommit    = "fluxid.commit"
	phaseCommandReview    = "fluxid.review"
)

// Test report templates.
const testPassReport = `command: test-command
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Test successful
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Complete
`

const testFailReport = `command: test-command
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
  - Retry
`

// Phase-specific test report templates for implement, commit, and review phases.
// These templates ensure tests properly validate phase-specific report handling.
//
// #nosec G101 -- False positive: test data, not credentials
const testImplementPassReport = `command: fluxid.implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Implement phase successful
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Continue to commit phase
`

const testImplementFailReport = `command: fluxid.implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: FAIL
summary: Implement phase failed
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Retry implement phase
`

// #nosec G101 -- False positive: test data, not credentials
const testCommitPassReport = `command: fluxid.commit
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Commit phase successful
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Continue to review phase
`
