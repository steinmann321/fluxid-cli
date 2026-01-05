# Quickstart: File-Based Report & History Integration

**Feature**: `001-report-history-refactor`
**Audience**: External agent developers (Claude, OpenCode, custom agents)
**Date**: 2026-01-05

## Overview

This guide shows how to integrate with fluxid's new file-based report and history interface. The old stdio-based IPC system (`fluxid ipc`) has been removed in favor of direct file I/O.

## Prerequisites

- Fluxid installed with version containing this refactor
- `FLUXID_SESSION_ID` environment variable set by fluxid workflow
- File system write permissions in session directory
- YAML writing capability in your agent

## Breaking Changes

**Removed Commands**:
- `fluxid ipc get-report-schema` → Use `fluxid report --get-schema`
- `fluxid ipc write-report` → Direct file write to path from `fluxid report --get-file`
- `fluxid ipc read-report` → Workflow reads automatically, validate with `fluxid report --validate`
- `fluxid ipc write-history` → Append to path from `fluxid history --get-file`
- `fluxid ipc view-history` → Read file directly from `fluxid history --get-file`
- `fluxid ipc abort` → Out of scope (separate evaluation)

## Quick Start: Writing Your First Report

### 1. Get Report File Path

```bash
# Query where to write the report
REPORT_FILE=$(fluxid report --get-file)
echo "Report file: $REPORT_FILE"
# Output: /path/to/.fluxid/sessions/<session-id>/report.yaml
```

**Notes**:
- File and parent directories are created automatically if missing
- Path is session-specific (scoped to `FLUXID_SESSION_ID`)
- You cannot specify a custom path (security: prevents path traversal)

### 2. Write Report YAML

```bash
# Write report with PASS status
cat > "$REPORT_FILE" << 'EOF'
command: fluxid implement
artifact: internal/storage/report.go
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns:
    - File validation error messages could be more specific
  observations:
    - Used xeipuuv/gojsonschema for JSON Schema validation
  enhancements:
    - Consider adding schema caching for performance
next_steps:
  - Run E2E tests to verify integration
summary: Successfully implemented report file operations
EOF
```

**Important**:
- All 5 issue categories are required (blockers, defects, concerns, observations, enhancements)
- Categories can be empty arrays: `blockers: []`
- `status` must be exactly `PASS` or `FAIL`
- `timestamp` must be valid ISO 8601 format

### 3. Validate Report (Optional but Recommended)

```bash
# Validate before fluxid reads it
fluxid report --validate

# Check exit code
if [ $? -eq 0 ]; then
    echo "Report is valid"
else
    echo "Report has validation errors (see stderr)"
fi
```

**Exit Codes**:
- `0`: Valid report
- `1`: Validation failed (see error messages on stderr)
- `2`: File not found or permission error
- `3`: Session ID missing
- `4`: Internal error

### 4. Done!

Fluxid workflow will read your report automatically. If `status: PASS`, workflow exits. If `status: FAIL`, workflow retries with history context.

---

## Recording History

History tracks what approaches failed to prevent repeating mistakes across workflow iterations.

### 1. Get History File Path

```bash
HISTORY_FILE=$(fluxid history --get-file)
echo "History file: $HISTORY_FILE"
# Output: /path/to/.fluxid/sessions/<session-id>/history.yaml
```

### 2. Append History Entry

```bash
# Append a FAIL entry documenting a failed approach
cat >> "$HISTORY_FILE" << 'EOF'
- timestamp: 2026-01-05T10:00:00Z
  step: implement
  status: FAIL
  summary: Initial implementation attempted using channels for file notification
  details: Tried implementing file change watcher using Go channels and fsnotify. Failed because requirements specify strictly sequential workflow with no concurrency. Channels introduce async behavior incompatible with constitution principle III.
EOF
```

**Important**:
- Use `>>` (append) not `>` (overwrite) to preserve previous history
- History is a YAML array - each entry starts with `-`
- `details` field is optional but **highly recommended** for FAIL entries
- File automatically truncated via FIFO eviction if exceeds 10MB

### 3. Validate History (Optional)

