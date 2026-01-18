# Research: Config-Driven Workflow System

**Feature**: 001-config-driven-workflow
**Date**: 2026-01-18
**Phase**: 0 (Outline & Research)

## Overview

This document consolidates research findings to resolve "NEEDS CLARIFICATION" items from the Technical Context section and documents technology decisions for implementing the config-driven workflow system.

## Research Tasks Completed

### 1. YAML Configuration Structure and Parsing

**Decision**: Use gopkg.in/yaml.v3 with strict struct tags for workflow configuration

**Rationale**:
- gopkg.in/yaml.v3 is already a project dependency (existing usage in internal/config/)
- Provides strong type safety through struct tags
- Supports custom validation via struct tag validation
- Clear error messages for malformed YAML

**Implementation Pattern**:
```go
type WorkflowConfig struct {
    Steps  []WorkflowStepConfig `yaml:"steps"`
    Review ReviewStepConfig     `yaml:"review"`
}

type WorkflowStepConfig struct {
    Name    string `yaml:"name"`
    Command string `yaml:"command"`
    Retries int    `yaml:"retries,omitempty"` // Default: 1
}

type ReviewStepConfig struct {
    Command string `yaml:"command"`
    Retries int    `yaml:"retries,omitempty"` // Default: 1
}
```

**Alternatives Considered**:
- **Manual map[string]interface{} parsing**: Rejected due to lack of type safety and verbose error-prone code
- **Third-party validation libraries (e.g., go-playground/validator)**: Rejected to minimize dependencies; custom validation is straightforward for this use case

---

### 2. Generic Step Execution with Retry Logic

**Decision**: Implement a single `executeStepWithRetry()` function that handles retry logic for any workflow step

**Rationale**:
- DRY principle: eliminates duplicate retry logic across implement/commit/review phases
- Current implementation has three separate retry implementations (runImplementPhase, runCommitPhaseWithRetry, review has none)
- Simplifies testing: single function to verify retry behavior
- Matches Constitution principle IV (SOLID, DRY, KISS)

**Implementation Pattern**:
```go
// Generic step executor with retry logic
func executeStepWithRetry(cfg types.Config, step WorkflowStep, iteration int) error {
    maxRetries := step.Retries
    if maxRetries == 0 {
        maxRetries = math.MaxInt // Infinite retries
    }

    for retry := 1; retry <= maxRetries; retry++ {
        // Execute step (agent invocation)
        exitCode, err := runPhase(cfg, step.Name, step.CommandFilePath)

        // Check report status
        status, err := waitForValidReport(cfg.SessionID, cfg.SessionRoot, step.Name)
        if err != nil {
            log.Printf("Failed to read %s report (treating as FAIL): %v", step.Name, err)
            status = statusFail
        }

        // Log attempt
        logStepAttempt(step.Name, retry, maxRetries, status)

        if status == statusPass {
            return nil // Success
        }

        // Status == FAIL: continue to next retry (or exit if exhausted)
    }

    // All retries exhausted with FAIL status
    logRetriesExhausted(step.Name, maxRetries)
    return nil // Continue to next step (FR-006)
}
```

**Alternatives Considered**:
- **Keep separate functions for each step type**: Rejected due to code duplication and violation of DRY principle
- **Async/concurrent step execution**: Rejected due to Constitution principle III (strictly sequential workflow)

---

### 3. Startup Validation Strategy

**Decision**: Implement comprehensive validation at startup before any workflow execution begins

**Rationale**:
- Constitution principle VII: fail fast with clear diagnostics
- Prevents wasted agent invocations on invalid configurations
- Enables clear, actionable error messages before resource allocation
- Matches existing pattern in internal/config/validate_*.go files

**Validation Checklist** (FR-011, FR-015, FR-018, FR-023):
1. ✅ Workflow section exists in config.yaml (FR-010)
2. ✅ Workflow.review section exists (FR-018)
3. ✅ Review.command path specified and readable (FR-017)
4. ✅ At least one custom workflow step defined (FR-011)
5. ✅ All step names non-empty and unique (FR-011)
6. ✅ All command file paths exist and readable (FR-015)
7. ✅ max_iterations is non-negative (FR-019)
8. ✅ All retries values are non-negative (FR-023)

