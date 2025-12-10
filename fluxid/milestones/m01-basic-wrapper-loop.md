---
id: m01
title: Users can execute automated implement-commit-review loops with Claude
status: pending
---

# Milestone: Users can execute automated implement-commit-review loops with Claude

## Deliverable
Users can run `fluxid --claude [args]` to automatically execute nested loops of implement → commit → review phases with the Claude agent, seeing real-time streamed output and completing multiple development iterations without manual intervention. The tool wraps the Claude CLI, executes configurable loops (default: 3 implement retries per iteration, 20 review iterations), and exits when all loops complete or Claude fails.

**What users can do:**
- Execute automated development workflows with systematic quality control through repeated iterations
- See real-time Claude output streamed directly to their terminal
- Configure loop counts via CLI flags (`--fluxid-implement-retries N`, `--fluxid-iterations N`)
- Get clear initialization status showing configuration and session details
- Rely on automatic workflow abort if Claude exits with error

## Success Criteria
- [ ] Users can invoke `fluxid --claude [claude-args]` to start automated workflow
- [ ] System generates unique UUID v4 session ID on startup
- [ ] System displays initialization status (loop counts, session ID, agent selection)
- [ ] System executes nested loops correctly:
  - [ ] Outer loop (iterations): up to 20 review cycles
  - [ ] Inner loop (implement retries): up to 3 implement→commit sequences per iteration
  - [ ] Each phase invokes Claude with appropriate command prompt
- [ ] Claude's stdout/stderr/stdin streams directly to/from user terminal (real-time passthrough)
- [ ] If Claude exits non-zero, wrapper aborts immediately with same exit code
- [ ] CLI flags override defaults: `--fluxid-implement-retries N`, `--fluxid-iterations N`
- [ ] Loop counters must be positive integers (≥1), validated at startup
- [ ] System propagates session ID to child processes via `FLUXID_SESSION_ID` environment variable
- [ ] Complete UI: terminal-based with clear status messages
- [ ] Full backend: process orchestration, loop control, agent invocation
- [ ] Can be deployed independently: single Go binary, no external dependencies except Claude CLI
- [ ] Requires no additional milestones: users get complete automated workflow execution

## Validation Questions
**Before marking this milestone complete, answer:**
- [ ] Can a real user perform complete automated development workflows with only this milestone? **YES** - they can execute full implement-commit-review loops from start to finish
- [ ] Is it polished enough to ship publicly? **YES** - clear initialization messages, error handling, and real-time output
- [ ] Does it solve a real problem end-to-end? **YES** - eliminates manual intervention between development phases
- [ ] Does it include both complete UI and functional backend integration? **YES** - terminal UI with full loop orchestration backend
- [ ] Can it run independently without waiting for other milestones? **YES** - uses built-in default prompts, no config files required
- [ ] Would you personally use this if it were released today? **YES** - delivers core automation value

## Vertical Slice - All Layers Included
This milestone includes:
- **CLI Interface**: Argument parsing, flag handling, help text
- **Session Management**: UUID generation, environment variable propagation
- **Loop Orchestration**: Nested loop control (iterations × implement retries)
- **Process Management**: Claude subprocess spawning, stream piping, exit code handling
- **Phase Execution**: Implement, commit, review phase invocation with built-in default prompts
- **Error Handling**: Claude failure detection, immediate abort, clean exit
- **Status Display**: Initialization status, loop progress visibility

## Notes
- This milestone uses **built-in default prompts** for implement, commit, and review phases (no config files required)
- Future milestones will add configuration system, IPC communication, and multi-agent support
- Default prompts should instruct Claude to use placeholder mechanisms (later replaced by IPC in m03)
- Focus is on **loop mechanics and process orchestration** - the foundation for all other features
- Claude must be available on PATH (prerequisite, not installed by fluxid)
