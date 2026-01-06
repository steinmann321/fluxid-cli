# fluxid

A workflow controller that orchestrates coding agents through structured IMPLEMENT → REVIEW → COMMIT loops, enabling agents to break through context window limits.

## Overview

**fluxid** lets coding agents (Claude, Codex, OpenCode) tackle complex tasks that exceed their context limits. It runs your agent through structured implement-review cycles, automatically tracking progress, validating outputs, and managing state between iterations until the task is complete.

**Key Features:**
- Orchestrates implement/review/commit cycles automatically
- File-based interface for agent communication
- Session-scoped storage with YAML validation
- Configurable retry limits and iteration counts
- Dry-run mode for workflow simulation

## Quick Start

**Install:**
```bash
# Build from source
make build

# Or install to $GOPATH/bin
go install ./cmd/fluxid
```

**Initialize:**
```bash
# Create home config and templates
fluxid init

# Or create project-specific config
fluxid init .
```

**Run a workflow:**
```bash
fluxid --claude --file=/path/to/task.md
```

## Workflow Overview

```
┌─────────────┐
│  IMPLEMENT  │  Agent implements changes
└──────┬──────┘
       │
       ↓
┌─────────────┐
│   COMMIT    │  Create git commit
└──────┬──────┘
       │
       ↓
┌─────────────┐
│   REVIEW    │  Agent evaluates quality and completeness
└──────┬──────┘
       │
       ├─ PASS → Workflow complete
       └─ FAIL → Return to IMPLEMENT (up to N iterations)
```

**How it works:**
1. Agent receives task via command file
2. Agent implements changes, writes `report.yaml` with status
3. If PASS: workflow creates commit and proceeds to review
4. If FAIL: workflow retries implementation (up to configured limit)
5. Review phase evaluates quality and completeness, loops back to IMPLEMENT if defects found

## Usage

**Basic commands:**
```bash
# Run workflow
fluxid {--claude|--codex|--opencode} --file=PATH [options]

# Initialize configuration
fluxid init [PATH]

# Agent file operations
fluxid report --get-file      # Get report path
fluxid report --validate      # Validate report
fluxid report --get-schema    # Get YAML schema
fluxid history --get-file     # Get history path
fluxid history --validate     # Validate history
fluxid history --get-schema   # Get YAML schema
```

**Common options:**
```bash
--file=PATH                    # Task file path (required)
--fluxid-iterations=N          # Max review cycles (default: 20)
--fluxid-implement-retries=R   # Max retries per cycle (default: 3)
--fluxid-dry-run               # Simulate without execution
--fluxid-output=FORMAT         # Output format: text|json|yaml
```

## Documentation

- **[CLI Reference](docs/cli-reference.md)** - Complete command-line options
- **[Configuration](docs/configuration.md)** - Config files and precedence
- **[Workflows](docs/workflows.md)** - Common usage patterns and examples
- **[Agent Integration](docs/agent-integration.md)** - File-based interface for agents
- **[Architecture](docs/architecture.md)** - Codebase structure and design

## Configuration

fluxid requires at least one config file:
- `~/.fluxid/config.yaml` (home config), OR
- `./.fluxid/config.yaml` (project config)

Create with `fluxid init`. See [Configuration](docs/configuration.md) for details.

## File-Based Interface

Agents communicate via file operations:

```bash
# Agent gets file paths
REPORT=$(fluxid report --get-file)
HISTORY=$(fluxid history --get-file)

# Agent writes YAML
cat > "$REPORT" <<EOF
command: implement
artifact: src/main.go
timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
EOF

# Optional validation
fluxid report --validate
```

See [Agent Integration](docs/agent-integration.md) for complete guide.

## Development

```bash
# Build
make build

# Run tests
make test

# Run with coverage
make coverage

# Clean artifacts
make clean
```

## License

See LICENSE file for details.
