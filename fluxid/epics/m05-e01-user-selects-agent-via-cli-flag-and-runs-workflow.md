---
id: m05-e01
title: User selects agent via CLI flag and runs workflow
milestone: m05
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User selects agent via CLI flag and runs workflow

## Overview
User invokes `fluxid --claude [agent-args]` or `--codex` or `--opencode` → System validates that exactly one agent flag is set → System resolves the corresponding agent binary from PATH and verifies it is executable → System displays initialization status showing selected agent and source=CLI → System spawns the agent using the same implement→commit→review orchestration, with real‑time stdin/stdout/stderr streaming → Workflow completes and exits successfully.

## Scope
- User actions: Run CLI with exactly one of `--claude`, `--codex`, `--opencode` and optional agent‑specific args; observe terminal output to completion.
- System responses: Parse flags; enforce mutual exclusion (exactly one); resolve agent binary via PATH; verify executable; set selection source to CLI; display initialization status; spawn child process with identical orchestration and IPC; forward agent‑specific args unchanged; propagate `FLUXID_SESSION_ID`; exit 0 on success.
- Screens/states: Terminal initialization status (agent name, selection source); loop progress and phase transitions; completion summary.

## Success Criteria
- [ ] Exactly one CLI agent flag is required when any flag is used [Test: run with 0/1/2 flags; only 1 proceeds; 0 falls back to other selection paths; 2+ prints error and exits non‑zero]
- [ ] Agent binary is resolved from PATH and is executable [Test: manipulate PATH; verify success when present and clear error when absent or non‑executable]
- [ ] Initialization status shows `agent` and `source=CLI` [Test: parse stdout; verify fields and values]
- [ ] Orchestration matches baseline behavior (implement→commit→review) [Test: instrument calls; compare sequence and counts vs baseline]
- [ ] Stdout/stderr/stdin stream in real time [Test: timestamp inter‑arrival; ensure no buffering regressions]
- [ ] Agent‑specific args are forwarded unchanged to the child [Test: stub agent echoes argv; verify pass‑through]
- [ ] Process exits 0 on successful end‑to‑end run [Test: assert exit code 0 and completion summary present]

## Dependencies
m01-e01, m01-e03, m02-e05, m03-e04

## E2E Test Mapping
**Test File**: `m05_e01_t01_cli-agent-selection.yaml`

**Test Flow**:
1. Place stub agent binary on PATH (e.g., `codex`) that echoes argv and succeeds
2. Run `fluxid --codex -- [agent-args]`
3. Verify initialization shows agent=codex, source=CLI
4. Observe orchestration and streaming; assert exit code 0

**Key Assertions**:
- PATH resolution selects the intended binary
- Initialization status includes agent and source
- Orchestration order and counts match baseline
- Agent args appear in child argv

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Do not hardcode binary locations; rely solely on PATH.
- Keep error messages actionable (suggest `which <agent>` to verify installation).