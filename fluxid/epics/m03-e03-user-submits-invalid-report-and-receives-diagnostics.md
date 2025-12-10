---
id: m03-e03
title: User submits invalid report and receives diagnostics
milestone: m03
status: pending
patterns: []
---

# Epic: User submits invalid report and receives diagnostics

## Overview
User attempts to write a malformed or schema-violating YAML report via `fluxid ipc write-report`; system rejects it with precise validation diagnostics that explain what failed and how to fix.

## Scope
- User actions: Pipe invalid YAML or schema-breaking content into `fluxid ipc write-report`
- System responses: Detect and print detailed validation errors; non-zero exit code; do not store report
- Screens/states: CLI stdout/stderr; no stored state on failure

## Success Criteria
- [ ] Validation failures enumerate mismatches and missing fields [Test: multiple invalid samples covering format, enums, categories]
- [ ] Clear guidance on remediation is included [Test: messages reference required keys and constraints]
- [ ] No report stored on failure [Test: `read-report` after failure returns previous or empty]

## Dependencies
m03-e01-user-retrieves-report-schema

## E2E Test Mapping
**Test File**: `m03_e03_t01_invalid-report-diagnostics.yaml`

**Test Flow**:
1. Run `fluxid ipc write-report < invalid.yaml`
2. Capture stderr and exit code
3. Run `fluxid ipc read-report`
4. Verify no new report is present

**Key Assertions**:
- Non-zero exit code on invalid input
- Error text lists exact issues and fix guidance
- Report state remains unchanged

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Cover all constraints: ISO timestamp, single-token artifact, PASS/FAIL enum, issue category completeness.