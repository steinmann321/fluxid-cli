---
id: m05-e04
title: User sees selected agent and selection source in initialization status
milestone: m05
status: pending
patterns: []
---

# Epic: User sees selected agent and selection source in initialization status

## Overview
User runs `fluxid` with any selection method (CLI flag, env var, or config) → System computes the winner per precedence and displays an initialization status that clearly shows the selected agent and the source that determined it (CLI/env/config/default) → User can verify at a glance which agent is active before execution proceeds → Workflow continues with standard orchestration.

## Scope
- User actions: Run `fluxid` with different selection methods; observe initialization.
- System responses: Resolve agent with precedence; prepare initialization status including `agent=<name>` and `source=<CLI|env|config|default>` and, for config, which file path; print before orchestration begins.
- Screens/states: Terminal initialization status section; subsequent normal run output.

## Success Criteria
- [ ] Initialization includes agent name and selection source [Test: parse stdout; check both fields]
- [ ] When source=config, the winning file path is displayed [Test: project vs home cases]
- [ ] Formatting is consistent across sources and readable in mono terminal [Test: alignment/labels]
- [ ] Display occurs prior to spawning agent [Test: ensure ordering in logs]

## Dependencies
m02-e05, m05-e01, m05-e02, m05-e03

## E2E Test Mapping
**Test File**: `m05_e04_t01_initialization-status-agent-source.yaml`

**Test Flow**:
1. Prepare three runs using CLI, env, and config selection
2. For each run, capture stdout from start
3. Verify agent and source appear in initialization status before any phase output
4. Confirm formatting and, when config, file path present

**Key Assertions**:
- Presence and correctness of agent and source fields
- Order before child process output
- Consistent formatting

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Keep messages concise; avoid leaking sensitive environment details beyond variable name and value origin.