**Implementation Pattern**:
```go
func ValidateWorkflowConfig(cfg *WorkflowConfig, configDir string) error {
    if cfg == nil {
        return fmt.Errorf("workflow section is required in config.yaml")
    }

    if cfg.Review == nil {
        return fmt.Errorf("workflow.review section is required")
    }

    if cfg.Review.Command == "" {
        return fmt.Errorf("workflow.review.command is required")
    }

    if len(cfg.Steps) == 0 {
        return fmt.Errorf("at least one custom workflow step is required before review")
    }

    // Validate uniqueness of step names
    stepNames := make(map[string]bool)
    for _, step := range cfg.Steps {
        if strings.TrimSpace(step.Name) == "" {
            return fmt.Errorf("step name cannot be empty or whitespace-only")
        }
        if stepNames[step.Name] {
            return fmt.Errorf("duplicate step name: %s", step.Name)
        }
        stepNames[step.Name] = true

        // Validate command file path
        if err := validateCommandPath(step.Command, configDir); err != nil {
            return fmt.Errorf("step %s: %w", step.Name, err)
        }

        // Validate retries
        if step.Retries < 0 {
            return fmt.Errorf("step %s: retries cannot be negative", step.Name)
        }
    }

    // Validate review command path
    if err := validateCommandPath(cfg.Review.Command, configDir); err != nil {
        return fmt.Errorf("review step: %w", err)
    }

    // Validate review retries
    if cfg.Review.Retries < 0 {
        return fmt.Errorf("review step: retries cannot be negative")
    }

    return nil
}

func validateCommandPath(path string, configDir string) error {
    // Resolve relative paths from config.yaml directory
    resolvedPath := path
    if !filepath.IsAbs(path) {
        resolvedPath = filepath.Join(configDir, path)
    }

    // Check file exists and is readable
    fileInfo, err := os.Stat(resolvedPath)
    if err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("command file not found: %s", path)
        }
        return fmt.Errorf("command file not readable: %s", path)
    }

    // Verify it's a file (not directory)
    if fileInfo.IsDir() {
        return fmt.Errorf("command path is a directory, not a file: %s", path)
    }

    return nil
}
```

**Alternatives Considered**:
- **Lazy validation during execution**: Rejected because it wastes resources and violates fail-fast principle
- **Lenient validation with warnings**: Rejected because unclear failures undermine trust (Constitution VII)

---

### 4. Relative Path Resolution for Command Files

**Decision**: Support both absolute and relative paths; resolve relative paths from config.yaml directory location

**Rationale**:
- Matches FR-015 and A-005 requirements
- Makes workflow configurations portable across environments
- Follows existing pattern in internal/config/resolve_commands.go
- Common convention in configuration files (e.g., Docker Compose, Kubernetes, etc.)

**Implementation Pattern**:
```go
func resolveCommandPath(commandPath string, configDir string) string {
    if filepath.IsAbs(commandPath) {
        return commandPath
    }
    return filepath.Clean(filepath.Join(configDir, commandPath))
}
```

**Alternatives Considered**:
- **Only absolute paths**: Rejected because it reduces portability and user convenience
- **Relative to working directory**: Rejected because config location should be the reference point

---

### 5. Logging Strategy for Workflow Execution

**Decision**: Implement structured logging with both human-readable (default) and JSON output formats

**Rationale**:
- FR-020 requires detailed logging: step transitions, retry attempts, report status, command paths, iteration boundaries
- FR-021 requires --out=json flag support for programmatic parsing
- Constitution principle VIII: CLI-first design requires parseable output
- Matches existing logging pattern in internal/workflow/workflow.go