```bash
fluxid history --validate
```

---

## Getting Schemas Programmatically

Agents can query schemas to understand structure without hardcoding.

### Report Schema

```bash
# Get JSON Schema for report structure
fluxid report --get-schema > report-schema.json

# Parse with jq to see required fields
fluxid report --get-schema | jq '.required'
# Output: ["command","artifact","timestamp","status","issues"]
```

### History Schema

```bash
# Get JSON Schema for history structure
fluxid history --get-schema > history-schema.json

# Parse with jq to see required fields per entry
fluxid history --get-schema | jq '.items.required'
# Output: ["timestamp","step","status","summary"]
```

---

## Complete Agent Integration Example

Here's a minimal agent script that integrates with fluxid:

```bash
#!/usr/bin/env bash
set -euo pipefail

# 1. Verify session context
if [ -z "${FLUXID_SESSION_ID:-}" ]; then
    echo "ERROR: FLUXID_SESSION_ID not set" >&2
    exit 3
fi

# 2. Get file paths
REPORT_FILE=$(fluxid report --get-file)
HISTORY_FILE=$(fluxid history --get-file)

# 3. Do work (implement or review)
echo "Agent performing work..."
WORK_RESULT="success"  # or "failure"

# 4. Write report based on outcome
if [ "$WORK_RESULT" = "success" ]; then
    cat > "$REPORT_FILE" << EOF
command: fluxid implement
artifact: src/feature.go
timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations:
    - Implementation complete
  enhancements: []
summary: Feature implemented successfully
EOF
else
    cat > "$REPORT_FILE" << EOF
command: fluxid implement
artifact: src/feature.go
timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
status: FAIL
issues:
  blockers:
    - Tests failing
  defects: []
  concerns: []
  observations: []
  enhancements: []
summary: Implementation incomplete due to test failures
EOF

    # 5. Record failure in history
    cat >> "$HISTORY_FILE" << EOF
- timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
  step: implement
  status: FAIL
  summary: First implementation attempt failed tests
  details: Tried approach X but tests showed issue Y. Need to revise approach.
EOF
fi

# 6. Validate report before exiting
if ! fluxid report --validate; then
    echo "ERROR: Report validation failed" >&2
    exit 1
fi

echo "Agent complete. Report written to: $REPORT_FILE"
```

---

## Error Handling

### Common Validation Errors

**Missing required field**:
```
Validation failed for report:
  - issues.blockers: field is required but missing

Expected schema available via: fluxid report --get-schema
```

**Invalid status value**:
```
Validation failed for report:
  - status: must be one of [PASS, FAIL], got "PENDING"

Expected schema available via: fluxid report --get-schema
```

**Invalid timestamp format**:
```
Validation failed for report:
  - timestamp: must be valid ISO 8601 format, got "2024-13-45"

Expected schema available via: fluxid report --get-schema
```

**YAML security violation**:
```
Validation failed for report:
  - YAML anchors not allowed (security constraint)

Expected schema available via: fluxid report --get-schema
```

### Session ID Not Set

```bash
$ fluxid report --get-file
ERROR: FLUXID_SESSION_ID environment variable not set
# Exit code: 3
```

**Solution**: Ensure your agent is launched by fluxid workflow, which sets `FLUXID_SESSION_ID`.

### File Permission Errors

```bash
$ fluxid report --validate
ERROR: permission denied: /path/to/report.yaml
# Exit code: 2
```

**Solution**: Check file system permissions. Session directory must be writable by agent.

---

## Advanced: Programmatic Schema Parsing

If your agent is written in Go, Python, or other languages with JSON Schema support:

### Go Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "os/exec"
)

type Schema struct {
    Required []string `json:"required"`
    Properties map[string]interface{} `json:"properties"`
}

func main() {
    // Get schema
    cmd := exec.Command("fluxid", "report", "--get-schema")
    output, err := cmd.Output()
    if err != nil {
        panic(err)
    }

    // Parse schema
    var schema Schema
    if err := json.Unmarshal(output, &schema); err != nil {
        panic(err)
    }

    // Use schema to validate your data structure
    fmt.Printf("Required fields: %v\n", schema.Required)
}
```

### Python Example

```python
import subprocess
import json

