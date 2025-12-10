---
id: m06-e02
title: User gets initialization status in JSON
milestone: m06
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User gets initialization status in JSON

## Overview
User runs `fluxid --fluxid-output json --[agent] [args]` to retrieve the initialization status in machine-readable JSON without altering agent output. System prints a structured JSON object covering configuration and resolved state fields, suitable for automation parsing.

## Scope
- User actions: Run CLI with `--fluxid-output json`
- System responses: Serialize initialization status to JSON; print to stdout; preserve normal agent output behavior when not in dry-run
- Screens/states: Terminal output (JSON payload for status; agent output unchanged)

## Success Criteria
- [ ] Default format remains human-readable when flag omitted [Test: invoking without `--fluxid-output` produces text]
- [ ] JSON format outputs a valid JSON object [Test: `jq` parses; schema keys present]
- [ ] Output includes session ID, agent, loop counts, command file paths, phase toggles [Test: keys exist and values reflect current config]
- [ ] Unknown format values are rejected with clear message [Test: `--fluxid-output xml` exits non-zero with guidance]
- [ ] Agent output remains unchanged when not in dry-run [Test: start a trivial run and confirm agent stream unaffected]

## Dependencies
None

## E2E Test Mapping
**Test File**: `m06_e02_t01_output-json-initialization-status.yaml`

**Test Flow**:
1. Invoke CLI with `--fluxid-output json` and minimal config
2. Capture initialization status output
3. Validate JSON parses and contains required keys
4. If executing non-dry-run, verify agent output stream is unaffected by the flag

**Key Assertions**:
- JSON parses without error
- Required fields present with correct values
- Unknown format triggers error and non-zero exit
- Agent output behavior unchanged in non-dry-run

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Scope limited to initialization status serialization; do not alter agent I/O semantics.
