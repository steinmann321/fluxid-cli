---
id: m04-e02
title: User writes history via IPC command
milestone: m04
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User writes history via IPC command

## Overview
User (or external tool) runs `fluxid ipc write-history <message>` during an active workflow session → System validates session context and message → System generates ISO 8601 timestamp and appends `[timestamp] message` to in-memory history → Command prints success response; workflow continues uninterrupted.

## Scope
- User actions: Execute `fluxid ipc write-history <message>` from a separate process or shell
- System responses: Validate `FLUXID_SESSION_ID`; handle concurrent writes safely; timestamp and append; respond with success or clear error
- Screens/states: Terminal output; in-memory per-session history buffer

## Success Criteria
- [ ] Appends entry with auto-generated ISO 8601 timestamp [Test: verify format and monotonic ordering across rapid writes]
- [ ] Concurrent writes do not corrupt history [Test: spawn multiple parallel writes; history remains complete and ordered]
- [ ] Command failures print clear error but do not abort wrapper [Test: simulate missing session; wrapper remains running]
- [ ] No disk persistence; session isolation honored [Test: distinct `FLUXID_SESSION_ID` values never leak entries]
- [ ] Supports UTF-8 messages without truncation beyond global size policy [Test: multi-byte characters round-trip]

## Dependencies
- External: Active session context via `FLUXID_SESSION_ID` (from earlier milestones m01–m03)
- Internal: None

## E2E Test Mapping
**Test File**: `m04_e02_t01_write_ipc_command.yaml`

**Test Flow**:
1. Export `FLUXID_SESSION_ID` for this shell
2. Run `fluxid ipc write-history "Decision: adopt FIFO eviction"`
3. Assert success output and zero exit code
4. Run several parallel writes; ensure no errors

**Key Assertions**:
- Each write produces one well-formed line in history
- No interleaving corruption or partial writes under concurrency
- Errors (e.g., missing session) are clear and non-fatal

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Use ordering guarantees suitable for the runtime (e.g., mutex/lock-free queue) to maintain chronological order on append.
- Enforce minimal validation: non-empty message input; trim trailing newlines before storage.