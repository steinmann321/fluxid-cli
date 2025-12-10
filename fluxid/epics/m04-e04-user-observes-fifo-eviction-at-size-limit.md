---
id: m04-e04
title: User observes FIFO eviction at size limit
milestone: m04
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User observes FIFO eviction at size limit

## Overview
User writes enough history to exceed the 32MB per-session limit → System tracks cumulative UTF-8 byte size and evicts oldest entries (FIFO) as needed to admit the new write → User views history and confirms that older entries were dropped while newer entries persist.

## Scope
- User actions: Perform repeated `--write-history` or `ipc write-history` with large messages; then run `ipc view-history`
- System responses: Track UTF-8 byte size; evict from oldest forward until under 32MB; maintain chronological order of remaining entries
- Screens/states: Terminal outputs for writes and view; in-memory session buffer with eviction

## Success Criteria
- [ ] Tracks size in UTF-8 bytes, not characters [Test: write multibyte content; confirm accurate byte accounting]
- [ ] Applies FIFO eviction only when write would exceed 32MB [Test: boundary write just under and just over limit]
- [ ] Preserves chronological order after eviction [Test: verify first retained entry is correct]
- [ ] No partial line corruption or split multibyte sequences [Test: verify full lines intact after eviction]
- [ ] Performance remains acceptable during eviction [Test: stress scenario with many entries]

## Dependencies
- External: Active session context via `FLUXID_SESSION_ID` (from earlier milestones m01–m03)
- Internal: Ability to write entries (m04-e01 and/or m04-e02); ability to view entries (m04-e03)

## E2E Test Mapping
**Test File**: `m04_e04_t01_fifo_eviction.yaml`

**Test Flow**:
1. Start session; write a sequence of identifiable entries totaling >32MB
2. Run `fluxid ipc view-history`
3. Verify that earliest entries were dropped until size ≤32MB
4. Confirm newest entries remain and ordering is intact

**Key Assertions**:
- Byte-accurate limit enforcement
- Correct FIFO removal behavior
- Stable formatting of remaining lines

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Eviction should be triggered atomically per write to avoid interleaved inconsistent states under concurrency.