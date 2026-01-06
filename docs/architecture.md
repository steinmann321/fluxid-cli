# Architecture

Codebase structure, design principles, and technical implementation details.

## Project Structure

```
fluxid-cli/
├── cmd/
│   └── fluxid/
│       └── main.go             # CLI entrypoint
├── internal/
│   ├── command/                # CLI command handlers
│   ├── config/                 # Configuration loading and validation
│   ├── storage/                # File-based storage operations
│   ├── assets/                 # Embedded schemas and templates
│   ├── output/                 # Output formatters (text/JSON/YAML)
│   ├── workflow/               # Core workflow orchestration
│   ├── errors/                 # Error definitions
│   ├── stream/                 # Streaming output formatters
│   └── types/                  # Shared type definitions
├── e2e-tests/                  # End-to-end integration tests
│   └── tests/
├── hooks/                      # Git hooks (dev use only)
├── specs/                      # Feature specifications
├── docs/                       # Documentation
├── Makefile                    # Build automation
├── go.mod                      # Go module definition
└── README.md                   # Quick start guide
```

---

## Design Principles

### 1. Pure Go Runtime
- No shell script dependencies at runtime
- Shell scripts limited to `hooks/` for development only
- Makefile for build automation, not runtime

### 2. Explicit Interfaces
- Clear boundaries between layers
- No implicit behavior
- Dependency injection where appropriate

### 3. File-Based Communication
- Replace stdio IPC with file operations
- Deterministic file paths
- YAML for structured data

### 4. Fail Fast
- Validate configuration at startup
- Clear error messages with actionable guidance
- Exit immediately on unrecoverable errors

### 5. Stateless Commands
- Each command run is independent
- State persisted in session files only
- No hidden global state

---

## Layer Architecture

```
┌─────────────────────────────────────┐
│         cmd/fluxid/main.go          │  CLI Entry
│    (delegates to command layer)     │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│      internal/command/              │  Command Layer
│  - Parse args                       │  - CLI interface
│  - Load config                      │  - Subcommand routing
│  - Execute workflow                 │  - Output formatting
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│      internal/workflow/             │  Workflow Layer
│  - Implement/Review/Commit loop     │  - Business logic
│  - Retry management                 │  - Phase orchestration
│  - Agent execution                  │  - Loop control
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│      internal/storage/              │  Storage Layer
│  - Read/write reports               │  - File operations
│  - Read/write history               │  - YAML validation
│  - Schema validation                │  - Security checks
└─────────────────────────────────────┘
```

---

## Key Components

### CLI Layer (`internal/command/`)

**Responsibility:** Parse arguments, route to appropriate handlers.

**Files:**
- `root.go` - Main command execution and routing
- `args.go` - Argument parsing
- `init.go` - Init command handler
- `report.go` - Report subcommand handlers
- `history.go` - History subcommand handlers
- `root_*.go` - Helper modules for main command

**Key functions:**
- `Execute()` - Main entry point
- `handleReportCommand()` - Report subcommand router
- `handleHistoryCommand()` - History subcommand router
- `handleInit()` - Init command implementation

**Design:**
- Thin layer, delegates to workflow or storage
- No business logic
- Focuses on CLI concerns (parsing, help, errors)

---

### Configuration Layer (`internal/config/`)

**Responsibility:** Load, validate, and resolve configuration.

**Files:**
- `config.go` - Core config loading logic
- `validation.go` - Config validation
- `resolve_commands.go` - Command file path resolution

**Key types:**
```go
type Config struct {
    Agent              string
    Iterations         int
    ImplementRetries   int
    Commands           Commands
    // ... other fields
}

type Commands struct {
    Implement string  // Absolute path
    Review    string  // Absolute path
    Commit    string  // Absolute path
}
```

**Precedence:**
```
Defaults → Home Config → Project Config → CLI Flags
```

**Design:**
- Field-by-field precedence (not file-level)
- Validates paths are absolute and files exist
- Fails fast on misconfiguration

---

### Storage Layer (`internal/storage/`)

**Responsibility:** File-based data persistence and validation.

