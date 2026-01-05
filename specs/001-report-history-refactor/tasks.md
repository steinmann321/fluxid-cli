---
description: "Tasks for Report & History File-Based Interface Refactor"
---

# Tasks: Report & History File-Based Interface

**Input**: Design documents from `/specs/001-report-history-refactor/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md (schemas at internal/assets/templates/)

**Tests**: Test-Driven Development enforced per Constitution I. Each implementation task has corresponding test task to be written first (Red-Green-Refactor cycle).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `- [ ] [ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Go CLI project structure:
- `internal/` for internal packages
- `cmd/fluxid/` for CLI entry point
- `e2e-tests/tests/` for E2E tests
- `internal/assets/schemas/` for embedded JSON schemas

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create foundational storage package and schema infrastructure

- [ ] T001 Create directory structure `internal/storage/` and `internal/assets/templates/`
- [ ] T002 [P] Verify existing YAML schema file for report in `internal/assets/templates/report-schema.yaml` includes: all required fields (command, artifact, timestamp, status, issues), all optional fields (next_steps, summary), type constraints, enum values, and strict validation rules per FR-025, FR-026, FR-027. File already exists - validate completeness only.
- [ ] T003 [P] Verify existing YAML schema file for history in `internal/assets/templates/history-schema.yaml` includes: array type, required fields (timestamp, step, status, summary), optional field (details), enum values, and strict validation rules per FR-028, FR-029, FR-030, FR-031. File already exists - validate completeness only.

---

## Phase 2: Foundational (Blocking Prerequisites) - TDD Red Phase

**Purpose**: Core storage infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete. TDD enforced: write failing tests FIRST.

**Step 1: Write Failing Tests (RED)**

- [ ] T005 [P] Add E2E test in `e2e-tests/tests/storage_report_test.go` for storage.ReadReport() (valid report file, malformed YAML, missing file scenarios)
- [ ] T006 [P] Add E2E test in `e2e-tests/tests/storage_history_test.go` for storage.ReadHistory() with FIFO eviction (valid history, >10MB truncation by removing oldest 30% of complete entries, empty file scenarios)
- [ ] T007 [P] Add E2E test in `e2e-tests/tests/security_yaml_test.go` for YAML security validator (reject anchors, aliases, merge keys with clear errors)
- [ ] T008 [P] Add E2E test in `e2e-tests/tests/path_validation_test.go` for session path validator (valid UUID paths, path traversal rejection, invalid session IDs)
- [ ] T008a [P] Add E2E test in `e2e-tests/tests/observability_test.go` for observability requirements (silent success produces no stderr output with exit 0, error messages include sufficient context with file paths, field paths, constraint details, and session IDs in stderr per FR-040, FR-041, FR-042)

**Step 2: Implement to Pass Tests (GREEN)**

- [ ] T009 Implement schema embedding and retrieval in `internal/storage/schema.go` using Go embed directive (//go:embed) to package report-schema.yaml and history-schema.yaml from internal/assets/templates/ into binary. Provide GetReportSchema() and GetHistorySchema() functions to retrieve embedded schemas.
- [ ] T010 [P] Implement YAML security validator in `internal/storage/security.go` (reject anchors &, aliases *, merge keys <<)
- [ ] T011 [P] Implement session path validator in `internal/storage/paths.go` (validate UUID format, prevent path traversal, ensure within session root, use os.TempDir() for cross-platform temp directory)
- [ ] T012 Implement report file reader in `internal/storage/report.go` with ReadReport(sessionID string) function that reads and parses existing report file at session-specific path
- [ ] T013 Implement history file reader in `internal/storage/history.go` with ReadHistory(sessionID string) function including FIFO eviction logic (removes oldest 30% of complete entries by count using ceiling function: ceiling(entry_count * 0.30) ensures minimum 1 entry removed when entry_count >= 4, removing only whole event objects when file >10MB, per FR-022)
- [ ] T014 Implement custom YAML validator in `internal/storage/validate.go` with ValidateReport(filePath string) and ValidateHistory(filePath string) functions that validate existing YAML files against embedded YAML schema documents (strict field, type, and constraint checking). Implement error formatting per FR-004 with format "[field_path]: [violation] (expected: [constraint], got: [value])" by default, or JSON array with {field, violation, constraint, value} objects when --format=json flag provided. Use exit codes: 0=success, 1=validation failure, 2=operational failure including schema load failure per FR-024
- [ ] T014a [P] Implement observability in all command handlers per FR-040, FR-041, FR-042: errors to stderr with sufficient context (file paths, field paths, constraint details, session IDs), silent success (no stderr output on success), stdout reserved for data output only

**Checkpoint**: Foundation ready with full E2E test coverage - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - External Agent Writes Report (Priority: P1) 🎯 MVP

**Goal**: Enable external agents to query report file path, write valid YAML reports, and have fluxid read them correctly

**Independent Test**: Agent runs `fluxid report --get-file`, writes valid YAML report to returned path, fluxid workflow reads and processes status correctly

**Step 1: Write Failing E2E Test (RED)**

- [ ] T015 [US1] Add E2E test in `e2e-tests/tests/report_write_test.go` for complete agent workflow (get-file → write report → fluxid reads → PASS/FAIL behavior)

**Step 2: Implement to Pass Test (GREEN)**

- [ ] T016 [P] [US1] Create report command handler in `internal/command/report.go` with cobra command structure
- [ ] T017 [P] [US1] Implement `--get-file` flag handler in `internal/command/report.go` (resolve session ID from existing session management, construct path using os.TempDir(), ensure file/dir exist, output absolute path)
- [ ] T018 [US1] Update root command in `internal/command/root.go` to register report command
- [ ] T019 [US1] Update workflow controller in `internal/workflow/workflow.go` to use storage.ReadReport() instead of IPC-based report reading
- [ ] T020 [US1] Add error handling in `internal/command/report.go` for missing FLUXID_SESSION_ID (exit code 3, reuses existing session validation)

**Checkpoint**: At this point, User Story 1 should be fully functional - agents can write reports and fluxid processes them

---

## Phase 4: User Story 2 - External Agent Validates Report (Priority: P1)

**Goal**: Enable agents to validate report file structure before fluxid reads it, preventing workflow failures

**Independent Test**: Agent writes valid/invalid report files and runs `fluxid report --validate`, receiving appropriate exit codes and error messages

**Step 1: Write Failing E2E Tests (RED)**

- [ ] T021 [US2] Add E2E test in `e2e-tests/tests/report_validate_test.go` covering valid report file (exit 0), missing required field (exit 1 with field name), invalid status value (exit 1 with enum values), file not found (exit 2)
- [ ] T022 [US2] Add E2E test in `e2e-tests/tests/report_validate_test.go` for schema mismatch scenarios (valid YAML but wrong structure, additional unexpected fields, wrong data types)

**Step 2: Implement to Pass Tests (GREEN)**

- [ ] T023 [P] [US2] Implement `--validate` flag handler in `internal/command/report.go` (resolve session path, read existing file, call storage.ValidateReport(filePath), format errors with field paths)
- [ ] T024 [US2] Implement error formatter in `internal/storage/validate.go` to convert JSON Schema validation errors to instructive messages with format "[field_path]: [violation] (expected: [constraint], got: [value])" per FR-004a, with exit codes per FR-004b (0=success, 1=validation failure, 2=operational failure)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - agents can write and validate report files

---

## Phase 5: User Story 3 - External Agent Retrieves Report Schema (Priority: P2)

**Goal**: Enable agents to programmatically discover report file structure without hardcoding

**Independent Test**: Agent runs `fluxid report --get-schema`, parses YAML schema output, verifies all required fields documented

**Step 1: Write Failing E2E Test (RED)**

- [ ] T025 [US3] Add E2E test in `e2e-tests/tests/report_schema_test.go` for schema retrieval (valid YAML output, parseable as YAML schema, all required fields present)

**Step 2: Implement to Pass Test (GREEN)**

- [ ] T026 [P] [US3] Implement `--get-schema` flag handler in `internal/command/report.go` (call storage.GetReportSchema(), output YAML schema to stdout)

**Checkpoint**: All P1/P2 report functionality complete - agents have full self-service report capabilities

---

## Phase 6: User Story 4 - External Agent Records History (Priority: P2)

**Goal**: Enable agents to append history entries documenting failed approaches to prevent repetition

**Independent Test**: Agent obtains history file path, appends valid entries, subsequent session retrieves history

**Step 1: Write Failing E2E Tests (RED)**

- [ ] T027 [US4] Add E2E test in `e2e-tests/tests/history_append_test.go` for history workflow (get-file → append entry → validate → subsequent session reads history)
- [ ] T028 [US4] Add E2E test in `e2e-tests/tests/history_fifo_test.go` for FIFO eviction (create >10MB history file, verify oldest 30% of complete entries removed by count on read, YAML structure preserved)

**Step 2: Implement to Pass Tests (GREEN)**

- [ ] T029 [P] [US4] Create history command handler in `internal/command/history.go` with cobra command structure
- [ ] T030 [P] [US4] Implement `--get-file` flag handler in `internal/command/history.go` (resolve session ID from existing session management, construct path using os.TempDir(), ensure file/dir exist, output absolute path)
- [ ] T031 [P] [US4] Implement `--validate` flag handler in `internal/command/history.go` (resolve session path, read existing file, call storage.ValidateHistory(filePath), format errors)
- [ ] T032 [US4] Update root command in `internal/command/root.go` to register history command
- [ ] T033 [US4] Update workflow controller in `internal/workflow/workflow.go` to use storage.ReadHistory() instead of IPC-based history reading

**Checkpoint**: User Story 4 complete - agents can record and retrieve history across sessions

---

## Phase 7: User Story 5 - External Agent Retrieves History Schema (Priority: P3)

**Goal**: Support agent autonomy for history recording by providing schema programmatically

**Independent Test**: Agent runs `fluxid history --get-schema`, parses YAML schema, verifies required fields documented

**Step 1: Write Failing E2E Test (RED)**

- [ ] T034 [US5] Add E2E test in `e2e-tests/tests/history_schema_test.go` for schema retrieval (valid YAML output, parseable as YAML schema, required fields present)

**Step 2: Implement to Pass Test (GREEN)**

- [ ] T035 [P] [US5] Implement `--get-schema` flag handler in `internal/command/history.go` (call storage.GetHistorySchema(), output YAML schema to stdout)

**Checkpoint**: All history functionality complete - agents have full self-service history capabilities

---

## Phase 8: User Story 6 - Developer Debugs Workflow (Priority: P3)

**Goal**: Enable developers to manually inspect and validate report/history files for troubleshooting

**Independent Test**: Developer creates invalid files, runs validation commands, receives actionable error messages

**Step 1: Write Failing E2E Tests (RED) - Implementation Already Complete**

- [ ] T036 [P] [US6] Add E2E test in `e2e-tests/tests/developer_workflow_test.go` for developer debugging scenarios (inspect file paths, validate suspect files, view errors for common mistakes)
- [ ] T037 [US6] Add security validation E2E test in `e2e-tests/tests/yaml_security_test.go` (YAML anchors rejected, aliases rejected, merge keys rejected with clear errors)

**Note**: No new implementation needed - US6 uses functionality from US1-US5

**Checkpoint**: All user stories implemented and independently testable

---

## Phase 9: Breaking Change Removal

**Purpose**: Remove deprecated IPC functionality after all new functionality is tested and working

**⚠️ CRITICAL**: Only proceed after ALL previous phases complete and ALL E2E tests pass

- [ ] T038 Remove IPC command files if present in main branch: `internal/command/ipc.go`, `internal/command/ipc_handlers.go`, `internal/command/ipc_abort.go`, `internal/command/ipc_history.go` (verify existence before removal)
- [ ] T039 Remove IPC storage package if present in main branch: `internal/ipc/storage.go`, `internal/ipc/schema.yaml`, and entire `internal/ipc/` directory (verify existence and exact file count before removal)
- [ ] T040 Remove IPC E2E tests if present in main branch: verify and document actual E2E test files for IPC functionality before removal (feature branch may not have IPC tests yet)
- [ ] T041 [P] Update `internal/command/root_signal.go` to remove abort flag integration if present
- [ ] T042 [P] Update `internal/command/root_config_loader.go` to remove IPC-specific session handling if present
- [ ] T043 Verify breaking change complete: `grep -r "ipc" internal/command internal/storage` returns 0 matches, and `go build` succeeds
- [ ] T044 Run full E2E test suite and verify all tests pass with new file-based interface

**Checkpoint**: Breaking change complete - old IPC system fully removed, new file-based system proven

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and validation

- [ ] T045 [P] Update agent integration documentation to reflect file-based interface (reference quickstart.md)
- [ ] T046 [P] Add edge case E2E tests in `e2e-tests/tests/edge_cases_test.go` (file permission errors with actionable messages, empty files, path validation, directory creation)
- [ ] T047 Run complete validation using scenarios from quickstart.md
- [ ] T048 Verify all success criteria from spec.md met (SC-001 through SC-010)
- [ ] T049 Final code review: check error messages are instructive with format "[field_path]: [violation] (expected: [constraint], got: [value])", exit codes correct, silent success implemented, observability requirements met (FR-041 stderr errors, FR-042 silent success, FR-043 sufficient context)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - User Story 1 (US1) can start after Foundational
  - User Story 2 (US2) depends on US1 (extends report command)
  - User Story 3 (US3) depends on US1 (extends report command)
  - User Story 4 (US4) can start after Foundational (independent of report stories)
  - User Story 5 (US5) depends on US4 (extends history command)
  - User Story 6 (US6) depends on US1-US5 (tests all functionality)
- **Breaking Change Removal (Phase 9)**: Depends on ALL user stories complete and tested
- **Polish (Phase 10)**: Depends on breaking change removal

### User Story Dependencies

- **User Story 1 (P1)**: Foundation → US1 (no other story dependencies)
- **User Story 2 (P1)**: Foundation → US1 → US2 (extends report command)
- **User Story 3 (P2)**: Foundation → US1 → US3 (extends report command)
- **User Story 4 (P2)**: Foundation → US4 (independent story, creates history command)
- **User Story 5 (P3)**: Foundation → US4 → US5 (extends history command)
- **User Story 6 (P3)**: Foundation → US1, US2, US3, US4, US5 → US6 (validates all)

### Within Each User Story

- Foundation components (storage, validation) before command handlers
- Command handlers before workflow integration
- Workflow integration before E2E tests
- Core implementation before error handling
- Story complete before moving to next

### Parallel Opportunities

- **Phase 1 (Setup)**: T002 [P] and T003 [P] can run in parallel (different schema files)
- **Phase 2 (Foundational)**: T005 [P], T006 [P], T007 [P], T008 [P], T008a [P] can run in parallel (independent E2E tests), then T010 [P], T011 [P], T014a [P] can run in parallel (independent utilities)
- **Phase 3 (US1)**: T011 [P] and T012 [P] can run in parallel initially (same file but different functions)
- **Phase 4 (US2)**: T017 [P] can be worked on independently (extends report command)
- **Phase 5 (US3)**: T020 [P] can be worked on independently (extends report command)
- **Phase 6 (US4)**: T022 [P], T023 [P], T024 [P] can run in parallel (different functions in history.go)
- **Phase 7 (US5)**: T029 [P] independent work
- **Phase 8 (US6)**: T031 [P] and T032 [P] can run in parallel (different test files)
- **Phase 10 (Polish)**: T040 [P] and T041 [P] can run in parallel (docs vs tests)

**Key Insight**: Report stories (US1-US3) can proceed sequentially while history stories (US4-US5) proceed in parallel track after foundation

---

## Parallel Example: Foundation Phase

```bash
# Launch independent foundation E2E tests in parallel (RED phase):
Task T005: "E2E test for storage.ReadReport() scenarios"
Task T006: "E2E test for storage.ReadHistory() with FIFO eviction"
Task T007: "E2E test for YAML security validator"
Task T008: "E2E test for session path validator"
Task T008a: "E2E test for observability requirements"

# Then launch independent implementations in parallel (GREEN phase):
Task T010: "YAML security validator in internal/storage/security.go"
Task T011: "Session path validator in internal/storage/paths.go"
Task T014a: "Observability implementation in command handlers"
# These have no dependencies on each other and work on different files
```

## Parallel Example: Report Track vs History Track

```bash
# After Foundation completes, two parallel tracks possible:
# Track 1: Report functionality (US1 → US2 → US3)
Task T011-T016: "User Story 1 - Agent writes report"
Task T017-T019: "User Story 2 - Agent validates report"
Task T020-T021: "User Story 3 - Agent retrieves schema"

# Track 2: History functionality (US4 → US5) - parallel to Track 1
Task T022-T028: "User Story 4 - Agent records history"
Task T029-T030: "User Story 5 - Agent retrieves history schema"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (3 tasks: T001-T003)
2. Complete Phase 2: Foundational (12 tasks: T005-T014a including T008a - CRITICAL)
3. Complete Phase 3: User Story 1 (6 tasks: T015-T020)
4. **STOP and VALIDATE**: Test US1 independently with E2E test
5. Verify agents can write reports and fluxid processes them

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Core workflow functional
3. Add User Story 2 → Test independently → Validation working
4. Add User Story 3 → Test independently → Full report self-service
5. Add User Story 4 → Test independently → History recording functional
6. Add User Story 5 → Test independently → Full history self-service
7. Add User Story 6 → Test independently → Developer debugging enabled
8. Remove deprecated code → Validate → Breaking change complete

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (15 tasks: T001-T003, T005-T014a including T008a)
2. Once Foundational is done:
   - **Developer A**: Report track (US1 → US2 → US3)
   - **Developer B**: History track (US4 → US5)
   - **Developer C**: Developer experience (US6) after A & B
3. Once all stories complete:
   - **Team**: Breaking change removal (Phase 9) together
   - **Team**: Polish (Phase 10) together

---

## Notes

- [P] tasks = different files or independent functions, no execution dependencies
- [Story] label maps task to specific user story from spec.md for traceability
- Each user story should be independently completable and testable
- Breaking change removal (Phase 9) MUST NOT proceed until all new functionality tested
- Verify E2E tests pass before removing old code
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Constitution compliance: TDD via E2E tests, sequential workflow, pure Go, explicit interfaces

---

## Task Summary

**Total Tasks**: 50 (T001-T049 + T008a + T014a)
- Phase 1 (Setup): 3 tasks (T001-T003)
- Phase 2 (Foundational - TDD): 12 tasks (5 E2E tests first [T005-T008a], then 7 implementations [T009-T014a]) ⚠️ BLOCKS ALL USER STORIES
- Phase 3 (US1 - P1 - TDD): 6 tasks (1 E2E test [T015], then 5 implementations [T016-T020]) 🎯 MVP
- Phase 4 (US2 - P1 - TDD): 4 tasks (2 E2E tests [T021-T022], then 2 implementations [T023-T024])
- Phase 5 (US3 - P2 - TDD): 2 tasks (1 E2E test [T025], then 1 implementation [T026])
- Phase 6 (US4 - P2 - TDD): 7 tasks (2 E2E tests [T027-T028], then 5 implementations [T029-T033])
- Phase 7 (US5 - P3 - TDD): 2 tasks (1 E2E test [T034], then 1 implementation [T035])
- Phase 8 (US6 - P3 - TDD): 2 tasks (2 E2E tests [T036-T037], no new implementation)
- Phase 9 (Breaking Change): 7 tasks (T038-T044) ⚠️ ONLY AFTER ALL USER STORIES COMPLETE
- Phase 10 (Polish): 5 tasks (T045-T049)

**TDD Compliance**: ✅ ALL phases now follow Red-Green-Refactor cycle with E2E tests written BEFORE implementation

**Parallel Opportunities**: 21 tasks marked [P] can run in parallel within their phase (T002, T003, T005, T006, T007, T008, T008a, T010, T011, T014a, T016, T017, T020, T022, T023, T024, T029, T030, T031, T036, T037, T040, T041)

**Independent Test Criteria**:
- US1: Agent writes report file → fluxid reads → workflow proceeds
- US2: Agent validates report file → receives instructive errors
- US3: Agent retrieves schema → parses successfully
- US4: Agent appends history → subsequent session reads
- US5: Agent retrieves history schema → parses successfully
- US6: Developer validates files → receives actionable errors

**MVP Scope (Recommended)**: Phase 1 + Phase 2 + Phase 3 (21 tasks: T001-T003, T005-T020, T008a, T014a) delivers core agent-to-fluxid report communication with full TDD E2E coverage

**Format Validation**: ✅ ALL tasks follow required checklist format: `- [ ] [ID] [P?] [Story?] Description with file paths`

**Constitution Compliance**: ✅ TDD violations fixed - all E2E tests now written BEFORE implementation
