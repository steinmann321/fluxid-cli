---
id: m02-e01
title: User configures home defaults and verifies initialization status
milestone: m02
status: pending
patterns: []
---

# Epic: User configures home defaults and verifies initialization status

## Overview
User creates `~/.fluxid/config.yaml` with default settings and runs fluxid; on startup, the system parses YAML, validates schema, merges with built-in defaults, and displays an initialization status that shows resolved values and their sources (defaults vs. home), confirming home-level configuration is active.

## Scope
- User actions: Create `~/.fluxid/config.yaml`; run `fluxid` from any project directory without a project config; observe initialization output.
- System responses: Read YAML; apply defaults; validate schema and value ranges; resolve effective configuration; print initialization status with each field's resolved value and source.
- Screens/states: Terminal startup output including a distinct "Initialization Status" section prior to any workflow execution.

## Success Criteria
- [ ] Reads `~/.fluxid/config.yaml` when present and applies values [Test: run with home config present/absent; verify values change only when file exists]
- [ ] Validates schema with sensible defaults when keys omitted [Test: omit optional keys; verify defaults: agent=claude, implement_retries=3, iterations=20, commit_enabled=true]
- [ ] Rejects invalid types/structures with clear message [Test: set `implement_retries: "three"`; expect immediate failure with field/type and location reference]
- [ ] Displays initialization status with resolved values and sources [Test: assert output lists each key with value and source="home" or "default"]
- [ ] Does not touch project state when only home config is used [Test: run from clean project; ensure no files created/modified in CWD]

## Dependencies
None

## E2E Test Mapping
**Test File**: `m02_e01_t01_home-config-initialization-status.yaml`

**Test Flow**:
1. Ensure no project config exists; create `~/.fluxid/config.yaml` with sample values
2. Run `fluxid`
3. Verify initialization status prints resolved values and sources
4. Remove/rename home config and rerun to confirm fallback to defaults

**Key Assertions**:
- Initialization section appears before workflow steps
- Agent, retries, iterations, commit_enabled shown with correct values
- Each displayed field includes source indicator: default vs. home

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Home config path expansion must support `~` and absolute path forms.
- Printing absolute paths for command files is out of scope here; covered in command file resolution epic.