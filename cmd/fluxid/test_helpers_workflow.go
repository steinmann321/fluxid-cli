package main

import (
	"fluxid-loop/internal/ipc"
	"time"
)

// Test report templates for deterministic tests.
const testPassReport = `command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: PASS
summary: Test
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

const testFailReport = `command: test
artifact: test
timestamp: 2025-12-15T10:00:00Z
status: FAIL
summary: Always fail
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`

// writeReportWithRetry simulates async report writing for testing.
// It writes reports after delays, supporting different statuses per attempt.
func writeReportWithRetry(
	sessionID string,
	attempts int,
	failUntilAttempt int,
) {
	attemptCount := 0
	for i := 0; i < attempts; i++ {
		time.Sleep(100 * time.Millisecond)
		attemptCount++

		var report string
		if attemptCount < failUntilAttempt {
			report = `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: FAIL
summary: Attempt failed
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Retry
`
		} else {
			report = `command: test-implement
artifact: test-artifact
timestamp: 2025-12-13T10:00:00Z
status: PASS
summary: Success
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
next_steps:
  - Continue
`
		}
		_ = ipc.WriteReport(sessionID, report)
	}
}
