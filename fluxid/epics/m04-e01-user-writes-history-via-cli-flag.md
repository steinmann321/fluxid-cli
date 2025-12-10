---
id: m04-e01
title: User writes history via CLI flag
milestone: m04
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User writes history via CLI flag

## Overview
User runs `fluxid --write-history <message>` during an active workflow session → System validates session context → System generates ISO 8601 timestamp and appends `[timestamp] message` to in-memory history for the session → System prints a clear confirmation without interrupting the workflow.

## Scope
- User actions: Invoke CLI with `--write-history <message>` while session is active
- System responses: Validate `FLUXID_SESSION_ID`; generate timestamp; append entry; report success or error without aborting workflow
- Screens/states: Terminal output; in-memory per-session history store

## Success Criteria
- [ ] Appends entry with ISO 8601 timestamp prefix [Test: assert output ack; verify timestamp format `YYYY-MM-DDTHH:MM:SSZ`]
- [ ] Uses current session from `FLUXID_SESSION_ID` [Test: with and without env var; missing var yields clear error]
- [ ] No file persistence or side effects outside memory [Test: verify nothing written to disk; new session starts empty]
- [ ] CLI returns non-zero only for unrecoverable CLI errors, not to abort wrapper [Test: simulate failure conditions; wrapper continues]
- [ ] Output contains no extraneous formatting beyond confirmation [Test: compare exact output against spec]

## Dependencies
- External: Active session context via `FLUXID_SESSION_ID` (from earlier milestones m01–m03)
- Internal: None

## E2E Test Mapping
**Test File**: `m04_e01_t01_write_cli_flag.yaml`

**Test Flow**:
1. Set `FLUXID_SESSION_ID` to a new value
2. Run `fluxid --write-history "First note"`
3. Assert confirmation output and zero exit code
4. Optionally run `fluxid ipc view-history` (after e03) and verify line `[timestamp] First note`

**Key Assertions**:
- Confirmation message present and well-formed
- Timestamp matches ISO 8601 with Z suffix
- Entry appears in session history when viewed later

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Timestamp must be system-generated only. Do not accept user-provided timestamps.
- Messages may include UTF-8 characters; treat size as UTF-8 bytes (see m04-e04 for eviction behavior).