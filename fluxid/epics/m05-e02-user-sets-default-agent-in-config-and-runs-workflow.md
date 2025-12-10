---
id: m05-e02
title: User sets default agent in config and runs workflow
milestone: m05
status: pending
patterns: []
---

# Epic: User sets default agent in config and runs workflow

## Overview
User defines `agent: claude|codex|opencode` in `~/.fluxid/config.yaml` or project `./.fluxid/config.yaml` → System loads configs and applies precedence (CLI > env > project > home > default) → System resolves the selected agent binary from PATH and validates executable → System shows initialization status indicating agent and source=config (with file path) → System spawns the agent under standard orchestration and streaming → Workflow completes successfully.

## Scope
- User actions: Add `agent:` field to home and/or project config; run `fluxid` without CLI agent flags.
- System responses: Load configs; compute precedence; select agent from highest‑priority config present; resolve binary via PATH; verify executable; display initialization with `source=config` and file origin; run implement→commit→review orchestration; propagate session/env; exit 0 on success.
- Screens/states: Initialization status highlighting agent and selection source; normal run states and completion.

## Success Criteria
- [ ] Precedence: CLI > env > project > home > default(claude) [Test: prepare conflicting sources; assert the correct winner]
- [ ] Reads `agent:` from both home and project config files [Test: presence/absence permutations; verify selection]
- [ ] Initialization status shows `source=config` and indicates which file path won [Test: parse output; path displayed]
- [ ] PATH resolution and executable validation succeed or fail with clear errors [Test: missing binary yields actionable message]
- [ ] Orchestration identical to baseline behavior [Test: phase order and counts]
- [ ] No config writes occur during selection [Test: filesystem invariance]

## Dependencies
m02-e01, m02-e02, m02-e05, m01-e01

## E2E Test Mapping
**Test File**: `m05_e02_t01_config-agent-selection.yaml`

**Test Flow**:
1. Create home and project configs with different `agent:` values
2. Run `fluxid` inside project without CLI flags
3. Confirm initialization shows the project value as winner and source=config
4. Verify end‑to‑end run exits 0 under the selected agent

**Key Assertions**:
- Correct precedence resolution among sources
- Display of file origin for config winner
- Successful PATH resolution or clear error

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Validate `agent:` allowed values strictly; fail fast on invalid entries with guidance.