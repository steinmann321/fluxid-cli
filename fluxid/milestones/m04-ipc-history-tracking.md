---
id: m04
title: Users can log and query execution history during workflow runtime
status: pending
---

# Milestone: Users can log and query execution history during workflow runtime

## Deliverable
Users and external processes can log important events during workflow execution via `fluxid ipc write-history <message>` or `fluxid --write-history <message>`, and query the accumulated history via `fluxid ipc view-history`. History entries are stored in-memory as plain text log entries with automatic ISO 8601 timestamps, queryable during workflow execution, and automatically cleared when the session ends. History size is limited to 32MB per session with FIFO eviction.

**What users can do:**
- Log events during workflow execution: `fluxid --write-history "Implemented feature X with approach Y"`
- Query history from external processes: `fluxid ipc view-history`
- See automatic timestamps in ISO 8601 format for all entries
- Track decisions, trade-offs, delegations, and postponed work during implementation
- Use history for debugging when workflow behaves unexpectedly
- Rely on automatic size management (oldest entries dropped when 32MB limit reached)
- Know that history is cleared automatically when workflow completes

## Success Criteria
- [ ] IPC commands for history management:
  - [ ] `fluxid ipc write-history <message>` - appends entry with auto-generated timestamp
  - [ ] `fluxid ipc view-history` - outputs all history entries in chronological order
- [ ] CLI flag for direct history writes: `fluxid --write-history <message>` (adds entry to current session)
- [ ] History format: plain text log entries, one per line
  - [ ] Format: `[ISO8601 timestamp] user message`
  - [ ] Example: `[2025-12-10T14:30:52Z] Implemented authentication with JWT approach`
- [ ] Timestamp generation:
  - [ ] System generates timestamp automatically (never from user input)
  - [ ] ISO 8601 format: YYYY-MM-DDTHH:MM:SSZ
  - [ ] Ensures accurate chronological ordering
- [ ] In-memory storage:
  - [ ] Map: session ID → history log (plain text, newline-separated entries)
  - [ ] No file persistence (cleared on session end)
  - [ ] Concurrent access handling for parallel writes
- [ ] Size management:
  - [ ] Track cumulative UTF-8 byte size of all entries
  - [ ] When write would exceed 32MB, drop oldest entries (FIFO) until space available
  - [ ] No limit on entry count, only total size
- [ ] Session context: inherits `FLUXID_SESSION_ID` from environment (no manual coordination)
- [ ] view-history output: chronological order, plain text, one entry per line
- [ ] IPC command failures print errors but don't abort wrapper
- [ ] Complete UI: clear error messages, formatted history output
- [ ] Full backend: in-memory log storage, timestamp generation, size tracking, FIFO eviction
- [ ] Can be deployed independently: works with m01+m02+m03
- [ ] Requires no additional milestones: full history tracking capability

## Validation Questions
**Before marking this milestone complete, answer:**
- [ ] Can a real user perform complete history tracking workflows with only this milestone? **YES** - they can log events and query history throughout workflow execution
- [ ] Is it polished enough to ship publicly? **YES** - clean log format, automatic timestamps, predictable size limits
- [ ] Does it solve a real problem end-to-end? **YES** - provides runtime visibility into agent decisions and workflow events
- [ ] Does it include both complete UI and functional backend integration? **YES** - IPC commands (UI) + in-memory log storage (backend)
- [ ] Can it run independently without waiting for other milestones? **YES** - builds on m01+m02+m03, adds history layer
- [ ] Would you personally use this if it were released today? **YES** - essential for understanding what happened during automated workflows

## Vertical Slice - All Layers Included
This milestone includes:
- **IPC Command Interface**: write-history and view-history subcommands
- **CLI Flag Handling**: --write-history flag parsing and execution
- **In-Memory Log Storage**: Session-to-log mapping, append operations
- **Timestamp Generation**: ISO 8601 timestamp creation, timezone handling
- **Size Tracking**: Cumulative byte size calculation (UTF-8 encoding)
- **FIFO Eviction**: Oldest entry removal when size limit exceeded
- **Chronological Ordering**: Maintaining insertion order for view-history output
- **Session Integration**: Automatic session ID inheritance, multi-session isolation

## Notes
- History is **in-memory only** - no file persistence (design decision for simplicity)
- History format is **plain text log** (not YAML, not markdown) for simplicity and parseability
- 32MB size limit prevents unbounded memory growth in very long workflows
- Timestamps are system-generated to prevent manipulation and ensure accuracy
- History provides **runtime visibility** for debugging and understanding agent behavior
- Future milestone (m05) will add multi-agent support
- Command files should be updated to show examples of history usage for documenting decisions
