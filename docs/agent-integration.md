# Agent Integration

How to integrate external coding agents (Claude, Codex, OpenCode) with fluxid's file-based workflow protocol.

## What is an Agent?

An **agent** is an external AI coding assistant like:
- `claude` (Claude CLI from Anthropic)
- `codex` (OpenAI Codex)
- `opencode` (OpenCode AI)

These are **separate command-line tools** that receive prompts and execute coding tasks.

## How fluxid Invokes Agents

When you run `fluxid --claude --file=task.md`, fluxid:

1. **Reads the command file** (e.g., `~/.fluxid/commands/fluxid.implement-e2e.md`)
2. **Composes a prompt** from:
   - Command file content
   - Task file path
3. **Executes the agent** as a subprocess:
   ```bash
   claude --dangerously-skip-permissions \
          --output-format stream-json \
          --verbose \
          -p "<composed-prompt>"
   ```
4. **Sets environment variables**:
   - `FLUXID_SESSION_ID=<session-id>`
   - `FLUXID_TASK_FILE=<task-file-path>`
   - `FLUXID_SESSION_ROOT=<session-root>` (optional)
5. **Streams agent output** to stdout/stderr
6. **Reads report.yaml** after agent completes to get status

## The File-Based Protocol

### Agent Must Write Report

After completing a task, the agent MUST write a valid `report.yaml` file. fluxid reads this file to determine workflow status (PASS/FAIL).

### Get Report File Path

**Agent command:**
```bash
fluxid report --get-file
```

**Output (to stdout):**
```
/Users/username/.fluxid/sessions/<session-id>/report.yaml
```

**Requirements:**
- `FLUXID_SESSION_ID` environment variable must be set (fluxid sets this automatically)
- Returns absolute path where agent should write report
- Creates session directory if it doesn't exist

### Get Report Schema

**Agent command:**
```bash
fluxid report --get-schema
```

**Output (to stdout):**
YAML schema defining required report structure (see Report Schema section below).

### Validate Report

**Agent command:**
```bash
fluxid report --validate
```

**Exit codes:**
- `0`: Report is valid
- Non-zero: Validation failed (error details on stderr)

**Requirements:**
- `FLUXID_SESSION_ID` must be set
- report.yaml must exist at session path

## Report Schema

### Required Fields

```yaml
command: implement              # Phase: implement, review, or commit
artifact: src/main.go          # Primary file or component modified
timestamp: 2026-01-07T10:00:00Z # ISO 8601 UTC timestamp
status: PASS                    # PASS or FAIL
issues:
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
summary: "Implementation complete"  # Brief summary
```

### Field Constraints

- **command**: string (required)
- **artifact**: string (required)
- **timestamp**: ISO 8601 UTC format `YYYY-MM-DDTHH:MM:SSZ` (required)
- **status**: enum `PASS` or `FAIL` (required)
- **issues**: object with 5 required arrays (required)
  - blockers, defects, concerns, observations, enhancements
- **next_steps**: array of strings (optional)
- **summary**: string (optional)
- **Additional properties**: NOT ALLOWED (schema strict validation)

## History File (Optional)

Agents can optionally append events to history.yaml to track workflow progression.

### Get History File Path

```bash
fluxid history --get-file
```

### Get History Schema

```bash
fluxid history --get-schema
```

### Validate History

```bash
fluxid history --validate
```

### History Event Structure

```yaml
- timestamp: 2026-01-07T10:00:00Z  # ISO 8601 UTC (required)
  step: implement                  # Step identifier (required)
  status: SUCCESS                  # SUCCESS or FAIL (required)
  summary: "Implementation complete" # Brief description (required)
  details: "Implemented login..."  # Detailed explanation (optional)
```

## Command File Templates

Command files are **markdown prompts** sent to the agent. They must instruct the agent to use the file-based protocol.

### Minimal Example

