---
id: m03-e04
title: User controls loop progression via PASS/FAIL reports
milestone: m03
status: pending
patterns: []
---

# Epic: User controls loop progression via PASS/FAIL reports

## Overview
During implement and review phases, user-provided reports determine loop flow: PASS ends the current retry/iteration; FAIL continues. Missing/invalid reports cause infinite retry of the active phase until a valid report is produced.

## Scope
- User actions: Produce PASS/FAIL reports at phase end; optionally omit or send invalid report
- System responses: Evaluate report status; break or continue loops accordingly; infinite retry on invalid/missing
- Screens/states: CLI loop status output; phase transitions; counters/logging

## Success Criteria
- [ ] PASS after implement moves to review; PASS after review completes workflow [Test: scripted sequence with two PASS reports]
- [ ] FAIL triggers next retry/iteration without incrementing when report invalid/missing [Test: invalid then valid FAIL then PASS]
- [ ] Infinite retry occurs only on invalid/missing report for the active phase [Test: observe repeated prompts/logs]

## Dependencies
m03-e02-user-writes-valid-report-and-verifies
m03-e03-user-submits-invalid-report-and-receives-diagnostics

## E2E Test Mapping
**Test File**: `m03_e04_t01_loop-control-pass-fail.yaml`

**Test Flow**:
1. Run wrapper workflow through implement → review
2. Submit FAIL after implement; observe retry of implement
3. Submit PASS after implement; transition to review
4. Submit PASS after review; workflow exits

**Key Assertions**:
- Correct phase transitions on report status
- No state corruption across retries/iterations
- Counters/logs reflect behavior accurately

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Ensure retries on invalid/missing do not increment iteration counters.