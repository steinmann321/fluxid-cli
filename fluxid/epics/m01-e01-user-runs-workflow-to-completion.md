---
id: m01-e01
title: User runs automated implement-commit-review loops to completion
milestone: m01
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User runs automated implement-commit-review loops to completion

## Overview
User invokes `fluxid --claude [claude-args]` → System validates defaults and generates a session UUID → System displays initialization status → System executes nested implement→commit→review loops → Streams pass through in real time → Workflow completes all iterations → Command exits with success.

## Scope
- User actions: Run CLI with `--claude` and optional Claude args; observe terminal output until completion
- System responses: Parse args; apply default loop counts; validate counters; generate UUID v4; display initialization (agent, session ID, iteration/retry counts); orchestrate nested loops; invoke Claude phases with default prompts; pipe stdout/stderr/stdin; propagate `FLUXID_SESSION_ID`; exit 0 when loops finish
- Screens/states: Terminal initialization status; running loop progress; phase transitions; completion summary

## Success Criteria
- [ ] Users can start automation with defaults and complete all loops [Test: run `fluxid --claude` with stubbed Claude success; verify exit code 0]
- [ ] Unique UUID v4 session ID is generated per run [Test: format validation, uniqueness across consecutive runs]
- [ ] Initialization status shows loop counts, session ID, agent selection [Test: parse stdout lines for required fields]
- [ ] Nested loops execute in correct order and counts [Test: instrument phases; assert 20 review cycles max, 3 implement retries max]
- [ ] Claude phases are invoked with correct default prompts [Test: capture child process args/env; verify prompt selection]
- [ ] Stdout/stderr/stdin are piped with low latency [Test: measure output timestamps; backpressure handling under burst output]
- [ ] `FLUXID_SESSION_ID` is propagated to child processes [Test: Claude echoes env; verify presence and value]
- [ ] Completion summary appears and process exits 0 [Test: last lines include completion; exit code assertion]

## Dependencies
- None (external prerequisite: Claude CLI available on PATH)

## E2E Test Mapping
**Test File**: `m01_e01_t01_user-runs-workflow-to-completion.yaml`

**Test Flow**:
1. Invoke `fluxid --claude` with a Claude stub that returns success
2. Verify initialization lines (agent, session ID, loop counts)
3. Observe loop progress and phase transitions until completion
4. Assert exit code 0 and presence of completion summary

**Key Assertions**:
- Initialization shows session ID (UUID v4) and counts
- Phases invoked in order with defaults
- Real-time output visible throughout run
- Exit code 0 at end

## Completion Checklist
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- No configuration files required; default prompts are built-in
- Focus on end-to-end orchestration correctness and observable UX