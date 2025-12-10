---
id: m04-e05
title: User handles history errors gracefully
milestone: m04
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User handles history errors gracefully

## Overview
User attempts history actions when preconditions are not met (e.g., missing `FLUXID_SESSION_ID`, invalid input, wrapper not running) → System prints clear, actionable error messages and returns appropriate exit codes → Wrapper workflow continues; user can correct the issue and retry successfully.

## Scope
- User actions: Invoke `--write-history` or `ipc view-history` without a valid session; provide empty message; attempt usage against non-running session
- System responses: Detect error condition; print concise error with remediation hint; do not abort wrapper; allow retry when corrected
- Screens/states: Terminal error output; wrapper remains running

## Success Criteria
- [ ] Missing or invalid `FLUXID_SESSION_ID` yields clear error [Test: unset env; verify message and non-zero exit]
- [ ] Empty message rejected with validation feedback [Test: `ipc write-history ""` returns validation error]
- [ ] IPC command failures never abort wrapper [Test: induce failures while wrapper runs; wrapper continues]
- [ ] Errors are plain text and consistent in format [Test: regex match across errors]
- [ ] Recovery works: after fixing issue, commands succeed [Test: set session and retry; success]

## Dependencies
- External: Wrapper that establishes `FLUXID_SESSION_ID` (earlier milestones m01–m03)
- Internal: None

## E2E Test Mapping
**Test File**: `m04_e05_t01_error_handling.yaml`

**Test Flow**:
1. Unset `FLUXID_SESSION_ID`; run `fluxid ipc view-history`; observe clear error
2. Run `fluxid ipc write-history ""`; observe validation error
3. Set `FLUXID_SESSION_ID` and retry both commands; both succeed
4. Confirm wrapper remains running throughout

**Key Assertions**:
- Specific and actionable error text
- Correct exit codes for error vs success
- No side effects on wrapper or session state

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Keep error messages user-friendly and consistent; avoid leaking internal details.