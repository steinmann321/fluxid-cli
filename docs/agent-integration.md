# Agent Integration

Guide for integrating coding agents with fluxid's file-based interface.

## Overview

Agents communicate with fluxid through **file operations**. This provides:
- **Reliability:** No parsing ambiguity
- **Debuggability:** Files can be inspected directly
- **Simplicity:** Standard file I/O, no custom protocol
- **Security:** Strict YAML validation prevents injection

---

## File-Based Interface

### Core Concept

```
┌──────────┐              ┌──────────┐              ┌──────────┐
│ fluxid   │   launches   │  Agent   │   writes     │  Files   │
│ workflow │─────────────>│  process │─────────────>│ (YAML)   │
└──────────┘              └──────────┘              └──────────┘
     │                                                    │
     └───────────────── reads and validates ─────────────┘
```

**Flow:**
1. fluxid sets `FLUXID_SESSION_ID` environment variable
2. Agent queries file paths using `fluxid report --get-file`
3. Agent writes YAML to those paths
4. fluxid reads and validates files
5. Workflow proceeds based on file contents

---

## Environment Variables

### `FLUXID_SESSION_ID`
Session identifier set by fluxid workflow.

**Usage in agent:**
```bash
# Check if running under fluxid
if [[ -z "$FLUXID_SESSION_ID" ]]; then
  echo "Error: Not running under fluxid workflow"
  exit 1
fi
```

### Other Environment Variables
Agents receive all environment variables from fluxid's execution context.

---

## Getting File Paths

### Report File Path

```bash
# Get absolute path to report file
REPORT_FILE=$(fluxid report --get-file)

# Example output:
# /Users/username/.fluxid/sessions/a1b2c3d4-e5f6-7890/report.yaml
```

**Requirement:** `FLUXID_SESSION_ID` must be set (automatically done by workflow).

### History File Path

```bash
# Get absolute path to history file
HISTORY_FILE=$(fluxid history --get-file)

# Example output:
# /Users/username/.fluxid/sessions/a1b2c3d4-e5f6-7890/history.yaml
```

---

## Report File Format

### Required Structure

```yaml
command: implement              # Workflow phase
artifact: src/main.go          # Primary artifact modified
timestamp: 2026-01-07T10:00:00Z # ISO 8601 UTC timestamp
status: PASS                    # PASS or FAIL
issues:
  blockers: []                  # Critical issues preventing progress
  defects: []                   # Bugs that need fixing
  concerns: []                  # Code smells or design issues
  observations: []              # Neutral observations
  enhancements: []              # Improvement opportunities
```

**Optional fields:**
```yaml
next_steps:                     # Array of next actions
  - "Add error handling"
  - "Write tests"
summary: "Implementation complete"  # Brief summary
```

### Field Specifications

#### `command` (required)
**Type:** string
**Purpose:** Workflow phase identifier
**Values:**
- `implement` - Implementation phase
- `review` - Review phase
- `commit` - Commit phase
- Custom values allowed

**Example:**
```yaml
command: implement
```

#### `artifact` (required)
**Type:** string
**Purpose:** Primary file or component modified
**Format:** File path or component name

**Examples:**
```yaml
artifact: src/auth/login.go
artifact: authentication-module
artifact: user-profile-api
```

#### `timestamp` (required)
**Type:** string
**Format:** ISO 8601 UTC timestamp
**Example:**
```yaml
timestamp: 2026-01-07T10:00:00Z
```

**Generate in bash:**
```bash
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
```

#### `status` (required)
**Type:** string (enum)
**Values:**
- `PASS` - Phase completed successfully
- `FAIL` - Phase failed, needs retry

**Example:**
```yaml
status: PASS
```

#### `issues` (required)
**Type:** object with 5 required categories
**Structure:**
```yaml
issues:
  blockers: []      # Array of strings
  defects: []       # Array of strings
  concerns: []      # Array of strings
  observations: []  # Array of strings
  enhancements: []  # Array of strings
```

**Category meanings:**
- **blockers:** Critical issues preventing completion
- **defects:** Bugs or errors that must be fixed
- **concerns:** Code smells, design issues
- **observations:** Neutral findings
- **enhancements:** Improvement opportunities

**Examples:**
```yaml
issues:
  blockers:
    - "Database migration failed"
  defects:
    - "Login endpoint returns 500 on invalid input"
  concerns:
    - "Tight coupling between auth and user modules"
  observations:
    - "Using bcrypt for password hashing"
  enhancements:
    - "Could add caching for user profiles"
```

