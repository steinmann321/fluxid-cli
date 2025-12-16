package main

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