**Implementation Pattern**:
```go
type WorkflowLogger struct {
    outputFormat string // "text" or "json"
}

func (l *WorkflowLogger) LogStepStart(stepName string, iteration int, retry int) {
    if l.outputFormat == "json" {
        log.Printf(`{"event":"step_start","step":"%s","iteration":%d,"retry":%d}`,
            stepName, iteration, retry)
    } else {
        log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        log.Printf("▶ Starting step: %s (iteration %d, retry %d)", stepName, iteration, retry)
    }
}

func (l *WorkflowLogger) LogStepComplete(stepName string, status string, retry int, maxRetries int) {
    if l.outputFormat == "json" {
        log.Printf(`{"event":"step_complete","step":"%s","status":"%s","retry":%d,"max_retries":%d}`,
            stepName, status, retry, maxRetries)
    } else {
        log.Printf("✓ Step complete: %s | Status: %s | Retry: %d/%d",
            stepName, status, retry, maxRetries)
    }
}

func (l *WorkflowLogger) LogIterationStart(iteration int, maxIterations int) {
    if l.outputFormat == "json" {
        log.Printf(`{"event":"iteration_start","iteration":%d,"max_iterations":%d}`,
            iteration, maxIterations)
    } else {
        log.Printf("\n")
        log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        log.Printf("🔄 DEVELOPMENT ITERATION %d/%d", iteration, maxIterations)
        log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    }
}
```

**Alternatives Considered**:
- **Third-party logging libraries (e.g., zap, logrus)**: Rejected to minimize dependencies; standard log package is sufficient
- **Verbose/debug levels**: Not required for MVP; current implementation uses log.Printf consistently

---

### 6. E2E Testing Strategy for Config-Driven Workflow

**Decision**: Create new E2E tests using fixture config.yaml files that define custom workflows; update/replace existing hardcoded tests

**Rationale**:
- User instruction: "ensure exiting e2e tests are droped/recreated as needed becuase they show a specific config"
- Constitution principle II: full E2E test coverage required
- Existing tests in e2e-tests/ use hardcoded 3-step workflow assumptions
- Fixture-based approach allows testing various workflow configurations without code changes

**Test Fixture Structure**:
```text
e2e-tests/fixtures/configs/
├── workflow_minimal.yaml              # 1 custom step + review
├── workflow_standard.yaml             # 3 custom steps + review (matches old behavior)
├── workflow_extended.yaml             # 5 custom steps + review
├── workflow_infinite_retries.yaml     # Step with retries=0
├── workflow_no_retries.yaml           # All steps with retries=1
└── workflow_review_only.yaml          # INVALID: no custom steps (validation test)
```

**Test Coverage Matrix**:

| Test Case | Fixture Config | Validates |
|-----------|---------------|-----------|
| TestConfigDrivenWorkflowMinimal | workflow_minimal.yaml | Single custom step execution, review gate |
| TestConfigDrivenWorkflowStandard | workflow_standard.yaml | 3-step workflow (backward compatibility) |
| TestConfigDrivenWorkflowExtended | workflow_extended.yaml | Extended workflow (5+ steps) |
| TestConfigDrivenWorkflowRetries | workflow_standard.yaml (modified) | Step retry exhaustion behavior |
| TestConfigDrivenWorkflowInfiniteRetries | workflow_infinite_retries.yaml | retries=0 infinite retry behavior |
| TestStartupValidationMissingWorkflow | (no workflow section) | FR-010: fail on missing workflow |
| TestStartupValidationMissingReview | (no review section) | FR-018: fail on missing review |
| TestStartupValidationDuplicateNames | (duplicate step names) | FR-011: fail on duplicate names |
| TestStartupValidationInvalidPaths | (nonexistent command files) | FR-015: fail on invalid paths |
| TestStartupValidationEmptyStepName | (empty step name) | FR-011: fail on empty names |
| TestStartupValidationZeroCustomSteps | workflow_review_only.yaml | FR-011: fail on zero custom steps |
| TestStartupValidationNegativeRetries | (retries=-1) | FR-023: fail on negative retries |
| TestStartupValidationNegativeIterations | (max_iterations=-1) | FR-019: fail on negative iterations |

**Existing Tests to Update/Replace**:
1. `m01-e01-user-runs-workflow-to-completion_test.go` → Replace with `workflow_config_driven_test.go` using fixture configs
2. `m01-e02-user-configures-loop-counts-and-runs-workflow_test.go` → Update to use workflow.steps from config
3. `m07-e01-implement-retries-exhausted-continues-through-commit_test.go` → Update to test generic step retry exhaustion
4. `m08-e01-workflow-completes-with-all-fail-reports_test.go` → Update to use custom workflow configs

**Alternatives Considered**:
- **Programmatic config generation in tests**: Rejected because fixture files are more readable and maintainable
- **Keep existing tests unchanged**: Rejected because they hardcode 3-step workflow assumptions (violates user instruction)

