---
id: m04-e03
title: User views history during execution
milestone: m04
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User views history during execution

## Overview
User runs `fluxid ipc view-history` during an active workflow session → System retrieves in-memory history entries for the session → System prints entries in chronological order as plain text, one per line, each prefixed by ISO 8601 timestamp → User reads the history to understand decisions and events.

## Scope
- User actions: Execute `fluxid ipc view-history`
- System responses: Validate `FLUXID_SESSION_ID`; stream entries in chronological order; produce stable, minimal formatting
- Screens/states: Terminal output of log lines; in-memory session history

## Success Criteria
- [ ] Outputs entries in chronological order, one per line [Test: write multiple entries; verify stable ordering]
- [ ] Each line uses `[YYYY-MM-DDTHH:MM:SSZ] message` format [Test: strict regex match]
- [ ] Handles empty history gracefully (no crash; deterministic output) [Test: fresh session prints blank or defined empty state]
- [ ] Efficient output up to 32MB limit [Test: performance check for large buffers]
- [ ] Safe under concurrent writes while viewing [Test: interleave writes and view; output is consistent and complete]

## Dependencies
- External: Active session context via `FLUXID_SESSION_ID` (from earlier milestones m01–m03)
- Internal: Populated history entries via `m04-e01` and/or `m04-e02`

## E2E Test Mapping
**Test File**: `m04_e03_t01_view_history.yaml`

**Test Flow**:
1. Set a new `FLUXID_SESSION_ID`
2. Write two entries via CLI flag and IPC write
3. Run `fluxid ipc view-history`
4. Verify two lines in order with ISO timestamps

**Key Assertions**:
- Exact per-line format and ordering
- No extraneous headers or footers; plain text only
- Works correctly when invoked multiple times within session

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Keep output deterministic to aid automation (e.g., piping, grep).