**Files:**
- `report.go` - Report read/write operations
- `history.go` - History read/write with FIFO eviction
- `validate.go` - Schema validation orchestration
- `validate_report.go` - Report-specific validation
- `validate_history.go` - History-specific validation
- `schema.go` - Schema loading from embedded assets
- `security.go` - YAML security checks
- `paths.go` - Session path resolution

**Key functions:**
```go
// Report operations
func ReadReport(sessionID, sessionRoot string) (*Report, error)
func ValidateReport(sessionID, sessionRoot string) error

// History operations
func ReadHistory(sessionID, sessionRoot string) (History, error)
func ValidateHistory(sessionID, sessionRoot string) error

// Path resolution
func ResolveSessionPath(sessionID, filename, sessionRoot string) (string, error)
```

**Features:**
- **FIFO eviction:** History files > 10MB auto-trim oldest 30%
- **Strict validation:** No additional properties allowed
- **Security:** Rejects YAML anchors/aliases/merge keys
- **Path safety:** UUID validation, no path traversal

---

### Workflow Layer (`internal/workflow/`)

**Responsibility:** Orchestrate implement/review/commit loops.

**Files:**
- `workflow.go` - Main workflow logic
- `workflow_helpers.go` - Workflow utilities

**Key types:**
```go
type Phase int
const (
    PhaseImplement Phase = iota
    PhaseCommit
    PhaseReview
)
```

**Flow:**
```
Start
  │
  ↓
Implement ──FAIL──> Retry (up to N times)
  │
  │ PASS
  ↓
Commit
  │
  ↓
Review ──FAIL──> Loop (up to M iterations)
  │
  │ PASS
  ↓
Done
```

**Features:**
- Configurable retry limits
- Configurable iteration counts
- Dry-run simulation mode
- Phase state management

---

### Assets Layer (`internal/assets/`)

**Responsibility:** Embed schemas, templates, defaults.

**Files:**
- `assets.go` - Template loading
- `schemas.go` - Schema embedding

**Embedded assets:**
```go
//go:embed templates/report-schema.yaml
var ReportSchemaYAML string

//go:embed templates/history-schema.yaml
var HistorySchemaYAML string
```

**Design:**
- All templates compiled into binary
- No runtime file dependencies
- Schemas versioned with binary

---

### Output Layer (`internal/output/`)

**Responsibility:** Format output (text/JSON/YAML).

**Features:**
- Text: Human-readable progress indicators
- JSON: Machine-parseable structured data
- YAML: Human-readable structured data

**Example outputs:**
```json
{
  "session_id": "abc123",
  "status": "success",
  "iterations": 5
}
```

---

## Data Flow

### Workflow Execution

```
1. User runs: fluxid --claude --file=/path/to/task.md

2. Command layer:
   - Parse args
   - Load config
   - Validate task file exists

3. Workflow layer:
   - Set FLUXID_SESSION_ID environment variable
   - Execute IMPLEMENT phase:
     → Launch agent with task file
     → Agent writes report.yaml
     → Read and validate report
     → If FAIL: retry (up to N times)

   - Execute COMMIT phase:
     → Create git commit

   - Execute REVIEW phase:
     → Launch agent with review prompt
     → Agent writes report.yaml
     → Read and validate report
     → If FAIL: loop to IMPLEMENT (up to M iterations)

4. Storage layer:
   - Read report: <session-root>/<session-id>/report.yaml
   - Read history: <session-root>/<session-id>/history.yaml
   - Validate YAML structure
   - Check schema compliance

5. Output results to user
```

---

## File-Based Interface

### Session Directory Structure

```
$HOME/.fluxid/sessions/           # Or custom via FLUXID_SESSION_ROOT
└── a1b2c3d4-e5f6-7890-abcd/     # Session ID (UUID)
    ├── report.yaml               # Current phase report
    └── history.yaml              # Workflow event history
```

### Report File

**Purpose:** Agent communicates phase status and findings.

**Schema:** `internal/assets/templates/report-schema.yaml`

**Security:**
- Strict schema validation (no additional properties)
- No YAML anchors/aliases/merge keys
- Size limit: 10MB

### History File

**Purpose:** Track workflow events across iterations.

**Schema:** `internal/assets/templates/history-schema.yaml`

