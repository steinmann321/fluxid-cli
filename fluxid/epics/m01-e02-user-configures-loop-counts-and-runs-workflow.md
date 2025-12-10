---
id: m01-e02
title: User configures loop counts and runs workflow successfully
milestone: m01
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User configures loop counts and runs workflow successfully

## Overview
User invokes `fluxid --claude --fluxid-iterations N --fluxid-implement-retries R` → System validates positive integers (≥1) → System displays applied configuration with session UUID → System runs loops using overrides → Workflow completes → Command exits with success.

## Scope
- User actions: Run CLI with `--fluxid-iterations N` and `--fluxid-implement-retries R`; observe status; complete run
- System responses: Parse flags; validate N,R ≥ 1; reject invalid values with clear error; display initialization showing overrides; orchestrate loops using provided counts; stream output; exit 0 on successful completion
- Screens/states: Terminal error message for invalid input; initialization status with overrides; running progress; completion summary

## Success Criteria
- [ ] CLI flags override defaults for iterations and implement retries [Test: run with N=5, R=2; assert applied counts in status]
- [ ] Invalid counts (0, negative, non-integer) are rejected with clear message and non-zero exit [Test: N=0, N=-1, N="abc"; assert validation error and exit code ≠ 0]
- [ ] Initialization reflects overrides (agent, session ID, counts) [Test: parse stdout for explicit N,R]
- [ ] Loops execute according to overrides [Test: instrument phases; assert exact iteration/retry counts]
- [ ] Successful completion exits 0 [Test: exit code assertion]

## Dependencies
- Depends on `m01-e01` (base orchestration and default-run path)
- External prerequisite: Claude CLI available on PATH

## E2E Test Mapping
**Test File**: `m01_e02_t01_user-configures-loop-counts.yaml`

**Test Flow**:
1. Invoke `fluxid --claude --fluxid-iterations 5 --fluxid-implement-retries 2`
2. Verify initialization displays overrides and session ID
3. Observe run and verify counts used
4. Assert exit code 0

**Key Assertions**:
- Overrides appear in status output
- Exact loop counts honored
- Exit code 0

## Completion Checklist
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Keep error messages actionable: show invalid value and expected format (positive integer)