# Research: Report & History File-Based Interface

**Feature**: `001-report-history-refactor`
**Date**: 2026-01-05
**Phase**: 0 - Research & Technology Selection

## Overview

This document captures research findings and technology decisions for replacing the stdio-based IPC system with a file-based report/history interface.

## Research Areas

### 1. JSON Schema Validation Library Selection

**Decision**: Use `github.com/xeipuuv/gojsonschema`

**Rationale**:
- Most popular Go JSON Schema validator (3.6k GitHub stars)
- Supports JSON Schema Draft 7 (current standard)
- Active maintenance (last commit within 6 months)
- Provides detailed validation error messages with field paths
- Zero external C dependencies (pure Go)
- Already used in Go ecosystem (kubernetes, docker, etc.)

**Alternatives Considered**:
1. `github.com/santhosh-tekuri/jsonschema` - Good performance but less detailed error messages
2. `github.com/qri-io/jsonschema` - Newer but less mature, smaller community
3. Manual validation - Would require significant effort and testing, doesn't leverage standard

**Implementation Notes**:
- Will validate YAML by unmarshaling to interface{} then validating against JSON Schema
- Error messages include field paths (e.g., "issues.blockers: field is required")
- Can embed schema as string or load from file (will use embedded for binary portability)

### 2. YAML Parser Security Configuration

**Decision**: Use existing `gopkg.in/yaml.v3` with security constraints configured

**Rationale**:
- Already a project dependency
- Supports disabling dangerous features via `KnownFields` and custom unmarshaler
- Requirements FR-011, FR-025 mandate rejecting YAML anchors, aliases, merge keys

**Security Constraints**:
```go
// Disable anchors, aliases, merge keys to prevent complexity attacks
decoder := yaml.NewDecoder(file)
decoder.KnownFields(true)  // Reject unknown fields
// Check for anchors/aliases by scanning for & and * characters before parsing
// Reject files containing YAML advanced features
```

**Alternatives Considered**:
1. Switch to `github.com/goccy/go-yaml` - Has performance benefits but would require migration
2. JSON-only format - Would break existing agent integrations expecting YAML

**Threat Model**:
- YAML billion laughs attack via anchors: `&x [*x, *x, ...]` causes exponential expansion
- Alias complexity attacks can cause parser to hang or consume excessive memory
- Merge keys can obscure actual data structure
- Prevention: Reject files containing `&`, `*`, `<<` at parse time with clear error

### 3. File Size Management Strategy

**Decision**: Implement FIFO eviction for history files before reading if size > 10MB

**Rationale**:
- Requirements FR-024 specifies 10MB limit with FIFO eviction
- Report files are single-shot (one per phase) so size limit is sufficient
- History files grow over time and need eviction strategy

**Implementation Approach**:
```go
// Before reading history:
1. Check file size
2. If > 10MB:
   a. Read entire file
   b. Parse as YAML array
   c. Remove oldest entries until size < 10MB
   d. Write truncated array back
   e. Return truncated history
```

**Alternatives Considered**:
1. Streaming parser - Complex for YAML array format, not needed for 10MB files
2. Line-based truncation - Would break YAML structure if entry spans multiple lines
3. Hard error on size limit - Requirement specifies eviction, not rejection

**Memory Safety**:
- 10MB fully loaded into memory is acceptable (modern systems have GB+ RAM)
- Go GC will reclaim memory after processing
- Size check happens before read to prevent OOM on malicious large files

### 4. Session Directory Structure

**Decision**: Use `<session-root>/<session-id>/` directory structure

**Rationale**:
- Requirements FR-010, FR-022 specify session-specific scoping
- Existing system uses `os.TempDir()/fluxid-reports/<session-id>.*`
- New system should use deterministic, user-controllable location

**Proposed Structure**:
```
<session-root>/              # Configured per-project or ~/.fluxid/sessions/
└── <session-id>/
    ├── report.yaml          # Current phase report (overwritten each phase)
    └── history.yaml         # Append-only history (FIFO eviction at 10MB)
```

**Path Resolution**:
1. Check `FLUXID_SESSION_ROOT` environment variable
2. Fall back to `.fluxid/sessions/` in current working directory
3. Fall back to `~/.fluxid/sessions/`

**Security**:
- Requirements FR-012, FR-026 mandate path validation within session directory
- Prevent path traversal: resolve symlinks, validate no `..` components
- Only fluxid creates/manages paths (agents cannot specify paths)

**Alternatives Considered**:
1. Continue using `os.TempDir()` - Not user-inspectable, harder to debug
2. Current working directory - Would clutter project root
3. XDG Base Directory - Overkill for simple session data

### 5. JSON Schema Structure for Report

**Decision**: Convert existing YAML schema to JSON Schema Draft 7

**Current Schema** (from `internal/ipc/schema.yaml`):
```yaml
command: string (required)
artifact: string (required)
timestamp: ISO 8601 string (required)
status: enum [PASS, FAIL] (required)
issues: object (required)
  blockers: array of strings
  defects: array of strings
  concerns: array of strings
  observations: array of strings
  enhancements: array of strings
next_steps: array of strings (optional)
summary: string (optional)
```

