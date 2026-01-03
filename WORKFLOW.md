# Fluxid Workflow

## Overview

Fluxid is a workflow controller that enables coding agents to break through context window limitations using a structured **IMPLEMENT → REVIEW** loop. The system orchestrates iterative development cycles where agents implement changes, commit them, and review their own work until the task is complete or iteration limits are reached.

## Core Workflow Loop

```
┌─────────────────────────────────────────────────────────────────┐
│                      FLUXID WORKFLOW LOOP                       │
│                  (Max Review Cycles: configurable)              │
└─────────────────────────────────────────────────────────────────┘

START
  │
  ├─ Initialize Session (UUID)
  ├─ Load Configuration (.fluxid/config.yaml)
  ├─ Resolve Command Files (implement.md, review.md, commit.md)
  │
  ▼
┌────────────────────── FOR EACH REVIEW CYCLE ──────────────────────┐
│                                                                    │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │              IMPLEMENT PHASE (with retries)                 │  │
│  │  ┌───────────────────────────────────────────────────────┐  │  │
│  │  │                                                         │  │  │
│  │  │  • Execute: agent --prompt implement.md <args>         │  │  │
│  │  │  • Agent implements changes using TDD                  │  │  │
│  │  │  • Agent writes YAML report via IPC                    │  │  │
│  │  │  • Report validated against schema                     │  │  │
│  │  │                                                         │  │  │
│  │  │  ┌─────────┐                                           │  │  │
│  │  │  │ PASS?   │───YES──> Break (continue to commit)      │  │  │
│  │  │  └─────────┘                                           │  │  │
│  │  │       │                                                │  │  │
│  │  │       NO                                               │  │  │
│  │  │       │                                                │  │  │
│  │  │  ┌─────────┐                                           │  │  │
│  │  │  │ Retries │                                           │  │  │
│  │  │  │  Left?  │───YES──> Retry (max: configurable)       │  │  │
│  │  │  └─────────┘                                           │  │  │
│  │  │       │                                                │  │  │
│  │  │       NO                                               │  │  │
│  │  │       │                                                │  │  │
│  │  │  Retries Exhausted → Continue to Review               │  │  │
│  │  │                                                         │  │  │
│  │  └─────────────────────────────────────────────────────────┘  │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                              │                                     │
│                              ▼                                     │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │                     COMMIT PHASE                            │  │
│  │                                                             │  │
│  │  • Execute: agent --prompt commit.md <args>                │  │
│  │  • Agent commits changes with descriptive message          │  │
│  │  • Git commit created                                      │  │
│  │                                                             │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                              │                                     │
│                              ▼                                     │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │                     REVIEW PHASE                            │  │
│  │                                                             │  │
│  │  • Execute: agent --prompt review.md <args>                │  │
│  │  • Agent reviews implementation & tests                    │  │
│  │  • Agent writes YAML report via IPC                        │  │
│  │  • Report validated against schema                         │  │
│  │                                                             │  │
│  │  ┌─────────┐                                               │  │
│  │  │ PASS?   │───YES──> SUCCESS (Exit Loop)                 │  │
│  │  └─────────┘                                               │  │
│  │       │                                                    │  │
│  │       NO                                                   │  │
│  │       │                                                    │  │
│  │  Continue to Next Review Cycle                            │  │
│  │                                                             │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                              │                                     │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │ Check Abort Flag (SIGINT/SIGTERM)                        │    │
│  │   • User pressed Ctrl+C?                                 │    │
│  │   • Agent requested abort?                               │    │
│  │   ───YES──> Exit with code 130                           │    │
│  │       │                                                   │    │
│  │       NO → Continue                                       │    │
│  └──────────────────────────────────────────────────────────┘    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
  │
  ▼
END
  • Exit Code 0: Success
  • Exit Code 1: Error
  • Exit Code 130: User Interrupt
```

## Phase Details

### IMPLEMENT PHASE
**Purpose:** Agent implements changes using Test-Driven Development (TDD)

**Process:**
1. Agent receives implement prompt (`.fluxid/commands/fluxid.implement.md`)
2. Implements changes incrementally with tests
3. Runs tests to verify implementation
4. Generates structured YAML report via IPC

**Retries:** Configurable (default: 3)
- On FAIL: Agent retries implementation
- After exhausting retries: Continues to commit phase anyway

**Report Status:**
- `PASS`: Implementation complete, tests passing
- `FAIL`: Issues found, needs retry or review

### COMMIT PHASE
**Purpose:** Create git commit with descriptive message

**Process:**
1. Agent receives commit prompt (`.fluxid/commands/fluxid.commit.md`)
2. Stages changes
3. Creates commit with clear message
4. No report generated (simple execution)

### REVIEW PHASE
**Purpose:** Agent reviews own implementation and identifies issues