#### `next_steps` (optional)
**Type:** array of strings
**Purpose:** Suggested next actions

**Example:**
```yaml
next_steps:
  - "Add integration tests"
  - "Update API documentation"
  - "Handle edge case for null values"
```

#### `summary` (optional)
**Type:** string
**Purpose:** Brief summary of phase outcome

**Example:**
```yaml
summary: "Authentication feature implemented with tests"
```

### Complete Example

```yaml
command: implement
artifact: src/auth/login.go
timestamp: 2026-01-07T10:30:45Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns:
    - "Password complexity validation could be stronger"
  observations:
    - "Using JWT tokens with 24h expiry"
    - "Login attempts rate-limited to 5 per minute"
  enhancements:
    - "Could add OAuth integration"
    - "Consider adding 2FA support"
next_steps:
  - "Add integration tests for auth flow"
  - "Document API endpoints"
summary: "Login feature implemented with rate limiting and JWT auth"
```

---

## History File Format

### Structure

History is an **array** of event objects:

```yaml
- timestamp: 2026-01-07T10:00:00Z
  step: implement
  status: SUCCESS
  summary: "Initial implementation"
  details: "Implemented user login with JWT tokens"

- timestamp: 2026-01-07T10:15:00Z
  step: review
  status: FAIL
  summary: "Security review failed"
  details: "Missing input validation on password field"

- timestamp: 2026-01-07T10:30:00Z
  step: implement
  status: SUCCESS
  summary: "Fixed validation issues"
  details: "Added password validation and rate limiting"
```

### Event Fields

#### `timestamp` (required)
**Type:** string (ISO 8601 UTC)
**Example:** `2026-01-07T10:00:00Z`

#### `step` (required)
**Type:** string
**Purpose:** Workflow step identifier
**Examples:** `implement`, `review`, `commit`, custom values

#### `status` (required)
**Type:** string (enum)
**Values:** `SUCCESS` or `FAIL`

#### `summary` (required)
**Type:** string
**Purpose:** Brief description of step goal

#### `details` (optional)
**Type:** string
**Purpose:** Detailed explanation of approach and outcome

### Appending to History

```bash
# Get history file path
HISTORY_FILE=$(fluxid history --get-file)

# Append event
cat >> "$HISTORY_FILE" <<EOF
- timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
  step: implement
  status: SUCCESS
  summary: "Feature implementation complete"
  details: "Implemented login with JWT auth and rate limiting"
EOF
```

**Important:** Use `>>` (append) not `>` (overwrite)

---

## Writing Reports

### Basic Agent Implementation

```bash
#!/bin/bash
# Agent script

set -e

# 1. Get file paths
REPORT_FILE=$(fluxid report --get-file)
HISTORY_FILE=$(fluxid history --get-file)

# 2. Do work
# ... implement changes ...

# 3. Write report
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
cat > "$REPORT_FILE" <<EOF
command: implement
artifact: src/main.go
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations:
    - "Implementation complete"
  enhancements: []
EOF

# 4. Optionally validate
fluxid report --validate

# 5. Append to history
cat >> "$HISTORY_FILE" <<EOF
- timestamp: $TIMESTAMP
  step: implement
  status: SUCCESS
  summary: "Implementation complete"
EOF

fluxid history --validate
```

### Handling Failures

```bash
# If agent encounters error
if [[ $? -ne 0 ]]; then
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  cat > "$REPORT_FILE" <<EOF
command: implement
artifact: src/main.go
timestamp: $TIMESTAMP
status: FAIL
issues:
  blockers:
    - "Failed to compile: syntax error in main.go"
  defects: []
  concerns: []
  observations: []
  enhancements: []
summary: "Compilation failed"
EOF

  cat >> "$HISTORY_FILE" <<EOF
- timestamp: $TIMESTAMP
  step: implement
  status: FAIL
  summary: "Compilation failed"
  details: "Syntax error in main.go line 42"
EOF

  exit 1
fi
```

---

## Validation

### Why Validate

- **Catch errors early:** Find issues before fluxid reads file
- **Better error messages:** Validation provides specific field-level errors
- **Schema compliance:** Ensure agent output meets requirements

### How to Validate

