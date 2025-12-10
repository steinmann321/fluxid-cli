---
id: m03-e05
title: User aborts runaway workflow gracefully
milestone: m03
status: pending
patterns: []
---

# Epic: User aborts runaway workflow gracefully

## Overview
User halts a running workflow either with Ctrl+C or `fluxid ipc abort [--session <id>]`. The system sets an abort flag, allows the current agent invocation to complete, then exits cleanly, with multiple signals forcing immediate exit.

## Scope
- User actions: Send SIGINT/SIGTERM (Ctrl+C) or run `fluxid ipc abort`
- System responses: Set abort flag for session; gracefully stop after phase; confirm abort; immediate exit on repeated signals
- Screens/states: CLI confirmations; exit codes; session-scoped abort state

## Success Criteria
- [ ] Ctrl+C sets session abort and stops workflow after safe point [Test: simulate signal; observe graceful shutdown]
- [ ] `fluxid ipc abort` marks session and causes clean exit post-phase [Test: run command; observe confirmation and exit]
- [ ] Multiple signals force immediate termination [Test: send repeated SIGINT; verify immediate exit]

## Dependencies
m03-e04-user-controls-loop-progression-via-pass-fail

## E2E Test Mapping
**Test File**: `m03_e05_t01_graceful-abort.yaml`

**Test Flow**:
1. Start workflow with long-running phase
2. Issue Ctrl+C once; observe graceful finish of current action, then exit
3. Repeat with two rapid Ctrl+C; verify immediate exit
4. Repeat using `fluxid ipc abort`; verify same behavior

**Key Assertions**:
- Abort flag set and respected
- User receives clear confirmation messages
- Exit codes are non-zero and consistent

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Ensure abort does not corrupt in-memory report state; respect session overrides.