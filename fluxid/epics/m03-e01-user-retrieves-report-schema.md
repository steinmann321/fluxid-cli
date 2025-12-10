---
id: m03-e01
title: User retrieves report schema
milestone: m03
status: pending
patterns: []
---

# Epic: User retrieves report schema

## Overview
User runs `fluxid ipc get-report-schema` to view the exact YAML schema required for IPC reports, confirming required fields and constraints to author valid reports.

## Scope
- User actions: Run `fluxid ipc get-report-schema`
- System responses: Output the full YAML schema to stdout; print actionable error if retrieval fails
- Screens/states: CLI stdout/stderr; no interactive prompts

## Success Criteria
- [ ] Full schema prints to stdout without truncation [Test: parse output as YAML; verify presence of required fields/constraints]
- [ ] Required fields and enums clearly identifiable [Test: assert keys `command`, `artifact`, `timestamp`, `status`, `issues`; `status` enum contains `PASS`, `FAIL`]
- [ ] Usage help available on `--help` [Test: `fluxid ipc get-report-schema --help` shows description and examples]

## Dependencies
None

## E2E Test Mapping
**Test File**: `m03_e01_t01_report-schema.yaml`

**Test Flow**:
1. Run `fluxid ipc get-report-schema`
2. Capture stdout and parse as YAML
3. Verify required keys and `status` enum values
4. Verify output is valid YAML and non-empty

**Key Assertions**:
- Schema YAML is syntactically valid
- Required keys exist; `status` enum is `PASS|FAIL`
- Command exits cleanly and prints to stdout

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Ensure schema matches milestone constraints exactly (keys, types, enums, categories).