# Agent Integration

How coding agents are integrated with fluxid.

## How fluxid Invokes Agents

When you run `fluxid --claude --file=task.md`, fluxid:

1. **Reads the command file** (e.g., `~/.fluxid/commands/fluxid.implement-e2e.md`)
2. **Composes a prompt** from:
   - Command file content
   - Task file path
3. **Executes the agent** as a subprocess:
   ```bash
   # Default invocation (agent_args defaults to --dangerously-skip-permissions)
   claude --dangerously-skip-permissions \
          --output-format stream-json \
          --verbose \
          -p "<composed-prompt>"

   # Can be customized via config.yaml agent_args field
   ```
4. **Sets environment variables**:
   - `FLUXID_SESSION_ID=<session-id>` - UUID identifying the current session
   - `FLUXID_TASK_FILE=<task-file-path>` - Absolute path to the task file
   - `FLUXID_SESSION_ROOT=<session-root>` - Override for session directory location (default: .fluxid/sessions/, fallback: $TMPDIR/fluxid)
5. **Streams agent output** to stdout/stderr
6. **Reads report.yaml** after agent completes to get status

## The File-Based Handover Protocol

### Agent Must Write Report

The agent MUST write a valid `report.yaml` file in ALL cases:

1. **Task completed successfully** → Write report with `status: PASS`
2. **Task partially done or blocked** → Write report with `status: FAIL`
3. **Agent exhausted/fatigued** → Write report with `status: FAIL` documenting current state

**Critical from implement-e2e command:**
> "If, after doing your best, you still cannot reach green, **stop implementing and produce a clear, validated FAIL report** that documents the current state and concrete next steps. Better stop working than start cheating."

A FAIL report must document:
- How far the implementation got
- What blocks further progress
- What work remains (concrete next steps)

fluxid reads this file to determine workflow status and decide whether to retry.

### Prompt Instructions for Report Writing

**Your command file must instruct the agent how to write the report. Example:**

```markdown
# Report Protocol

After completing work (whether successful, blocked, or exhausted), write a report:

1. Get report file path: `fluxid report --get-file`
2. Get schema: `fluxid report --get-schema`
3. Write YAML report following the schema
4. Validate: `fluxid report --validate` (fix and re-validate if it fails)

## When to use PASS vs FAIL

**PASS:**
- Task fully completed
- All tests passing
- Implementation working as expected

**FAIL:**
- Task partially completed
- Tests failing
- Blocked by external issues
- Agent exhausted/fatigued
- Cannot make further progress without disproportionate effort

**CRITICAL**: If you cannot complete the task, write a FAIL report documenting:
- What you accomplished
- What blocks further progress (in issues.blockers)
- What needs to be done next (in next_steps)

DO NOT:
- Continue indefinitely when exhausted
- Mark PASS by weakening quality
- Skip the report when stuck

Better to stop and document current state clearly than to continue producing low-quality work.
```

### Example: Exhaustion Scenario in Prompt

```markdown
## If You Hit Limits

If after your best effort you cannot complete the task:

1. **Stop working** - Don't continue indefinitely
2. **Write FAIL report** following the report protocol above, documenting what blocks further progress
3. **Exit** - fluxid will read your FAIL report and decide whether to retry
```

## Command File Templates

Command files are **markdown prompts** sent to the agent. They must instruct the agent to use the file-based protocol.

### Minimal Command File Example

```markdown
# Task

Read the task file at: $FLUXID_TASK_FILE

Implement the required changes.

# Report Protocol

After completing the task (or if blocked/exhausted), write a report:

1. Get file path: `fluxid report --get-file`
2. Get schema: `fluxid report --get-schema`
3. Write YAML report following the schema
4. Validate: `fluxid report --validate`

Set status to PASS only if task completed successfully.
Set status to FAIL if blocked, incomplete, or exhausted - document current state clearly.
```

### Complete Example

