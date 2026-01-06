# fluxid

A workflow controller that orchestrates coding agents through structured IMPLEMENT → REVIEW → COMMIT loops, enabling agents to break through context window limits.

## Overview

**fluxid** is a CLI tool that wraps coding agents (Claude, Codex, OpenCode) and manages iterative development workflows. It handles session management, tracks workflow history, validates agent outputs, and provides file-based communication primitives.

**Key Features:**
- Orchestrates implement/review/commit cycles automatically
- File-based interface for agent communication (no stdio IPC)
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
│   REVIEW    │  Agent reviews implementation
└──────┬──────┘
       │
       ├─ PASS → Workflow complete
       └─ FAIL → Return to IMPLEMENT (up to N iterations)
```

**How it works:**
1. Agent receives task via command file
2. Agent implements, writes `report.yaml` with status
3. If PASS: workflow creates commit and proceeds to review
4. If FAIL: workflow retries implementation
5. Review phase evaluates quality, loops back if needed

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

Agents communicate via file operations instead of stdio:

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
