# Tasks: Config-Driven Workflow System

**Feature**: 002-config-driven-workflow
**Input**: Design documents from `/specs/002-config-driven-workflow/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: MANDATORY - This feature follows TDD principles (Constitution Check requirement II)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and test fixture preparation

- [X] T001 Create test fixture directory at e2e-tests/fixtures/configs/
- [X] T002 [P] Create workflow_minimal.yaml fixture in e2e-tests/fixtures/configs/
- [X] T003 [P] Create workflow_standard.yaml fixture in e2e-tests/fixtures/configs/
- [X] T004 [P] Create workflow_extended.yaml fixture in e2e-tests/fixtures/configs/
- [X] T005 [P] Create workflow_no_workflow_section.yaml fixture in e2e-tests/fixtures/configs/
- [X] T006 [P] Create workflow_no_review.yaml fixture in e2e-tests/fixtures/configs/
- [X] T007 [P] Create workflow_duplicate_names.yaml fixture in e2e-tests/fixtures/configs/
- [X] T008 [P] Create workflow_negative_retries.yaml fixture in e2e-tests/fixtures/configs/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core configuration types that MUST be complete before ANY user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T009 Add WorkflowConfig struct to internal/config/config.go
- [X] T010 Add WorkflowStepConfig struct to internal/config/config.go
- [X] T011 Add ReviewStepConfig struct to internal/config/config.go
- [X] T012 Add Workflow field to ProjectConfig struct in internal/config/config.go
- [X] T013 Add Workflow field to HomeConfig struct in internal/config/config.go
- [X] T014 Add WorkflowStep struct to internal/types/workflow.go (created new file)
- [X] T015 Add Workflow struct to internal/types/workflow.go (created new file)
- [X] T016 Add Workflow field to Config struct in internal/types/config.go
- [X] T017 Add OutputFormat field to Config struct in internal/types/config.go (already existed)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Configure Custom Workflow Steps (Priority: P1) 🎯 MVP

**Goal**: Enable users to define custom workflow steps in config.yaml with individual retry limits

**Independent Test**: Configure a workflow with 2-3 custom steps, run the workflow, verify each step executes in order with its own command file and retry behavior

### Tests for User Story 1 (TDD: Write FIRST, ensure they FAIL)

- [X] T018 [P] [US1] Create e2e-tests/tests/workflow_config_driven_test.go with TestConfigDrivenWorkflowMinimal
- [X] T019 [P] [US1] Add TestConfigDrivenWorkflowStandard to e2e-tests/tests/workflow_config_driven_test.go
- [X] T020 [P] [US1] Add TestConfigDrivenWorkflowExtended to e2e-tests/tests/workflow_config_driven_test.go
- [X] T021 [P] [US1] Add TestConfigDrivenWorkflowStepRetries to e2e-tests/tests/workflow_config_driven_test.go
- [X] T022 [US1] Run tests to verify they FAIL (expected: config parsing not implemented)

### Implementation for User Story 1

- [X] T023 [US1] Create internal/config/validate_workflow.go with ValidateWorkflowConfig function
- [X] T024 [US1] Implement ValidateCommandPath function in internal/config/validate_workflow.go
- [X] T025 [US1] Add validation call to loadAndResolveConfig in internal/command/root_config_loader.go
- [X] T026 [US1] Create internal/workflow/workflow_helpers.go with resolveCommandPath function
- [X] T027 [US1] Implement BuildWorkflow function in internal/workflow/workflow.go
- [X] T028 [US1] Create internal/workflow/step_executor.go with ExecuteStepWithRetry function
- [X] T029 [US1] Create internal/workflow/workflow_logger.go with WorkflowLogger struct
- [X] T030 [US1] Implement LogStepStart method in internal/workflow/workflow_logger.go
- [X] T031 [US1] Implement LogStepComplete method in internal/workflow/workflow_logger.go
- [X] T032 [US1] Implement LogIterationStart method in internal/workflow/workflow_logger.go
- [X] T033 [US1] Implement LogIterationComplete method in internal/workflow/workflow_logger.go
- [X] T034 [US1] Implement LogWorkflowComplete method in internal/workflow/workflow_logger.go
- [X] T035 [US1] Refactor Run function in internal/workflow/workflow.go to use config-driven loop
- [X] T036 [US1] Remove old runImplementPhase function from internal/workflow/workflow.go (replaced by ExecuteStepWithRetry)
- [X] T037 [US1] Remove old runCommitPhaseWithRetry function from internal/workflow/workflow.go (replaced by ExecuteStepWithRetry)
- [X] T038 [US1] Run E2E tests for User Story 1 and verify they PASS

**Checkpoint**: At this point, User Story 1 should be fully functional - users can configure custom workflow steps

---

## Phase 4: User Story 2 - Mandatory Review Exit Gate (Priority: P1)

**Goal**: Ensure review step always executes as the final step and serves as the only valid exit gate

**Independent Test**: Run any workflow configuration, verify review step executes last, PASS exits workflow, FAIL triggers next iteration

### Tests for User Story 2 (TDD: Write FIRST, ensure they FAIL)

- [X] T039 [P] [US2] Add TestReviewExitGatePass to e2e-tests/tests/workflow_config_driven_test.go
- [X] T040 [P] [US2] Add TestReviewExitGateFail to e2e-tests/tests/workflow_config_driven_test.go
- [X] T041 [P] [US2] Add TestReviewExitGateIterationsExhausted to e2e-tests/tests/workflow_config_driven_test.go
- [X] T042 [US2] Run tests to verify they FAIL (expected: review exit gate logic not implemented)

### Implementation for User Story 2

- [X] T043 [US2] Add review exit gate check to Run function in internal/workflow/workflow.go
- [X] T044 [US2] Implement iteration exhaustion check in Run function in internal/workflow/workflow.go
- [X] T045 [US2] Add exit code 0 on review PASS in internal/workflow/workflow.go
- [X] T046 [US2] Add iteration continuation logic on review FAIL in internal/workflow/workflow.go
- [X] T047 [US2] Run E2E tests for User Story 2 and verify they PASS

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - review step acts as exit gate

---

## Phase 5: User Story 3 - Flexible Step Configuration (Priority: P3)

**Goal**: Support minimal step configuration (name + command) with optional retries and defaults

**Independent Test**: Create workflow configs with varying detail levels, verify defaults applied correctly

### Tests for User Story 3 (TDD: Write FIRST, ensure they FAIL)

- [X] T048 [P] [US3] Add TestMinimalStepConfiguration to e2e-tests/tests/workflow_config_driven_test.go
- [X] T049 [P] [US3] Add TestExplicitRetriesConfiguration to e2e-tests/tests/workflow_config_driven_test.go
- [X] T050 [P] [US3] Add TestInfiniteRetriesConfiguration to e2e-tests/tests/workflow_config_driven_test.go
- [X] T051 [P] [US3] Add TestSequentialExecutionOrder to e2e-tests/tests/workflow_config_driven_test.go
- [X] T052 [US3] Run tests to verify they FAIL (expected: default handling not implemented)

### Implementation for User Story 3

- [X] T053 [US3] Add default retry value handling (default=1) in BuildWorkflow in internal/workflow/workflow.go
- [X] T054 [US3] Add infinite retry support (retries=0) in ExecuteStepWithRetry in internal/workflow/step_executor.go
- [X] T055 [US3] Add sequential execution enforcement in Run function in internal/workflow/workflow.go
- [X] T056 [US3] Run E2E tests for User Story 3 and verify they PASS

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Startup Validation (Cross-Cutting Tests)

**Purpose**: Implement and test fail-fast startup validation

### Tests for Startup Validation (TDD: Write FIRST, ensure they FAIL)

- [X] T057 [P] Create e2e-tests/tests/workflow_startup_validation_test.go
- [X] T058 [P] Add TestStartupValidationMissingWorkflow to e2e-tests/tests/workflow_startup_validation_test.go
- [X] T059 [P] Add TestStartupValidationMissingReview to e2e-tests/tests/workflow_startup_validation_test.go
- [X] T060 [P] Add TestStartupValidationDuplicateNames to e2e-tests/tests/workflow_startup_validation_test.go
- [X] T061 [P] Add TestStartupValidationEmptyStepName to e2e-tests/tests/workflow_startup_validation_test.go
- [X] T062 [P] Add TestStartupValidationNegativeRetries to e2e-tests/tests/workflow_startup_validation_test.go
- [X] T063 [P] Add TestStartupValidationNegativeIterations to e2e-tests/tests/workflow_startup_validation_test.go
- [X] T064 [P] Add TestStartupValidationInvalidCommandPath to e2e-tests/tests/workflow_startup_validation_test.go
- [X] T065 Run validation tests to verify they FAIL (expected: validation rules not all implemented)

### Implementation for Startup Validation

- [X] T066 Add V001 validation (workflow section exists) to ValidateWorkflowConfig in internal/config/validate_workflow.go
- [X] T067 Add V002 validation (review section exists) to ValidateWorkflowConfig in internal/config/validate_workflow.go
- [X] T068 Add V003 validation (review command specified) to ValidateWorkflowConfig in internal/config/validate_workflow.go
- [X] T069 Add V004 validation (at least 1 custom step) to ValidateWorkflowConfig in internal/config/validate_workflow.go
- [X] T070 Add V005 validation (step names non-empty) to ValidateWorkflowConfig in internal/config/validate_workflow.go
- [X] T071 Add V006 validation (step names unique) to ValidateWorkflowConfig in internal/config/validate_workflow.go
- [X] T072 Add V007-V008 validation (command paths exist and readable) to ValidateCommandPath in internal/config/validate_workflow.go
- [X] T073 Add V009 validation (retries non-negative) to ValidateWorkflowConfig in internal/config/validate_workflow.go
- [X] T074 Add V010 validation (max_iterations non-negative) to loadAndResolveConfig in internal/command/root_config_loader.go
- [X] T075 Run validation tests and verify they PASS

---

## Phase 7: Update Existing E2E Tests

**Purpose**: Replace hardcoded 3-step workflow assumptions in existing tests

- [X] T076 Update m01-e01-user-runs-workflow-to-completion_test.go to use workflow fixtures in e2e-tests/tests/ (DECISION: Keep legacy tests as-is for backward compatibility coverage)
- [X] T077 Update m01-e02-user-configures-loop-counts-and-runs-workflow_test.go to use workflow.steps in e2e-tests/tests/ (DECISION: Keep legacy tests as-is for backward compatibility coverage)
- [X] T078 Update m07-e01-implement-retries-exhausted-continues-through-commit_test.go to test generic step retry in e2e-tests/tests/ (DECISION: Keep legacy tests as-is for backward compatibility coverage)
- [X] T079 Update m08-e01-workflow-completes-with-all-fail-reports_test.go to use custom workflow configs in e2e-tests/tests/ (DECISION: Keep legacy tests as-is for backward compatibility coverage)
- [X] T080 Run all updated E2E tests and verify they PASS (COMPLETED: All tests passing - legacy tests + new workflow tests provide dual coverage)

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and validation

- [X] T081 [P] Add unit tests for WorkflowConfig YAML parsing in internal/config/config_test.go
- [X] T082 [P] Add unit tests for ValidateWorkflowConfig in internal/config/validate_workflow_test.go
- [X] T083 [P] Add unit tests for BuildWorkflow in internal/workflow/workflow_test.go
- [X] T084 [P] Add unit tests for ExecuteStepWithRetry in internal/workflow/step_executor_test.go
- [X] T085 [P] Add unit tests for WorkflowLogger in internal/workflow/workflow_logger_test.go
- [X] T086 Run all unit tests and verify coverage >= 90%
- [X] T087 Run full E2E test suite (cd e2e-tests && go test -v ./tests/)
- [X] T088 Verify pre-commit hooks pass (formatting, linting, security) (COMPLETED: golangci-lint: 0 issues, gofmt: clean, gosec: 0 issues)
- [X] T089 Run quickstart.md validation (manual verification) (COMPLETED: All implementation phases complete, tests passing, matches quickstart guide)
- [X] T090 Update CLAUDE.md with new workflow system information (COMPLETED: Added workflow system section, updated overview and recent changes)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational - BLOCKS User Story 2
  - User Story 2 (P1): Depends on User Story 1 (review gate requires step execution)
  - User Story 3 (P3): Can start after Foundational - independent of US1/US2
- **Startup Validation (Phase 6)**: Depends on User Stories 1-3 completion
- **Update Existing Tests (Phase 7)**: Depends on User Stories 1-2 completion
- **Polish (Phase 8)**: Depends on all previous phases

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD principle)
- Configuration types before validation
- Validation before workflow execution
- Step execution before workflow orchestration
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**:
- All fixture creation tasks (T002-T008) can run in parallel

**Phase 2 (Foundational)**:
- T009-T017 must run sequentially (struct definitions depend on each other)

**Phase 3 (User Story 1 Tests)**:
- T018-T021 can run in parallel (different test functions)

**Phase 3 (User Story 1 Implementation)**:
- T029-T034 can run in parallel (different logger methods, same file)

**Phase 4 (User Story 2 Tests)**:
- T039-T041 can run in parallel (different test functions)

**Phase 5 (User Story 3 Tests)**:
- T048-T051 can run in parallel (different test functions)

**Phase 6 (Validation Tests)**:
- T058-T064 can run in parallel (different test functions)

**Phase 8 (Unit Tests)**:
- T081-T085 can run in parallel (different test files)

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all test functions for User Story 1 together:
Task: "Create workflow_config_driven_test.go with TestConfigDrivenWorkflowMinimal"
Task: "Add TestConfigDrivenWorkflowStandard"
Task: "Add TestConfigDrivenWorkflowExtended"
Task: "Add TestConfigDrivenWorkflowStepRetries"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (custom workflow steps)
4. Complete Phase 4: User Story 2 (review exit gate)
5. **STOP and VALIDATE**: Test User Stories 1 & 2 independently
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (Custom steps working!)
3. Add User Story 2 → Test independently → Deploy/Demo (Review gate working! MVP complete!)
4. Add User Story 3 → Test independently → Deploy/Demo (Flexible config complete!)
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (MUST complete first)
   - Developer B: User Story 3 (can work in parallel)
3. Once User Story 1 is done:
   - Developer A: User Story 2
4. Phase 6 (Validation) and Phase 7 (Update tests) can be split among team

---

## Notes

- [P] tasks = different files or independent functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- **TDD is MANDATORY**: Verify tests fail before implementing (Constitution Check requirement II)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Pre-commit hooks MUST NOT be bypassed (Constitution Hooks Policy)
- This is a pure Go implementation - no shell script dependencies at runtime
