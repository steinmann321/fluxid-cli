# fluxid Loop Workflow

## Overview

**fluxid** is a workflow controller that enables coding agents to break through context window limitations using a structured **IMPLEMENT → REVIEW** loop with retry logic and graceful abort handling.

## Workflow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          FLUXID WORKFLOW CONTROLLER                      │
│                                                                          │
│  Entry: CLI Command + Configuration (CLI > Env > Project > Home)        │
└──────────────────────────────────┬──────────────────────────────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │   Initialize Session        │
                    │  - Generate Session ID      │
                    │  - Setup Signal Handler     │
                    │  - Print Config Status      │
                    └──────────────┬──────────────┘
                                   │
        ┌──────────────────────────▼──────────────────────────┐
        │         REVIEW CYCLE LOOP (1..MaxReviewCycles)      │
        │                                                      │
        │  ┌────────────────────────────────────────────────┐ │
        │  │  IMPLEMENT PHASE (1..MaxImplementRetries)     │ │
        │  │                                                │ │
        │  │  ┌──────────────────────────────────────────┐ │ │
        │  │  │  Execute Agent: Implement                │ │ │
        │  │  │  Command: agent --print [args] prompt    │ │ │
        │  │  │  Environment: FLUXID_SESSION_ID set      │ │ │
        │  │  └───────────────┬──────────────────────────┘ │ │
        │  │                  │                             │ │
        │  │  ┌───────────────▼──────────────────────────┐ │ │
        │  │  │  Agent Writes Report (IPC)               │ │ │
        │  │  │  fluxid ipc write-report < report.yaml   │ │ │
        │  │  │  Location: /tmp/fluxid-reports/          │ │ │
        │  │  │            <session-id>.yaml              │ │ │
        │  │  └───────────────┬──────────────────────────┘ │ │
        │  │                  │                             │ │
        │  │  ┌───────────────▼──────────────────────────┐ │ │
        │  │  │  Poll & Validate Report                  │ │ │
        │  │  │  - Poll every 2s (max 150 attempts)      │ │ │
        │  │  │  - Validate schema (required fields)     │ │ │
        │  │  │  - Extract status: PASS or FAIL          │ │ │
        │  │  └───────────────┬──────────────────────────┘ │ │
        │  │                  │                             │ │
        │  │         ┌────────▼─────────┐                  │ │
        │  │         │  Status = PASS?  │                  │ │
        │  │         └────┬────────┬────┘                  │ │
        │  │              │ YES    │ NO                     │ │
        │  │              │        │                        │ │
        │  │              │   ┌────▼──────────────────────┐ │ │
        │  │              │   │  Retry < MaxRetries?     │ │ │
        │  │              │   └────┬────────┬────────────┘ │ │
        │  │              │        │ YES    │ NO           │ │
        │  │              │   ┌────▼────┐   │              │ │
        │  │              │   │  Retry  │   │              │ │
        │  │              │   │  Again  │   │              │ │
        │  │              │   └────┬────┘   │              │ │
        │  │              │        │        │              │ │
        │  │              │        └────┐   │              │ │
        │  │              │             │   │              │ │
        │  │              │             │   │ (Continue    │ │
        │  │              │             │   │  to commit/  │ │
        │  │              │             │   │  review)     │ │
        │  └──────────────┼─────────────┴───┘              │ │
        │                 │                                │ │
        │  ┌──────────────▼───────────┐                    │ │
        │  │  COMMIT PHASE (optional) │                    │ │
        │  │  - Only if CommitEnabled │                    │ │
        │  │  - No report validation  │                    │ │
        │  └──────────────┬───────────┘                    │ │
        │                 │                                │ │
        │  ┌──────────────▼─────────────────────────┐      │ │
        │  │  REVIEW PHASE                          │      │ │
        │  │                                │         │    │ │
        │  │  ┌───────────────────────┐     │         │    │ │
        │  │  │  Execute Agent: Review│     │         │    │ │
        │  │  │  Evaluate impl quality│     │         │    │ │
        │  │  └────────┬──────────────┘     │         │    │ │
        │  │           │                    │         │    │ │
        │  │  ┌────────▼──────────────┐     │         │    │ │
        │  │  │  Agent Writes Report  │     │         │    │ │
        │  │  │  (IPC)                │     │         │    │ │
        │  │  └────────┬──────────────┘     │         │    │ │
        │  │           │                    │         │    │ │
        │  │  ┌────────▼──────────────┐     │         │    │ │
        │  │  │  Poll & Validate      │     │         │    │ │
        │  │  │  Report               │     │         │    │ │
        │  │  └────────┬──────────────┘     │         │    │ │
        │  │           │                    │         │    │ │
        │  │  ┌────────▼──────────┐         │         │    │ │
        │  │  │  Status = PASS?   │         │         │    │ │
        │  │  └────┬─────────┬────┘         │         │    │ │
        │  │       │ YES     │ NO           │         │    │ │
        │  │       │         │              │         │    │ │
        │  └───────┼─────────┼──────────────┘         │    │ │
        │          │         │                        │    │ │
        └──────────┼─────────┴────────────────────────┘    │ │
                   │                                       │ │
                   │  ┌────────────────────────────────────┘ │
                   │  │  Next Review Cycle                   │
                   │  │  (if < MaxReviewCycles)              │
                   │  └──────────────────────────────────────┘
                   │
    ┌──────────────▼──────────────┐
    │  WORKFLOW COMPLETE          │
    │  - Exit code: 0 (success)   │
    │  - Exit code: 1 (failure)   │
    │  - Exit code: 130 (aborted) │
    └─────────────────────────────┘
