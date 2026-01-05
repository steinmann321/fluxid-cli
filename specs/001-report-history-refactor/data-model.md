# Data Model: Report & History File-Based Interface

**Feature**: `001-report-history-refactor`
**Date**: 2026-01-05
**Phase**: 1 - Design & Contracts

## Overview

This document defines the data structures for report and history files in the new file-based interface. These structures replace the stdio-based IPC protocol.

## Entities

### 1. Report File

**Purpose**: Captures the outcome of a single workflow phase (implement or review). Written by external agents, read by fluxid workflow controller.

**File Location**: `<session-root>/<session-id>/report.yaml`

**Lifecycle**: Overwritten each workflow phase. Not preserved across phases.

**Structure**:

```yaml
# Required fields
command: string             # Command that generated the report (e.g., "fluxid implement")
artifact: string            # Primary artifact produced (e.g., "src/main.go")
timestamp: string           # ISO 8601 timestamp (e.g., "2026-01-05T10:30:00Z")
status: string              # Enum: [PASS, FAIL]

# Required nested object
issues:
  blockers: [string]        # Critical issues preventing progress
  defects: [string]         # Bugs or incorrect behavior
  concerns: [string]        # Potential problems requiring attention
  observations: [string]    # Notable findings (not problems)
  enhancements: [string]    # Improvement suggestions

# Optional fields
next_steps: [string]        # Recommended follow-up actions
summary: string             # Human-readable summary of report

# Extensibility
# Additional properties allowed for agent-specific metadata
```

**Validation Rules**:

| Field | Required | Type | Constraints |
|-------|----------|------|-------------|
| command | Yes | string | Non-empty |
| artifact | Yes | string | Non-empty |
| timestamp | Yes | string | Valid ISO 8601 format |
| status | Yes | string | Exactly "PASS" or "FAIL" |
| issues | Yes | object | Must contain all 5 categories |
| issues.blockers | Yes | array | May be empty array |
| issues.defects | Yes | array | May be empty array |
| issues.concerns | Yes | array | May be empty array |
| issues.observations | Yes | array | May be empty array |
| issues.enhancements | Yes | array | May be empty array |
| next_steps | No | array | Items must be strings |
| summary | No | string | - |

**State Transitions**:

```
Agent writes report → fluxid validates → fluxid reads status
  ├─ status: PASS  → Workflow exits loop
  └─ status: FAIL  → Workflow retries with history context
```

**Example - PASS Report**:

```yaml
command: fluxid implement
artifact: internal/storage/report.go
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns:
    - "File validation error messages could be more specific"
  observations:
    - "Used xeipuuv/gojsonschema for JSON Schema validation"
  enhancements:
    - "Consider adding schema caching for performance"
next_steps:
  - "Run E2E tests to verify integration"
  - "Update documentation with new commands"
summary: "Successfully implemented report file operations with JSON Schema validation"
```

**Example - FAIL Report**:

```yaml
command: fluxid review
artifact: internal/storage/validate.go
timestamp: 2026-01-05T15:45:22Z
status: FAIL
issues:
  blockers:
    - "Test TestValidateReport_InvalidStatus failing due to missing error message"
  defects:
    - "Validation doesn't check for YAML anchors/aliases (security requirement)"
  concerns:
    - "Error messages don't include field paths for nested validation failures"
  observations:
    - "JSON Schema validator provides detailed errors but we're not surfacing them"
  enhancements: []
next_steps:
  - "Fix security validation to reject YAML anchors before parsing"
  - "Update error formatter to include JSON Schema field paths"
  - "Add test cases for anchor/alias rejection"
summary: "Review identified security gap and test failure requiring fixes"
```

---

### 2. History File

**Purpose**: Records workflow events across multiple phases to prevent repeating failed approaches. Append-only log written by external agents, read by fluxid workflow controller.

**File Location**: `<session-root>/<session-id>/history.yaml`

**Lifecycle**: Append-only. Survives across workflow phases. FIFO eviction at 10MB limit.

**Structure**:

```yaml
# Array of history events
- timestamp: string         # ISO 8601 timestamp
  step: string             # Workflow step name (e.g., "implement", "review")
  status: string           # Enum: [SUCCESS, FAIL]
  summary: string          # Brief outcome description
  details: string          # Optional: detailed approach and failure reason

- timestamp: string
  step: string
  status: string
  summary: string
  details: string

# ... more entries
```

**Validation Rules**:

