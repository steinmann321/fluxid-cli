---
id: m05-e03
title: User selects agent via environment variable and runs workflow
milestone: m05
status: pending
patterns: []
---

# Epic: User selects agent via environment variable and runs workflow

## Overview
User exports `FLUXID_AGENT=claude|codex|opencode` and runs `fluxid` without agent flags → System reads the environment variable, applies precedence (CLI > env > config > default) → System resolves the selected agent binary from PATH and validates executable → System shows initialization status indicating agent and source=env → System spawns the agent with standard orchestration and streaming → Workflow completes successfully.

## Scope
- User actions: Set `FLUXID_AGENT` in shell environment; run `fluxid` with no agent flags.
- System responses: Read env; validate value; select agent per precedence; resolve binary via PATH; verify executable; display initialization with `source=env`; run implement→commit→review orchestration; propagate session/env; exit 0 on success.
- Screens/states: Initialization status including agent and source; normal progress and completion screens.

## Success Criteria
- [ ] Environment variable `FLUXID_AGENT` recognized with allowed values only [Test: invalid value errors with guidance]
- [ ] Precedence places env between CLI and config [Test: conflicting sources resolve deterministically]
- [ ] Initialization status shows `source=env` [Test: parse stdout for field]
- [ ] PATH resolution and exec validation succeed/fail with clear messages [Test: missing binary case]
- [ ] Orchestration parity with baseline [Test: phase order and counts]

## Dependencies
m02-e04, m02-e05, m01-e01

## E2E Test Mapping
**Test File**: `m05_e03_t01_env-agent-selection.yaml`

**Test Flow**:
1. Export `FLUXID_AGENT=opencode`
2. Ensure `opencode` stub is on PATH and executable
3. Run `fluxid` without flags
4. Verify initialization shows agent=opencode, source=env; assert successful run

**Key Assertions**:
- Correct precedence outcome
- Source displayed as env
- Child process spawned with selected agent

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Document shell‑specific export guidance in CLI help text if applicable.