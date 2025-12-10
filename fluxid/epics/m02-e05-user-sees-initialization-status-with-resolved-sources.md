---
id: m02-e05
title: User sees initialization status with resolved sources
milestone: m02
status: pending
patterns: []
---

# Epic: User sees initialization status with resolved sources

## Overview
On running fluxid, the user is presented with a clear initialization status that summarizes all resolved configuration values, their sources (default, home, project, env, cli), and the resolved command file paths. This flow focuses on the presentation and clarity of the status output, enabling users to verify configuration before the workflow proceeds.

## Scope
- User actions: Run `fluxid` with any combination of configuration sources present.
- System responses: Collate resolved configuration and sources; format and print a human-readable status section including command file absolute paths.
- Screens/states: Terminal "Initialization Status" block with tabular or enumerated entries per key and a dedicated "Command Files" subsection.

## Success Criteria
- [ ] Displays each configuration key with its final value and source [Test: matrix of scenarios covering all precedence paths]
- [ ] Shows absolute paths for `implement`, `review`, `commit` command files [Test: presence, correctness, and readability validated]
- [ ] Output is structured and scannable [Test: check headings, spacing, key=value formatting; no ambiguity]
- [ ] Appears consistently before any workflow actions [Test: ensure status prints immediately on startup]

## Dependencies
m02-e01, m02-e02, m02-e03, m02-e04

## E2E Test Mapping
**Test File**: `m02_e05_t01_initialization-status-display.yaml`

**Test Flow**:
1. Prepare mixed configuration sources
2. Run `fluxid`
3. Capture and inspect the initialization status output
4. Confirm clarity, completeness, and correctness

**Key Assertions**:
- All keys present with values and sources
- Command file paths correct and absolute
- Status prints prior to any loop execution

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Keep presentation neutral and non-verbose; aim for operator clarity. Consider color or icons only if supported consistently across environments.