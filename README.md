# fluxid

A workflow controller for coding agents that enables breaking through context window limits via structured IMPLEMENT (→ COMMIT) → REVIEW loops.

## What

`fluxid` is a thin CLI wrapper around coding agents (Claude, Codex, etc.) that orchestrates iterative development workflows. It manages sessions, tracks history, validates reports, and provides IPC primitives for agent communication.

## Quick Start

**Build and install:**
```bash
make build    # Creates bin/fluxid
go install ./cmd/fluxid  # Install to $GOPATH/bin
```

**Basic workflow execution:**
```bash
# Initialize config (first time only)
fluxid init

# Run a workflow with Claude
fluxid --claude --file=/absolute/path/to/task.md

# Or with other agents
fluxid --codex --file=/absolute/path/to/task.md
fluxid --opencode --file=/absolute/path/to/task.md
```

## Usage

**Run workflow:**
```bash
fluxid {--claude|--codex|--opencode} --file=PATH [options] [agent-args]
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

**All Command-Line Options:**

*Agent Selection (required):*
- `--claude` - Use Claude agent
- `--codex` - Use Codex agent
- `--opencode` - Use OpenCode agent

*Workflow Options:*
- `--file=PATH` - **Required.** Absolute path to task file
- `--fluxid-iterations=N` - Max review cycles (default: 20)
- `--fluxid-implement-retries=R` - Max implement retries (default: 3)
- `--fluxid-dry-run` - Run simulation without executing agent
- `--dry-run` - Alias for `--fluxid-dry-run`
- `--fluxid-output=FORMAT` - Set initialization output format: `text`, `json`, or `yaml` (default: text)
- `--output=FORMAT` - Alias for `--fluxid-output`

*Configuration:*
- `--config=PATH` - Load custom configuration file (parsed, integration in progress)
- `--implement-command=PATH` - Override implement command file (parsed, integration in progress)
- `--review-command=PATH` - Override review command file (parsed, integration in progress)
- `--commit-command=PATH` - Override commit command file (parsed, integration in progress)

*Session Management:*
- `--session=ID` - Specify session ID (overrides FLUXID_SESSION_ID env)

*Other:*
- `--help` - Show help information

**Note:** All value flags require equals syntax (`--flag=value`). Space syntax (`--flag value`) is not supported.

## Configuration

### Configuration Precedence

Configuration is resolved using a layered precedence system (lowest to highest):

1. **Built-in Defaults** - Hardcoded fallback values
2. **Home Config** - User-wide settings (`~/.fluxid/config.yaml`)
3. **Project Config** - Project-specific settings (`./.fluxid/config.yaml`)
4. **Custom Config** - Optional config via `--config` flag (when implemented)
5. **CLI Flags** - Command-line arguments (highest priority)

**At least one config file is required** (home or project). Both may exist; fields are resolved independently with the precedence above.

### Configuration Files

All configuration files use the same YAML structure:

```yaml
agent: claude                # Agent to use: claude, codex, or opencode
iterations: 20               # Max review cycles
implement_retries: 3         # Max implement retries per cycle
commands:
  implement: /absolute/path/to/implement.md  # MUST be absolute path
  review: /absolute/path/to/review.md        # MUST be absolute path
  commit: /absolute/path/to/commit.md        # MUST be absolute path
```

**Important:** Command file paths **must be absolute paths**. Relative paths are not supported.

### Home Config (`~/.fluxid/config.yaml`)

User-wide default configuration. Create with:

```bash
fluxid init  # Creates ~/.fluxid/config.yaml and template command files
```

Example:
```yaml
agent: claude
iterations: 20
implement_retries: 3
commands:
  implement: /Users/username/.fluxid/implement.md
  review: /Users/username/.fluxid/review.md
  commit: /Users/username/.fluxid/commit.md
```

### Project Config (`./.fluxid/config.yaml`)

Project-specific configuration that overrides home config. Create with:

```bash
fluxid init .  # Creates ./.fluxid/config.yaml in current directory
```

Example (override agent for this project):
```yaml
agent: codex
# iterations and implement_retries will fall back to home config or defaults
commands:
  implement: /absolute/path/to/project/.fluxid/implement.md
  review: /absolute/path/to/project/.fluxid/review.md
  commit: /absolute/path/to/project/.fluxid/commit.md