# Get schema
result = subprocess.run(
    ["fluxid", "report", "--get-schema"],
    capture_output=True,
    text=True,
    check=True
)

# Parse schema
schema = json.loads(result.stdout)

# Use schema
print(f"Required fields: {schema['required']}")

# Validate with jsonschema library
from jsonschema import validate
report_data = {
    "command": "fluxid implement",
    "artifact": "src/main.py",
    "timestamp": "2026-01-05T10:00:00Z",
    "status": "PASS",
    "issues": {
        "blockers": [],
        "defects": [],
        "concerns": [],
        "observations": [],
        "enhancements": []
    }
}
validate(instance=report_data, schema=schema)
print("Report data is valid!")
```

---

## Best Practices

1. **Always validate before fluxid reads**: Use `fluxid report --validate` to catch errors early

2. **Provide detailed history for failures**: Include `details` field explaining what failed and why, so future iterations avoid the same approach

3. **Use ISO 8601 timestamps**:
   - Bash: `date -u +"%Y-%m-%dT%H:%M:%SZ"`
   - Go: `time.Now().UTC().Format(time.RFC3339)`
   - Python: `datetime.utcnow().isoformat() + "Z"`

4. **Handle all 5 issue categories**: Even if empty, include all: `blockers`, `defects`, `concerns`, `observations`, `enhancements`

5. **Append to history, never overwrite**: Use `>>` not `>` to preserve previous entries

6. **Check exit codes**: Non-zero exit from validation commands indicates errors

7. **Security**: Never use YAML anchors (`&`), aliases (`*`), or merge keys (`<<`) - they are rejected for security reasons

8. **File size**: Keep reports under 10MB. History auto-truncates at 10MB via FIFO eviction, but don't rely on this for normal operation.

---

## Migration from Old IPC System

| Old Command | New Approach |
|-------------|--------------|
| `fluxid ipc get-report-schema` | `fluxid report --get-schema` |
| `fluxid ipc write-report <data>` | Write YAML to `$(fluxid report --get-file)` |
| `fluxid ipc read-report` | Validate with `fluxid report --validate` (optional) |
| `fluxid ipc write-history <entry>` | Append to `$(fluxid history --get-file)` |
| `fluxid ipc view-history` | Read from `$(fluxid history --get-file)` directly |

**Key Differences**:
- No stdin/stdout IPC protocol - direct file I/O
- No session ID flag - uses `FLUXID_SESSION_ID` environment variable
- No fluxid-managed writes - agents write files directly
- Validation is optional but recommended

---

## Troubleshooting

### Problem: "command not found: fluxid ipc"

**Cause**: Old IPC commands removed in this refactor.

**Solution**: Update agent to use new file-based interface (see migration table above).

### Problem: Report validation fails with "YAML anchors not allowed"

**Cause**: Security constraint rejects YAML advanced features.

**Solution**: Remove `&` (anchors), `*` (aliases), `<<` (merge keys) from YAML. Use plain YAML structures only.

### Problem: History file grows too large

**Cause**: Many workflow iterations without cleanup.

**Solution**: Automatic FIFO eviction at 10MB. No action needed unless you want manual cleanup between sessions.

### Problem: "Session ID not set" error

**Cause**: Agent not launched by fluxid workflow.

**Solution**: Ensure fluxid workflow launches your agent, which sets `FLUXID_SESSION_ID`. For testing, manually set: `export FLUXID_SESSION_ID=$(uuidgen)`

---

## Next Steps

- Review [data-model.md](./data-model.md) for detailed structure documentation
- Review [contracts/report-schema.json](./contracts/report-schema.json) for full JSON Schema
- Review [contracts/history-schema.json](./contracts/history-schema.json) for history schema
- See [spec.md](./spec.md) for complete requirements and acceptance scenarios

## Support

For issues or questions:
- Check schema with `fluxid report --get-schema` or `fluxid history --get-schema`
- Validate files with `--validate` commands for specific error messages
- Refer to spec.md edge cases section for unusual scenarios
