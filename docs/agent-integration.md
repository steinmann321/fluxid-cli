# Agent Integration

Guide for LLM agents (Claude, Codex, OpenCode) integrating with fluxid's file-based interface.

## Overview

Agents are LLMs that receive task instructions and generate structured YAML outputs. fluxid provides:
- **File paths:** Where to write results (`--get-file`)
- **Schemas:** What structure to generate (`--get-schema`)
- **Validation:** How to verify correctness (`--validate`)

---

## Integration Flow

```
┌──────────┐              ┌──────────┐              ┌──────────┐
│ fluxid   │   launches   │   LLM    │  generates   │  Files   │
│ workflow │─────────────>│  Agent   │─────────────>│ (YAML)   │
└──────────┘              └──────────┘              └──────────┘
     │                                                    │
     └───────────────── reads and validates ─────────────┘
```

**Workflow:**
1. fluxid launches agent with task file and command template
2. Agent receives `FLUXID_SESSION_ID` environment variable
3. Agent queries file path: `fluxid report --get-file`
4. Agent queries schema: `fluxid report --get-schema`
5. Agent generates YAML conforming to schema
6. Agent writes YAML to file path
7. Agent validates output: `fluxid report --validate`
8. fluxid reads and validates report, proceeds with workflow

---

## Interface Commands

### Get File Path

```bash
fluxid report --get-file
# Output: /Users/username/.fluxid/sessions/<session-id>/report.yaml

fluxid history --get-file
# Output: /Users/username/.fluxid/sessions/<session-id>/history.yaml
```

Returns absolute path where agent should write results.

### Get Schema

```bash
fluxid report --get-schema
# Outputs YAML schema defining required structure

fluxid history --get-schema
# Outputs YAML schema for history events
```

Returns the JSON schema in YAML format that defines required fields, types, and constraints.

### Validate Output

```bash
fluxid report --validate
# Validates report.yaml in current session

fluxid history --validate
# Validates history.yaml in current session
```

Returns exit code 0 if valid, non-zero with error details if invalid.

---

## Report Schema

### Required Structure

```yaml
command: implement              # Workflow phase (implement, review, commit)
artifact: src/main.go          # Primary file or component modified
timestamp: 2026-01-07T10:00:00Z # ISO 8601 UTC timestamp
status: PASS                    # PASS or FAIL
issues:                         # Required object with 5 categories
  blockers: []                  # Critical issues preventing progress
  defects: []                   # Bugs that need fixing
  concerns: []                  # Code smells or design issues
  observations: []              # Neutral observations
  enhancements: []              # Improvement opportunities
```

### Optional Fields

```yaml
next_steps:                     # Array of suggested actions
  - "Add error handling"
  - "Write tests"
summary: "Implementation complete"  # Brief summary of phase
```

### Field Specifications

#### `command` (required)
- **Type:** string
- **Purpose:** Workflow phase identifier
- **Values:** `implement`, `review`, `commit`, or custom values

#### `artifact` (required)
- **Type:** string
- **Purpose:** Primary file or component modified
- **Examples:** `src/auth/login.go`, `authentication-module`, `user-profile-api`

#### `timestamp` (required)
- **Type:** string
- **Format:** ISO 8601 UTC (`YYYY-MM-DDTHH:MM:SSZ`)
- **Example:** `2026-01-07T10:30:45Z`

#### `status` (required)
- **Type:** enum
- **Values:** `PASS` (success) or `FAIL` (retry needed)

#### `issues` (required)
- **Type:** object with 5 required arrays
- **Categories:**
  - **blockers:** Critical issues preventing completion
  - **defects:** Bugs or errors that must be fixed
  - **concerns:** Code smells, design issues
  - **observations:** Neutral findings
  - **enhancements:** Improvement opportunities

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

## History Schema

### Structure

History is an array of event objects:

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
- **Type:** string (ISO 8601 UTC)
- **Example:** `2026-01-07T10:00:00Z`

#### `step` (required)
- **Type:** string
- **Purpose:** Workflow step identifier
- **Examples:** `implement`, `review`, `commit`, custom values

#### `status` (required)
- **Type:** enum
- **Values:** `SUCCESS` or `FAIL`