```

## Flow Description

### 1. Initialization Phase

- **Configuration Loading**: Resolves config with precedence: CLI flags > Environment > Project config (.fluxid/config.yaml) > Home config (~/.fluxid/config.yaml) > Defaults
- **Session Setup**: Generates unique session ID (UUID) for IPC isolation
- **Signal Handler**: Installs SIGINT/SIGTERM handler for graceful abort (first signal = graceful, second = force exit)
- **Status Output**: Prints initialization status with agent, retries, commit setting, and config sources

### 2. Review Cycle Loop (Outer Loop)

Executes up to `MaxReviewCycles` times (default: 20)

**Purpose**: Iterates until review phase reports PASS or max cycles reached

**Abort Checks**: Before each cycle, checks for abort flag (set by Ctrl+C or `fluxid ipc abort`)

### 3. Implement Phase (Inner Loop)

Executes up to `MaxImplementRetries` times (default: 3) per review cycle

**Steps**:
1. **Agent Execution**: Spawns coding agent (e.g., `claude --print [args] "Implement the required changes..."`)
   - Agent process blocks until completion
2. **IPC Report**: Agent writes report synchronously via `fluxid ipc write-report < report.yaml` to `/tmp/fluxid-reports/<session-id>.yaml`
   - Report write is synchronous - file is guaranteed to exist when IPC command completes
3. **Immediate Report Check**: After agent exits, controller immediately checks for report (no polling)
4. **Validation**: Validates report schema (required fields: command, artifact, timestamp, status, issues)
   - **Report exists and valid**: Use its status (PASS or FAIL)
   - **Report missing**: Treat as FAIL (agent didn't write report or IPC failed)
   - **Report invalid/malformed**: Treat as FAIL (agent wrote bad YAML)
5. **Status Check**:
   - **PASS**: Proceed to commit phase (if enabled) then review phase
   - **FAIL**: Retry implement (if retries remaining), or continue to commit/review phases if all retries exhausted

**Important**: When all implement retries are exhausted (all attempts report FAIL), the workflow continues to the commit phase (if enabled) and review phase instead of terminating. This allows the review phase to assess the incomplete implementation and decide whether to continue to the next review cycle.

**Report Schema**:
```yaml
command: "fluxid.implement"
artifact: "path/to/changed/file"
timestamp: "2025-12-25T10:00:00Z"
status: PASS  # or FAIL
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
```

### 4. Commit Phase (Optional)

**Activation**: Only runs if `CommitEnabled=true` in config

**Steps**:
1. Executes agent with commit prompt: `"Create a git commit with all changes."`
2. No report validation (fire-and-forget)

### 5. Review Phase

Executes once per review cycle after successful implement

**Steps**:
1. **Agent Execution**: Spawns agent with review prompt: `"Review the implementation and report status."`
2. **IPC Report**: Same report mechanism as implement phase
3. **Report Polling & Validation**: Same polling/validation as implement
4. **Status Check**:
   - **PASS**: Workflow succeeds, exit code 0
   - **FAIL**: Continue to next review cycle (if cycles remaining) or fail workflow

### 6. Graceful Abort Handling

**Trigger**: User presses Ctrl+C (SIGINT) or sends SIGTERM

**Behavior**:
- **First Signal**: Sets abort flag (`/tmp/fluxid-reports/<session-id>.abort`), prints "graceful abort requested, press Ctrl+C again to force exit"
- **Abort Checks**: Controller checks abort flag:
  - Before each review cycle
  - Before each implement retry
  - During report polling
- **Exit**: When abort detected, exits with code 130 after current phase completes
- **Second Signal**: Forces immediate exit with code 130 (no cleanup)

## Key Components

### Configuration Resolution

```
Precedence (highest to lowest):
1. CLI flags (--fluxid-iterations, --fluxid-implement-retries, etc.)
2. Environment variables (FLUXID_ITERATIONS, FLUXID_IMPLEMENT_RETRIES, etc.)
3. Project config (./.fluxid/config.yaml)
4. Home config (~/.fluxid/config.yaml)
5. Built-in defaults (iterations=20, retries=3, commit=false)
```

### IPC (Inter-Process Communication)

**Storage**: `/tmp/fluxid-reports/` directory

**Files**:
- `<session-id>.yaml` - Current report
- `<session-id>.abort` - Abort flag (presence = abort requested)
- `<session-id>.history` - Timestamped history entries (max 32MB, FIFO eviction)
- `<session-id>.history.lock` - File lock for concurrent writes

**Commands**:
```bash
fluxid ipc get-report-schema           # Print report schema
fluxid ipc write-report                # Validate and store report (reads stdin)
fluxid ipc read-report                 # Output stored report
fluxid ipc abort                       # Request graceful abort
fluxid ipc write-history <message>     # Append history entry
fluxid ipc view-history                # Read history
```

### Agent Integration

**Execution Pattern**:
```bash
# Controller spawns agent with:
claude --print [agent-args] "Phase-specific prompt"

