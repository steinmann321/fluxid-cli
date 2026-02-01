# Implementation Plan: Config-Driven Workflow System

**Branch**: `002-config-driven-workflow` | **Date**: 2026-01-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-config-driven-workflow/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

**BREAKING CHANGE**: Replace the hardcoded 3-step workflow (implement, commit, review) with a config-driven workflow system that allows users to define 1..N custom workflow steps in config.yaml. No backward compatibility - users must explicitly configure workflow steps in config.yaml. Each step has a name, command file path, and retry limit. The workflow always ends with a mandatory review step that acts as the only exit gate (PASS exits successfully, FAIL triggers the next development iteration until max iterations exhausted). This enables workflow customization without code changes while maintaining report status checking and development iteration loop behavior.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: gopkg.in/yaml.v3 (YAML parsing), github.com/google/uuid (Session ID generation)
**Storage**: File-based YAML storage (report.yaml, history.yaml) in session-specific directories
**Testing**: Go testing framework (`go test`), E2E tests in e2e-tests/ directory with mock agent invocations
**Target Platform**: macOS, Linux, Windows (cross-platform CLI)
**Project Type**: Single CLI application (pure Go implementation)
**Performance Goals**: Immediate workflow validation (<100ms startup), no timeout enforcement on agent execution
**Constraints**: Sequential execution only (no parallelism), single report.yaml per session (overwritten by each step), strict fail-fast validation at startup
**Scale/Scope**: 1..10 workflow steps per configuration, 20+ default development iterations, 100+ default commit retries

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-Driven Development (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Implementation will follow TDD: Write failing E2E tests for config-driven workflow BEFORE refactoring existing workflow.go
- Tests will verify workflow step execution order, retry behavior, review gate enforcement, and startup validation

### II. Full E2E Test Coverage (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Existing E2E tests in e2e-tests/ directory will be updated/replaced to reflect config-driven workflow
- New E2E tests will cover: workflow.steps YAML parsing, custom step execution, review step enforcement, startup validation failures
- User instruction: "ensure exiting e2e tests are droped/recreated as needed becuase they show a specific config"

### III. Strictly Sequential Workflow (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Config-driven workflow maintains sequential execution: step[0] → step[1] → ... → step[N] → review
- No parallel execution introduced
- Single report.yaml file overwritten by each step (maintains current behavior per FR-002 clarification)

### IV. Strict Code Quality Enforcement (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Pre-commit hooks will enforce code quality gates (formatting, linting, security, coverage)
- No bypass or weakening of existing hooks
- Implementation will follow SOLID, DRY (generic step execution function replaces duplicate retry logic), KISS, YAGNI principles

### V. Pure Go Implementation (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- All workflow logic implemented in Go (internal/workflow/ directory)
- YAML parsing using gopkg.in/yaml.v3 (existing dependency)
- No shell script dependencies at runtime

### VI. Explicit Interfaces Over Implicit Behavior (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Explicit YAML schema for workflow configuration (exact field names documented in FR-001, FR-002, FR-017)
- Clear validation rules at startup (FR-011, FR-015, FR-018)
- No "magic" conventions - all step properties defined in config.yaml

### VII. Fail Fast with Clear Diagnostics (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Startup validation failures prevent workflow execution (FR-010, FR-011, FR-015, FR-018)
- Clear error messages for: missing workflow section, duplicate step names, invalid command paths, empty step names
- Detailed logging during execution (FR-020, FR-021)

### VIII. Command-Line First, Scriptable Always (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Existing CLI interface maintained (FR-012: `fluxid run` unchanged)
- JSON output format support via --out=json flag (FR-021)
- Exit codes indicate success/failure
- No interactive prompts required

**GATE RESULT**: ✅ **ALL CHECKS PASS** - Proceed to Phase 0

---

## Constitution Check (Post-Design)

*Re-evaluation after Phase 1 design completion*

### I. Test-Driven Development (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- E2E test fixtures created (workflow_minimal.yaml, workflow_standard.yaml, workflow_extended.yaml, invalid configs)
- Test structure defined in quickstart.md Phase 1 (write failing tests first)
- TDD workflow maintained: Red (failing tests) → Green (implementation) → Refactor

### II. Full E2E Test Coverage (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- New tests: workflow_config_driven_test.go, workflow_startup_validation_test.go
- Existing tests to be updated: m01-e01, m01-e02, m07-e01, m08-e01
- Coverage: config parsing, step execution, retry logic, validation failures, review exit gate

### III. Strictly Sequential Workflow (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Design maintains sequential execution: outer iteration loop → inner step loop
- No parallelism in workflow orchestration
- Single report.yaml overwritten by each step (data-model.md confirmed)

### IV. Strict Code Quality Enforcement (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- DRY: Generic ExecuteStepWithRetry() replaces duplicate phase-specific retry logic
- SOLID: Single responsibility (validation, execution, logging in separate modules)
- KISS: Simple YAML structure, straightforward validation rules
- YAGNI: Only implements required features (no speculative abstractions)

### V. Pure Go Implementation (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- All implementation in Go (no shell script dependencies)
- Uses existing dependencies: gopkg.in/yaml.v3, github.com/google/uuid
- No new external dependencies required (research.md confirmed)

### VI. Explicit Interfaces Over Implicit Behavior (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Explicit YAML schema (contracts/workflow-config-schema.yaml)
- Explicit Go type signatures (contracts/go-types.go)
- Clear validation rules (V001-V010 documented)
- No "magic" conventions or implicit behavior

### VII. Fail Fast with Clear Diagnostics (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- Startup validation prevents workflow execution (validate_workflow.go)
- Clear error messages for each validation rule (schema documented)
- Validation errors include context (step names, file paths, config values)

### VIII. Command-Line First, Scriptable Always (NON-NEGOTIABLE)
**Status**: ✅ **PASS**
- JSON output support (--out=json flag, WorkflowLogger implementation)
- Existing CLI interface maintained (fluxid run unchanged)
- Exit codes preserved (0 for success/completion)
- No interactive prompts required

**GATE RESULT**: ✅ **ALL CHECKS PASS POST-DESIGN** - Proceed to Implementation

---

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── config/
│   ├── config.go                    # MODIFIED: Add workflow config structs
│   ├── resolve_commands.go          # MODIFIED: Adapt for dynamic workflow steps
│   └── validate_workflow.go         # NEW: Startup validation for workflow config
├── workflow/
│   ├── workflow.go                  # MODIFIED: Replace hardcoded 3-step loop
│   ├── workflow_helpers.go          # MODIFIED: Generic step execution function
│   └── step_executor.go             # NEW: DRY retry logic for any step
├── command/
│   └── root_config_loader.go        # MODIFIED: Load workflow config
├── storage/
│   └── report.go                    # UNCHANGED: Existing report read/write
└── types/
    └── types.go                     # MODIFIED: Add workflow types (WorkflowStep, etc.)

e2e-tests/
├── tests/
│   ├── workflow_config_driven_test.go           # NEW: Config-driven workflow E2E
│   ├── workflow_startup_validation_test.go      # NEW: Validation failure scenarios
│   ├── m01-e01-*_test.go                        # MODIFIED/REPLACED: Update for new config
│   ├── m01-e02-*_test.go                        # MODIFIED/REPLACED: Update for new config
│   ├── m07-e01-*_test.go                        # MODIFIED/REPLACED: Update for new config
│   └── m08-e01-*_test.go                        # MODIFIED/REPLACED: Update for new config
└── fixtures/
    └── configs/                                  # NEW: Test config.yaml fixtures
```

**Structure Decision**: Single CLI project structure (Option 1). All workflow logic lives in `internal/workflow/` directory. Configuration handling in `internal/config/`. E2E tests in dedicated `e2e-tests/` directory with fixture configs for testing various workflow scenarios. This matches the existing codebase structure and maintains the pure Go implementation principle.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitution violations detected. All gates pass.
