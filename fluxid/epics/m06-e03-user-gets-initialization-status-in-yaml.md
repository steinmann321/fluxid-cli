---
id: m06-e03
title: User gets initialization status in YAML
milestone: m06
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User gets initialization status in YAML

## Overview
User runs `fluxid --fluxid-output yaml --[agent] [args]` to retrieve the initialization status as YAML for tooling that prefers YAML. System prints a structured YAML document representing configuration and resolved state fields; agent output remains unchanged.

## Scope
- User actions: Run CLI with `--fluxid-output yaml`
- System responses: Serialize initialization status to YAML; print to stdout; preserve agent output behavior when not in dry-run
- Screens/states: Terminal output (YAML payload for status; agent output unchanged)

## Success Criteria
- [ ] Default format is human-readable when flag omitted [Test: no flag → text]
- [ ] YAML format outputs valid YAML [Test: `yq` parses; schema keys present]
- [ ] Includes session ID, agent, loop counts, command file paths, phase toggles [Test: keys/paths align with current config]
- [ ] Unknown format values are rejected with clear message [Test: `--fluxid-output toml` errors clearly]
- [ ] Agent output remains unchanged in non-dry-run [Test: confirm agent stream unaffected]

## Dependencies
None

## E2E Test Mapping
**Test File**: `m06_e03_t01_output-yaml-initialization-status.yaml`

**Test Flow**:
1. Invoke CLI with `--fluxid-output yaml`
2. Capture initialization status
3. Validate YAML parses and required keys exist
4. If non-dry-run, verify agent output unaffected by the flag

**Key Assertions**:
- YAML parses without error
- Required fields present with correct values
- Unknown format triggers informative error and non-zero exit
- Agent output behavior unchanged in non-dry-run

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Use a consistent field set and ordering across JSON/YAML to simplify tooling.
