# Quickstart: Config-Driven Workflow System

**Feature**: 001-config-driven-workflow
**Date**: 2026-01-18
**Audience**: Developers implementing the config-driven workflow feature

## Overview

This quickstart provides a step-by-step guide for implementing the config-driven workflow system in fluxid. Follow TDD principles: write tests first, then implement to pass those tests.

---

## Prerequisites

- Go 1.25 installed
- Existing fluxid codebase cloned
- Familiarity with TDD (Red-Green-Refactor cycle)
- Read the following documents:
  - `spec.md` (feature requirements)
  - `research.md` (technology decisions)
  - `data-model.md` (entity definitions)
  - `contracts/` (schemas and types)

---

## Implementation Phases

### Phase 0: Setup and Planning ✅

**Status**: Complete (documents generated)

**Artifacts**:
- `plan.md` - Implementation plan
- `research.md` - Technology decisions
- `data-model.md` - Entity definitions
- `contracts/` - Schemas and type contracts

---

### Phase 1: E2E Test Fixtures (TDD: Red)

**Goal**: Create test fixture configs and failing E2E tests BEFORE implementing any production code

**Duration Estimate**: DO NOT ESTIMATE (see constitution)

**Steps**:

1. **Create test fixture directory**:
   ```bash
   mkdir -p e2e-tests/fixtures/configs
   ```

2. **Create fixture config files** (based on `contracts/workflow-config-schema.yaml` examples):

   **Minimal config** (`e2e-tests/fixtures/configs/workflow_minimal.yaml`):
   ```yaml
   agent: claude
   iterations: 2
   workflow:
     steps:
       - name: implement
         command: commands/implement.md
     review:
       command: commands/review.md
   ```

   **Standard config** (`e2e-tests/fixtures/configs/workflow_standard.yaml`):
   ```yaml
   agent: claude
   iterations: 2
   workflow:
     steps:
       - name: implement
         command: commands/implement.md
         retries: 3
       - name: commit
         command: commands/commit.md
         retries: 5
     review:
       command: commands/review.md
       retries: 1
   ```

   **Extended config** (`e2e-tests/fixtures/configs/workflow_extended.yaml`):
   ```yaml
   agent: claude
   iterations: 2
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
     review:
       command: commands/review.md
       retries: 1
   ```

   **Invalid configs** (for validation testing):
   - `workflow_no_workflow_section.yaml` (missing workflow)
   - `workflow_no_review.yaml` (missing review section)
   - `workflow_duplicate_names.yaml` (duplicate step names)
   - `workflow_negative_retries.yaml` (retries = -1)

