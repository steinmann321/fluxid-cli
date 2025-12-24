package workflow

import (
	"context"
	"time"
)

// Test helper constants for workflow tests.
const (
	testAgentEcho  = "echo"
	testAgentTrue  = "true"
	testAgentFalse = "false"
)

// testContext creates a context with timeout for testing.
// This helper avoids direct context.Background() calls in test files.
//
//nolint:unused // Reserved for future test refactoring
func testContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

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
