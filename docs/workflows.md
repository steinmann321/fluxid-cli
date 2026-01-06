# Workflows

Common usage patterns and workflow examples for fluxid.

## Basic Workflows

### Simple Feature Implementation

```bash
# 1. Initialize (first time only)
fluxid init

# 2. Create task file
cat > /tmp/add-feature.md <<EOF
# Add Login Feature

Implement user authentication with:
- Login form UI
- API endpoint
- Session management
EOF

# 3. Run workflow
fluxid --claude --file=/tmp/add-feature.md
```

### Project-Specific Configuration

```bash
cd /path/to/project

# 1. Create project config
fluxid init .

# 2. Customize for project
vim ./.fluxid/config.yaml
vim ./.fluxid/commands/implement.md

# 3. Run with project settings
fluxid --claude --file=/tmp/task.md
```

### Named Sessions

```bash
# Use named session for tracking
export FLUXID_SESSION_ID=feature-auth
fluxid --claude --file=/tmp/auth-feature.md

# Later, inspect session files
REPORT=$(fluxid report --get-file)
HISTORY=$(fluxid history --get-file)
cat "$REPORT"
cat "$HISTORY"
```

---

## Dry-Run Workflows

### Preview Without Execution

```bash
# See what would happen
fluxid --claude --file=/tmp/task.md --fluxid-dry-run

# Get structured preview
fluxid --claude --file=/tmp/task.md --fluxid-dry-run --output=json

# Check iteration plan
fluxid --claude --file=/tmp/task.md --fluxid-dry-run --fluxid-iterations=5
```

---

## Multi-Agent Workflows

### Use Different Agents for Different Tasks

```bash
# Claude for implementation
fluxid --claude --file=/tmp/implement-api.md --session=api-impl

# Codex for code generation
fluxid --codex --file=/tmp/generate-tests.md --session=test-gen

# OpenCode for refactoring
fluxid --opencode --file=/tmp/refactor.md --session=refactor-code
```

### Agent Selection Strategy

**Claude:** Best for complex reasoning, architectural decisions
```bash
fluxid --claude --file=/tmp/design-system.md
```

**Codex:** Best for code generation, boilerplate
```bash
fluxid --codex --file=/tmp/generate-crud.md
```

**OpenCode:** Best for refactoring, optimization
```bash
fluxid --opencode --file=/tmp/optimize-performance.md
```

---

## Iteration Control

### Tight Iteration Limits

```bash
# Quick task with strict limits
fluxid --claude --file=/tmp/small-fix.md \
  --fluxid-iterations=3 \
  --fluxid-implement-retries=1
```

### Generous Limits for Exploration

```bash
# Complex task with room for iteration
fluxid --claude --file=/tmp/complex-feature.md \
  --fluxid-iterations=30 \
  --fluxid-implement-retries=5
```

---

## Output Formats

### JSON for Scripting

```bash
# Capture workflow result
RESULT=$(fluxid --claude --file=/tmp/task.md --output=json)

# Extract fields
SESSION=$(echo "$RESULT" | jq -r '.session_id')
STATUS=$(echo "$RESULT" | jq -r '.status')
ITERATIONS=$(echo "$RESULT" | jq -r '.iterations')

echo "Session $SESSION completed with status: $STATUS after $ITERATIONS iterations"
```

### YAML for Readability

```bash
# Human-readable structured output
fluxid --claude --file=/tmp/task.md --output=yaml
```

### Text for Interactive Use

```bash
# Default: plain text with progress indicators
fluxid --claude --file=/tmp/task.md
```

---

## Session Management

### Working with Sessions

```bash
# 1. Start workflow with named session
export FLUXID_SESSION_ID=feature-123
fluxid --claude --file=/tmp/task.md

# 2. Access session files (in another terminal or after completion)
REPORT=$(fluxid report --get-file)
HISTORY=$(fluxid history --get-file)

# 3. Inspect results
cat "$REPORT"
cat "$HISTORY"

# 4. Validate outputs
fluxid report --validate
fluxid history --validate
```

### Session Directory Structure

```
$HOME/.fluxid/sessions/           # Or custom via FLUXID_SESSION_ROOT
└── feature-123/                  # Session ID
    ├── report.yaml               # Agent status report
    └── history.yaml              # Workflow event history
```

### Custom Session Root

```bash
# Use project-local sessions
export FLUXID_SESSION_ROOT=./sessions
fluxid --claude --file=/tmp/task.md

# Sessions created in:
# ./sessions/<session-id>/report.yaml
# ./sessions/<session-id>/history.yaml
```

---

## Integration Patterns

### CI/CD Pipeline

```bash
#!/bin/bash
# .gitlab-ci.yml or GitHub Actions workflow

set -e

# Run fluxid with JSON output
RESULT=$(fluxid --claude --file=task.md --output=json)

# Check if workflow succeeded
STATUS=$(echo "$RESULT" | jq -r '.status')
if [[ "$STATUS" != "success" ]]; then
  echo "Workflow failed"
  exit 1
fi

# Get session for artifact collection
SESSION=$(echo "$RESULT" | jq -r '.session_id')
echo "Session: $SESSION"
```

### Monitoring and Observability