3. **Write failing E2E tests**:

   **File**: `e2e-tests/tests/workflow_config_driven_test.go`

   ```go
   package tests

   import (
       "os/exec"
       "strings"
       "testing"
   )

   func TestConfigDrivenWorkflowMinimal(t *testing.T) {
       // Test: Minimal workflow (1 custom step + review)
       // Expected: implement → review → exit
       cmd := exec.Command("fluxid", "run",
           "--config", "fixtures/configs/workflow_minimal.yaml",
           "test-task.txt")
       output, err := cmd.CombinedOutput()
       if err != nil {
           t.Fatalf("Command failed: %v\nOutput: %s", err, output)
       }

       outputStr := string(output)

       // Verify steps executed in order
       if !strings.Contains(outputStr, "Starting step: implement") {
           t.Error("Implement step not executed")
       }
       if !strings.Contains(outputStr, "Starting step: review") {
           t.Error("Review step not executed")
       }

       // Verify review is last step
       implementIdx := strings.Index(outputStr, "Starting step: implement")
       reviewIdx := strings.Index(outputStr, "Starting step: review")
       if implementIdx >= reviewIdx {
           t.Error("Review step did not execute after implement step")
       }
   }

   func TestConfigDrivenWorkflowStandard(t *testing.T) {
       // Test: Standard workflow (matches old 3-step behavior)
       // Expected: implement → commit → review → exit
       // Similar structure to TestConfigDrivenWorkflowMinimal
   }

   func TestConfigDrivenWorkflowExtended(t *testing.T) {
       // Test: Extended workflow (5 steps)
       // Expected: design → implement → test → review → exit
   }

   func TestConfigDrivenWorkflowStepRetries(t *testing.T) {
       // Test: Step retry exhaustion behavior
       // Use mock agent that returns FAIL for specific step
       // Verify step retries up to configured limit
       // Verify workflow continues to next step after exhaustion
   }
   ```

   **File**: `e2e-tests/tests/workflow_startup_validation_test.go`

   ```go
   package tests

   import (
       "os/exec"
       "strings"
       "testing"
   )

   func TestStartupValidationMissingWorkflow(t *testing.T) {
       // Test: Missing workflow section
       // Expected: Exit with error "workflow section is required"
       cmd := exec.Command("fluxid", "run",
           "--config", "fixtures/configs/workflow_no_workflow_section.yaml",
           "test-task.txt")
       output, err := cmd.CombinedOutput()

       if err == nil {
           t.Fatal("Expected command to fail, but it succeeded")
       }

       outputStr := string(output)
       if !strings.Contains(outputStr, "workflow section is required") {
           t.Errorf("Expected error message not found. Got: %s", outputStr)
       }
   }

   func TestStartupValidationMissingReview(t *testing.T) {
       // Test: Missing review section
       // Expected: Exit with error "workflow.review section is required"
   }

   func TestStartupValidationDuplicateNames(t *testing.T) {
       // Test: Duplicate step names
       // Expected: Exit with error "duplicate step name: {name}"
   }

   func TestStartupValidationNegativeRetries(t *testing.T) {
       // Test: Negative retries value
       // Expected: Exit with error "retries cannot be negative"
   }
   ```

4. **Run tests to verify they fail (RED)**:
   ```bash
   cd e2e-tests
   go test -v ./tests/workflow_config_driven_test.go
   go test -v ./tests/workflow_startup_validation_test.go
   ```

   **Expected**: All tests FAIL (config parsing not implemented yet)

---

### Phase 2: Configuration Types (TDD: Green - Part 1)

**Goal**: Add workflow config types and YAML parsing

**Steps**:

