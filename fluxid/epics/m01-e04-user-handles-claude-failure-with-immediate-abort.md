---
id: m01-e04
title: User encounters Claude failure and wrapper aborts immediately
milestone: m01
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User encounters Claude failure and wrapper aborts immediately

## Overview
User runs `fluxid --claude` → Claude exits non-zero during a phase → System detects failure → Wrapper aborts workflow instantly, mirrors Claude's exit code, and displays clear error status → Command terminates without further loops.

## Scope
- User actions: Run CLI; observe failure; note exit code and message
- System responses: Monitor child exit status; stop orchestration immediately on non-zero; print concise error explanation; propagate same exit code; avoid partial state
- Screens/states: Running; error status; terminated state

## Success Criteria
- [ ] Non-zero Claude exit is detected and triggers immediate abort [Test: stub Claude exit 2; assert no further phases run]
- [ ] Wrapper returns the same exit code as Claude [Test: exit code equality]
- [ ] Error message explains what happened and next steps [Test: presence of agent failure summary]
- [ ] Streams flush gracefully on abort [Test: no hanging pipes; process exits promptly]
- [ ] No residual state or orphaned child processes [Test: child process cleanup]

## Dependencies
- Depends on `m01-e01` (base orchestration and process management)
- External prerequisite: Claude CLI available on PATH

## E2E Test Mapping
**Test File**: `m01_e04_t01_user-handles-claude-failure.yaml`

**Test Flow**:
1. Invoke `fluxid --claude` with a stubbed Claude that exits non-zero mid-run
2. Verify immediate abort message and mirrored exit code
3. Confirm no further phases or iterations execute
4. Assert process terminates cleanly

**Key Assertions**:
- Abort occurs instantly on failure
- Exit code matches
- No further loop activity
- Clean termination

## Completion Checklist
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Maintain user trust with clear, actionable failure messaging