```bash
# Stream workflow status
fluxid --claude --file=/tmp/task.md --output=json > workflow.json

# Monitor in real-time (separate terminal)
watch -n 1 'cat $(fluxid report --get-file 2>/dev/null) 2>/dev/null | head -20'
```

---

## Advanced Workflows

### Multi-Phase Development

```bash
# Phase 1: Implementation
export FLUXID_SESSION_ID=project-impl
fluxid --claude --file=/tmp/phase1-impl.md

# Phase 2: Tests
export FLUXID_SESSION_ID=project-tests
fluxid --codex --file=/tmp/phase2-tests.md

# Phase 3: Documentation
export FLUXID_SESSION_ID=project-docs
fluxid --claude --file=/tmp/phase3-docs.md
```

### Iterative Refinement

```bash
# Round 1: Initial implementation
fluxid --claude --file=/tmp/v1-requirements.md --session=feature-v1

# Review results
cat $(fluxid report --get-file)

# Round 2: Refinement based on review
fluxid --claude --file=/tmp/v2-refinements.md --session=feature-v2

# Round 3: Final polish
fluxid --claude --file=/tmp/v3-polish.md --session=feature-v3
```

---

## Error Recovery

### Handling Failed Workflows

```bash
# Workflow may fail after exhausting retries
fluxid --claude --file=/tmp/task.md --session=attempt-1

# Check what went wrong
cat $(fluxid report --get-file)
cat $(fluxid history --get-file)

# Adjust task and retry
fluxid --claude --file=/tmp/task-revised.md --session=attempt-2
```

### Validation Failures

```bash
# Run workflow
fluxid --claude --file=/tmp/task.md

# Validate outputs
if ! fluxid report --validate; then
  echo "Report validation failed"
  cat $(fluxid report --get-file)
  exit 1
fi

if ! fluxid history --validate; then
  echo "History validation failed"
  cat $(fluxid history --get-file)
  exit 1
fi
```

---

## Testing Workflows

### Test with Dry-Run

```bash
# Always dry-run first for new tasks
fluxid --claude --file=/tmp/new-task.md --fluxid-dry-run

# Verify plan looks correct
fluxid --claude --file=/tmp/new-task.md --fluxid-dry-run --output=yaml

# Then run for real
fluxid --claude --file=/tmp/new-task.md
```

### Test with Tight Limits

```bash
# Quick test run with minimal iterations
fluxid --claude --file=/tmp/task.md \
  --fluxid-iterations=2 \
  --fluxid-implement-retries=1 \
  --fluxid-dry-run
```

---

## Configuration Patterns

### Per-Task Configuration

```bash
# Override config for specific task
fluxid --codex --file=/tmp/task.md \
  --fluxid-iterations=10 \
  --fluxid-implement-retries=2
```

### Environment-Specific Workflows

```bash
# Development: generous limits
if [[ "$ENV" == "development" ]]; then
  ITERATIONS=50
  RETRIES=5
else
  # Production: strict limits
  ITERATIONS=10
  RETRIES=2
fi

fluxid --claude --file=/tmp/task.md \
  --fluxid-iterations=$ITERATIONS \
  --fluxid-implement-retries=$RETRIES
```

---

## Complete Example: Feature Development

```bash
#!/bin/bash
# Complete workflow for adding a feature

set -e

PROJECT_ROOT="/path/to/project"
cd "$PROJECT_ROOT"

# 1. Setup
echo "Initializing project config..."
fluxid init .

# 2. Create task file
echo "Creating task specification..."
cat > /tmp/feature-task.md <<EOF
# Add User Profile Feature

## Requirements
- User profile page showing user info
- Edit profile functionality
- Avatar upload
- Privacy settings

## Technical Requirements
- Use existing auth system
- Follow project conventions
- Add tests
- Update documentation
EOF

# 3. Run workflow with named session
echo "Running implementation workflow..."
export FLUXID_SESSION_ID="feature-user-profile"
fluxid --claude --file=/tmp/feature-task.md \
  --fluxid-iterations=20 \
  --output=json > /tmp/workflow-result.json

# 4. Check results
STATUS=$(cat /tmp/workflow-result.json | jq -r '.status')
echo "Workflow status: $STATUS"

if [[ "$STATUS" != "success" ]]; then
  echo "Workflow failed. Checking reports..."
  cat $(fluxid report --get-file)
  cat $(fluxid history --get-file)
  exit 1
fi

# 5. Validate outputs
echo "Validating workflow outputs..."
fluxid report --validate || exit 1
fluxid history --validate || exit 1

# 6. Success
echo "Feature implementation complete!"
echo "Session: feature-user-profile"
echo "Report: $(fluxid report --get-file)"
echo "History: $(fluxid history --get-file)"
```

---

## Best Practices

1. **Always initialize config first:** Run `fluxid init` before first use
2. **Use named sessions:** Makes debugging and tracking easier
3. **Dry-run new tasks:** Preview workflow before execution
4. **Validate outputs:** Use `--validate` to catch issues early
5. **Use JSON for automation:** Easier to parse than text output
6. **Set appropriate limits:** Tighter for simple tasks, generous for complex
7. **Check history on failure:** Understand what went wrong
8. **Keep task files:** Version control task specifications
9. **Use project configs:** Keep team aligned on workflow settings
10. **Monitor sessions:** Watch report files for progress tracking
