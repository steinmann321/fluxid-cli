---
id: m03-e02
title: User writes a valid report and verifies it
milestone: m03
status: pending
patterns: []
---

# Epic: User writes a valid report and verifies it

## Overview
User pipes a compliant YAML report into `fluxid ipc write-report`, receives success feedback, and confirms persistence by reading it back via `fluxid ipc read-report` within the same session.

## Scope
- User actions: Prepare YAML; run `fluxid ipc write-report < report.yaml`; run `fluxid ipc read-report`
- System responses: Validate report against schema; store per `FLUXID_SESSION_ID`; output report on read
- Screens/states: CLI stdout/stderr; session context via env var or `--session`

## Success Criteria
- [ ] Valid report is accepted and stored for the session [Test: write then read; deep-compare YAML]
- [ ] Command exits with zero status and concise success message [Test: capture exit code and stdout]
- [ ] Session override via `--session` works [Test: write with explicit session; read back matches]

## Dependencies
m03-e01-user-retrieves-report-schema

## E2E Test Mapping
**Test File**: `m03_e02_t01_write-and-read-valid-report.yaml`

**Test Flow**:
1. Export `FLUXID_SESSION_ID` and run `fluxid ipc write-report < valid.yaml`
2. Run `fluxid ipc read-report`
3. Parse stdout and compare with `valid.yaml`
4. Assert exit codes are zero

**Key Assertions**:
- Report round-trips without mutation
- Exit codes indicate success
- Session scoping is respected

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Store in-memory keyed by session; no persistence beyond process lifetime.