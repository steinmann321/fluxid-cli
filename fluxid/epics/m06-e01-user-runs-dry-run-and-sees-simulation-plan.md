---
id: m06-e01
title: User runs dry-run and sees simulation plan
milestone: m06
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User runs dry-run and sees simulation plan

## Overview
User runs `fluxid --fluxid-dry-run --[agent] [args]` to simulate a workflow without invoking any agent. System resolves and validates configuration, prints the full execution plan (iterations, retries, phases) with the command file that would be used per phase, generates synthetic PASS reports to simulate loop progression, skips agent process spawning entirely, and exits successfully.

## Scope
- User actions: Run CLI with `--fluxid-dry-run` and optional configuration/flags
- System responses: Resolve config; validate values and file paths; print simulated execution plan with phase order and command files; generate synthetic PASS reports; skip agent spawn; exit 0
- Screens/states: Terminal output in simulation mode; initialization status; simulated loop state

## Success Criteria
- [ ] Dry-run flag activates simulation mode [Test: `--fluxid-dry-run` prints a header indicating simulation mode]
- [ ] Configuration is fully resolved and validated before simulation [Test: invalid values cause clear errors before plan prints]
- [ ] Command file for each phase is shown in the plan [Test: output includes file paths per phase]
- [ ] Plan lines include iteration, retry, and phase [Test: match lines like `Would execute: Iteration 1, Retry 1, Phase: IMPLEMENT`]
- [ ] Agent process is not spawned in dry-run [Test: assert no external process is invoked; no agent I/O markers appear]
- [ ] Synthetic PASS reports drive loop progression [Test: loop advances until completion using generated PASS results]
- [ ] Process exits with status code 0 on successful simulation [Test: shell exit code equals 0]

## Dependencies
None

## E2E Test Mapping
**Test File**: `m06_e01_t01_dry-run-simulation-plan.yaml`

**Test Flow**:
1. Invoke CLI with `--fluxid-dry-run` and minimal valid config
2. Verify printed initialization summary and simulation header
3. Verify plan lines for each phase with iteration/retry and command files
4. Confirm exit code 0 and no agent process was started

**Key Assertions**:
- Output contains "Would execute:" lines with correct iteration/retry/phase
- Output lists command file path per phase
- No agent output markers appear
- Exit code 0 on success

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Ensure dry-run path shares configuration resolution/validation with normal execution, then diverges before agent spawn.