#### `summary` (required)
- **Type:** string
- **Purpose:** Brief description of step goal

#### `details` (optional)
- **Type:** string
- **Purpose:** Detailed explanation of approach and outcome

---

## Validation

### Why Validate

- **Catch errors early:** Find schema violations before workflow reads file
- **Better error messages:** Validation provides field-level errors
- **Schema compliance:** Ensure output meets requirements

### Validation Errors

**Malformed YAML:**
```
error: [file]: malformed YAML (expected: valid YAML, got: parse error: yaml: line 5: did not find expected key)
```

**Invalid enum value:**
```
error: [status]: invalid value (expected: enum[PASS,FAIL], got: PENDING)
```

**Missing required field:**
```
error: [issues]: missing required field (expected: object with blockers/defects/concerns/observations/enhancements)
```

**Additional properties:**
```
error: [report]: additional property not allowed: 'extra_field'
```

---

## Security

### YAML Restrictions

fluxid rejects YAML files containing:
- **Anchors** (`&anchor`)
- **Aliases** (`*alias`)
- **Merge keys** (`<<:`)

**Reason:** Prevent YAML deserialization attacks.

### Strict Validation

Reports must:
- Only contain defined schema fields (`additionalProperties: false`)
- Match exact field types
- Use valid enum values

---

## Environment Variables

### `FLUXID_SESSION_ID`

Session identifier set by fluxid. Agent can access this to reference the current session.

**Example usage:**
```bash
echo "Current session: $FLUXID_SESSION_ID"
```

### `FLUXID_SESSION_ROOT`

Session storage root directory (defaults to `$HOME/.fluxid/sessions`).

---

## Agent Implementation Guide

### 1. Receive Task

Agent receives task description via command file template with variables like `{task_content}`.

### 2. Query Interface

```bash
# Get where to write
REPORT_FILE=$(fluxid report --get-file)

# Get what structure to generate
fluxid report --get-schema > schema.yaml
```

### 3. Generate YAML

Agent generates YAML conforming to schema:
- All required fields present
- Correct types and enums
- ISO 8601 timestamps
- No additional properties

### 4. Write Output

Agent writes generated YAML to file path from step 2.

### 5. Validate

```bash
fluxid report --validate
```

If validation fails, agent should fix errors and retry.

### 6. Append History (Optional)

Agent can append workflow events to history:

```bash
HISTORY_FILE=$(fluxid history --get-file)
# Append event to $HISTORY_FILE
fluxid history --validate
```

---

## Common Patterns

### Success Path

1. Get file path and schema
2. Generate valid YAML
3. Write to file
4. Validate
5. Return success (exit 0)

### Failure Path

1. Get file path and schema
2. Generate YAML with status: FAIL
3. Populate `issues.blockers` or `issues.defects`
4. Write to file
5. Validate
6. Return failure (exit 1)

### Retry After Validation Error

1. Generate YAML
2. Validate: fails with error
3. Parse error message
4. Fix YAML structure
5. Validate again
6. Continue when valid

---

## Best Practices

1. **Always validate:** Call `--validate` before completing
2. **Use schemas:** Query `--get-schema` to understand format
3. **ISO timestamps:** Use `YYYY-MM-DDTHH:MM:SSZ` format
4. **Be specific:** Provide detailed issue descriptions
5. **Categorize correctly:** Use appropriate issue categories
6. **Handle errors:** Write FAIL report when encountering problems
7. **No extra fields:** Stick to schema-defined fields only
8. **Test locally:** Validate outputs during development

---

## Troubleshooting

### "FLUXID_SESSION_ID not set"
**Cause:** Agent not launched by fluxid workflow.
**Fix:** Only run agent via `fluxid --<agent> --file=...`

### "File not found"
**Cause:** File path query failed.
**Fix:** Ensure session ID is set, check fluxid is installed.

### "Validation failed"
**Cause:** YAML doesn't match schema.
**Fix:** Compare output with schema using `fluxid report --get-schema`.

### "YAML anchor detected"
**Cause:** Used YAML anchor/alias/merge.
**Fix:** Remove YAML anchors, use plain YAML only.

### "Additional property not allowed"
**Cause:** Added field not in schema.
**Fix:** Remove custom fields, use only schema-defined fields.