| Field | Required | Type | Constraints |
|-------|----------|------|-------------|
| (root) | Yes | array | Array of event objects |
| timestamp | Yes | string | Valid ISO 8601 format |
| step | Yes | string | Non-empty |
| status | Yes | string | Exactly "SUCCESS" or "FAIL" |
| summary | Yes | string | Non-empty |
| details | No | string | - |

**FIFO Eviction Logic**:

```
1. Before reading history file:
   a. Check file size
   b. If size > 10MB:
      - Read entire file
      - Parse as YAML array
      - Calculate per-entry size
      - Remove entries from index 0 (oldest) until total size < 10MB
      - Write truncated array back to file
   c. Return history array (truncated or full)
```

**Event Ordering**: Chronological (oldest first). Newest entries appended to end of array.

**Example - History File**:

```yaml
- timestamp: 2026-01-05T10:00:00Z
  step: implement
  status: FAIL
  summary: "Initial implementation attempted using channels for file notification"
  details: "Tried implementing file change watcher using Go channels and fsnotify. Failed because requirements specify strictly sequential workflow with no concurrency. Channels introduce async behavior incompatible with constitution principle III."

- timestamp: 2026-01-05T11:15:00Z
  step: implement
  status: FAIL
  summary: "Second attempt using file polling with goroutines"
  details: "Attempted file polling with background goroutines to check report file existence. Failed review because goroutines violate sequential workflow requirement. Constitution check rejected parallel operations."

- timestamp: 2026-01-05T13:30:00Z
  step: implement
  status: SUCCESS
  summary: "Successfully implemented file-based report reading with direct file I/O"
  details: "Removed all async operations. Workflow calls storage.ReadReport() synchronously after agent exits. Simple file read with error handling. Passes constitution check."

- timestamp: 2026-01-05T14:00:00Z
  step: review
  status: SUCCESS
  summary: "Review passed with minor concerns addressed"
  details: "Added file existence checks and improved error messages. All E2E tests passing."
```

---

### 3. Session Context

**Purpose**: Scopes report/history files to a single workflow execution. Prevents cross-session contamination.

**Representation**: String identifier (UUID v4)

**Source**: `FLUXID_SESSION_ID` environment variable (set by fluxid workflow, read by agents and commands)

**Usage**:

```go
// In fluxid workflow:
sessionID := uuid.New().String()
os.Setenv("FLUXID_SESSION_ID", sessionID)

// In report/history commands:
sessionID := os.Getenv("FLUXID_SESSION_ID")
if sessionID == "" {
    return errors.New("FLUXID_SESSION_ID environment variable not set")
}

// File path resolution:
reportPath := filepath.Join(sessionRoot, sessionID, "report.yaml")
historyPath := filepath.Join(sessionRoot, sessionID, "history.yaml")
```