1. **Add workflow config types** to `internal/config/config.go`:

   ```go
   // Copy from contracts/go-types.go
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

2. **Add Workflow field** to `ProjectConfig` and `HomeConfig`:

   ```go
   type ProjectConfig struct {
       // ... existing fields ...
       Workflow *WorkflowConfig `yaml:"workflow"` // NEW
   }

   type HomeConfig struct {
       // ... existing fields ...
       Workflow *WorkflowConfig `yaml:"workflow"` // NEW
   }
   ```

3. **Test YAML parsing**:
   ```bash
   go test -v internal/config/config_test.go
   ```

   Write unit tests to verify WorkflowConfig unmarshals correctly from YAML.

---

### Phase 3: Validation Logic (TDD: Green - Part 2)

**Goal**: Implement startup validation for workflow config

**Steps**:

1. **Create validation file**: `internal/config/validate_workflow.go`

   ```go
   package config

   import (
       "fmt"
       "os"
       "path/filepath"
       "strings"
   )

   // ValidateWorkflowConfig validates the workflow configuration at startup
   func ValidateWorkflowConfig(cfg *WorkflowConfig, configDir string) error {
       // V001: Workflow section exists
       if cfg == nil {
           return fmt.Errorf("workflow section is required in config.yaml")
       }

       // V002: Review section exists
       if cfg.Review == nil {
           return fmt.Errorf("workflow.review section is required")
       }

       // V003: Review command specified
       if cfg.Review.Command == "" {
           return fmt.Errorf("workflow.review.command is required")
       }

       // V004: At least 1 custom step
       if len(cfg.Steps) == 0 {
           return fmt.Errorf("at least one custom workflow step is required before review")
       }

       // V005: Step names non-empty
       // V006: Step names unique
       stepNames := make(map[string]bool)
       for _, step := range cfg.Steps {
           if strings.TrimSpace(step.Name) == "" {
               return fmt.Errorf("step name cannot be empty or whitespace-only")
           }
           if stepNames[step.Name] {
               return fmt.Errorf("duplicate step name: %s", step.Name)
           }
           stepNames[step.Name] = true

           // V007, V008: Command file exists and readable
           if err := ValidateCommandPath(step.Command, configDir); err != nil {
               return fmt.Errorf("step %s: %w", step.Name, err)
           }

           // V009: Retries non-negative
           if step.Retries < 0 {
               return fmt.Errorf("step %s: retries cannot be negative", step.Name)
           }
       }

       // Validate review command path
       if err := ValidateCommandPath(cfg.Review.Command, configDir); err != nil {
           return fmt.Errorf("review step: %w", err)
       }

       // Validate review retries
       if cfg.Review.Retries < 0 {
           return fmt.Errorf("review step: retries cannot be negative")
       }

       return nil
   }

   // ValidateCommandPath validates a command file path
   func ValidateCommandPath(path string, configDir string) error {
       resolvedPath := path
       if !filepath.IsAbs(path) {
           resolvedPath = filepath.Join(configDir, path)
       }

       fileInfo, err := os.Stat(resolvedPath)
       if err != nil {
           if os.IsNotExist(err) {
               return fmt.Errorf("command file not found: %s", path)
           }
           return fmt.Errorf("command file not readable: %s", path)
       }

       if fileInfo.IsDir() {
           return fmt.Errorf("command path is a directory, not a file: %s", path)
       }

       return nil
   }
   ```

2. **Add validation call** to `internal/command/root_config_loader.go`:

   ```go
   func loadAndResolveConfig() (types.Config, int) {
       // ... existing config loading ...

       // NEW: Validate workflow config
       configDir := getConfigDir(projectConfig, homeConfig)
       workflowCfg := getWorkflowConfig(projectConfig, homeConfig)

       if err := config.ValidateWorkflowConfig(workflowCfg, configDir); err != nil {
           log.Fatalf("Workflow configuration error: %v", err)
           return types.Config{}, 1
       }

       // ... rest of config resolution ...
   }
   ```

3. **Run validation tests**:
   ```bash
   go test -v e2e-tests/tests/workflow_startup_validation_test.go
   ```

   **Expected**: Validation tests now PASS (startup validation working)

---

### Phase 4: Runtime Types (TDD: Green - Part 3)

**Goal**: Add runtime workflow types and initialization logic

**Steps**:

1. **Add runtime types** to `internal/types/types.go`:

   ```go
   // Copy from contracts/go-types.go
   type WorkflowStep struct {
       Name            string
       CommandFilePath string
       Retries         int
       IsReview        bool
       Order           int
   }

   type Workflow struct {
       Steps            []WorkflowStep
       MaxIterations    int
       CurrentIteration int
   }
   ```

2. **Add Workflow field** to `Config`:

   ```go
   type Config struct {
       // ... existing fields ...
       Workflow     *Workflow // NEW
       OutputFormat string    // NEW (for JSON output support)
   }
   ```

3. **Create workflow initialization** in `internal/workflow/workflow.go`:

   ```go
   func BuildWorkflow(cfg *config.WorkflowConfig, configDir string, maxIterations int) (*Workflow, error) {
       steps := make([]WorkflowStep, 0, len(cfg.Steps)+1)

       // Add custom steps
       for i, stepCfg := range cfg.Steps {
           retries := stepCfg.Retries
           if retries == 0 {
               retries = math.MaxInt // Infinite retries
           } else if retries < 0 {
               retries = 1 // Default
           }

           steps = append(steps, WorkflowStep{
               Name:            stepCfg.Name,
               CommandFilePath: resolveCommandPath(stepCfg.Command, configDir),
               Retries:         retries,
               IsReview:        false,
               Order:           i,
           })
       }

       // Add review step (mandatory, always last)
       reviewRetries := cfg.Review.Retries
       if reviewRetries == 0 {
           reviewRetries = math.MaxInt
       } else if reviewRetries < 0 {
           reviewRetries = 1
       }

       steps = append(steps, WorkflowStep{
           Name:            "review",
           CommandFilePath: resolveCommandPath(cfg.Review.Command, configDir),
           Retries:         reviewRetries,
           IsReview:        true,
           Order:           len(cfg.Steps),
       })

       return &Workflow{
           Steps:            steps,
           MaxIterations:    maxIterations,
           CurrentIteration: 0,
       }, nil
   }

   func resolveCommandPath(path string, configDir string) string {
       if filepath.IsAbs(path) {
           return path
       }
       return filepath.Clean(filepath.Join(configDir, path))
   }
   ```

---

### Phase 5: Generic Step Executor (TDD: Green - Part 4)

**Goal**: Implement DRY step execution with retry logic

**Steps**:

1. **Create step executor**: `internal/workflow/step_executor.go`

   ```go
   package workflow

   import (
       "log"
       "math"

       "github.com/fluxid/internal/storage"
       "github.com/fluxid/internal/types"
   )

   const (
       statusPass = "PASS"
       statusFail = "FAIL"
   )

   // ExecuteStepWithRetry executes a workflow step with retry logic
   func ExecuteStepWithRetry(cfg types.Config, step types.WorkflowStep, iteration int, logger *WorkflowLogger) error {
       maxRetries := step.Retries
       if maxRetries == 0 {
           maxRetries = math.MaxInt // Infinite retries
       }

       for retry := 1; retry <= maxRetries; retry++ {
           logger.LogStepStart(step.Name, iteration, retry)

           // Execute step (agent invocation)
           exitCode, execErr := runPhase(cfg, step.Name, step.CommandFilePath)
           if execErr != nil {
               log.Printf("Agent invocation failed for step %s (retry %d/%d): %v",
                   step.Name, retry, maxRetries, execErr)
               // Treat as FAIL, continue to retry logic
               logger.LogStepComplete(step.Name, statusFail, retry, maxRetries)
               continue
           }

           // Check report status
           status, err := waitForValidReport(cfg.SessionID, cfg.SessionRoot, step.Name)
           if err != nil {
               log.Printf("Failed to read %s report (treating as FAIL): %v", step.Name, err)
               status = statusFail
           }

           logger.LogStepComplete(step.Name, status, retry, maxRetries)

           if status == statusPass {
               return nil // Success
           }

           // Status == FAIL: continue to next retry
       }

       // All retries exhausted with FAIL status
       log.Printf("Step %s exhausted all %d retries (continuing to next step)", step.Name, maxRetries)
       return nil // Continue to next step (FR-006)
   }
   ```

2. **Test step executor**:
   ```bash
   go test -v internal/workflow/step_executor_test.go
   ```

---

### Phase 6: Workflow Logger (TDD: Green - Part 5)

**Goal**: Implement structured logging with JSON/text output

**Steps**:

1. **Create logger**: `internal/workflow/workflow_logger.go`

   ```go
   package workflow

   import (
       "encoding/json"
       "log"
   )

   type WorkflowLogger struct {
       OutputFormat string // "text" or "json"
   }

   func (l *WorkflowLogger) LogStepStart(stepName string, iteration int, retry int) {
       if l.OutputFormat == "json" {
           event := map[string]interface{}{
               "event":     "step_start",
               "step":      stepName,
               "iteration": iteration,
               "retry":     retry,
           }
           jsonBytes, _ := json.Marshal(event)
           log.Println(string(jsonBytes))
       } else {
           log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
           log.Printf("▶ Starting step: %s (iteration %d, retry %d)", stepName, iteration, retry)
       }
   }

   func (l *WorkflowLogger) LogStepComplete(stepName string, status string, retry int, maxRetries int) {
       if l.OutputFormat == "json" {
           event := map[string]interface{}{
               "event":       "step_complete",
               "step":        stepName,
               "status":      status,
               "retry":       retry,
               "max_retries": maxRetries,
           }
           jsonBytes, _ := json.Marshal(event)
           log.Println(string(jsonBytes))
       } else {
           log.Printf("✓ Step complete: %s | Status: %s | Retry: %d/%d",
               stepName, status, retry, maxRetries)
       }
   }

   func (l *WorkflowLogger) LogIterationStart(iteration int, maxIterations int) {
       if l.OutputFormat == "json" {
           event := map[string]interface{}{
               "event":          "iteration_start",
               "iteration":      iteration,
               "max_iterations": maxIterations,
           }
           jsonBytes, _ := json.Marshal(event)
           log.Println(string(jsonBytes))
       } else {
           log.Printf("\n")
           log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
           log.Printf("🔄 DEVELOPMENT ITERATION %d/%d", iteration, maxIterations)
           log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
       }
   }
   ```

---

### Phase 7: Main Workflow Loop Refactor (TDD: Green - Part 6)

**Goal**: Replace hardcoded 3-step loop with config-driven loop

**Steps**:

1. **Modify `Run()` function** in `internal/workflow/workflow.go`:

   ```go
   func Run(cfg types.Config) (int, error) {
       logger := &WorkflowLogger{OutputFormat: cfg.OutputFormat}

       // Outer loop: Development iterations
       maxIter := cfg.Workflow.MaxIterations
       if maxIter == 0 {
           maxIter = math.MaxInt // Infinite iterations
       }

       for iteration := 1; iteration <= maxIter; iteration++ {
           cfg.Workflow.CurrentIteration = iteration
           logger.LogIterationStart(iteration, cfg.Workflow.MaxIterations)

           // Inner loop: Execute all workflow steps (custom + review)
           for _, step := range cfg.Workflow.Steps {
               if err := ExecuteStepWithRetry(cfg, step, iteration, logger); err != nil {
                   return 1, err
               }

               // If this is the review step, check status for exit gate
               if step.IsReview {
                   report, err := storage.ReadReport(cfg.SessionID, cfg.SessionRoot)
                   if err != nil {
                       log.Printf("Failed to read review report: %v", err)
                       continue // Treat as FAIL, continue to next iteration
                   }

                   if report.Status == statusPass {
                       logger.LogWorkflowComplete(true, iteration)
                       return 0, nil // Workflow succeeded
                   }
                   // FAIL: continue to next iteration
               }
           }

           // Check iteration exhaustion
           if cfg.Workflow.MaxIterations > 0 && iteration >= cfg.Workflow.MaxIterations {
               logger.LogWorkflowComplete(false, iteration)
               return 0, nil // Iterations exhausted
           }
       }

       return 0, nil
   }
   ```

2. **Delete old phase functions**:
   - Remove `runImplementPhase()`
   - Remove `runCommitPhaseWithRetry()`
   - Remove phase-specific retry logic

3. **Run all E2E tests**:
   ```bash
   cd e2e-tests
   go test -v ./tests/
   ```

   **Expected**: All workflow E2E tests now PASS

---

### Phase 8: Update Existing E2E Tests (TDD: Refactor)

**Goal**: Update or replace existing hardcoded tests with config-driven versions

**Steps**:

1. **Review existing tests**:
   - `m01-e01-user-runs-workflow-to-completion_test.go`
   - `m01-e02-user-configures-loop-counts-and-runs-workflow_test.go`
   - `m07-e01-implement-retries-exhausted-continues-through-commit_test.go`
   - `m08-e01-workflow-completes-with-all-fail-reports_test.go`

2. **Replace hardcoded assumptions**:
   - Update tests to use fixture configs with workflow section
   - Update assertion patterns (generic step names vs. hardcoded "implement", "commit", "review")
   - Verify test intent still valid (e.g., retry exhaustion behavior)

3. **Run updated tests**:
   ```bash
   go test -v e2e-tests/tests/m01-*.go
   go test -v e2e-tests/tests/m07-*.go
   go test -v e2e-tests/tests/m08-*.go
   ```

   **Expected**: All tests PASS with config-driven workflow

---

### Phase 9: Agent Context Update

**Goal**: Update agent-specific context files with new workflow information

**Steps**:

1. **Run agent context update script**:
   ```bash
   .specify/scripts/bash/update-agent-context.sh claude
   ```

2. **Verify context file updated**:
   - Check `.specify/agents/claude-context.md` (or similar)
   - Ensure workflow configuration information added

---

### Phase 10: Re-evaluate Constitution Check

**Goal**: Verify all constitution principles still satisfied after implementation

**Steps**:

1. **Review constitution compliance**:
   - TDD followed? (tests written first)
   - E2E coverage complete?
   - Sequential execution maintained?
   - Code quality gates passed?
   - Pure Go implementation?
   - Explicit interfaces used?
   - Fail-fast behavior working?
   - CLI-first design maintained?

2. **Update plan.md** with post-implementation constitution check results

---

## Testing Strategy

### Unit Tests

**Location**: `internal/config/`, `internal/workflow/`, `internal/types/`

**Coverage**:
- YAML parsing (WorkflowConfig unmarshaling)
- Validation logic (all 10 validation rules)
- Workflow initialization (BuildWorkflow)
- Step execution retry logic
- Logger output formats (text/JSON)

**Run**:
```bash
go test -v ./internal/...
```

### E2E Tests

**Location**: `e2e-tests/tests/`

**Coverage**:
- Config-driven workflow execution (minimal, standard, extended)
- Step retry exhaustion behavior
- Startup validation failures
- Iteration loop behavior
- Review exit gate behavior

**Run**:
```bash
cd e2e-tests
go test -v ./tests/
```

### Integration Tests

**Location**: `e2e-tests/tests/`

**Coverage**:
- Full workflow with mock agent invocations
- Report read/write behavior
- Session management integration
- Config precedence (project > home > defaults)

**Run**:
```bash
go test -v -tags=integration ./e2e-tests/tests/
```

---

## Troubleshooting

### Common Issues

**Issue**: Tests fail with "workflow section is required"
**Solution**: Ensure test fixture configs include workflow section

**Issue**: Command file not found errors
**Solution**: Verify command file paths are relative to config.yaml directory

**Issue**: Review step not executing last
**Solution**: Check BuildWorkflow appends review step with IsReview=true

**Issue**: Infinite retry loop not working (retries=0)
**Solution**: Verify ExecuteStepWithRetry treats retries=0 as math.MaxInt

**Issue**: JSON output not working
**Solution**: Check OutputFormat field set correctly and logger uses it

---

## Definition of Done

Implementation is complete when:

- ✅ All E2E tests pass (new + updated existing tests)
- ✅ All unit tests pass
- ✅ Code coverage >= 90% (existing project threshold)
- ✅ Pre-commit hooks pass (formatting, linting, security)
- ✅ Constitution check passes (all principles satisfied)
- ✅ Agent context updated
- ✅ Documentation generated (this quickstart + contracts)

---

## Next Steps After Implementation

1. Run full test suite: `go test -v ./...`
2. Run E2E tests: `cd e2e-tests && go test -v ./tests/`
3. Verify pre-commit hooks: `git commit` (should run hooks)
4. Create pull request with implementation
5. Request code review (verify constitution compliance)

---

## Resources

- **Specification**: `spec.md`
- **Research**: `research.md`
- **Data Model**: `data-model.md`
- **Contracts**: `contracts/`
- **Constitution**: `.specify/memory/constitution.md`
- **Existing Implementation**: `internal/workflow/workflow.go` (to be refactored)

---

## Support

For questions or issues during implementation:
1. Review spec.md for requirements clarifications
2. Check research.md for technology decision rationale
3. Consult contracts/ for type signatures and examples
4. Refer to constitution for governance and principles
