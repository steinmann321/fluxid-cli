---
id: m01-e03
title: Developer monitors coverage and enforces quality gates
milestone: m01
status: pending
patterns: []
---

# Epic: Developer monitors coverage and enforces quality gates

## Overview
Developer runs coverage reports to monitor test completeness and has quality automatically enforced through pre-commit hooks. Developer → runs tests with coverage flag → sees detailed coverage report → identifies any gaps → attempts commit → pre-commit hook validates coverage → commit succeeds or fails based on 90% threshold → developer has continuous quality enforcement.

## Scope
- **User actions**: Developer runs `go test -cover`, developer reads coverage report, developer identifies uncovered lines, developer attempts git commit
- **System responses**: Display coverage percentage and uncovered line numbers, pre-commit hook executes tests and validates coverage, hook blocks commit if < 90% or allows commit if >= 90%
- **Screens/states**: Terminal showing coverage report with percentages and line numbers, pre-commit hook output showing test execution and coverage validation, commit success or failure message

## Success Criteria
- [ ] Developer can run `go test -cover ./cmd/fluxid/...` and see coverage percentage [Test: execute command, verify percentage displayed]
- [ ] Coverage report shows which specific lines are uncovered [Test: verify line numbers of uncovered code are shown]
- [ ] Coverage achieves 90%+ for cmd/fluxid package [Test: verify coverage >= 90% threshold]
- [ ] Developer can identify exactly what needs testing from coverage report [Test: uncovered lines point to specific code that lacks tests]
- [ ] Pre-commit hook executes tests automatically on commit attempt [Test: attempt commit, verify tests run]
- [ ] Pre-commit hook enforces 90% coverage threshold [Test: mock < 90% coverage, verify commit blocked; mock >= 90%, verify commit allowed]
- [ ] Pre-commit hook provides clear diagnostic if coverage insufficient [Test: verify error message shows current coverage and required threshold]
- [ ] Coverage validation executes in seconds for fast commit feedback [Test: measure pre-commit execution time, verify < 5s]
- [ ] Coverage metrics are tracked consistently across test runs [Test: run coverage multiple times, verify consistent results]

## Dependencies
**Requires**: m01-e01 and m01-e02 (tests must exist to generate meaningful coverage)

This epic adds coverage monitoring and enforcement on top of the test suite created in e01 and e02.

## E2E Test Mapping
**Test File**: `m01_e03_t01_coverage-monitoring.yaml`

**Test Flow**:
1. Developer runs `go test -cover ./cmd/fluxid/...`
2. System displays coverage percentage (should be 90%+)
3. Developer sees which lines are covered vs uncovered
4. Developer attempts `git commit -m "test commit"`
5. Pre-commit hook executes automatically
6. Hook runs tests and checks coverage
7. If >= 90%: commit succeeds, developer sees success message
8. If < 90%: commit blocked, developer sees coverage diagnostic

**Key Assertions**:
- Coverage report shows percentage >= 90%
- Coverage report includes line numbers of any uncovered code
- Pre-commit hook output shows "Running tests..." message
- Pre-commit hook output shows coverage check result
- Commit succeeds only when coverage >= 90%
- Error messages clearly state current coverage and threshold

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone (enables continuous quality enforcement)
- [ ] No regressions
- [ ] ONE atomic flow (developer monitors coverage and has quality enforced)

## Notes
**Coverage Tools**:
- Use `go test -cover` for basic coverage percentage
- Use `go test -coverprofile=coverage.out` to generate detailed coverage data
- Use `go tool cover -func=coverage.out` to see per-function coverage
- Use `go tool cover -html=coverage.out` to visualize coverage in browser

**Pre-commit Hook Integration**:
The pre-commit hook should:
1. Run `go test -cover ./cmd/fluxid/...`
2. Parse coverage output to extract percentage
3. Compare against 90% threshold
4. Block commit if below threshold with clear message
5. Allow commit if at or above threshold

**Hook Policy Compliance**:
Per CLAUDE.md, the pre-commit hook must be strictly read-only regarding repository contents. It validates but never modifies code. This epic upholds that policy - the hook only reads test output and blocks/allows commits based on coverage.

**Coverage Reporting Best Practices**:
- Make coverage reports easily accessible (can run locally anytime)
- Provide actionable information (exact line numbers to test)
- Fast feedback (coverage check completes in seconds)
- Clear pass/fail criteria (90% is the bright line)

**Integration with e01 and e02**:
This epic assumes tests from e01 (happy paths) and e02 (error paths) already exist and achieve 90%+ coverage together. This epic adds the monitoring and enforcement layer on top of that foundation.
