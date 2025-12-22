# fluxid

A workflow controller for coding agents that enables breaking through context window limits via structured IMPLEMENT (→ COMMIT) → REVIEW loops.

## What

`fluxid` is a thin CLI wrapper around coding agents (Claude, Codex, etc.) that orchestrates iterative development workflows. It manages sessions, tracks history, validates reports, and provides IPC primitives for agent communication.

## Usage

**Build:**
```bash
make build    # Creates bin/fluxid
```

**Run workflow:**
```bash
fluxid --claude [--fluxid-iterations N] [--fluxid-implement-retries R]
```

**IPC commands:**

These commands are primarily used for controller ↔ agent communication during a workflow, but they can also be called by any external CLI tools or scripts to read/write session state (history, reports) independently of the agent.

```bash
fluxid ipc get-report-schema              # Print YAML schema for reports
fluxid ipc write-report [--session ID]    # Validate + store report from stdin
fluxid ipc read-report [--session ID]     # Print stored report as YAML
fluxid ipc abort [--session ID]           # Request graceful abort for session
fluxid ipc write-history <msg>            # Append timestamped entry to history
fluxid ipc view-history [--session ID]    # Print session history lines
```

Session scoping and ID:
- All IPC operations are scoped to a session. Provide it via `--session ID` or `FLUXID_SESSION_ID`; `--session` overrides the env var.
- fluxid generates a UUID v4 session automatically if `FLUXID_SESSION_ID` is not set.
- To obtain the current session ID:
  - Human-readable: fluxid prints “Session ID: …” at startup and on completion in text output.
  - Machine-readable: run with `--fluxid-output json|yaml` and read `session_id`.

Examples:
- Capture session ID once and reuse:
  - `SID=$(fluxid --claude --fluxid-output json | jq -r '.session_id')`
  - `fluxid ipc view-history --session "$SID"`
  - `fluxid ipc read-report --session "$SID"`
- Pin a session you choose (env var):
  - `export FLUXID_SESSION_ID=my-session && fluxid --claude`
  - `fluxid ipc write-report < report.yaml`  # writes report into that session
  - `fluxid ipc read-report`                 # reads report using env var
  - `fluxid ipc view-history`                # prints history using env var
- Or pin via flag per call:
  - `fluxid ipc write-history "Started build" --session my-session`
  - `fluxid ipc view-history --session my-session`
  - `fluxid ipc read-report --session my-session`

**Options:**
- `--fluxid-iterations N` - Max review cycles (default: 20)
- `--fluxid-implement-retries R` - Max implement retries (default: 3)
- `--fluxid-no-commit` - Disable commit phase
- `--session ID` - Specify session ID (overrides FLUXID_SESSION_ID env)

## Configuration

Configuration follows precedence: **CLI flags** > **env vars** > **project** > **home** > **defaults**

**Home config** (`~/.fluxid/config.yaml`):
```yaml
agent: claude
iterations: 20
implement_retries: 3
commit_enabled: false
commands:
  implement: /path/to/implement.md
  review: /path/to/review.md
  commit: /path/to/commit.md
```

**Project config** (`./.fluxid/config.yaml`):
```yaml
# Same structure as home config
# Overrides home settings for this project
```

**Environment variables:**
- `FLUXID_AGENT` - Agent to use
- `FLUXID_ITERATIONS` - Max implement - review cycles
- `FLUXID_IMPLEMENT_RETRIES` - Max implement retries
- `FLUXID_COMMIT_ENABLED` - Enable/disable commit phase
- `FLUXID_SESSION_ID` - Session identifier

**Defaults:**
- Agent: `claude`
- Iterations: `20`
- Implement retries: `3`
- Commit enabled: `false`

## Application Layout

```
fluxid-loop/
├── cmd/
│   └── fluxid/
│       └── main.go             # Thin entrypoint; calls internal/command.Execute()
├── internal/
│   ├── command/                # CLI parsing, flags, IPC subcommands, signal handling
│   ├── config/                 # Load, resolve, validate config (home, project, env)
│   ├── ipc/                    # In-memory report/history storage and schema validation
│   ├── output/                 # Text/JSON/YAML initialization status formatting
│   └── workflow/               # Core implement→review loop and dry-run simulation
├── e2e-tests/                  # End-to-end CLI tests
├── hooks/                      # Dev hooks (read-only policy)
└── Makefile, go.mod, README.md
```

Notes:
- Runtime is pure Go; shell scripts in `hooks/` are development helpers only.
- The default `cmd/fluxid/main.go` is intentionally minimal.

## Key Flags & Formats

- `--claude | --codex | --opencode` select the agent; subsequent args are passed through to the agent.
- `--fluxid-iterations N` set max review cycles (default 20).
- `--fluxid-implement-retries R` set max implement retries (default 3).
- `--fluxid-commit-enabled` enable commit phase; `--fluxid-no-commit` disables it.
- `--fluxid-dry-run` run a non-executing simulation and print the plan.
- `--fluxid-output text|json|yaml` choose initialization status format.
- `--session ID` override `FLUXID_SESSION_ID` for IPC commands.

Examples:
- Text simulation plan:
  - `fluxid --claude --fluxid-dry-run`
- JSON init status for scripting:
  - `fluxid --claude --fluxid-output json | jq -r '.session_id'`
- YAML init status:
  - `fluxid --claude --fluxid-output yaml`

Top-level helper:
- `fluxid --write-history <message>` appends to history for the active session (requires `FLUXID_SESSION_ID`).
