---
id: m02-e02
title: User applies project overrides and sees precedence in status
milestone: m02
status: pending
patterns: []
---

# Epic: User applies project overrides and sees precedence in status

## Overview
User adds `./.fluxid/config.yaml` in a project to override home defaults. On running fluxid within that project, the system loads both home and project config, resolves precedence (project over home over defaults), validates values, and shows an initialization status that clearly indicates which source won for each field.

## Scope
- User actions: Create `~/.fluxid/config.yaml`; create project `./.fluxid/config.yaml` overriding at least one field; run `fluxid` inside the project.
- System responses: Read both YAML files; merge with default precedence; validate values; compute resolved config; print initialization status with winning source per field.
- Screens/states: Terminal initialization display highlighting precedence outcomes for each key.

## Success Criteria
- [ ] Detects and reads `./.fluxid/config.yaml` when present [Test: run inside/outside project; verify project config only applied inside project]
- [ ] Applies precedence: project > home > defaults [Test: set different values in home and project; verify project wins]
- [ ] Initialization status shows, per field, the winning source [Test: parse output lines for each key with source=project/home/default]
- [ ] Gracefully handles missing home config [Test: only project present; still works and indicates sources]
- [ ] No side effects outside project directory [Test: ensure no writes under home or project during resolution]

## Dependencies
m02-e01

## E2E Test Mapping
**Test File**: `m02_e02_t01_project-overrides-precedence.yaml`

**Test Flow**:
1. Create home config with baseline values
2. Create project config overriding specific fields
3. Run `fluxid` inside project
4. Confirm initialization status reflects project values and reports sources

**Key Assertions**:
- Correct values match project config
- Every field shows its source as project or home/default as appropriate
- No errors when home missing

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Avoid leaking absolute home path in errors unless necessary; keep messages clear and actionable.