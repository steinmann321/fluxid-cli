# Configuration

Complete guide to fluxid configuration files, precedence, and environment variables.

## Overview

fluxid uses a layered configuration system where settings are resolved from multiple sources with clear precedence rules.

**Configuration sources** (lowest to highest priority):
1. Built-in defaults
2. Home config (`~/.fluxid/config.yaml`)
3. Project config (`./.fluxid/config.yaml`)
4. Custom config (`--config` flag) *(in progress)*
5. CLI flags

**Requirement:** At least one config file (home or project) must exist.

---

## Configuration File Format

All configuration files use the same YAML structure:

```yaml
agent: claude                # Agent: claude, codex, or opencode
iterations: 20               # Max review cycles
implement_retries: 3         # Max retries per cycle
commands:
  implement: /absolute/path/to/implement.md  # MUST be absolute
  review: /absolute/path/to/review.md        # MUST be absolute
  commit: /absolute/path/to/commit.md        # MUST be absolute
```

**Requirements:**
- All command file paths **must be absolute paths**
- Command files must exist and be readable
- At least one config file (home or project) must exist

---

## Home Configuration

**Location:** `~/.fluxid/config.yaml`

**Purpose:** User-wide default settings.

**Create:**
```bash
fluxid init
```

**What it creates:**
```
~/.fluxid/
├── config.yaml                 # Configuration
├── commands/                   # Command templates
│   ├── implement.md
│   ├── review.md
│   ├── commit.md
│   └── ... (additional templates)
└── templates/                  # Workflow templates
```

**Example config:**
```yaml
agent: claude
iterations: 20
implement_retries: 3
commands:
  implement: /Users/username/.fluxid/commands/implement.md
  review: /Users/username/.fluxid/commands/review.md
  commit: /Users/username/.fluxid/commands/commit.md
```

**Use case:** Default settings for all projects for a user.

---

## Project Configuration

**Location:** `./.fluxid/config.yaml` (in project root)

**Purpose:** Project-specific settings that override home config.

**Create:**
```bash
cd /path/to/project
fluxid init .
```

**What it creates:**
```
./.fluxid/
├── config.yaml
├── commands/
│   ├── implement.md
│   ├── review.md
│   └── commit.md
└── templates/
```

**Example config:**
```yaml
# Override agent for this project
agent: codex

# Override iterations
iterations: 10

# implement_retries will fall back to home config or default

commands:
  implement: /absolute/path/to/project/.fluxid/commands/implement.md
  review: /absolute/path/to/project/.fluxid/commands/review.md
  commit: /absolute/path/to/project/.fluxid/commands/commit.md
```

**Use case:** Project-specific overrides (different agent, tighter limits, custom prompts).

---

## Custom Configuration

**Status:** Flag parsed but integration in progress.

**Usage (when implemented):**
```bash
fluxid --claude --config=/path/to/custom-config.yaml --file=/path/to/task.md
```

**Use case:** Ad-hoc configuration for special workflows.

---

## Configuration Precedence

Settings are resolved **field-by-field** (not file-level override).

### Precedence Examples

**Example 1: Basic layering**
```
Built-in defaults:  agent=claude, iterations=20, implement_retries=3
Home config:        agent=claude, iterations=30
Project config:     (none)
Result:             agent=claude, iterations=30, implement_retries=3
```

**Example 2: Project overrides home**
```
Home config:        agent=claude, iterations=20, implement_retries=3
Project config:     agent=codex, iterations=10
Result:             agent=codex, iterations=10, implement_retries=3
```

**Example 3: CLI flags override everything**
```
Home config:        agent=claude, iterations=20
Project config:     agent=codex
CLI flags:          --opencode --fluxid-iterations=5
Result:             agent=opencode, iterations=5, implement_retries=3 (from default)
```

**Example 4: Commands are not merged**
```
Home config:
  commands:
    implement: /home/.fluxid/implement.md
    review: /home/.fluxid/review.md
    commit: /home/.fluxid/commit.md

Project config:
  commands:
    implement: /project/.fluxid/implement.md
    # review and commit not specified

Result: ERROR - project config must provide ALL command paths if it specifies any
```

**Rule:** Command file paths are **all-or-nothing** per config level. If a config file specifies any command path, it must specify all three.

---

## Built-in Defaults

When no config value is provided:

