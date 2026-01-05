# Implementation Plan: Report & History File-Based Interface

**Branch**: `001-report-history-refactor` | **Date**: 2026-01-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-report-history-refactor/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Replace stdio-based IPC system with file-based report/history interface. External agents will write YAML report/history files directly to session-specific paths obtained via new CLI commands (`fluxid report --get-file/--get-schema/--validate` and `fluxid history --get-file/--get-schema/--validate`). This breaking change removes all existing IPC functionality, eliminates temp file stdio storage, and simplifies agent integration by making file I/O explicit. Workflow controller continues reading report/history files but provides no write operations (delegated to agents).

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**:
- `gopkg.in/yaml.v3` (existing - YAML parsing, needs security constraints configured)
- `github.com/xeipuuv/gojsonschema` (new - JSON Schema validation for report/history validation commands)
- `github.com/google/uuid` (existing - session ID generation)

**Storage**: File-based YAML storage in session-specific directories (report.yaml, history.yaml)
**Testing**: Go testing (`go test`), E2E test suite (existing framework in `e2e-tests/`)
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single CLI project with workflow orchestration
**Performance Goals**:
- File validation < 100ms for 10MB files
- Schema retrieval < 10ms (embedded in binary)
- File path resolution < 5ms

**Constraints**:
- Report files read entirely into memory (< 10MB enforced)
- History files truncated via FIFO eviction if > 10MB before reading
- Sequential workflow (no concurrent file access between agent and fluxid)
- YAML security: disable anchors, aliases, merge keys to prevent complexity attacks

**Scale/Scope**:
- Single CLI binary with embedded schemas
- 6 new CLI commands (report --get-file/--get-schema/--validate, history --get-file/--get-schema/--validate)
- Remove 4 command files + 1 directory (internal/ipc) + 7 E2E test files
- Update 5 integration points (signal handler, workflow, config loader, root command, output)
- 2 JSON Schema documents (report schema, history schema)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-Driven Development (NON-NEGOTIABLE)
**Status**: ✅ PASS

**Compliance**: Breaking change requires comprehensive test coverage. User explicitly requested "ensure full e2e coverage, less BIG flows are better than lots of small flows." Implementation enforces TDD per tasks.md:
1. Write failing E2E tests BEFORE implementation (Red-Green-Refactor cycle)
2. Implement new commands to pass tests (Green)
3. Refactor while keeping tests green
4. Remove deprecated IPC code only after new tests pass
5. Remove deprecated E2E tests (M03, M04) after replacement coverage verified

**Test Strategy**:
- User Story 1 (P1): E2E test for agent writing report via --get-file
- User Story 2 (P1): E2E test for report validation (valid/invalid cases)
- User Story 3 (P2): E2E test for schema retrieval and parsing
- User Story 4 (P2): E2E test for history recording and retrieval
- User Story 5 (P3): E2E test for history schema retrieval
- User Story 6 (P3): Developer workflow test for manual validation

User requested "less BIG flows are better than lots of small flows" → consolidate into comprehensive workflow tests rather than granular unit tests for each subcommand.

---

### II. Full E2E Test Coverage (NON-NEGOTIABLE)
**Status**: ✅ PASS

**Compliance**: All 6 user stories have E2E acceptance scenarios defined in spec.md. Current plan removes 6 E2E tests (M03-E05, M04-E01 through M04-E05) and replaces with new comprehensive E2E tests covering:
- Complete agent workflow: get-file → write → validate → fluxid reads
- Error scenarios: missing files, invalid YAML, schema violations
- Edge cases: file size limits, path validation, security constraints
- Integration: workflow loop progression with file-based approach

**Verification**: No existing E2E tests will be removed until replacement coverage is implemented and passing.

---

### III. Strictly Sequential Workflow (NON-NEGOTIABLE)
**Status**: ✅ PASS

**Compliance**: This refactor reinforces sequential workflow. Spec.md explicitly states (A-005): "Workflow execution is 100% sequential - agent writes report/history and exits, then fluxid reads files. No concurrent file access between agent and fluxid is possible."

**Changes**:
- Remove file locking logic from internal/ipc/storage.go (no longer needed)
- No concurrent access patterns introduced
- Agent writes → exits → fluxid reads (strict ordering)
- No async operations, no parallelism

---

### IV. Strict Code Quality Enforcement (NON-NEGOTIABLE)
**Status**: ✅ PASS

**Compliance**: Pre-commit hooks remain unchanged. All new code subject to:
- gofmt/goimports formatting
- golangci-lint strict checks
- gosec security scanning
- Test coverage threshold (90%)