```bash
# Write report
cat > "$REPORT_FILE" <<EOF
...
EOF

# Validate before proceeding
if ! fluxid report --validate; then
  echo "Report validation failed"
  exit 1
fi

# Same for history
if ! fluxid history --validate; then
  echo "History validation failed"
  exit 1
fi
```

### Validation Errors

Validation failures print detailed errors:

```
error: [file]: malformed YAML (expected: valid YAML, got: parse error: yaml: line 5: did not find expected key)
```

```
error: [status]: invalid value (expected: enum[PASS,FAIL], got: PENDING)
```

```
error: [issues]: missing required field (expected: object with blockers/defects/concerns/observations/enhancements)
```

---

## Security Considerations

### YAML Security

fluxid **rejects** YAML files containing:
- Anchors (`&anchor`)
- Aliases (`*alias`)
- Merge keys (`<<:`)

**Reason:** Prevent YAML deserialization attacks

**Error example:**
```
error: [file]: YAML anchor/alias/merge detected (security: anchors, aliases, and merge keys are forbidden)
```

### Strict Validation

Reports must:
- Only contain defined fields (`additionalProperties: false`)
- Match exact field types
- Use valid enum values

**Agent must not:**
- Add custom fields
- Use wrong types
- Deviate from schema

---

## Schema Inspection

### Get Schemas

```bash
# Report schema
fluxid report --get-schema > report-schema.yaml

# History schema
fluxid history --get-schema > history-schema.yaml
```

### Use Cases

- **Agent development:** Understand required format
- **Testing:** Validate test data
- **Documentation:** Reference for integration guides

---

## Complete Agent Example

```bash
#!/bin/bash
# Complete agent implementation

set -euo pipefail

# Configuration
readonly AGENT_NAME="my-agent"

# Get file paths
REPORT_FILE=$(fluxid report --get-file)
HISTORY_FILE=$(fluxid history --get-file)

# Helper: Write report
write_report() {
  local status="$1"
  local summary="$2"
  local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  cat > "$REPORT_FILE" <<EOF
command: implement
artifact: src/main.go
timestamp: $timestamp
status: $status
issues:
  blockers: []
  defects: []
  concerns: []
  observations:
    - "$summary"
  enhancements: []
summary: "$summary"
EOF
}

# Helper: Append history
append_history() {
  local status="$1"
  local summary="$2"
  local details="$3"
  local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  cat >> "$HISTORY_FILE" <<EOF
- timestamp: $timestamp
  step: implement
  status: $status
  summary: "$summary"
  details: "$details"
EOF
}

# Main logic
main() {
  echo "[$AGENT_NAME] Starting implementation..."

  # Do work
  if perform_implementation; then
    write_report "PASS" "Implementation complete"
    append_history "SUCCESS" "Implementation complete" "All changes implemented successfully"

    # Validate
    fluxid report --validate || exit 1
    fluxid history --validate || exit 1

    echo "[$AGENT_NAME] Success!"
    exit 0
  else
    write_report "FAIL" "Implementation failed"
    append_history "FAIL" "Implementation failed" "Error details here"

    echo "[$AGENT_NAME] Failed!"
    exit 1
  fi
}

perform_implementation() {
  # Actual implementation logic
  # Return 0 for success, 1 for failure
  return 0
}

main
```

---

## Troubleshooting

### "FLUXID_SESSION_ID not set"
**Cause:** Agent not launched by fluxid workflow

**Fix:** Only run agent via `fluxid --agent --file=...`

### "File not found"
**Cause:** File path query failed

**Fix:** Ensure session ID is set, check fluxid is installed

### "Validation failed"
**Cause:** YAML doesn't match schema

**Fix:** Compare output with schema using `fluxid report --get-schema`

### "YAML anchor detected"
**Cause:** Used YAML anchor/alias/merge

**Fix:** Remove YAML anchors, use plain YAML

---

## Best Practices

1. **Always validate:** Use `--validate` before finishing
2. **Use timestamps:** Generate with `date -u`
3. **Be specific:** Provide detailed issue descriptions
4. **Categorize correctly:** Use appropriate issue categories
5. **Write incrementally:** Append to history as work progresses
6. **Handle errors:** Write FAIL report on error
7. **Check schema:** Use `--get-schema` for reference
8. **Test locally:** Validate reports before integration
9. **Use helpers:** Create utility functions for common operations
10. **Follow format:** Match examples exactly