```

### Custom Config (via `--config` flag)

**Status:** Flag is parsed but integration is in progress.

When implemented, will allow loading a config from an arbitrary path:
```bash
fluxid --claude --config=/path/to/custom-config.yaml --file=/path/to/task.md
```

### Field Resolution Examples

**Example 1: Basic layering**
```
Defaults:        agent=claude, iterations=20, implement_retries=3
Home config:     agent=claude, iterations=30
Project config:  (none)
Result:          agent=claude, iterations=30, implement_retries=3
```

**Example 2: Project overrides home**
```
Home config:     agent=claude, iterations=20, implement_retries=3
Project config:  agent=codex, iterations=10
Result:          agent=codex, iterations=10, implement_retries=3
```

**Example 3: CLI flags override everything**
```
Home config:     agent=claude, iterations=20
Project config:  agent=codex
CLI flags:       --opencode --fluxid-iterations=5
Result:          agent=opencode, iterations=5, implement_retries=3 (from default)
```

### Environment Variables

- `FLUXID_SESSION_ID` - Session identifier for IPC operations (optional, auto-generated if not set)

### Built-in Defaults

- **Agent:** `claude`
- **Iterations:** `20`
- **Implement retries:** `3`
- **Commands:** **No defaults** - must be specified in at least one config file

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

## Common Workflows & Examples

### Basic Execution

**Run a simple workflow:**
```bash
# Basic execution with Claude
fluxid --claude --file=/absolute/path/to/task.md

# With custom iterations
fluxid --claude --file=/path/to/task.md --fluxid-iterations=10

# With different agent
fluxid --codex --file=/path/to/task.md
```

### Dry Run (Simulation Mode)

**Preview workflow without execution:**
```bash
# See the execution plan without running the agent
fluxid --claude --file=/path/to/task.md --fluxid-dry-run

# Short alias
fluxid --claude --file=/path/to/task.md --dry-run
```

### Output Formats

**Get structured output for scripting:**
```bash
# JSON output - useful for scripting
fluxid --claude --file=/path/to/task.md --fluxid-output=json

# Extract session ID for later use
SID=$(fluxid --claude --file=/path/to/task.md --output=json | jq -r '.session_id')
echo "Session: $SID"

# YAML output
fluxid --claude --file=/path/to/task.md --output=yaml
```

### Session Management

**Work with specific sessions:**
```bash
# Use a custom session ID
export FLUXID_SESSION_ID=my-feature-work
fluxid --claude --file=/path/to/task.md

# Or pass via flag
fluxid --claude --file=/path/to/task.md --session=my-feature-work

# View history for a session
fluxid ipc view-history --session=my-feature-work

# Read report from a session
fluxid ipc read-report --session=my-feature-work
```

### Complete Example: Feature Development

```bash
# 1. Initialize fluxid (first time only)
fluxid init

# 2. Create a task file
cat > /tmp/add-login-feature.md <<EOF
# Add User Login Feature

Implement a user login system with:
- Login form UI
- Authentication API endpoint
- Session management
- Error handling
EOF

# 3. Run the workflow with a named session
export FLUXID_SESSION_ID=feature-login
fluxid --claude --file=/tmp/add-login-feature.md --fluxid-iterations=15

# 4. Check progress during execution (from another terminal)
fluxid ipc view-history --session=feature-login
fluxid ipc read-report --session=feature-login

# 5. Request graceful abort if needed
fluxid ipc abort --session=feature-login
```

### Advanced: Multiple Agents for Different Tasks

```bash
# Use Claude for initial implementation
fluxid --claude --file=/tmp/implement-api.md --session=api-impl

# Use Codex for code generation tasks
fluxid --codex --file=/tmp/generate-tests.md --session=test-gen

# Use OpenCode for refactoring
fluxid --opencode --file=/tmp/refactor-auth.md --session=refactor-auth
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

**Required task file path:**
```bash
fluxid --claude --file=/absolute/path/to/task.md  # Now required with absolute path
```

**Multiple agent support:**
```bash
fluxid --claude --file=/path/to/task.md   # Claude agent
fluxid --codex --file=/path/to/task.md    # Codex agent
fluxid --opencode --file=/path/to/task.md # OpenCode agent
```

**Dry-run mode:**
```bash
fluxid --claude --file=/path/to/task.md --fluxid-dry-run
```

**Output format options:**
```bash
fluxid --claude --file=/path/to/task.md --fluxid-output=json
fluxid --claude --file=/path/to/task.md --output=yaml
```

**Configuration precedence chain:**
```
Defaults → Home config → Project config → CLI flags
```

**Command file validation:**
- All command file paths must now be absolute paths
- Files are validated at startup for existence and readability

### In Progress Features

The following flags are parsed but not fully integrated yet:
- `--config=PATH` - Custom config file loading
- `--implement-command=PATH` - Override implement command
- `--review-command=PATH` - Override review command
- `--commit-command=PATH` - Override commit command