**Process:**
1. Agent receives review prompt (`.fluxid/commands/fluxid.review-implementation.md`)
2. Reviews code changes, test coverage, and quality
3. Identifies blockers, defects, concerns, observations, enhancements
4. Generates structured YAML report via IPC

**Report Status:**
- `PASS`: Review complete, no critical issues → Exit loop
- `FAIL`: Issues found → Start next review cycle

## Report Structure

Agents communicate with fluxid via IPC by writing YAML reports:

```yaml
command: "implement" | "review"
artifact: "path/to/main/artifact"
timestamp: "2025-12-27T10:30:00Z"
status: "PASS" | "FAIL"

issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []

summary: "Optional summary text"
next_steps: "Optional next steps"
```

Each issue contains:
- `message`: Description (required)
- `location`: File/line reference (optional)
- `code`: Code snippet (optional)
- `suggestion`: Fix recommendation (optional)
- `reference`: Documentation link (optional)

## Configuration

### Hierarchy (lowest to highest priority):
1. Built-in defaults
2. Home config: `~/.fluxid/config.yaml`
3. Project config: `./.fluxid/config.yaml`
4. Custom config: `--config=PATH`
5. CLI flags: `--fluxid-iterations=N`

### Key Settings:

```yaml
agent: "claude"                  # or "codex", "opencode"
iterations: 20                   # Max review cycles
implement_retries: 3             # Max implement phase retries

commands:
  implement: "commands/fluxid.implement.md"
  commit: "commands/fluxid.commit.md"
  review: "commands/fluxid.review-implementation.md"
```

## Usage Examples

### Basic Usage
```bash
# Initialize project
fluxid init

# Run workflow with Claude
fluxid --claude --prompt "implement user authentication"

# Custom iteration limits
fluxid --claude --fluxid-iterations=10 --prompt "fix bug in login"

# Dry run (simulation)
fluxid --claude --fluxid-dry-run --prompt "refactor database layer"
```

### Agent Communication (IPC)
```bash
# Inside agent: Get report schema
fluxid ipc get-report-schema

# Inside agent: Submit report
cat report.yaml | fluxid ipc write-report

# Inside agent: Add history entry
echo "Starting implementation phase" | fluxid ipc write-history

# Inside agent: Request abort
fluxid ipc abort
```

### Output Formats
```bash
# Human-readable text (default)
fluxid --claude --prompt "task"

# Machine-readable JSON
fluxid --claude --fluxid-output=json --prompt "task"

# YAML output
fluxid --claude --fluxid-output=yaml --prompt "task"
```

## Session Management

Each workflow run creates a unique session:
- **Session ID:** Auto-generated UUID v4
- **Environment:** `FLUXID_SESSION_ID` exported to agent
- **Storage:** Reports/history/abort flags in `${TMPDIR}/fluxid-reports/`

### Graceful Abort
Press `Ctrl+C` or send `SIGTERM` to gracefully abort:
1. Signal handler sets abort flag
2. Workflow checks flag before each phase
3. Exits with code 130 (interrupted)
4. Agent can also request abort via IPC

## Workflow Variants

### CLI Implementation Workflow
Use `fluxid.implement-cli.md` for CLI-specific development:
- Focuses on command parsing and user interface
- Emphasizes error messages and help text
- Tests for various input combinations

### E2E Test Workflow
Use `fluxid.implement-e2e.md` and `fluxid.review-implementation-e2e.md`:
- Black-box integration testing
- Full system validation
- User scenario coverage

## Design Philosophy

**Breaking Context Windows**
- Agents work in focused iterations
- Each cycle has clear input/output boundaries
- History preserved across cycles

**Structured Communication**
- Schema-validated YAML reports
- Typed issue categories (blockers → enhancements)
- Machine-readable and human-readable

**Fail-Fast with Retries**
- Implement phase allows retries (transient failures)
- Review phase triggers new cycles (iterative improvement)
- Clear exit conditions (PASS or max iterations)

**Stateless Commands, Stateful Sessions**
- Each command execution is independent
- Session state maintained via IPC storage
- Git provides durable state between runs

## Exit Codes

- `0`: Success (review passed or simulation complete)
- `1`: Error (validation failure, missing config, agent crash)
- `130`: User interrupt (Ctrl+C, SIGTERM, or agent abort request)

## Hooks Integration

Fluxid works with git hooks (read-only policy):
- Pre-commit hooks validate formatting, linting, coverage
- Hooks must never modify repository contents
- Hooks must never be weakened or bypassed

## Future Extensions

Potential workflow enhancements:
- Parallel phase execution
- Custom phase ordering
- Conditional branching logic
- Historical analysis and metrics
- Multi-agent collaboration
- Distributed execution support
