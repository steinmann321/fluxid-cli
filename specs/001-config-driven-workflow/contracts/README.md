# Contracts: Config-Driven Workflow System

**Feature**: 001-config-driven-workflow
**Date**: 2026-01-18
**Phase**: 1 (Design & Contracts)

## Overview

This directory contains API contracts, schemas, and type definitions for the config-driven workflow system.

## Files

### 1. `workflow-config-schema.yaml`

YAML schema defining the structure and validation rules for the workflow configuration section in `config.yaml`.

**Purpose**:
- Defines required/optional fields
- Specifies validation rules (min/max lengths, patterns, etc.)
- Provides examples of valid and invalid configurations
- Documents error messages for each validation rule

**Usage**:
- Reference for implementing `ValidateWorkflowConfig()` function
- Documentation for users configuring workflows
- Test fixture generation (valid/invalid examples)

**Key Sections**:
- `workflow` root structure
- `steps` array schema (1..10 custom steps)
- `review` object schema (mandatory review step)
- Validation rules (V001-V010)
- Valid examples (minimal, standard, extended, infinite_retries)
- Invalid examples (for testing validation failures)

---

### 2. `go-types.go`

Go type signatures and function contracts for the config-driven workflow implementation.

**Purpose**:
- Reference implementation for types in `internal/config/` and `internal/types/`
- Function signatures for validation, initialization, execution, and logging
- Example integration patterns

**Key Sections**:
- **Configuration Types**: `WorkflowConfig`, `WorkflowStepConfig`, `ReviewStepConfig`
- **Runtime Types**: `WorkflowStep`, `Workflow`
- **Validation Functions**: `ValidateWorkflowConfig()`, `ValidateCommandPath()`
- **Workflow Functions**: `BuildWorkflow()`, `ExecuteStepWithRetry()`
- **Logging Functions**: `WorkflowLogger` with JSON/text output support
- **Example Usage**: Integration patterns for loading, validating, and running workflows

**Usage**:
- Copy type definitions to actual implementation files
- Reference function signatures during TDD test writing
- Verify implementation matches contract

---

## Contract Guarantees

### Structural Guarantees

1. **Config Structure**:
   - `workflow` section contains `steps` array and `review` object
   - `steps` array has 1..10 `WorkflowStepConfig` objects
   - Each `WorkflowStepConfig` has `name` (required), `command` (required), `retries` (optional, default 1)
   - `review` object has `command` (required), `retries` (optional, default 1)

2. **Runtime Structure**:
   - `Workflow` contains 2..11 `WorkflowStep` objects (1..10 custom + 1 review)
   - Last `WorkflowStep` has `IsReview == true`
   - All `WorkflowStep.CommandFilePath` values are absolute paths
   - `WorkflowStep.Order` reflects sequential array position (0-based)

### Validation Guarantees

1. **Startup Validation** (before workflow execution):
   - Workflow section exists (V001)
   - Review section exists (V002)
   - Review command specified (V003)
   - At least 1 custom step (V004)
   - All step names non-empty (V005)
   - All step names unique (V006)
   - All command file paths exist (V007)
   - All command files readable (V008)
   - All retries non-negative (V009)
   - Max iterations non-negative (V010)

2. **Runtime Validation**:
   - Report status is PASS or FAIL (missing/invalid treated as FAIL)
   - Agent invocation failures treated as FAIL
   - Report read failures treated as FAIL
   - Sequential execution enforced (no parallelism)

### Behavioral Guarantees

1. **Step Execution**:
   - Steps execute in array order (sequential, no parallelism)
   - Each step retries up to `Retries` times on FAIL (0 = infinite)
   - After retry exhaustion, continue to next step (not abort)
   - Review step PASS exits workflow (exit code 0)
   - Review step FAIL continues to next iteration (if iterations remain)

2. **Iteration Loop**:
   - Outer loop runs up to `MaxIterations` times (0 = infinite)
   - Each iteration executes all steps (custom + review)
   - Review PASS exits immediately (success)
   - Review FAIL continues loop (until iterations exhausted)
   - Iterations exhausted exits workflow (exit code 0)

3. **Logging**:
   - Step transitions logged (start/complete)
   - Retry attempts logged (current/max)
   - Report status logged (PASS/FAIL)
   - Command file paths logged
   - Iteration boundaries logged
   - Human-readable format (default) or JSON format (--out=json)

---

## Migration from Hardcoded Workflow

### Old Behavior (Hardcoded 3-Step)

```go
// Hardcoded loop structure
for iteration := 1; iteration <= MaxReviewCycles; iteration++ {
    runImplementPhase() // Has its own retry logic
    runCommitPhaseWithRetry() // Has its own retry logic
    runReviewPhase() // No retry logic
    if reviewStatus == PASS {
        return 0 // Success
    }
}
```

**Issues**:
- Hardcoded 3 steps (implement, commit, review)
- Duplicate retry logic in each phase
- Review has NO retries
- Not configurable without code changes

### New Behavior (Config-Driven)

```go
// Config-driven loop structure
for iteration := 1; iteration <= cfg.Workflow.MaxIterations; iteration++ {
    for _, step := range cfg.Workflow.Steps {
        ExecuteStepWithRetry(cfg, step, iteration) // Generic retry logic

        if step.IsReview && reportStatus == PASS {
            return 0 // Success
        }
    }
}
```

