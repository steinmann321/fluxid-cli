---
id: m06-e04
title: User runs dry-run with missing command files and sees clear errors
milestone: m06
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User runs dry-run with missing command files and sees clear errors

## Overview
User invokes `fluxid --fluxid-dry-run` with configuration referencing non-existent command files. System performs validation, reports each missing file with clear paths and phase context, stops simulation before plan execution, and exits non-zero with actionable guidance.

## Scope
- User actions: Run dry-run referencing one or more missing command files
- System responses: Validate file existence/readability; aggregate and report missing paths with phase names; halt simulation; exit non-zero
- Screens/states: Terminal error output; no simulation plan printed beyond validation failure

## Success Criteria
- [ ] All referenced command files are validated before simulation [Test: missing files detected pre-plan]
- [ ] Error messages include phase, expected path, and how to fix [Test: message format checked]
- [ ] Multiple missing files are reported together [Test: show all errors, not first-only]
- [ ] Exit status is non-zero on validation failure [Test: shell exit code != 0]
- [ ] No agent process is spawned [Test: assert no external process invoked]

## Dependencies
m06-e01

## E2E Test Mapping
**Test File**: `m06_e04_t01_dry-run-missing-command-files.yaml`

**Test Flow**:
1. Invoke dry-run with config pointing to missing files
2. Verify validation stops simulation and prints detailed errors
3. Confirm exit code non-zero
4. Confirm no plan lines nor agent output produced

**Key Assertions**:
- Errors enumerate all missing file paths with phase context
- No "Would execute:" lines appear
- Exit code non-zero
- No agent process spawned

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Prefer aggregating validation errors to reduce fix/retry cycles for users.