- **agent:** `claude`
- **iterations:** `20`
- **implement_retries:** `3`
- **commands:** **No default** - must be provided in config

---

## Environment Variables

### `FLUXID_SESSION_ID`
Session identifier for file operations.

**Default:** Auto-generated UUID if not set

**Usage:**
```bash
export FLUXID_SESSION_ID=feature-auth
fluxid --claude --file=/path/to/task.md
```

**Affects:**
- Report file location: `<session-root>/<session-id>/report.yaml`
- History file location: `<session-root>/<session-id>/history.yaml`

### `FLUXID_SESSION_ROOT`
Session storage root directory.

**Default:**
- `$HOME/.fluxid/sessions` (for home config users)
- `./.fluxid/sessions` (for project config users)

**Usage:**
```bash
export FLUXID_SESSION_ROOT=/custom/sessions
fluxid --claude --file=/path/to/task.md
```

**Affects:**
- Where session directories are created

---

## Configuration Validation

fluxid validates configuration at startup:

**Checks:**
1. At least one config file exists (home or project)
2. Config files are valid YAML
3. Required fields are present
4. Field values are correct types
5. Command file paths are absolute
6. Command files exist and are readable
7. Agent value is valid (`claude`, `codex`, or `opencode`)
8. Numeric values are positive

**Errors:**
- Missing required config: Exit with error message
- Invalid YAML: Exit with parse error
- Missing command files: Exit with file not found error
- Relative command paths: Exit with validation error

---

## Command Files

Command files are markdown templates passed to agents.

**Structure:**
```markdown
# Implement Changes

You are an expert software engineer. Your task is to implement the required changes.

## Task
{task_content}

## Requirements
- Write clean, tested code
- Follow project conventions
- Document changes

## Output
Write a report to: {report_file}
```

**Variables available:**
- `{task_content}` - Content of task file
- `{report_file}` - Path to report file
- `{history_file}` - Path to history file
- `{session_id}` - Current session ID
- `{iteration}` - Current iteration number

---

## Configuration Workflow

### First-Time Setup
```bash
# Create home config
fluxid init

# Customize if needed
vim ~/.fluxid/config.yaml
vim ~/.fluxid/commands/implement.md
```

### Per-Project Setup
```bash
cd /path/to/project

# Create project config
fluxid init .

# Customize for project
vim ./.fluxid/config.yaml
vim ./.fluxid/commands/implement.md
```

### Temporary Override
```bash
# Use different agent for this run
fluxid --codex --file=/path/to/task.md

# Use tighter iteration limit
fluxid --claude --file=/path/to/task.md --fluxid-iterations=5
```

---

## Common Patterns

### Development vs Production
```yaml
# Development config (~/.fluxid/config.yaml)
agent: claude
iterations: 50           # Generous iterations for exploration
implement_retries: 5     # Allow more attempts

# Production config (project/.fluxid/config.yaml)
agent: claude
iterations: 10           # Stricter limits
implement_retries: 2     # Fail faster
```

### Team Consistency
```yaml
# Committed to repo: .fluxid/config.yaml
agent: claude
iterations: 20
implement_retries: 3
commands:
  implement: /absolute/path/to/.fluxid/commands/implement.md
  review: /absolute/path/to/.fluxid/commands/review.md
  commit: /absolute/path/to/.fluxid/commands/commit.md

# Note: Use absolute paths or document setup process
```

### Testing
```yaml
# Test config with strict limits
agent: claude
iterations: 3            # Fast failure for tests
implement_retries: 1     # Single attempt
```

---

## Troubleshooting

### "No config file found"
**Cause:** Neither home nor project config exists.

**Fix:**
```bash
fluxid init  # Create home config
# or
fluxid init .  # Create project config
```

### "Command file not found"
**Cause:** Config references non-existent command file.

**Fix:**
1. Check path in config is absolute
2. Verify file exists: `ls /path/from/config`
3. Run `fluxid init` to regenerate templates

### "Command path must be absolute"
**Cause:** Config uses relative path (e.g., `./implement.md`).

**Fix:** Use absolute path:
```yaml
# Wrong
commands:
  implement: ./commands/implement.md

# Correct
commands:
  implement: /absolute/path/to/commands/implement.md
```

### "Invalid agent value"
**Cause:** Config specifies unsupported agent.

**Fix:** Use `claude`, `codex`, or `opencode`:
```yaml
agent: claude  # Must be one of these three
```
