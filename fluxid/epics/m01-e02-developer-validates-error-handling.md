---
id: m01-e02
title: Developer validates error handling and failure scenarios
milestone: m01
status: pending
patterns: []
---

# Epic: Developer validates error handling and failure scenarios

## Overview
Developer runs error path tests that validate main.go correctly propagates non-zero exit codes when command.Execute() fails. Developer → runs `go test ./cmd/fluxid` → sees error path tests execute → verifies exit code propagation → sees detailed failure scenarios covered → has confidence error handling works correctly with 2:1 error-to-success test ratio.

## Scope
- **User actions**: Developer runs error-focused tests, developer reviews which failure scenarios are validated, developer confirms exit code propagation
- **System responses**: Tests execute, mock command.Execute() to return non-zero codes, verify os.Exit(non-zero) would be called, display test results with clear error scenario descriptions
- **Screens/states**: Terminal output showing error test execution, detailed failure scenario names, coverage report showing error paths tested

## Success Criteria
- [ ] Developer can run tests that validate config loading errors return exit code 1 [Test: mock config errors, verify exit code 1]
- [ ] Developer can run tests that validate arg parsing errors return exit code 1 [Test: invalid arguments, verify exit code 1]
- [ ] Developer can run tests that validate agent validation errors return exit code 1 [Test: missing agent binary, verify exit code 1]
- [ ] Developer can run tests that validate workflow execution errors propagate correctly [Test: command.Execute() returns various non-zero codes, verify propagation]
- [ ] Error path tests outnumber success tests at 2:1 ratio minimum [Test: count test cases, verify ratio >= 2:1]
- [ ] Test failures provide clear diagnostics about which error scenario failed [Test: descriptive names, assertion messages indicate exact failure point]
- [ ] Error path tests contribute majority of coverage toward 90% target [Test: run with -cover, verify error paths account for 50-60% coverage]
- [ ] Tests handle edge cases like nil values and panic recovery [Test: nil config scenarios, recover from panics if applicable]

## Dependencies
**Builds on**: m01-e01 (happy path test framework established)

The test framework and helper utilities from e01 are reused, but error scenarios are tested independently.

## E2E Test Mapping
**Test File**: `m01_e02_t01_error-handling.yaml`

**Test Flow**:
1. Developer opens terminal and navigates to project
2. Developer runs `go test ./cmd/fluxid/... -v` to see detailed error test names
3. Tests execute simulating various command.Execute() failures
4. Developer sees each error scenario test pass
5. Developer runs `go test -cover ./cmd/fluxid/...`
6. Developer sees coverage report showing error paths thoroughly tested
7. Developer verifies 2:1 error-to-success ratio in test output

**Key Assertions**:
- Test output contains error scenario names (e.g., "TestMain_ConfigLoadingError", "TestMain_ArgParsingError")
- Each error test validates specific exit code returned
- Coverage report shows 90%+ total coverage when combined with happy path tests
- Error test count is at least 2x happy path test count

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone (achieves 2:1 error ratio, drives coverage to 90%+)
- [ ] No regressions
- [ ] ONE atomic flow (developer validates error scenarios work)

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone (achieves 2:1 error ratio, drives coverage to 90%+)
- [ ] No regressions
- [ ] ONE atomic flow (developer validates error scenarios work)

## Notes
**Error Scenarios to Cover**:
- Config loading failures (home config error, project config error, env config error)
- Argument parsing failures (invalid flags, malformed arguments)
- Agent validation failures (agent not in PATH, agent not executable, unsupported agent)
- Workflow execution failures (command.Execute() returns various non-zero exit codes)
- Special command failures if applicable (ipc command errors, --write-history errors)

**2:1 Ratio Implementation**:
If happy path has 2 test cases, error path should have minimum 4 test cases. Use table-driven tests to efficiently cover multiple error scenarios.

**Integration with e01**:
Reuse test helpers and mocking utilities from e01, but focus exclusively on failure paths. The combination of e01 (success) + e02 (errors) should achieve 90%+ coverage.

**Coverage Target**:
Error path tests should cover approximately 50-60% of main.go execution paths. Combined with e01 (30-40%), total should exceed 90% threshold.