```markdown
# Task

Read the task file at: $FLUXID_TASK_FILE

Implement the required changes.

# Report Protocol

After completing the task, write a report using fluxid's file-based interface:

1. Get report file path:
   ```bash
   REPORT_FILE=$(fluxid report --get-file)
   ```

2. Get schema to understand structure:
   ```bash
   fluxid report --get-schema
   ```

3. Write YAML conforming to schema to $REPORT_FILE

4. Validate before finishing:
   ```bash
   fluxid report --validate
   ```

Required report fields:
- command: "implement"
- artifact: <main file modified>
- timestamp: <ISO 8601 UTC>
- status: PASS or FAIL
- issues: object with blockers, defects, concerns, observations, enhancements arrays

Set status to PASS if task completed successfully, FAIL if blocked or incomplete.
```

### Complete Example

See `internal/assets/commands/*.md` for full command templates, but note that default templates may reference outdated shell script approaches. Modern templates should use the fluxid CLI commands documented above.

## Extending fluxid

### Custom Command Files

1. Create custom command templates in `~/.fluxid/commands/` or `.fluxid/commands/`
2. Update `config.yaml` to reference your command files:
   ```yaml
   commands:
     implement: /absolute/path/to/implement.md
     review: /absolute/path/to/review.md
     commit: /absolute/path/to/commit.md
   ```

### Command File Structure

Your command files should:
1. **Define the role/context** for the agent
2. **Reference task file** via `$FLUXID_TASK_FILE` environment variable
3. **Instruct agent** to use fluxid CLI commands:
   - `fluxid report --get-file`
   - `fluxid report --get-schema`
   - `fluxid report --validate`
4. **Specify report requirements** (required fields, PASS/FAIL criteria)
5. **Include examples** of valid reports (optional but helpful)

### Critical Requirements

Your command file MUST ensure the agent:
- Writes a valid `report.yaml` file
- Uses `fluxid report --get-file` to get the path (not hardcoded paths)
- Validates the report before completing (`fluxid report --validate`)
- Sets `status: PASS` only if task completed successfully
- Sets `status: FAIL` if blocked, incomplete, or errors occurred

## Environment Variables

Agents receive these from fluxid automatically:

- **FLUXID_SESSION_ID**: Current session identifier (UUID)
- **FLUXID_TASK_FILE**: Absolute path to task file
- **FLUXID_SESSION_ROOT**: Session storage root directory (optional)

## Error Handling

### "FLUXID_SESSION_ID not set"

**Cause**: Agent called fluxid commands outside of fluxid workflow.

**Fix**: Only run agent via `fluxid --<agent> --file=...`

### "Validation failed"

**Cause**: Report YAML doesn't match schema.

**Fix**: Run `fluxid report --get-schema` to see requirements.

### "Additional property not allowed"

**Cause**: Report contains fields not in schema.

**Fix**: Remove custom fields, use only schema-defined fields.

## Security

### YAML Security

fluxid rejects YAML with:
- Anchors (`&anchor`)
- Aliases (`*alias`)
- Merge keys (`<<:`)

**Reason**: Prevent YAML deserialization attacks.

### Path Security

- Session IDs must be valid UUIDs
- Paths validated to prevent traversal attacks
- No `../` or absolute paths in session IDs

## Best Practices

1. **Always instruct agents** to use CLI commands, not hardcoded paths
2. **Include validation step** in command templates
3. **Define clear PASS/FAIL criteria** in command files
4. **Test command files** with `--fluxid-dry-run` before production use
5. **Keep command files simple** and focused on protocol compliance
6. **Document your command files** for team consistency

## Troubleshooting

### Report not found after agent completes

**Possible causes:**
- Agent didn't write report.yaml
- Agent wrote to wrong path
- Agent command file doesn't include protocol instructions

**Debug steps:**
1. Check if `report.yaml` exists in session directory
2. Review command file to ensure it instructs agent to write report
3. Check agent output for errors

### Agent ignores protocol instructions

**Possible causes:**
- Command file prompt is unclear
- Agent doesn't have access to fluxid CLI
- Agent fails before reaching protocol section

**Debug steps:**
1. Simplify command file to focus on protocol
2. Verify fluxid is in agent's PATH
3. Add explicit error handling in command file

### Validation always fails

**Possible causes:**
- Report uses additional fields not in schema
- Wrong field types or values
- Missing required fields

**Debug steps:**
1. Compare report.yaml with schema from `fluxid report --get-schema`
2. Check validation error message for specific field issues
3. Ensure timestamp is ISO 8601 UTC format
4. Verify status is exactly `PASS` or `FAIL` (case-sensitive)