**Improvements**:
- N custom steps (1..10) from config.yaml
- Single generic retry function (DRY)
- Review step supports retries (configurable)
- Fully configurable via YAML (no code changes)

---

## Example Configurations

### Minimal (1 Custom Step + Review)

```yaml
workflow:
  steps:
    - name: implement
      command: commands/implement.md
  review:
    command: commands/review.md
```

### Standard (Matches Old 3-Step Behavior)

```yaml
workflow:
  steps:
    - name: implement
      command: commands/implement.md
      retries: 3
    - name: commit
      command: commands/commit.md
      retries: 100
  review:
    command: commands/review.md
    retries: 1
```

### Extended (5 Custom Steps)

```yaml
workflow:
  steps:
    - name: design
      command: commands/design.md
      retries: 2
    - name: implement
      command: commands/implement.md
      retries: 3
    - name: test
      command: commands/test.md
      retries: 1
    - name: lint
      command: commands/lint.md
      retries: 1
    - name: commit
      command: commands/commit.md
      retries: 100
  review:
    command: commands/review.md
    retries: 1
```

### Infinite Retries Example

```yaml
workflow:
  steps:
    - name: critical_step
      command: commands/critical.md
      retries: 0  # Infinite retries until PASS
  review:
    command: commands/review.md
    retries: 1
```

---

## Validation Error Examples

### Missing Workflow Section

```yaml
# Invalid: no workflow section
agent: claude
iterations: 20
```

**Error**: `workflow section is required in config.yaml`

### Missing Review Section

```yaml
# Invalid: no review section
workflow:
  steps:
    - name: implement
      command: commands/implement.md
```

**Error**: `workflow.review section is required`

### Duplicate Step Names

```yaml
# Invalid: duplicate step name
workflow:
  steps:
    - name: implement
      command: commands/implement.md
    - name: implement
      command: commands/other.md
  review:
    command: commands/review.md
```

**Error**: `duplicate step name: implement`

### Negative Retries

```yaml
# Invalid: negative retries
workflow:
  steps:
    - name: implement
      command: commands/implement.md
      retries: -1
  review:
    command: commands/review.md
```

**Error**: `retries cannot be negative`

---

## Implementation Checklist

Use this checklist during implementation to verify contract adherence:

### Configuration Loading
- [ ] `WorkflowConfig` struct added to `internal/config/config.go`
- [ ] `WorkflowStepConfig` struct added
- [ ] `ReviewStepConfig` struct added
- [ ] `ProjectConfig.Workflow` field added
- [ ] `HomeConfig.Workflow` field added
- [ ] YAML parsing handles workflow section

### Validation
- [ ] `internal/config/validate_workflow.go` created
- [ ] `ValidateWorkflowConfig()` implemented with all 10 validation rules (V001-V010)
- [ ] `ValidateCommandPath()` implemented (absolute/relative path support)
- [ ] Validation called at startup (before workflow execution)
- [ ] Clear error messages for each validation failure

### Runtime Types
- [ ] `WorkflowStep` struct added to `internal/types/types.go`
- [ ] `Workflow` struct added
- [ ] `Config.Workflow` field added
- [ ] `Config.OutputFormat` field added (for JSON output support)

### Workflow Initialization
- [ ] `BuildWorkflow()` function implemented in `internal/workflow/workflow.go`
- [ ] Command paths resolved to absolute paths
- [ ] Review step appended as final step with `IsReview == true`
- [ ] Step order assigned correctly (0-based index)

### Step Execution
- [ ] `internal/workflow/step_executor.go` created
- [ ] `ExecuteStepWithRetry()` implemented with generic retry logic
- [ ] Handles infinite retries (retries == 0)
- [ ] Treats agent failures as FAIL (FR-022)
- [ ] Continues to next step after retry exhaustion (FR-006)

### Logging
- [ ] `internal/workflow/workflow_logger.go` created
- [ ] `WorkflowLogger` struct implemented
- [ ] Human-readable output (default)
- [ ] JSON output (--out=json flag)
- [ ] All required log events (FR-020)

### Main Workflow Loop
- [ ] `internal/workflow/workflow.go` modified
- [ ] Hardcoded 3-step loop replaced with config-driven loop
- [ ] Iteration loop uses `cfg.Workflow.MaxIterations`
- [ ] Step loop iterates `cfg.Workflow.Steps`
- [ ] Review exit gate logic implemented (PASS exits, FAIL continues)

### E2E Tests
- [ ] `e2e-tests/fixtures/configs/` directory created
- [ ] Test fixture configs created (minimal, standard, extended, invalid)
- [ ] `workflow_config_driven_test.go` created
- [ ] `workflow_startup_validation_test.go` created
- [ ] Existing tests updated/replaced (m01-e01, m01-e02, m07-e01, m08-e01)

---

## Next Steps

Contracts are complete. Proceed to:
1. Generate `quickstart.md` (Phase 1)
2. Update agent context (Phase 1)
3. Re-evaluate Constitution Check (Phase 1)
