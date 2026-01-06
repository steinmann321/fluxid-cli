# CLI Reference

Complete reference for all fluxid command-line options and subcommands.

## Commands

### `fluxid` (Main Command)

Run a workflow with an agent.

**Syntax:**
```bash
fluxid {--claude|--codex|--opencode} --file=PATH [options]
```

**Required Flags:**
- **Agent selection** (exactly one required):
  - `--claude` - Use Claude agent
  - `--codex` - Use Codex agent
  - `--opencode` - Use OpenCode agent
- `--file=PATH` - Absolute path to task description file

**Workflow Options:**
- `--fluxid-iterations=N` - Maximum review cycles (default: 20)
- `--fluxid-implement-retries=R` - Maximum implementation retries per cycle (default: 3)
- `--fluxid-dry-run` - Run simulation without executing agent
- `--dry-run` - Alias for `--fluxid-dry-run`

**Output Options:**
- `--fluxid-output=FORMAT` - Output format: `text`, `json`, or `yaml` (default: `text`)
- `--output=FORMAT` - Alias for `--fluxid-output`

**Configuration Overrides:**
- `--config=PATH` - Load custom configuration file *(parsed, integration in progress)*
- `--implement-command=PATH` - Override implement command file *(parsed, integration in progress)*
- `--review-command=PATH` - Override review command file *(parsed, integration in progress)*
- `--commit-command=PATH` - Override commit command file *(parsed, integration in progress)*

**Session Management:**
- `--session=ID` - Specify session ID (overrides `FLUXID_SESSION_ID` environment variable)

**Help:**
- `--help`, `-h` - Show help information

**Examples:**
```bash
# Basic execution
fluxid --claude --file=/path/to/task.md

# With custom iterations
fluxid --claude --file=/path/to/task.md --fluxid-iterations=10

# Dry-run simulation
fluxid --claude --file=/path/to/task.md --fluxid-dry-run

# JSON output for scripting
fluxid --claude --file=/path/to/task.md --fluxid-output=json

# Named session
fluxid --claude --file=/path/to/task.md --session=my-feature
```

---

### `fluxid init`

Initialize fluxid configuration and templates.

**Syntax:**
```bash
fluxid init [PATH]
```

**Arguments:**
- `PATH` - Optional directory path:
  - If omitted: creates `~/.fluxid/config.yaml` (home config)
  - If `.`: creates `./.fluxid/config.yaml` (project config)
  - If path: creates config at specified location

**What it creates:**
- `config.yaml` - Configuration file
- `commands/` - Directory with command template files:
  - `implement.md` - Implementation instructions
  - `review.md` - Review instructions
  - `commit.md` - Commit instructions
  - Plus additional templates
- `templates/` - Additional workflow templates

**Examples:**
```bash
# Create home config
fluxid init

# Create project config in current directory
fluxid init .

# Create config at specific location
fluxid init /path/to/project
```

---

### `fluxid report`

Manage agent report files.

**Commands:**

#### `fluxid report --get-file`
Get absolute path to the report file for the current session.

**Usage:**
```bash
REPORT_FILE=$(fluxid report --get-file)
```

**Output:** Absolute path to `<session-root>/<session-id>/report.yaml`

**Session:** Uses `FLUXID_SESSION_ID` environment variable (must be set by workflow or user)

---

#### `fluxid report --validate`
Validate existing report file against schema.

**Usage:**
```bash
fluxid report --validate
```

**Exit codes:**
- `0` - Report valid
- `1` - Validation error (details printed to stderr)

**Validation checks:**
- YAML structure correctness
- Required fields present
- Field types match schema
- Status enum values valid
- No additional properties (strict validation)
- No YAML anchors/aliases/merge keys (security)

---

#### `fluxid report --get-schema`
Output the report YAML schema.

**Usage:**
```bash
fluxid report --get-schema > report-schema.yaml
```

**Output:** Complete YAML schema definition for reports

---

### `fluxid history`

Manage workflow history files.

**Commands:**

#### `fluxid history --get-file`
Get absolute path to the history file for the current session.

**Usage:**
```bash
HISTORY_FILE=$(fluxid history --get-file)
```

**Output:** Absolute path to `<session-root>/<session-id>/history.yaml`

**Session:** Uses `FLUXID_SESSION_ID` environment variable

---

#### `fluxid history --validate`
Validate existing history file against schema.

**Usage:**
```bash
fluxid history --validate
```

**Exit codes:**
- `0` - History valid
- `1` - Validation error (details printed to stderr)

**Validation checks:**
- YAML array structure
- Each event has required fields
- Timestamp format (ISO 8601)
- Status enum values valid
- No additional properties per event
- No YAML security issues

---

#### `fluxid history --get-schema`
Output the history YAML schema.

**Usage:**
```bash
fluxid history --get-schema > history-schema.yaml
```

**Output:** Complete YAML schema definition for history events

---

## Environment Variables

### `FLUXID_SESSION_ID`
Session identifier for file operations.

**Usage:**
```bash
export FLUXID_SESSION_ID=my-feature-123
fluxid --claude --file=/path/to/task.md
```

**Behavior:**
- If not set: fluxid auto-generates a UUID
- If set by user: used as-is (no validation, user responsibility)
- If set by workflow: passed to agent for file operations

**Affects:**
- Report file path: `<session-root>/<session-id>/report.yaml`
- History file path: `<session-root>/<session-id>/history.yaml`

### `FLUXID_SESSION_ROOT`
Override session storage root directory.

**Default:** `$HOME/.fluxid/sessions` or `./.fluxid/sessions`

**Usage:**
```bash
export FLUXID_SESSION_ROOT=/custom/path
fluxid --claude --file=/path/to/task.md
```

---

## Flag Syntax Requirements

**⚠️ Important:** All value flags require equals syntax.

**Correct:**
```bash
--fluxid-iterations=20
--file=/path/to/task.md
--output=json
```

**Incorrect:**
```bash
--fluxid-iterations 20   # ❌ Not supported
--file /path/to/task.md  # ❌ Not supported
--output json            # ❌ Not supported
```

---

## Exit Codes

- `0` - Success
- `1` - General error (configuration, validation, workflow failure)
- `2` - Command-line argument error
- `3` - Agent execution error

---

## Common Patterns

### Scripting with JSON Output
```bash
# Extract session ID
SESSION=$(fluxid --claude --file=/path/to/task.md --output=json | jq -r '.session_id')

# Extract final status
STATUS=$(fluxid --claude --file=/path/to/task.md --output=json | jq -r '.status')
```

### Session Management
```bash
# Named session
export FLUXID_SESSION_ID=feature-auth
fluxid --claude --file=/path/to/task.md

# Get session files
REPORT=$(fluxid report --get-file)
HISTORY=$(fluxid history --get-file)

# Validate session outputs
fluxid report --validate && echo "Report valid"
fluxid history --validate && echo "History valid"
```

### Dry-Run Testing
```bash
# Test workflow without agent execution
fluxid --claude --file=/path/to/task.md --fluxid-dry-run

# See what would happen
fluxid --claude --file=/path/to/task.md --fluxid-dry-run --output=yaml
```
