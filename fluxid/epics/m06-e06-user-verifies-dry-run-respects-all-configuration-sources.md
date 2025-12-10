---
id: m06-e06
title: User verifies dry-run respects all configuration sources
milestone: m06
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User verifies dry-run respects all configuration sources

## Overview
User configures values via config file, environment variables, and CLI flags, then runs `fluxid --fluxid-dry-run`. System resolves configuration with correct precedence, validates the combined result, and prints a simulation plan reflecting the resolved values.

## Scope
- User actions: Provide config file; set env vars; pass CLI flags; run dry-run
- System responses: Resolve values across sources with defined precedence; validate; simulate using resolved configuration
- Screens/states: Terminal output showing effective configuration and matching simulation behavior

## Success Criteria
- [ ] Dry-run reads from file, env, and CLI sources [Test: all sources recognized]
- [ ] Precedence is applied correctly (CLI > env > file) [Test: conflicting values resolve to highest precedence]
- [ ] Simulation reflects the resolved configuration [Test: plan lines and fields match effective values]
- [ ] Clear display or log of effective configuration used [Test: output includes an "Effective configuration" section or equivalent]
- [ ] Exit status 0 on successful simulation [Test: shell exit code 0]

## Dependencies
m06-e01

## E2E Test Mapping
**Test File**: `m06_e06_t01_dry-run-config-sources-precedence.yaml`

**Test Flow**:
1. Prepare conflicting values across file, env, and CLI
2. Run dry-run and capture output
3. Verify effective configuration reflects precedence
4. Verify simulation behavior (counts, toggles, paths) matches effective values

**Key Assertions**:
- Effective config values visible and correct
- Plan matches resolved values
- Exit code 0

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Keep precedence rules explicit and documented in the output to reduce confusion.