**JSON Schema Design**:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["command", "artifact", "timestamp", "status", "issues"],
  "properties": {
    "command": {"type": "string"},
    "artifact": {"type": "string"},
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 timestamp"
    },
    "status": {
      "type": "string",
      "enum": ["PASS", "FAIL"]
    },
    "issues": {
      "type": "object",
      "required": ["blockers", "defects", "concerns", "observations", "enhancements"],
      "properties": {
        "blockers": {"type": "array", "items": {"type": "string"}},
        "defects": {"type": "array", "items": {"type": "string"}},
        "concerns": {"type": "array", "items": {"type": "string"}},
        "observations": {"type": "array", "items": {"type": "string"}},
        "enhancements": {"type": "array", "items": {"type": "string"}}
      }
    },
    "next_steps": {"type": "array", "items": {"type": "string"}},
    "summary": {"type": "string"}
  },
  "additionalProperties": true
}
```

**Rationale for `additionalProperties: true`**: Requirement FR-030 specifies "allow additional properties for extensibility"

### 6. JSON Schema Structure for History

**Decision**: Define history as array of event objects

**Design** (based on requirements FR-031, FR-032, FR-033):
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["timestamp", "step", "status", "summary"],
    "properties": {
      "timestamp": {
        "type": "string",
        "format": "date-time",
        "description": "ISO 8601 timestamp"
      },
      "step": {
        "type": "string",
        "description": "Workflow step name"
      },
      "status": {
        "type": "string",
        "enum": ["SUCCESS", "FAIL"]
      },
      "summary": {
        "type": "string",
        "description": "Brief outcome description"
      },
      "details": {
        "type": "string",
        "description": "Detailed approach and failure reason"
      }
    },
    "additionalProperties": false
  }
}
```

**Rationale**:
- Array format allows append-only writes by agents
- Each entry is independent (no cross-entry references)
- FIFO eviction removes oldest entries (array index 0)
- `additionalProperties: false` for stricter validation (history is internal format)

### 7. Error Message Format

**Decision**: Use structured error messages with field paths

**Format**:
```
Validation failed for <file-type>:
  - <field-path>: <error-message>
  - <field-path>: <error-message>

Expected schema available via: fluxid <file-type> --get-schema
```

**Example**:
```
Validation failed for report:
  - status: must be one of [PASS, FAIL], got "UNKNOWN"
  - issues.blockers: field is required but missing
  - timestamp: must be valid ISO 8601 format, got "2024-13-45"

Expected schema available via: fluxid report --get-schema
```

**Rationale**:
- Field paths make errors actionable (agent knows exactly what to fix)
- Reference to --get-schema helps agents self-correct
- Multiple errors shown at once (don't fail fast on first error)
- Aligns with requirement FR-005 (instructive error messages)

### 8. Command Exit Codes

**Decision**: Use standard Unix exit code conventions

**Exit Codes**:
- `0`: Success (file valid, schema retrieved, file path returned)
- `1`: Validation failed (invalid YAML structure)
- `2`: File not found or permission error
- `3`: Session ID missing or invalid
- `4`: Internal error (schema missing, unexpected failure)

**Rationale**:
- Standard Unix conventions for scriptability
- Different codes allow agents to distinguish error types
- Aligns with requirement FR-006, FR-007 (exit code 0 for success, non-zero for failure)

### 9. Backward Compatibility Approach

**Decision**: Hard breaking change with no compatibility layer

**Rationale**:
- User explicitly requested "breaking change" with "no backward compatibility"
- Spec.md migration path states "hard breaking change" preferred for simplicity
- Old `fluxid ipc` commands return "command not found" error
- Agents using old interface fail immediately and visibly

**Migration Support**:
- Update documentation with clear migration examples
- Error message for `fluxid ipc` suggests new commands
- Agent integration guides updated in Phase 1

**Alternatives Considered**:
1. Dual support (both IPC and file-based) - Adds complexity, delays cleanup
2. Gradual deprecation - Requires version detection, migration period coordination

## Technology Stack Summary

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Language | Go | 1.25 | Core implementation |
| YAML Parser | gopkg.in/yaml.v3 | v3.0.1 | Read report/history files |
| JSON Schema | github.com/xeipuuv/gojsonschema | latest | Validate against schema |
| File I/O | Go stdlib (os, io) | - | File operations |
| Session ID | github.com/google/uuid | v1.6.0 | Generate session IDs |

## Open Questions Resolved

All "NEEDS CLARIFICATION" items from Technical Context have been resolved:

1. ✅ **JSON Schema library**: Selected `github.com/xeipuuv/gojsonschema`
2. ✅ **YAML security**: Configured to reject anchors, aliases, merge keys
3. ✅ **File size handling**: FIFO eviction at 10MB for history files
4. ✅ **Session directory**: `<session-root>/<session-id>/` structure
5. ✅ **Schema format**: JSON Schema Draft 7 for both report and history
6. ✅ **Error messages**: Structured with field paths and actionable guidance
7. ✅ **Exit codes**: Standard Unix conventions (0 success, 1-4 error types)
8. ✅ **Backward compatibility**: Hard break, no compatibility layer

## Next Steps

Phase 0 complete. Proceed to Phase 1: Design & Contracts
- Generate `data-model.md` defining report/history structures
- Schemas already exist at `internal/assets/templates/report-schema.yaml` and `internal/assets/templates/history-schema.yaml`
- Generate `quickstart.md` for agent integration guide
- Update agent context files