**Security Constraints**:
- Session ID used in file paths MUST be validated (no path traversal characters: `..`, `/`, `\`)
- Resolved paths MUST be within session root directory
- Symlinks resolved before path validation

---

### 4. Report Schema (JSON Schema)

**Purpose**: Defines structure for report validation. Embedded in fluxid binary, output via `fluxid report --get-schema`.

**Format**: JSON Schema Draft 7

**Location**: Embedded in `internal/assets/schemas/report.json`, exposed via `internal/storage/schema.go`

**Key Properties**:
- `$schema`: "http://json-schema.org/draft-07/schema#"
- `required`: ["command", "artifact", "timestamp", "status", "issues"]
- `additionalProperties`: true (for extensibility per FR-030)
- Enum validation for `status` field
- Nested object validation for `issues` structure
- ISO 8601 format validation for `timestamp`

---

### 5. History Schema (JSON Schema)

**Purpose**: Defines structure for history validation. Embedded in fluxid binary, output via `fluxid history --get-schema`.

**Format**: JSON Schema Draft 7

**Location**: Embedded in `internal/assets/schemas/history.json`, exposed via `internal/storage/schema.go`

**Key Properties**:
- `$schema`: "http://json-schema.org/draft-07/schema#"
- `type`: "array"
- `items`: Object schema for history events
- `items.required`: ["timestamp", "step", "status", "summary"]
- `additionalProperties`: false (stricter validation for internal format)
- Enum validation for `status` field
- ISO 8601 format validation for `timestamp`

---

## Relationships

```
Session Context (1) ──┬── (1) Report File
                      └── (1) History File

Report File (1) ── validated by ── (1) Report Schema
History File (1) ── validated by ── (1) History Schema

Workflow Controller ── reads ──> Report File
Workflow Controller ── reads ──> History File
External Agent ── writes ──> Report File
External Agent ── appends ──> History File

Report File ── contains ── Issues (5 categories)
History File ── contains ── History Events (array)
```

---

## Data Flow

### Report Workflow

```
1. Agent queries file path:
   $ fluxid report --get-file
   → /path/to/session-root/<session-id>/report.yaml

2. Agent writes report file:
   $ echo "command: fluxid implement\n..." > /path/to/report.yaml

3. Agent validates report (optional):
   $ fluxid report --validate
   → Exit 0 (success) or Exit 1 (validation errors)

4. Workflow reads report:
   report, err := storage.ReadReport(sessionID)
   if report.Status == "PASS" {
       // Exit loop
   } else {
       // Retry with history
   }
```

### History Workflow

```
1. Agent queries file path:
   $ fluxid history --get-file
   → /path/to/session-root/<session-id>/history.yaml

2. Agent appends history entry:
   $ echo "- timestamp: 2026-01-05T10:00:00Z\n..." >> /path/to/history.yaml

3. Agent validates history (optional):
   $ fluxid history --validate
   → Exit 0 (success) or Exit 1 (validation errors)

4. Workflow reads history:
   history, err := storage.ReadHistory(sessionID)
   // Pass history to agent for context in next iteration
```

---

## File Size Constraints

| File | Max Size | Behavior on Exceed |
|------|----------|-------------------|
| Report | 10MB | Reject read with error (reports should never exceed 10MB) |
| History | 10MB | FIFO eviction before read (remove oldest entries) |

**Rationale**: Reports are single-phase outputs (bounded size). History grows over time and needs eviction.

---

## Security Properties

### Path Validation

```go
// Pseudo-code for path validation
func validateSessionPath(sessionID string, sessionRoot string) error {
    // 1. Validate session ID format (UUID only)
    if !isValidUUID(sessionID) {
        return errors.New("invalid session ID format")
    }

    // 2. Construct path
    path := filepath.Join(sessionRoot, sessionID)

    // 3. Resolve symlinks
    resolved, err := filepath.EvalSymlinks(path)
    if err != nil && !os.IsNotExist(err) {
        return err
    }

    // 4. Verify within session root
    if !strings.HasPrefix(resolved, sessionRoot) {
        return errors.New("path traversal detected")
    }

    return nil
}
```

### YAML Security

```go
// Pseudo-code for YAML security checks
func readSecureYAML(filePath string, v interface{}) error {
    // 1. Check file size
    stat, err := os.Stat(filePath)
    if err != nil {
        return err
    }
    if stat.Size() > 10*1024*1024 {  // 10MB
        return errors.New("file size exceeds 10MB limit")
    }

    // 2. Read file content
    content, err := os.ReadFile(filePath)
    if err != nil {
        return err
    }

    // 3. Check for dangerous YAML features
    if bytes.Contains(content, []byte("&")) {
        return errors.New("YAML anchors not allowed")
    }
    if bytes.Contains(content, []byte("*")) {
        return errors.New("YAML aliases not allowed")
    }
    if bytes.Contains(content, []byte("<<")) {
        return errors.New("YAML merge keys not allowed")
    }

    // 4. Parse with strict decoding
    decoder := yaml.NewDecoder(bytes.NewReader(content))
    decoder.KnownFields(true)  // Reject unknown fields
    return decoder.Decode(v)
}
```

---

## Migration Notes

### Old IPC Storage Format

**Location**: `os.TempDir()/fluxid-reports/<session-id>.yaml`

**Structure**: Same YAML structure as new report format

**Migration**: File location changes, but YAML structure remains compatible. Agents need to:
1. Replace `fluxid ipc write-report` with direct file write to path from `fluxid report --get-file`
2. Replace `fluxid ipc read-report` with `fluxid report --validate` (optional)
3. Update to use `FLUXID_SESSION_ID` environment variable instead of passing session ID as flag

### Removed Entities

- **IPC Storage Abstraction**: Removed entire `internal/ipc/storage.go` with functions:
  - `WriteReport()`, `ReadReport()`, `SetAbortFlag()`, `CheckAbortFlag()`, `ClearAbortFlag()`, `WriteHistoryEntry()`, `ReadHistory()`, `ClearHistory()`
- **Abort Flag**: `.abort` file removed. Abort mechanism moved to signal handler (out of scope for this refactor, separate evaluation needed per spec)

---

## Next Steps

Data model complete. Continue Phase 1:
- JSON Schema files already exist at `internal/assets/templates/report-schema.yaml` and `internal/assets/templates/history-schema.yaml`
- Generate quickstart guide for agent integration
- Update agent context files
