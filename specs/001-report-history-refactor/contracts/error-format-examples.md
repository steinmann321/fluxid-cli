# Error Format Examples

This document provides concrete examples of error messages that validation commands should produce, as required by FR-005, FR-018, and SC-009.

## Error Message Format

**Human-Readable Format (stderr output):**
```
[field_path]: [violation] (expected: [constraint], got: [value])
```

**Programmatic Format (JSON for machine parsing):**
```json
{
  "field": "[field_path]",
  "violation": "[violation description]",
  "constraint": "[expected constraint]",
  "value": "[actual value]"
}
```

## Report Validation Examples

### Missing Required Field

**Human-Readable:**
```
status: missing required field (expected: PASS or FAIL, got: <not present>)
```

**JSON:**
```json
{
  "field": "status",
  "violation": "missing required field",
  "constraint": "PASS or FAIL",
  "value": "<not present>"
}
```

### Invalid Enum Value

**Human-Readable:**
```
status: invalid value (expected: PASS or FAIL, got: INVALID)
```

**JSON:**
```json
{
  "field": "status",
  "violation": "invalid value",
  "constraint": "PASS or FAIL",
  "value": "INVALID"
}
```

### Wrong Data Type

**Human-Readable:**
```
timestamp: wrong type (expected: string (ISO 8601), got: number)
```

**JSON:**
```json
{
  "field": "timestamp",
  "violation": "wrong type",
  "constraint": "string (ISO 8601)",
  "value": "number"
}
```

### Additional Unexpected Fields (FR-028)

**Human-Readable:**
```
extra_field: additional property not allowed (expected: no additional properties, got: "some value")
```

**JSON:**
```json
{
  "field": "extra_field",
  "violation": "additional property not allowed",
  "constraint": "no additional properties",
  "value": "some value"
}
```

### Malformed YAML

**Human-Readable:**
```
[file]: malformed YAML (expected: valid YAML document, got: parse error at line 5: mapping values are not allowed here)
```

**JSON:**
```json
{
  "field": "[file]",
  "violation": "malformed YAML",
  "constraint": "valid YAML document",
  "value": "parse error at line 5: mapping values are not allowed here"
}
```

### YAML Security Violation (FR-011)

**Human-Readable:**
```
[file]: YAML anchor not allowed (expected: no anchors, aliases, or merge keys, got: anchor '&anchor_name' at line 3)
```

**JSON:**
```json
{
  "field": "[file]",
  "violation": "YAML anchor not allowed",
  "constraint": "no anchors, aliases, or merge keys",
  "value": "anchor '&anchor_name' at line 3"
}
```

### File Not Found

**Human-Readable:**
```
[file]: file not found (expected: file at /tmp/fluxid/report-<session-id>.yaml, got: file does not exist)
```

**JSON:**
```json
{
  "field": "[file]",
  "violation": "file not found",
  "constraint": "file at /tmp/fluxid/report-<session-id>.yaml",
  "value": "file does not exist"
}
```

### Empty File

**Human-Readable:**
```
[file]: empty file (expected: non-empty YAML document, got: 0 bytes)
```

**JSON:**
```json
{
  "field": "[file]",
  "violation": "empty file",
  "constraint": "non-empty YAML document",
  "value": "0 bytes"
}
```

## History Validation Examples

### Invalid Timestamp Format

**Human-Readable:**
```
events[0].timestamp: invalid format (expected: ISO 8601 format (e.g., 2026-01-05T12:00:00Z), got: "2026/01/05 12:00:00")
```

**JSON:**
```json
{
  "field": "events[0].timestamp",
  "violation": "invalid format",
  "constraint": "ISO 8601 format (e.g., 2026-01-05T12:00:00Z)",
  "value": "2026/01/05 12:00:00"
}
```

### Single Entry Exceeds Size Limit (FR-025)

**Human-Readable:**
```
events[3]: single entry too large (expected: entry size < 10MB, got: 12.5MB) - split entry into multiple smaller entries
```

**JSON:**
```json
{
  "field": "events[3]",
  "violation": "single entry too large",
  "constraint": "entry size < 10MB",
  "value": "12.5MB - split entry into multiple smaller entries"
}
```

### Invalid Event Status

**Human-Readable:**
```
events[2].status: invalid value (expected: SUCCESS or FAIL, got: PENDING)
```

**JSON:**
```json
{
  "field": "events[2].status",
  "violation": "invalid value",
  "constraint": "SUCCESS or FAIL",
  "value": "PENDING"
}
```

## Session Validation Examples

### Missing Session ID Environment Variable

**Human-Readable:**
```
FLUXID_SESSION_ID: environment variable not set (expected: valid UUID, got: <not set>)
```

**JSON:**
```json
{
  "field": "FLUXID_SESSION_ID",
  "violation": "environment variable not set",
  "constraint": "valid UUID",
  "value": "<not set>"
}
```

### Invalid Session ID Format

**Human-Readable:**
```
session_id: invalid format (expected: valid UUID v4, got: "not-a-uuid")
```

**JSON:**
```json
{
  "field": "session_id",
  "violation": "invalid format",
  "constraint": "valid UUID v4",
  "value": "not-a-uuid"
}
```

### Path Traversal Attempt (FR-012)

**Human-Readable:**
```
path: path traversal not allowed (expected: path within /tmp/fluxid/, got: "../../etc/passwd")
```

**JSON:**
```json
{
  "field": "path",
  "violation": "path traversal not allowed",
  "constraint": "path within /tmp/fluxid/",
  "value": "../../etc/passwd"
}
```

## Permission Errors

**Human-Readable:**
```
[file]: permission denied (expected: read permission on /tmp/fluxid/report-<session-id>.yaml, got: permission denied - check file ownership and permissions)
```

**JSON:**
```json
{
  "field": "[file]",
  "violation": "permission denied",
  "constraint": "read permission on /tmp/fluxid/report-<session-id>.yaml",
  "value": "permission denied - check file ownership and permissions"
}
```

## Multiple Errors

When multiple validation errors occur, output should include all errors:

**Human-Readable:**
```
status: missing required field (expected: PASS or FAIL, got: <not present>)
timestamp: wrong type (expected: string (ISO 8601), got: number)
command: missing required field (expected: string, got: <not present>)
```

**JSON:**
```json
[
  {
    "field": "status",
    "violation": "missing required field",
    "constraint": "PASS or FAIL",
    "value": "<not present>"
  },
  {
    "field": "timestamp",
    "violation": "wrong type",
    "constraint": "string (ISO 8601)",
    "value": "number"
  },
  {
    "field": "command",
    "violation": "missing required field",
    "constraint": "string",
    "value": "<not present>"
  }
]
```

## Exit Codes

- **0**: Validation successful (no errors)
- **1**: Validation failed (schema violations)
- **2**: File system error (file not found, permission denied)
- **3**: Configuration error (missing session ID, invalid session ID)

## Notes

- All error messages should be written to stderr (per FR-041)
- Successful validation should produce no output except exit code 0 (per FR-042)
- Error messages must include sufficient context to diagnose failures without verbose logging (per FR-043)
- Field paths use dot notation for nested fields: `issues.blockers[0].description`
- Array indices are zero-based: `events[0]`, `events[1]`, etc.
