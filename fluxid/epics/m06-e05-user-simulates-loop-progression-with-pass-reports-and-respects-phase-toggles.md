---
id: m06-e05
title: User simulates loop progression with PASS reports and respects phase toggles
milestone: m06
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User simulates loop progression with PASS reports and respects phase togggles

## Overview
User runs `fluxid --fluxid-dry-run` with specific loop counts and phase toggles (e.g., `--fluxid-no-commit`). System uses synthetic PASS reports to advance through iterations and retries, prints a plan that respects disabled phases, and completes the simulated loop accordingly.

## Scope
- User actions: Run dry-run with loop count/retry flags and phase toggle flags
- System responses: Generate PASS reports; iterate through configured counts; omit disabled phases from plan; finish when loop criteria met
- Screens/states: Terminal simulation output reflecting loop and toggles

## Success Criteria
- [ ] Loop counts (iterations/retries) are honored in simulation [Test: number of plan entries matches configured counts]
- [ ] Disabled phases do not appear in the plan [Test: `--fluxid-no-commit` removes COMMIT lines]
- [ ] Synthetic PASS outcomes drive advancement until termination [Test: loop ends when success criteria met]
- [ ] Summary indicates completion state consistent with settings [Test: final line shows simulated completion]
- [ ] Exit status is 0 on successful simulated run [Test: shell exit code 0]

## Dependencies
m06-e01

## E2E Test Mapping
**Test File**: `m06_e05_t01_dry-run-loop-progression-with-toggles.yaml`

**Test Flow**:
1. Invoke dry-run with explicit iteration/retry counts and `--fluxid-no-commit`
2. Verify plan prints only enabled phases per iteration/retry
3. Verify synthetic PASS reports advance loop and finish
4. Confirm exit code 0

**Key Assertions**:
- Plan count equals configured iterations × enabled phases
- No disabled phase lines appear
- Final completion message present
- Exit code 0

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Ensure phase toggles are enforced consistently across validation and plan printing.
