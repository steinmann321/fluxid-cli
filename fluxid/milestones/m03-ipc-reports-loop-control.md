---
id: m03
title: Users can exchange structured reports via IPC to control loop flow
status: pending
---

# Milestone: Users can exchange structured reports via IPC to control loop flow

## Deliverable
External processes (typically called by the agent during workflow phases) can write structured YAML reports via `fluxid ipc write-report`, and fluxid uses these reports to control loop progression (PASS status breaks loops, FAIL status continues). Reports are validated against schema, and invalid/missing reports trigger infinite retry of the phase until a valid report is produced. Users can escape infinite retries via Ctrl+C (SIGINT/SIGTERM) or programmatic abort via `fluxid ipc abort`.

**What users can do:**
- Call `fluxid ipc write-report < report.yaml` from external processes during workflow execution
- Generate reports following schema (status: PASS/FAIL, categorized issues, metadata)
- Control loop behavior through report status: PASS breaks retry/iteration loops, FAIL continues
- Retrieve report schema via `fluxid ipc get-report-schema` to understand required structure
- Read current report via `fluxid ipc read-report` for decision-making in later phases
- Rely on automatic schema validation with clear error messages on validation failure
- Handle infinite retry scenarios when prompts fail to produce valid reports
- Abort runaway workflows via Ctrl+C or `fluxid ipc abort` command

## Success Criteria
- [ ] IPC subcommand interface implemented:
  - [ ] `fluxid ipc write-report` - reads YAML from stdin, validates, stores for session
  - [ ] `fluxid ipc read-report` - outputs current session report to stdout
  - [ ] `fluxid ipc get-report-schema` - outputs report schema YAML to stdout
  - [ ] `fluxid ipc abort [--session <id>]` - triggers graceful workflow abort
- [ ] Report schema validation (based on bash implementation requirements):
  - [ ] Required fields: `command`, `artifact`, `timestamp`, `status`, `issues`
  - [ ] `status` enum: `PASS` or `FAIL`
  - [ ] `artifact` must be single token (no `/` or `.` characters)
  - [ ] `timestamp` must be ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)
  - [ ] `issues` must contain all 5 categories: `blockers`, `defects`, `concerns`, `observations`, `enhancements`
  - [ ] Each category is array of issue objects
  - [ ] Each issue requires `message` field
  - [ ] Optional fields: `next_steps` (array), `summary` (string)
- [ ] Validation failure prints clear diagnostic: what failed, why, how to fix (schema mismatch details)
- [ ] Loop control based on reports:
  - [ ] After implement phase: if report status == PASS, break retry loop and proceed to review
  - [ ] After review phase: if report status == PASS, exit iteration loop (workflow complete)
  - [ ] If report status == FAIL, continue to next retry/iteration
  - [ ] If report invalid/missing: infinite retry of that phase (don't increment counter)
- [ ] Session context inherited from `FLUXID_SESSION_ID` environment variable (no manual session management)
- [ ] IPC commands support explicit `--session <id>` override when needed
- [ ] Graceful abort mechanism:
  - [ ] Ctrl+C (SIGINT/SIGTERM) triggers abort flag for session
  - [ ] `fluxid ipc abort` sets abort flag programmatically
  - [ ] Allows current agent invocation to complete
  - [ ] Exits cleanly after phase completion
  - [ ] Multiple abort signals force immediate exit
- [ ] IPC command failures print errors to stderr but don't abort wrapper
- [ ] Reports stored in-memory per session (map: session ID → report)
- [ ] Complete UI: clear validation error messages, abort confirmation
- [ ] Full backend: YAML validation, in-memory storage, loop control logic, signal handling
- [ ] Can be deployed independently: works with m01 loop + m02 config
- [ ] Requires no additional milestones: full report-driven workflow control

## Validation Questions
**Before marking this milestone complete, answer:**
- [ ] Can a real user perform complete report-driven workflows with only this milestone? **YES** - external processes can write reports that control loop behavior end-to-end
- [ ] Is it polished enough to ship publicly? **YES** - clear validation errors, graceful abort, helpful schema diagnostics
- [ ] Does it solve a real problem end-to-end? **YES** - enables programmatic workflow control based on test results and validation
- [ ] Does it include both complete UI and functional backend integration? **YES** - IPC CLI commands (UI) + validation/storage/loop-control (backend)
- [ ] Can it run independently without waiting for other milestones? **YES** - builds on m01+m02, adds IPC report layer
- [ ] Would you personally use this if it were released today? **YES** - critical for structured feedback and automated decision-making

## Vertical Slice - All Layers Included
This milestone includes:
- **IPC Command Interface**: Subcommand routing, stdin/stdout handling
- **YAML Schema Validation**: Schema definition, validation library integration, error reporting
- **In-Memory Storage**: Session-to-report mapping, concurrent access handling
- **Loop Control Logic**: Report-driven retry/iteration breaking, infinite retry on invalid reports
- **Signal Handling**: SIGINT/SIGTERM capture, graceful shutdown, abort flag management
- **Abort Command**: Programmatic abort trigger, session targeting, cleanup
- **Session Context**: Automatic session ID inheritance from environment
- **Error Reporting**: Detailed validation failures, actionable error messages

## Notes
- This milestone enables **report-driven workflow orchestration** - the core value proposition
- Command files from m02 should now instruct agents to use `fluxid ipc write-report` (update examples/templates)
- Infinite retry behavior ensures reports are mandatory for workflow correctness
- Abort mechanism provides escape hatch for misconfigured prompts
- Future milestones will add history tracking (m04) and multi-agent support (m05)
- Report schema matches bash implementation requirements for compatibility