# Environment variables set:
FLUXID_SESSION_ID=<session-id>

# Agent streams output to user (stdout/stderr passthrough)
# Agent writes report via IPC when done
```

## Example Execution

```bash
# Run workflow with Claude agent, 5 review cycles, 2 implement retries
fluxid --claude --fluxid-iterations 5 --fluxid-implement-retries 2

# Dry-run to see execution plan
fluxid --claude --fluxid-dry-run

# Custom output format
fluxid --claude --fluxid-output json
```

## Exit Codes

- **0**: Workflow succeeded (review phase reported PASS)
- **1**: Workflow failed (review phase exhausted all cycles with FAIL, or agent command failed)
- **130**: Graceful abort (user pressed Ctrl+C or sent SIGTERM)

**Note**: Exhausting implement retries (all FAIL) does NOT cause the workflow to fail. The workflow continues to the review phase, which makes the final determination.

## Configuration Files

**Home Config** (`~/.fluxid/config.yaml`):
```yaml
agent: claude
iterations: 20
implement_retries: 3
commit_enabled: false
commands:
  implement: implement.md
  review: review.md
  commit: commit.md
```

**Project Config** (`./.fluxid/config.yaml`):
```yaml
agent: claude
iterations: 5
implement_retries: 2
commit_enabled: true
commands:
  implement: custom-impl.md
  review: custom-review.md
```

## Design Principles

1. **Stateless Commands, Stateful Sessions**: Commands are idempotent; state managed via IPC
2. **Explicit Interfaces**: Reports must match schema (no implicit fields)
3. **Fail Fast with Diagnostics**: Clear error messages with session ID for debugging
4. **Graceful Abort**: Two-signal pattern (first=graceful, second=force)
5. **Configuration Precedence**: CLI overrides all, env overrides files
6. **Pure Go Runtime**: No shell scripts as runtime dependencies
7. **File-Based IPC**: Reports/history/abort via files in /tmp with flock for concurrency

## Key Files Reference

| File | Purpose |
|------|---------|
| `internal/workflow/workflow.go` | Core loop logic (Run, runImplementPhase, runReviewPhase) |
| `internal/command/root.go` | CLI orchestration and main Execute() entry point |
| `internal/ipc/storage.go` | Report/history/abort storage and file operations |
| `internal/config/config.go` | Configuration resolution with precedence |
| `internal/command/root_signal.go` | Signal handler for graceful abort |
| `internal/ipc/schema.yaml` | Report schema definition |
| `cmd/fluxid/main.go` | Minimal entry point |
