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
fluxid --claude [--fluxid-iterations=N] [--fluxid-implement-retries=R]
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
  - `SID=$(fluxid --claude --fluxid-output=json | jq -r '.session_id')`
  - `fluxid ipc view-history --session="$SID"`
  - `fluxid ipc read-report --session="$SID"`
- Pin a session you choose (env var):
  - `export FLUXID_SESSION_ID=my-session && fluxid --claude`
  - `fluxid ipc write-report < report.yaml`  # writes report into that session
  - `fluxid ipc read-report`                 # reads report using env var
  - `fluxid ipc view-history`                # prints history using env var
- Or pin via flag per call:
  - `fluxid ipc write-history "Started build" --session=my-session`
  - `fluxid ipc view-history --session=my-session`
  - `fluxid ipc read-report --session=my-session`

**Options:**
- `--fluxid-iterations=N` - Max review cycles (default: 20)
- `--fluxid-implement-retries=R` - Max implement retries (default: 3)
- `--config=PATH` - Load custom configuration file
- `--implement-command=PATH` - Override implement command file
- `--review-command=PATH` - Override review command file
- `--commit-command=PATH` - Override commit command file
- `--session=ID` - Specify session ID (overrides FLUXID_SESSION_ID env)

## Configuration

Configuration follows precedence (lowest to highest): **Defaults** → **Home config** → **Project config** → **Custom config** (`--config`) → **CLI flags**

**At least one default config is required** (home or project). Both may exist; project takes precedence.

**Home config** (`~/.fluxid/config.yaml`):
```yaml
agent: claude
iterations: 20
implement_retries: 3
commands:
  implement: implement.md    # Relative to ~/.fluxid/
  review: review.md
  commit: commit.md
```

**Project config** (`./.fluxid/config.yaml`):
```yaml
# Same structure as home config
# Overrides home settings for this project
# Command paths are relative to ./.fluxid/
agent: codex
commands:
  implement: implement.md
  review: review.md
  commit: commit.md
```

**Custom config** (`--config=path/to/config.yaml`):
```yaml
# Optional custom config file
# Inherits missing fields from default configs
# Command paths are relative to the config file's directory
iterations: 10
commands:
  implement: custom-implement.md
```

**Environment variables:**
- `FLUXID_SESSION_ID` - Session identifier for IPC operations (optional, auto-generated if not set)

**Defaults:**
- Agent: `claude`
- Iterations: `20`
- Implement retries: `3`
- Commands: **No defaults** - must be specified in config file

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

**Agent selection:**
- `--claude | --codex | --opencode` - Select the agent; subsequent args are passed through to the agent

**Workflow control:**
- `--fluxid-iterations=N` - Set max review cycles (default: 20)
- `--fluxid-implement-retries=R` - Set max implement retries (default: 3)
- `--fluxid-dry-run` - Run a non-executing simulation and print the plan

**Configuration:**
- `--config=PATH` - Load custom configuration file
- `--implement-command=PATH` - Override implement command file
- `--review-command=PATH` - Override review command file
- `--commit-command=PATH` - Override commit command file

**Output & session:**
- `--fluxid-output=text|json|yaml` - Choose initialization status format (default: text)
- `--session=ID` - Override `FLUXID_SESSION_ID` for IPC commands

**Note:** All value flags require equals syntax (`--flag=value`). Space syntax (`--flag value`) is not supported.

**Examples:**

Run workflow with custom iterations:
```bash
fluxid --claude --fluxid-iterations=10
```

Use custom config file:
```bash
fluxid --claude --config=my-config.yaml
```

Text simulation plan:
```bash
fluxid --claude --fluxid-dry-run
```

JSON init status for scripting:
```bash
SID=$(fluxid --claude --fluxid-output=json | jq -r '.session_id')
```

YAML init status:
```bash
fluxid --claude --fluxid-output=yaml
```

Override command files:
```bash
fluxid --claude --implement-command=custom-implement.md
```

## Migration Guide (v1.x → v2.0)

**v2.0 introduces breaking changes.** This is a major version update with backwards-incompatible changes.

### Removed Features

**Environment variables for configuration:**
- ❌ `FLUXID_AGENT` - Use CLI flags (`--claude`, `--codex`, `--opencode`) or config files
- ❌ `FLUXID_ITERATIONS` - Use `--fluxid-iterations=N` or config file
- ❌ `FLUXID_IMPLEMENT_RETRIES` - Use `--fluxid-implement-retries=R` or config file
- ❌ `FLUXID_COMMIT_ENABLED` - Commits always run (cannot be disabled)
- ✅ `FLUXID_SESSION_ID` - Still supported for IPC operations

**Commit toggle flags:**
- ❌ `--fluxid-commit` - Removed (commits always run)
- ❌ `--fluxid-no-commit` - Removed (commits always run)
- ❌ `commit_enabled` config field - Removed from config files

**Space syntax for value flags:**
- ❌ `--fluxid-iterations 20` - Not supported
- ✅ `--fluxid-iterations=20` - Required equals syntax

**Source tracking in output:**
- ❌ JSON/YAML output no longer includes `*_source` fields
- Example removed fields: `agent_source`, `review_cycles_source`, `commit_enabled_source`

### Required Changes

**1. Update config files:**

Before (v1.x):
```yaml
agent: claude
iterations: 20
commit_enabled: false  # ❌ Remove this field
commands:
  implement: implement.md
  review: review.md
  commit: commit.md
```

After (v2.0):
```yaml
agent: claude
iterations: 20
# commit_enabled removed - commits always run
commands:
  implement: implement.md
  review: review.md
  commit: commit.md
```

**2. Update CLI invocations:**

Before (v1.x):
```bash
export FLUXID_AGENT=claude              # ❌ Environment variables removed
fluxid --fluxid-iterations 20           # ❌ Space syntax not supported
fluxid --fluxid-no-commit               # ❌ Commit toggle removed
```

After (v2.0):
```bash
fluxid --claude                         # ✅ Use CLI flags for agent
fluxid --fluxid-iterations=20           # ✅ Equals syntax required
# Commits always run automatically      # ✅ No toggle needed
```

**3. At least one config file required:**

v2.0 requires at least one default config file to exist:
- `~/.fluxid/config.yaml` (home config), OR
- `./.fluxid/config.yaml` (project config)

If neither exists, fluxid will exit with an error. Create a minimal config:

```yaml
agent: claude
iterations: 20
implement_retries: 3
commands:
  implement: implement.md
  review: review.md
  commit: commit.md
```

And create the command files in the same directory as the config.

**4. Update scripts parsing JSON/YAML output:**

Before (v1.x):
```bash
fluxid --claude --fluxid-output json | jq -r '.agent_source'  # ❌ *_source fields removed
```

After (v2.0):
```bash
fluxid --claude --fluxid-output=json | jq -r '.agent'  # ✅ Use direct fields only
```

### New Features in v2.0

**Custom config files:**
```bash
fluxid --claude --config=custom-config.yaml
```

**Command override flags:**
```bash
fluxid --claude --implement-command=my-implement.md
```

**Precedence chain:**
Defaults → Home config → Project config → Custom config → CLI flags
