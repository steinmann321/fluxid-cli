---
id: m05-e05
title: User receives clear errors for conflicts and missing agents
milestone: m05
status: pending
patterns: []
---

# Epic: User receives clear errors for conflicts and missing agents

## Overview
User runs `fluxid` with conflicting agent selections (e.g., multiple flags) or with a selection whose binary is not on PATH → System detects the issue before orchestration, emits a clear actionable error (e.g., "Multiple agents specified" or "Agent 'codex' not found on PATH") including brief remediation guidance → Process exits non‑zero without partial execution.

## Scope
- User actions: Invoke CLI with conflicting flags; set invalid `FLUXID_AGENT`; configure unsupported agent value; run with agent not installed on PATH.
- System responses: Validate mutual exclusion and allowed values; if multiple sources specify agents simultaneously, apply precedence while still warning when multiple CLI flags are present; verify PATH presence and executability; emit precise error messages; exit with non‑zero code and no orchestration.
- Screens/states: Error output in terminal; immediate termination state.

## Success Criteria
- [ ] Conflicting CLI flags result in error and exit >0 [Test: `--claude --codex`; expect "Multiple agents specified" and exit code 2]
- [ ] Unsupported value in env/config yields validation error [Test: `FLUXID_AGENT=foo`; config `agent: foo`]
- [ ] Missing or non‑executable binary yields PATH error [Test: remove from PATH or `chmod -x`; verify message]
- [ ] No child process is spawned on error conditions [Test: stub agent logs; ensure not invoked]
- [ ] Messages are concise and actionable [Test: include hints like `which <agent>`]

## Dependencies
m05-e01, m05-e02, m05-e03

## E2E Test Mapping
**Test File**: `m05_e05_t01_agent-selection-errors.yaml`

**Test Flow**:
1. Run with `--claude --codex`; verify error message and exit code >0
2. Export `FLUXID_AGENT=foo`; run; verify validation error and no spawn
3. Ensure `opencode` missing from PATH; select it; verify PATH error and no spawn

**Key Assertions**:
- Correct error text for each scenario
- Non‑zero exit codes
- No orchestration side effects

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Keep precedence logic deterministic and documented in error contexts when helpful.