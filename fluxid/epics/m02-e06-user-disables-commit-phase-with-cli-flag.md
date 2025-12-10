---
id: m02-e06
title: User disables commit phase with CLI flag
milestone: m02
status: pending
patterns: []
---

# Epic: User disables commit phase with CLI flag

## Overview
User runs `fluxid` with `--fluxid-no-commit` to disable the commit phase while keeping the review phase mandatory. The system reflects the change in initialization status, adjusts loop orchestration accordingly, and never executes commit steps during this run.

## Scope
- User actions: Run `fluxid --fluxid-no-commit` within a configured project.
- System responses: Parse the flag; set `commit_enabled=false`; ensure any commit steps are skipped; show status reflecting the disabled commit phase; preserve review phase as mandatory.
- Screens/states: Terminal output showing initialization status and subsequent workflow without commit execution.

## Success Criteria
- [ ] `--fluxid-no-commit` reliably disables commit phase [Test: flag present vs. absent; commit steps skipped only when present]
- [ ] Initialization status shows `commit_enabled=false` with source=cli [Test: verify source labeling]
- [ ] Review phase remains mandatory and executes [Test: ensure review steps still run]
- [ ] No commit-related side effects occur [Test: logs and hooks not invoked]

## Dependencies
m02-e04

## E2E Test Mapping
**Test File**: `m02_e06_t01_no-commit-flag.yaml`

**Test Flow**:
1. Prepare a project with standard configuration
2. Run with `--fluxid-no-commit`
3. Verify status and observe run to confirm commit never executes
4. Run without the flag to confirm normal behavior

**Key Assertions**:
- Commit phase skipped only when flag set
- Review phase always runs
- Status accurately reflects the toggle

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Follow business rule: review cannot be disabled.