See `internal/assets/commands/*.md` for full command templates, noting that default templates may reference outdated shell script approaches. Modern templates should use the fluxid CLI commands documented above.

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
   - `fluxid report --get-file` - Get path where to write report
   - `fluxid report --get-schema` - Get YAML schema
   - `fluxid report --validate` - Validate written report
4. **Specify PASS/FAIL criteria** (when to mark each status)
5. **Handle exhaustion scenario** (what to do when agent can't complete task)

### Critical Requirements

Your command file MUST ensure the agent:
- Writes a valid `report.yaml` file
- Uses `fluxid report --get-file` to get the path (not hardcoded paths)
- Validates the report before completing (`fluxid report --validate`)
- Sets `status: PASS` only if task completed successfully
- Sets `status: FAIL` if blocked, incomplete, or errors occurred
- Documents current state clearly in FAIL reports (what's done, what blocks, what's next)

## History File: The Session Memory

The history file is **ESSENTIAL**, not optional. It's the memory mechanism that enables subsequent sessions in the implement-review loop to learn from previous attempts and avoid repeating mistakes.

### Why History is Critical

**From implement-e2e command (section 1):**
> "Read existing report and history (if present) to understand prior attempts, decisions, and known gaps."

**Purpose:**
- **Avoid repeating failures** - Document what was tried and why it failed
- **Track decisions** - Log reasoning behind architectural choices
- **Record trade-offs** - Document what was postponed and why
- **Enable continuity** - Give next iteration context to build on

Without history, each implement iteration starts blind, repeating the same failed approaches.

### What Must Go in History

**From implement-e2e command (section 4, 5, 7):**
- **Planning decisions** - "Record the plan and any explicit scope cuts or postponed items in the history file, including reasoning and trade-offs"
- **Failed attempts** - Document what approaches didn't work and why
- **Design decisions** - "Briefly log design decisions, assumptions, and trade-offs in the history file"
- **Scope changes** - "Ensure the history file contains a clear log of key decisions, scope changes, and postponed items, with enough context for future sessions to continue effectively"

### Prompt Instructions for History

**Your command file should instruct the agent to use history. Example:**

```markdown
# History

At session start, read existing history (if present) to understand prior attempts, decisions, and known gaps:
- Get path: `fluxid history --get-file`
- Get schema: `fluxid history --get-schema`

During implementation, log key decisions, failed approaches, and blockers to the history file.

Before finishing, validate: `fluxid history --validate`
```

### Example: What to Document

**Failed approach:**
```
Tried inline password validation in UI component. Caused performance issues (UI lag).
Learned: Validation must be on backend. Moving validation logic to API endpoint.
```

**Design decision based on constraints:**
```
Chose JWT tokens over session cookies for authentication.
Reason: Existing API endpoints already expect JWT bearer tokens in Authorization header.
Trade-off: Cannot invalidate tokens before expiry.
```

**Blocker that can be fixed in next iteration:**
```
Database migration failed due to foreign key constraint violation.
Cause: Migration creates child records before parent table exists.
Fix: Reorder migrations - create parent table first, then child table.
```

Next iteration reads these history events and avoids repeating the same mistakes.

## Environment Variables

fluxid sets these environment variables before executing the agent subprocess:

- **FLUXID_SESSION_ID**: Current session identifier (UUID)
- **FLUXID_TASK_FILE**: Absolute path to task file
- **FLUXID_SESSION_ROOT**: Override for session directory location
  - If set: Uses this directory for session storage
  - If not set: Default is `.fluxid/sessions/` in current directory, with fallback to `$TMPDIR/fluxid`
  - Example: `FLUXID_SESSION_ROOT=/custom/path` stores sessions in `/custom/path/<session-id>/`

## Best Practices

1. **Always instruct agents** to use CLI commands, not hardcoded paths
2. **Include validation step** in command templates
3. **Define clear PASS/FAIL criteria** in command files
4. **Handle exhaustion explicitly** - tell agent what to do when stuck
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