**Breaking Change Impact**: Removes ~25 files, improving code quality by eliminating stdio-based IPC complexity. Reduction in codebase size aligns with SOLID (single responsibility), DRY (eliminates duplication between IPC and direct file I/O), KISS (simpler interface), YAGNI (removes unused abstraction).

---

### V. Pure Go Implementation
**Status**: ✅ PASS

**Compliance**: All implementation in Go. No new shell script dependencies. Breaking change removes stdio-based IPC that used temp file indirection - replacement is pure Go file I/O via `os` package.

---

### VI. Explicit Interfaces Over Implicit Behavior
**Status**: ✅ PASS

**Compliance**: New file-based interface is more explicit than stdio IPC:
- Agent obtains file paths via explicit CLI commands (`--get-file`)
- Schema structure documented via JSON Schema (`--get-schema`)
- Validation explicit via CLI command (`--validate`)
- No implicit temp directory management, no hidden stdio protocols

FR-012, FR-013 enforce explicit path management: fluxid alone determines and manages file locations, agents cannot provide custom paths.

---

### VII. Fail Fast with Clear Diagnostics
**Status**: ✅ PASS

**Compliance**: Requirements explicitly mandate instructive errors:
- FR-005: Report validation error message format requirements
- FR-018: History validation error message format requirements
- FR-041/FR-042/FR-043: Error output requirements (stderr, silent success, sufficient context)

New validation using JSON Schema validator will provide field-level error messages per specified format.

---

### VIII. Command-Line First, Scriptable Always
**Status**: ✅ PASS

**Compliance**: New commands follow CLI-first design:
- Parseable output: JSON Schema to stdout, file paths to stdout
- Exit codes: 0 for success, non-zero for failure
- Stdin/stdout/stderr adherence: data to stdout, errors to stderr
- No interactive prompts
- Session context via environment variable (FLUXID_SESSION_ID)

Breaking change improves scriptability by replacing stdio IPC protocol with standard file I/O.

---

### Summary
**All Constitution Gates**: ✅ PASS

No violations identified. Breaking change aligns with constitution principles by:
1. Requiring comprehensive E2E test coverage (TDD, full coverage)
2. Reinforcing sequential workflow (eliminates concurrency concerns)
3. Improving code quality (removes complex stdio abstraction)
4. Maintaining pure Go implementation
5. Making interfaces more explicit
6. Enforcing fail-fast with clear errors
7. Following CLI-first design

**Proceed to Phase 0 Research**: ✅ APPROVED

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
fluxid-cli/
├── cmd/fluxid/              # CLI entry point (main.go)
├── internal/
│   ├── command/             # Command handlers
│   │   ├── report.go       # NEW: report --get-file/--get-schema/--validate
│   │   ├── history.go      # NEW: history --get-file/--get-schema/--validate
│   │   ├── root.go         # MODIFY: remove IPC command routing
│   │   ├── root_signal.go  # MODIFY: remove abort flag integration
│   │   └── root_config_loader.go  # MODIFY: session ID handling
│   ├── storage/             # NEW: file-based storage layer
│   │   ├── report.go       # Report file operations
│   │   ├── history.go      # History file operations
│   │   ├── schema.go       # Schema embedding and retrieval
│   │   └── validate.go     # YAML validation with JSON Schema
│   ├── workflow/            # MODIFY: remove abort checks, update report reading
│   │   └── workflow.go
│   └── types/               # Shared types
│       └── config.go
├── internal/assets/
│   └── schemas/             # NEW: embedded JSON Schema files
│       ├── report.json     # Report JSON Schema
│       └── history.json    # History JSON Schema
└── e2e-tests/
    └── tests/
        ├── report_*.go     # NEW: E2E tests for report commands
        └── history_*.go    # NEW: E2E tests for history commands

# Files to REMOVE (breaking change):
internal/command/ipc*.go               # All IPC command handlers (6 files)
internal/ipc/                          # Entire IPC package (12 files)
e2e-tests/tests/m03-e05*.go           # Abort E2E test (1 file)
e2e-tests/tests/m04-e0*.go            # History IPC E2E tests (6 files)
```

**Structure Decision**: Single Go CLI project. New storage package replaces internal/ipc package with file-based operations. Command layer gets new report.go and history.go handlers. JSON Schemas embedded in binary via internal/assets/schemas/. E2E tests reorganized: remove 7 IPC-related tests, add comprehensive file-based workflow tests.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
