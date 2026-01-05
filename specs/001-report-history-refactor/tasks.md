---
description: "Tasks for Report & History File-Based Interface Refactor"
---

# Tasks: Report & History File-Based Interface

**Input**: Design documents from `/specs/001-report-history-refactor/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

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

- [ ] T001 Create directory structure `internal/storage/` and `internal/assets/schemas/`
- [ ] T002 [P] Create JSON Schema file for report in `internal/assets/schemas/report.json` using content from contracts/report-schema.json
- [ ] T003 [P] Create JSON Schema file for history in `internal/assets/schemas/history.json` using content from contracts/history-schema.json
- [ ] T004 Add `github.com/xeipuuv/gojsonschema` dependency via `go get`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core storage infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T005 Implement schema embedding and retrieval in `internal/storage/schema.go` (embed report.json and history.json, provide GetReportSchema() and GetHistorySchema() functions)
- [ ] T006 [P] Implement YAML security validator in `internal/storage/security.go` (reject anchors &, aliases *, merge keys <<)
- [ ] T007 [P] Implement session path validator in `internal/storage/paths.go` (validate UUID format, prevent path traversal, ensure within session root)
- [ ] T008 Implement report file reader in `internal/storage/report.go` with ReadReport(sessionID) function
- [ ] T009 Implement history file reader in `internal/storage/history.go` with ReadHistory(sessionID) and FIFO eviction logic for 10MB limit
- [ ] T010 Implement JSON Schema validator in `internal/storage/validate.go` with ValidateReport() and ValidateHistory() functions
- [ ] T010b [P] Add E2E test in `e2e-tests/tests/storage_report_test.go` for storage.ReadReport() (valid report, malformed YAML, missing file scenarios)
- [ ] T010c [P] Add E2E test in `e2e-tests/tests/storage_history_test.go` for storage.ReadHistory() with FIFO eviction (valid history, >10MB truncation, empty file scenarios)
- [ ] T010d [P] Add E2E test in `e2e-tests/tests/security_yaml_test.go` for YAML security validator (reject anchors, aliases, merge keys with clear errors)
- [ ] T010e [P] Add E2E test in `e2e-tests/tests/path_validation_test.go` for session path validator (valid UUID paths, path traversal rejection, invalid session IDs)

**Checkpoint**: Foundation ready with full E2E test coverage - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - External Agent Writes Report (Priority: P1) 🎯 MVP

**Goal**: Enable external agents to query report file path, write valid YAML reports, and have fluxid read them correctly

**Independent Test**: Agent runs `fluxid report --get-file`, writes valid YAML report to returned path, fluxid workflow reads and processes status correctly

### Implementation for User Story 1

- [ ] T011 [P] [US1] Create report command handler in `internal/command/report.go` with cobra command structure
- [ ] T012 [P] [US1] Implement `--get-file` flag handler in `internal/command/report.go` (resolve session ID from env, construct path, ensure file/dir exist, output absolute path)
- [ ] T013 [US1] Update root command in `internal/command/root.go` to register report command
- [ ] T014 [US1] Update workflow controller in `internal/workflow/workflow.go` to use storage.ReadReport() instead of IPC-based report reading
- [ ] T015 [US1] Add error handling in `internal/command/report.go` for missing FLUXID_SESSION_ID (exit code 3)
- [ ] T016 [US1] Add E2E test in `e2e-tests/tests/report_write_test.go` for complete agent workflow (get-file → write report → fluxid reads → PASS/FAIL behavior) with performance assertion (< 5 seconds per SC-001)

**Checkpoint**: At this point, User Story 1 should be fully functional - agents can write reports and fluxid processes them

---

## Phase 4: User Story 2 - External Agent Validates Report (Priority: P1)

**Goal**: Enable agents to validate report structure before fluxid reads it, preventing workflow failures

**Independent Test**: Agent writes valid/invalid reports and runs `fluxid report --validate`, receiving appropriate exit codes and error messages

### Implementation for User Story 2

- [ ] T017 [P] [US2] Implement `--validate` flag handler in `internal/command/report.go` (read file from session path, call storage.ValidateReport(), format errors with field paths)
- [ ] T018 [US2] Implement error formatter in `internal/storage/validate.go` to convert JSON Schema errors to instructive messages with field paths
- [ ] T019 [US2] Add E2E test in `e2e-tests/tests/report_validate_test.go` covering valid report (exit 0), missing required field (exit 1 with field name), invalid status value (exit 1 with enum values), file not found (exit 2)
- [ ] T019b [US2] Add E2E test in `e2e-tests/tests/report_validate_test.go` for schema mismatch scenarios (valid YAML but wrong structure, additional unexpected fields, wrong data types)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - agents can write and validate reports

---

## Phase 5: User Story 3 - External Agent Retrieves Report Schema (Priority: P2)

**Goal**: Enable agents to programmatically discover report structure without hardcoding

**Independent Test**: Agent runs `fluxid report --get-schema`, parses JSON Schema output, verifies all required fields documented

### Implementation for User Story 3

- [ ] T020 [P] [US3] Implement `--get-schema` flag handler in `internal/command/report.go` (call storage.GetReportSchema(), output JSON to stdout)
- [ ] T021 [US3] Add E2E test in `e2e-tests/tests/report_schema_test.go` for schema retrieval (valid JSON output, parseable as JSON Schema Draft 7, all required fields present)

**Checkpoint**: All P1/P2 report functionality complete - agents have full self-service report capabilities

---

## Phase 6: User Story 4 - External Agent Records History (Priority: P2)

**Goal**: Enable agents to append history entries documenting failed approaches to prevent repetition

**Independent Test**: Agent obtains history file path, appends valid entries, subsequent session retrieves history

### Implementation for User Story 4

- [ ] T022 [P] [US4] Create history command handler in `internal/command/history.go` with cobra command structure
- [ ] T023 [P] [US4] Implement `--get-file` flag handler in `internal/command/history.go` (resolve session ID, construct path, ensure file/dir exist, output absolute path)
- [ ] T024 [P] [US4] Implement `--validate` flag handler in `internal/command/history.go` (read file, call storage.ValidateHistory(), format errors)
- [ ] T025 [US4] Update root command in `internal/command/root.go` to register history command
- [ ] T026 [US4] Update workflow controller in `internal/workflow/workflow.go` to use storage.ReadHistory() instead of IPC-based history reading
- [ ] T027 [US4] Add E2E test in `e2e-tests/tests/history_append_test.go` for history workflow (get-file → append entry → validate → subsequent session reads history)
- [ ] T028 [US4] Add E2E test in `e2e-tests/tests/history_fifo_test.go` for FIFO eviction (create >10MB history file, verify oldest entries removed on read)

**Checkpoint**: User Story 4 complete - agents can record and retrieve history across sessions

---

## Phase 7: User Story 5 - External Agent Retrieves History Schema (Priority: P3)

**Goal**: Support agent autonomy for history recording by providing schema programmatically

**Independent Test**: Agent runs `fluxid history --get-schema`, parses JSON Schema, verifies required fields documented

### Implementation for User Story 5

- [ ] T029 [P] [US5] Implement `--get-schema` flag handler in `internal/command/history.go` (call storage.GetHistorySchema(), output JSON to stdout)
- [ ] T030 [US5] Add E2E test in `e2e-tests/tests/history_schema_test.go` for schema retrieval (valid JSON output, parseable as JSON Schema Draft 7, required fields present)

**Checkpoint**: All history functionality complete - agents have full self-service history capabilities

---

## Phase 8: User Story 6 - Developer Debugs Workflow (Priority: P3)

**Goal**: Enable developers to manually inspect and validate report/history files for troubleshooting

**Independent Test**: Developer creates invalid files, runs validation commands, receives actionable error messages

### Implementation for User Story 6

- [ ] T031 [P] [US6] Add E2E test in `e2e-tests/tests/developer_workflow_test.go` for developer debugging scenarios (inspect file paths, validate suspect files, view errors for common mistakes)
- [ ] T032 [US6] Add security validation E2E test in `e2e-tests/tests/yaml_security_test.go` (YAML anchors rejected, aliases rejected, merge keys rejected with clear errors)

**Checkpoint**: All user stories implemented and independently testable

---

## Phase 9: Breaking Change Removal

**Purpose**: Remove deprecated IPC functionality after all new functionality is tested and working

**⚠️ CRITICAL**: Only proceed after ALL previous phases complete and ALL E2E tests pass

- [ ] T033 Remove IPC command files: `internal/command/ipc.go`, `internal/command/ipc_handlers.go`, `internal/command/ipc_abort.go`, `internal/command/ipc_history.go`
- [ ] T034 Remove IPC storage package: `internal/ipc/storage.go`, `internal/ipc/schema.yaml`, and entire `internal/ipc/` directory
- [ ] T035 Remove IPC E2E tests: `e2e-tests/tests/m03_e05_*.go` (abort test), `e2e-tests/tests/m04_e0*_*.go` (history IPC tests - 5 files)
- [ ] T036 [P] Update `internal/command/root_signal.go` to remove abort flag integration if present
- [ ] T037 [P] Update `internal/command/root_config_loader.go` to remove IPC-specific session handling
- [ ] T038 Verify breaking change complete: `grep -r "ipc" internal/command internal/storage` returns 0 matches, E2E test count reduced by exactly 7 tests (1 from M03, 6 from M04), and `go build` succeeds
- [ ] T039 Run full E2E test suite and verify all tests pass with new file-based interface

**Checkpoint**: Breaking change complete - old IPC system fully removed, new file-based system proven

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and validation

- [ ] T040 [P] Update agent integration documentation to reflect file-based interface (reference quickstart.md)
- [ ] T041 [P] Add edge case E2E tests in `e2e-tests/tests/edge_cases_test.go` (file permission errors with actionable messages, empty files, missing session ID, path validation, directory creation)
- [ ] T042 Run complete validation using scenarios from quickstart.md
- [ ] T043 Verify all success criteria from spec.md met (SC-001 through SC-010)
- [ ] T044 Final code review: check error messages are instructive, exit codes correct, silent success implemented, observability requirements met (FR-041 stderr errors, FR-042 silent success, FR-043 sufficient context)

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
- **Phase 2 (Foundational)**: T006 [P] and T007 [P] can run in parallel (independent utilities)
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
# Launch independent foundation components in parallel:
Task T006: "YAML security validator in internal/storage/security.go"
Task T007: "Session path validator in internal/storage/paths.go"
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

1. Complete Phase 1: Setup (4 tasks)
2. Complete Phase 2: Foundational (6 tasks - CRITICAL)
3. Complete Phase 3: User Story 1 (6 tasks)
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

1. Team completes Setup + Foundational together (10 tasks)
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

**Total Tasks**: 49
- Phase 1 (Setup): 4 tasks
- Phase 2 (Foundational): 10 tasks (6 implementation + 4 E2E tests) ⚠️ BLOCKS ALL USER STORIES
- Phase 3 (US1 - P1): 6 tasks 🎯 MVP
- Phase 4 (US2 - P1): 4 tasks (includes schema mismatch test)
- Phase 5 (US3 - P2): 2 tasks
- Phase 6 (US4 - P2): 7 tasks
- Phase 7 (US5 - P3): 2 tasks
- Phase 8 (US6 - P3): 2 tasks
- Phase 9 (Breaking Change): 7 tasks ⚠️ ONLY AFTER ALL USER STORIES COMPLETE
- Phase 10 (Polish): 5 tasks

**Parallel Opportunities**: 19 tasks marked [P] can run in parallel within their phase

**Independent Test Criteria**:
- US1: Agent writes report → fluxid reads → workflow proceeds
- US2: Agent validates report → receives instructive errors
- US3: Agent retrieves schema → parses successfully
- US4: Agent appends history → subsequent session reads
- US5: Agent retrieves history schema → parses successfully
- US6: Developer validates files → receives actionable errors

**MVP Scope (Recommended)**: Phase 1 + Phase 2 + Phase 3 (20 tasks) delivers core agent-to-fluxid report communication with full E2E coverage

**Format Validation**: ✅ ALL tasks follow required checklist format: `- [ ] [ID] [P?] [Story?] Description with file paths`