---

### 7. Backward Compatibility Strategy

**Decision**: NO backward compatibility - workflow section is required in config.yaml

**Rationale**:
- User clarification (spec.md line 19): "No backward compatibility - workflow section is required; users must explicitly configure workflow steps including the mandatory review step"
- FR-010 explicitly requires workflow section
- Simplifies implementation: no complex fallback logic
- Clearer user intent: explicit configuration prevents surprises

**Migration Path**:
Users must update their config.yaml to include workflow section. Example migration:

**Old config.yaml (implicit 3-step workflow)**:
```yaml
agent: claude
implement_retries: 3
commit_retries: 100
iterations: 20
commands:
  implement: commands/implement.md
  commit: commands/commit.md
  review: commands/review.md
```

**New config.yaml (explicit workflow)**:
```yaml
agent: claude
iterations: 20
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

**Alternatives Considered**:
- **Auto-generate default workflow from old config**: Rejected per user requirement (line 19 in spec.md)
- **Support both old and new config formats**: Rejected to avoid complexity and confusion

---

### 8. Agent Invocation Error Handling

**Decision**: Treat agent invocation failures as FAIL status and retry according to step's retry limit

**Rationale**:
- FR-022: catastrophic execution failures follow same retry-then-continue behavior as report-based FAIL
- Consistent error handling across all failure modes (report FAIL vs. agent crash)
- Transient errors (network issues, temporary agent unavailability) get retries
- Permanent failures eventually proceed to next step (FR-006)

**Implementation Pattern**:
```go
func executeStepWithRetry(cfg types.Config, step WorkflowStep, iteration int) error {
    for retry := 1; retry <= step.Retries; retry++ {
        // Execute step
        exitCode, execErr := runPhase(cfg, step.Name, step.CommandFilePath)

        // If agent invocation failed catastrophically, treat as FAIL
        if execErr != nil {
            log.Printf("Agent invocation failed for step %s (retry %d/%d): %v",
                step.Name, retry, step.Retries, execErr)
            // Continue to retry logic (same as FAIL status)
            continue
        }

        // Check report status
        status, err := waitForValidReport(cfg.SessionID, cfg.SessionRoot, step.Name)
        if err != nil {
            log.Printf("Failed to read %s report (treating as FAIL): %v", step.Name, err)
            status = statusFail
        }

        logStepAttempt(step.Name, retry, step.Retries, status)

        if status == statusPass {
            return nil
        }
    }

    return nil // Retries exhausted, continue to next step
}
```

**Alternatives Considered**:
- **Abort workflow on agent invocation failures**: Rejected because it's too aggressive (transient errors should be retryable)
- **Different retry behavior for agent failures vs. report FAIL**: Rejected for consistency (FR-022)

---

## Technology Stack Summary

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| Language | Go | 1.25 | Pure Go implementation (Constitution V) |
| YAML Parser | gopkg.in/yaml.v3 | v3 | Existing dependency, strong type safety |
| Session ID Generation | github.com/google/uuid | v1 | Existing dependency, UUID v4 standard |
| Testing Framework | go test | stdlib | Native Go testing, no external deps |
| Logging | log | stdlib | Simple, sufficient for requirements |
| File I/O | os, filepath | stdlib | Cross-platform path handling |

**No New Dependencies Required** - All functionality can be implemented using existing dependencies and Go standard library.

---

## Open Questions (Resolved)

All "NEEDS CLARIFICATION" items from Technical Context have been resolved:

1. ✅ YAML structure and parsing → gopkg.in/yaml.v3 with strict structs
2. ✅ Generic step execution → Single `executeStepWithRetry()` function
3. ✅ Startup validation → Comprehensive validation before workflow execution
4. ✅ Relative path resolution → Resolve from config.yaml directory
5. ✅ Logging strategy → Human-readable (default) + JSON (--out=json)
6. ✅ E2E testing → Fixture-based configs, replace existing hardcoded tests
7. ✅ Backward compatibility → None required (explicit workflow config mandatory)
8. ✅ Agent error handling → Treat as FAIL status, retry according to step config

---

## Next Steps

Phase 0 research is complete. All unknowns resolved. Proceed to Phase 1: Design & Contracts.