**Features:**
- FIFO eviction when > 10MB (removes oldest 30%)
- Append-only semantic
- Per-event strict validation

---

## Error Handling

### Error Categories

```go
// internal/errors/errors.go
type ComponentError struct {
    Component   string  // config, args, workflow, etc.
    Description string
}
```

**Format:**
```
error: component: description
```

**Examples:**
```
error: config: home config file not found: ~/.fluxid/config.yaml
error: args: required flag --file not provided
error: workflow: agent execution failed: exit code 1
error: [file]: malformed YAML (expected: valid YAML, got: parse error)
error: [status]: invalid value (expected: enum[PASS,FAIL], got: PENDING)
```

**Design:**
- Consistent format across all components
- Actionable error messages
- Field-level validation errors for schemas

---

## Testing Strategy

### Unit Tests
- Location: `*_test.go` files alongside implementation
- Coverage target: >90%
- Focus: Individual functions, edge cases

### Integration Tests
- Location: `e2e-tests/tests/`
- Scope: Full CLI workflow execution
- Features: Real agent stubs, file I/O, validation

### Test Helpers
- `e2e-tests/tests/test_helpers*.go`
- Shared utilities for E2E tests
- Build binary, create stubs, run commands

---

## Security

### YAML Security

**Threat:** YAML deserialization attacks via anchors/aliases.

**Mitigation:**
- `internal/storage/security.go` - Pre-validation security checks
- Reject files containing `&`, `*`, `<<:`
- Fail immediately on detection

**Code:**
```go
func ValidateYAMLSecurity(filePath string) error {
    // Scan for anchors (&), aliases (*), merge keys (<<:)
    // Return error if found
}
```

### Path Security

**Threat:** Path traversal attacks via session IDs.

**Mitigation:**
- `internal/storage/paths.go` - Path validation
- UUID validation on session IDs
- Reject `../`, absolute paths in session ID
- Construct paths safely

**Code:**
```go
func ResolveSessionPath(sessionID, filename, root string) (string, error) {
    // Validate sessionID is safe
    // Build path: root/sessionID/filename
    // Return absolute path
}
```

### Input Validation

**All user inputs validated:**
- Config files: YAML schema validation
- CLI args: Type and range checks
- Session IDs: UUID format or user-provided string (no path components)
- File paths: Absolute path requirements

---

## Performance Considerations

### File I/O

- **Reports:** Read once per phase (cached in memory during phase)
- **History:** Read once at workflow start, append as needed
- **FIFO eviction:** Only when file > 10MB (infrequent)

### Memory

- **Schemas:** Embedded in binary, loaded once
- **Templates:** Embedded in binary, loaded on-demand
- **History:** Full file loaded (10MB max after eviction)

### Concurrency

- **Single workflow:** One active phase at a time
- **No parallelism:** Sequential phase execution
- **Race-free:** No shared state between invocations

---

## Extensibility

### Adding New Agents

1. Add agent name to config validation
2. Update documentation
3. No code changes required (agents are external processes)

### Adding New Phases

1. Define phase constant in `internal/workflow/`
2. Implement phase execution logic
3. Update workflow loop
4. Add tests

### Custom Commands

Users can customize command templates:
- Edit `.fluxid/commands/*.md`
- Use variables: `{task_content}`, `{report_file}`, etc.
- No code changes to fluxid

---

## Build and Release

### Build Process

```bash
make build  # Produces bin/fluxid
```

**Artifacts:**
- Single binary: `bin/fluxid`
- No runtime dependencies
- Embeds all templates and schemas

### Installation

```bash
go install ./cmd/fluxid  # Installs to $GOPATH/bin
```

---

## Future Architecture Considerations

### Planned Improvements

1. **Custom config via `--config` flag:** Parse and integrate custom config files
2. **Command file overrides:** Allow per-run command file customization
3. **Plugin system:** External validators, custom phases
4. **Remote sessions:** Share sessions across machines
5. **Session cleanup:** Auto-delete old sessions

### Design Constraints

- **Maintain pure Go:** No runtime shell dependencies
- **Keep file-based interface:** No return to stdio IPC
- **Preserve security:** Strict validation always
- **Stay stateless:** No persistent daemon or database
