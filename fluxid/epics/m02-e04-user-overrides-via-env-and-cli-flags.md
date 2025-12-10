---
id: m02-e04
title: User overrides via environment and CLI flags
milestone: m02
status: pending
patterns: []
---

# Epic: User overrides via environment and CLI flags

## Overview
User sets environment variables and/or passes CLI flags to override configuration at runtime. On startup, the system applies precedence (defaults → home → project → env → CLI), validates values and ranges, and reports the final resolved configuration with the correct winning source for each field.

## Scope
- User actions: Export env vars (e.g., `FLUXID_AGENT`, `FLUXID_ITERATIONS`, `FLUXID_IMPLEMENT_RETRIES`, `FLUXID_COMMIT_ENABLED`); run `fluxid` with flags `--fluxid-implement-retries`, `--fluxid-iterations`, and `--fluxid-no-commit`.
- System responses: Parse env and flags; merge using defined precedence; respect data types and constraints; include in initialization status the source for each field.
- Screens/states: Terminal output showing initialization status with source annotations including `env` and `cli` as applicable.

## Success Criteria
- [ ] Environment variables override file-based configs [Test: set env vars; verify values supersede project/home configs]
- [ ] CLI flags override environment variables [Test: set env var and flag; flag wins]
- [ ] Type and range validation enforced for env/CLI inputs [Test: negative iterations or retries rejected with clear error]
- [ ] `--fluxid-no-commit` maps to `commit_enabled=false` [Test: pass flag; status shows commit_enabled=false with source=cli]
- [ ] Unrecognized flags/envs do not alter config and produce helpful guidance [Test: set `FLUXID_UNKNOWN`; ensure ignored with a warning]

## Dependencies
m02-e01, m02-e02

## E2E Test Mapping
**Test File**: `m02_e04_t01_env-and-cli-precedence.yaml`

**Test Flow**:
1. Define conflicting values in home/project
2. Set env vars overriding those
3. Run with flags overriding env
4. Verify final values and sources reflect CLI dominance

**Key Assertions**:
- Precedence order strictly maintained
- Error messages for invalid values identify the input source (env or cli)
- `commit_enabled` correctly toggled by `--fluxid-no-commit`

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Ensure flags are documented in `--help` output with clear descriptions and defaults.