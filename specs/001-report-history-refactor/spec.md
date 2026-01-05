# Feature Specification: Report & History File-Based Interface

**Feature Branch**: `001-report-history-refactor`
**Created**: 2026-01-05
**Status**: Draft
**Input**: User description: "Breaking refactor: Replace stdio-based IPC with file-based report/history system. New CLI: fluxid report --get-schema/--get-file/--validate and fluxid history --get-file/--get-schema/--validate. Remove all existing IPC functionality and affected E2E tests."

## Clarifications

### Session 2026-01-05

- Q: What should happen when report/history files exceed expected size during validation or read operations? → A: Read entire file into memory, but cut off old entries if size exceeds 10MB
- Q: Should the YAML parser enforce security restrictions beyond file size limits when reading report/history files from external agents? → A: Disable advanced YAML features (anchors, aliases, merge keys) to prevent complexity attacks
- Q: What level of observability (logging/error reporting) should be included when processing report/history files? → A: Minimal - only output errors to stderr when commands fail, no logging of successful operations
- Q: How should the system handle file path validation to prevent security issues when external agents interact with report/history files? → A: Strict - validate that report/history file paths are within session-specific directory, reject path traversal attempts. Fluxid alone manages the file path/existence.
- Q: Should the system provide guidance or guarantees about how external agents should write report/history files to prevent race conditions with fluxid reads? → A: 100% sequential workflow - agent writes, then exits, then fluxid reads. No race condition possible.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - External Agent Writes Report (Priority: P1)

An external coding agent (Claude, OpenCode, etc.) needs to write a workflow report after completing an implement or review phase. The agent must know where to write the report and how to structure it correctly.

**Why this priority**: This is the core workflow integration point. Without this, agents cannot report their work status back to fluxid, blocking the entire implement-review loop.

**Independent Test**: Can be fully tested by having an agent query the report file path, write a valid YAML report to that location, and verify fluxid reads it correctly. Delivers immediate value by enabling basic agent-to-fluxid communication.

**Acceptance Scenarios**:

1. **Given** a fluxid workflow session is active, **When** agent runs `fluxid report --get-file`, **Then** system returns absolute path to report file and ensures the file exists (creates if needed)
2. **Given** agent has report file path, **When** agent writes valid YAML report following schema, **Then** fluxid workflow reads report and progresses based on PASS/FAIL status
3. **Given** agent writes malformed report, **When** agent runs `fluxid report --validate`, **Then** system displays specific validation errors with field names and reasons

---

### User Story 2 - External Agent Validates Report (Priority: P1)

An external agent wants to verify its report is correctly formatted before fluxid attempts to read it, preventing workflow failures due to validation errors.

**Why this priority**: Immediate feedback prevents workflow failures. Agents can self-correct before submission, reducing retry loops and improving user experience.

**Independent Test**: Can be fully tested by providing various report files (valid, invalid, missing fields) to `fluxid report --validate` and verifying error messages are instructive and actionable.

**Acceptance Scenarios**:

1. **Given** a valid report file, **When** agent runs `fluxid report --validate`, **Then** system confirms validation success with exit code 0
2. **Given** an invalid report (missing required field), **When** agent runs `fluxid report --validate`, **Then** system displays error message specifying the missing field name and requirement
3. **Given** an invalid report (wrong status value), **When** agent runs `fluxid report --validate`, **Then** system displays error explaining allowed values (PASS, FAIL)
4. **Given** no report file exists, **When** agent runs `fluxid report --validate`, **Then** system displays error indicating file not found with expected path

---

### User Story 3 - External Agent Retrieves Report Schema (Priority: P2)

An external agent needs the report schema to understand required fields and structure when generating reports.

**Why this priority**: Enables agent autonomy and reduces documentation dependency. Agents can discover schema programmatically.

**Independent Test**: Can be fully tested by running `fluxid report --get-schema`, parsing output as YAML schema, and verifying all required fields are documented.

**Acceptance Scenarios**:

1. **Given** any context, **When** agent runs `fluxid report --get-schema`, **Then** system outputs complete YAML schema to stdout
2. **Given** schema output, **When** agent parses it, **Then** all required fields (command, artifact, timestamp, status, issues) are defined with types and constraints
3. **Given** schema output, **When** agent examines status field, **Then** enum values PASS and FAIL are specified

---

### User Story 4 - External Agent Records History (Priority: P2)

An external agent needs to record why a particular approach failed to prevent future sessions from repeating the same mistake.

**Why this priority**: Enables learning across sessions, reducing wasted effort. Secondary to basic workflow operation but critical for multi-session efficiency.

**Independent Test**: Can be fully tested by agent obtaining history file path, writing structured entries with timestamps and failure reasons, then verifying fluxid surfaces this history to subsequent sessions.

**Acceptance Scenarios**:

1. **Given** a fluxid workflow session is active, **When** agent runs `fluxid history --get-file`, **Then** system returns absolute path to history file and ensures it exists (creates if needed)
2. **Given** agent has history file path, **When** agent appends valid history entry following schema, **Then** subsequent session can retrieve this entry
3. **Given** agent writes history entry, **When** agent runs `fluxid history --validate`, **Then** system confirms entry structure is valid or reports specific errors

---

### User Story 5 - External Agent Retrieves History Schema (Priority: P3)

An external agent needs the history schema to understand required structure for recording workflow history entries.

**Why this priority**: Supports agent autonomy for history recording. Less critical than report schema since history is advisory rather than workflow-critical.

**Independent Test**: Can be fully tested by running `fluxid history --get-schema`, parsing output as YAML schema, and verifying timestamp, step, status, and summary fields are documented.

**Acceptance Scenarios**:

1. **Given** any context, **When** agent runs `fluxid history --get-schema`, **Then** system outputs complete YAML schema to stdout
2. **Given** schema output, **When** agent parses it, **Then** required fields (timestamp, step, status, summary) are defined with types
3. **Given** schema output, **When** agent examines status field, **Then** enum values SUCCESS and FAIL are specified

---

### User Story 6 - Developer Debugs Workflow (Priority: P3)

A developer troubleshooting workflow failures needs to manually inspect or validate report/history files to understand what's happening.

**Why this priority**: Developer experience and debuggability. Not required for basic workflow operation but improves maintainability.

**Independent Test**: Can be fully tested by manually creating invalid report/history files and running validation commands to get actionable error messages.

**Acceptance Scenarios**:

1. **Given** a workflow session directory, **When** developer runs `fluxid report --get-file`, **Then** developer knows exact file path to inspect
2. **Given** a suspect report file, **When** developer runs `fluxid report --validate`, **Then** developer receives specific validation errors
3. **Given** multiple workflow sessions, **When** developer runs `fluxid history --get-file` in each, **Then** developer can inspect session-specific history

---

### Edge Cases

- What happens when report file path directory doesn't exist? System must create parent directories within `<OS temp folder>/fluxid/` directory only.
- What happens when report/history files have incorrect permissions? System must handle permission errors with clear messages.
- What happens when validation is run on empty files? System must report "empty file" error.
- What happens when schema is requested but internal schema file is missing? System must fail fast with clear error.
- What happens when multiple processes write to history file simultaneously? Cannot occur - workflow is 100% sequential (agent writes and exits, then fluxid reads). No concurrent access possible.
- What happens when report file contains YAML but doesn't match schema? Validation must report specific field mismatches.
- What happens when history schema requires timestamp format but entry has invalid timestamp? Validation must report timestamp format error with example.
- What happens when session ID is not set? Commands must fail with clear error indicating session context is required.
- What happens when history file exceeds 10MB? System must truncate by removing oldest 30% of entries (FIFO eviction) to preserve valid YAML structure before reading.
- What happens when a single history entry exceeds 10MB? System must reject with error instructing agent to split entries.
- What happens when report/history files contain advanced YAML features (anchors, aliases, merge keys)? Parser must reject these features with clear error to prevent complexity attacks.
- What happens if an agent attempts path traversal (e.g., ../../etc/passwd)? System determines all paths internally; agents receive paths via --get-file and cannot specify custom paths.
- What happens if fluxid reads a report/history file while agent is still writing? Cannot occur - workflow guarantees agent exits before fluxid reads. No partial read protection needed.

## Requirements *(mandatory)*

### Functional Requirements

**Report Management:**

- **FR-001**: System MUST provide `fluxid report --get-schema` command that outputs complete JSON Schema (draft-07) for report file structure as YAML to stdout. Complete schema includes: all required fields with types and constraints, all optional fields, examples, and additionalProperties:false enforcement
- **FR-002**: System MUST provide `fluxid report --get-file` command that returns absolute path to session's report file
- **FR-003**: System MUST ensure report file and parent directories exist when `--get-file` is called (create if missing within `<OS temp folder>/fluxid/` directory)
- **FR-004**: System MUST provide `fluxid report --validate` command that validates current session's report against schema
- **FR-004a**: All validation commands MUST produce instructive error messages in format "[field_path]: [violation] (expected: [constraint], got: [value])" for human readability and as JSON array with {field, violation, constraint, value} objects for programmatic parsing
- **FR-004b**: System MUST use standardized exit codes: 0 for successful validation, 1 for validation failures (schema violations), 2 for operational failures (file not found, permission denied, schema load failure)
- **FR-005**: Report validation MUST produce error messages per FR-004a format and exit codes per FR-004b
- **FR-006**: Report validation MUST succeed with exit code 0 for valid reports
- **FR-007**: Report validation MUST fail with exit codes 1 or 2 per FR-004b and error message for invalid reports
- **FR-008**: System MUST NOT provide any report write functionality (delegated to external systems)
- **FR-009**: System MUST read report file during workflow execution to determine PASS/FAIL status
- **FR-010**: Report file location MUST be deterministic based on session ID in format `<OS temp folder>/fluxid/report-<session-id>.yaml`
- **FR-011**: System MUST reject report and history files containing YAML anchors, aliases, or merge keys with clear error message to prevent complexity attacks
- **FR-012**: System MUST validate all report and history file paths are within `<OS temp folder>/fluxid/` directory boundaries (reject path traversal attempts, validate session ID is valid UUID)
- **FR-013**: System MUST NOT accept agent-provided file paths; fluxid alone determines and manages report and history file locations

**History Management:**

- **FR-014**: System MUST provide `fluxid history --get-schema` command that outputs complete JSON Schema (draft-07) for history array structure as YAML to stdout. Complete schema includes: all required fields with types and constraints, all optional fields, examples, and additionalProperties:false enforcement
- **FR-015**: System MUST provide `fluxid history --get-file` command that returns absolute path to session's history file
- **FR-016**: System MUST ensure history file and parent directories exist when `--get-file` is called (create if missing within `<OS temp folder>/fluxid/` directory)
- **FR-017**: System MUST provide `fluxid history --validate` command that validates current session's history against schema
- **FR-018**: History validation MUST produce error messages per FR-004a format and exit codes per FR-004b
- **FR-019**: History validation MUST succeed with exit code 0 for valid history
- **FR-020**: History validation MUST fail with exit codes 1 or 2 per FR-004b and error message for invalid history
- **FR-021**: System MUST NOT provide any history write functionality (delegated to external systems)
- **FR-022**: History file location MUST be deterministic based on session ID in format `<OS temp folder>/fluxid/history-<session-id>.yaml`
- **FR-023**: System MUST read history file during workflow execution to provide context to agents
- **FR-024**: System MUST enforce 10MB maximum size for history files by removing oldest 30% of complete entries (FIFO eviction by entry count using ceiling function for minimum 1 entry removal, removing only whole event objects to preserve valid YAML array structure) before reading when size exceeds limit
- **FR-025**: System MUST reject during validation (via `--validate` command) when a single history entry exceeds 10MB with clear error instructing agent to split entries
- **FR-025a**: System MUST fail fast with exit code 2 and clear error message when embedded schema files (report-schema.yaml, history-schema.yaml) fail to load from binary assets, indicating internal system error requiring binary rebuild

**Schema Definitions:**

- **FR-026**: Report schema MUST define required fields: command (string), artifact (string), timestamp (ISO 8601 string), status (enum: PASS, FAIL), issues (object with 5 categories: blockers, defects, concerns, observations, enhancements)
- **FR-027**: Report schema MUST define optional fields: next_steps (array), summary (string)
- **FR-028**: Report schema MUST NOT allow additional properties; agents MUST use --validate to ensure validated YAML conforms strictly to provided schema
- **FR-029**: History schema MUST define array of event objects
- **FR-030**: History event schema MUST define required fields: timestamp (ISO 8601 string), step (string), status (enum: SUCCESS, FAIL), summary (string)
- **FR-031**: History event schema MUST define optional field: details (string describing approach and failure reason)
- **FR-032**: History event schema MUST NOT allow additional properties; agents MUST use --validate to ensure validated YAML conforms strictly to provided schema

**Breaking Changes:**

- **FR-033**: System MUST completely remove `fluxid ipc` command and all subcommands (get-report-schema, write-report, read-report, write-history, view-history, abort)
- **FR-034**: System MUST remove all stdio-based IPC functionality from codebase
- **FR-035**: System MUST remove all E2E tests that validate removed IPC functionality
- **FR-036**: System MUST remove internal storage abstraction for stdio-based operations (temp file management for IPC)

**Workflow Integration:**

- **FR-037**: Workflow execution MUST read report file after agent process exits and before retry decision to determine PASS/FAIL status for loop progression (PASS = exit loop, FAIL = retry with history context)
- **FR-038**: Workflow execution MUST continue to provide session ID to agents via environment variable
- **FR-039**: Workflow execution MUST continue to read history files to provide failure context to agents
- **FR-040**: Workflow execution behavior MUST remain unchanged (only interface to report/history changes)

**Observability:**

- **FR-041**: Commands MUST output all errors to stderr when operations fail (validation errors, file I/O errors, schema load errors). Stdout reserved exclusively for data output (file paths, schemas)
- **FR-042**: Commands MUST NOT log or output any messages to stderr for successful operations (silent success). Exit code 0 is the only success indicator besides stdout data output
- **FR-043**: Error messages MUST include sufficient context to diagnose failures: file path for I/O errors, field path for validation errors, constraint details for schema violations, session ID for context errors

### Key Entities

- **Report File**: YAML file containing workflow phase results (implement or review). Located at `<OS temp folder>/fluxid/report-<session-id>.yaml` where OS temp folder is platform-specific (e.g., `/tmp` on Unix, `%TEMP%` on Windows). Contains status (PASS/FAIL), issues categorized by severity, and optional next steps. Written by external agents, read by fluxid workflow.

- **History File**: YAML array of workflow events. Located at `<OS temp folder>/fluxid/history-<session-id>.yaml` where OS temp folder is platform-specific (e.g., `/tmp` on Unix, `%TEMP%` on Windows). Each entry records timestamp, step name, outcome (SUCCESS/FAIL), summary, and failure details. Written by external agents, read by fluxid workflow.

- **Report Schema**: JSON Schema (draft-07) document defining report file structure. Embedded in fluxid binary at `internal/assets/templates/report-schema.yaml` (JSON Schema outputted as YAML format). Outputs via `--get-schema` command. Used for validation.

- **History Schema**: JSON Schema (draft-07) document defining history file structure. Embedded in fluxid binary at `internal/assets/templates/history-schema.yaml` (JSON Schema outputted as YAML format). Outputs via `--get-schema` command. Used for validation.

- **Session Context**: Unique identifier (UUID format) scoping report/history files to single workflow run. Passed to agents via environment variable (FLUXID_SESSION_ID). Determines file paths in format `<OS temp folder>/fluxid/{report,history}-<session-id>.yaml`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: External agents can obtain report file path and write valid reports
- **SC-002**: Report validation detects and reports schema violations with specific field-level errors in 100% of invalid cases
- **SC-003**: History recording by external agents succeeds without fluxid-provided write commands
- **SC-004**: Workflow loop progression continues to work correctly (PASS status exits, FAIL status retries) with file-based approach
- **SC-005**: Zero IPC-related code remains in codebase after refactor (verified by grep for "ipc" in command/storage layers)
- **SC-006**: All existing E2E tests for removed IPC functionality are deleted from repository
- **SC-007**: Schema retrieval commands (`--get-schema`) output valid, parseable YAML schema in 100% of cases
- **SC-008**: File existence guarantee (`--get-file` creates file if missing) succeeds in 100% of cases
- **SC-009**: Validation commands provide actionable error messages with specific field names and expected values (format: "[field_path]: [violation] (expected: [constraint], got: [value])")
- **SC-010**: Successful command operations produce no output to stderr (silent success except for data output to stdout)

## Assumptions *(mandatory)*

- **A-001**: External agents (Claude, OpenCode, etc.) are capable of file I/O operations to write YAML
- **A-002**: External agents can parse YAML schema to understand report/history structure
- **A-003**: Session ID is available via environment variable (FLUXID_SESSION_ID) during agent execution
- **A-004**: File system supports standard POSIX operations (create directory, write file, read file)
- **A-005**: Workflow execution is 100% sequential - agent writes report/history and exits, then fluxid reads files. No concurrent file access between agent and fluxid is possible, eliminating need for file locking or atomic write patterns.
- **A-006**: Report and history files under 10MB can be read entirely into memory; files exceeding 10MB will have oldest entries truncated (FIFO eviction) before reading
- **A-007**: External agents start, execute once, and exit with report status; fluxid workflow controller handles retry logic by restarting agents when report status is FAIL (limited by retry count configuration)
- **A-008**: YAML schema format is sufficient for schema documentation (agents can parse YAML schema)
- **A-009**: Validation is performed on demand via CLI command rather than automatically on write (since fluxid doesn't write)
- **A-010**: Agents obtain file paths exclusively from `--get-file` commands; agents do not specify or construct paths themselves

## Dependencies *(include if relevant)*

- **D-001**: Go YAML parser library (e.g., `gopkg.in/yaml.v3`) for reading report/history files during workflow execution; must support disabling anchors/aliases/merge keys for security
- **D-002**: JSON Schema validator library `github.com/xeipuuv/gojsonschema` for implementing `--validate` commands. Validates YAML files against JSON Schema (draft-07) documents. Ensures 100% schema compatibility including additionalProperties:false enforcement
- **D-003**: Go standard library `os` package for file path operations and directory creation
- **D-004**: Existing workflow execution logic remains unchanged (dependencies preserved)

## Migration Path *(include if breaking change)*

This is a **breaking change** that removes existing functionality. Migration requires:

### For External Agent Developers

1. **Remove** all calls to `fluxid ipc write-report` and `fluxid ipc read-report`
2. **Replace** with:
   - Call `fluxid report --get-file` to obtain report file path
   - Write report file directly to that path using file I/O
   - Optionally call `fluxid report --validate` before fluxid reads it
3. **Remove** all calls to `fluxid ipc write-history` and `fluxid ipc view-history`
4. **Replace** with:
   - Call `fluxid history --get-file` to obtain history file path
   - Append history entries directly to that file using file I/O
   - Follow history schema for entry structure
5. **Update** schema discovery from `fluxid ipc get-report-schema` to `fluxid report --get-schema`

### For fluxid Maintainers

1. **Phase 1 - Add New Commands** (before removal):
   - Implement `fluxid report --get-schema/--get-file/--validate`
   - Implement `fluxid history --get-schema/--get-file/--validate`
   - Create report YAML schema at `internal/assets/templates/report-schema.yaml`
   - Create history YAML schema at `internal/assets/templates/history-schema.yaml`
   - Embed both schemas in binary
   - Update workflow to read report/history files directly (remove IPC calls)

2. **Phase 2 - Remove Old Commands** (after Phase 1 complete):
   - Delete `internal/command/ipc.go`
   - Delete `internal/command/ipc_handlers.go`
   - Delete `internal/command/ipc_abort.go`
   - Delete `internal/command/ipc_history.go`
   - Delete entire `internal/ipc/` directory (all files including storage.go, schema.yaml)
   - Delete E2E tests for IPC functionality if present in main branch (verify actual file count before removal; feature branch may not have IPC tests yet)
   - **Total removal**: 4 command files + 1 directory (internal/ipc) + any existing E2E test files for deprecated IPC commands

3. **Phase 3 - Update Documentation**:
   - Update agent integration guides
   - Update workflow documentation
   - Add migration examples
   - Document new file-based interface

### Backwards Compatibility

- **None**: This is a hard breaking change
- Old `fluxid ipc` commands will return "command not found" error
- Agents using old interface will fail immediately and visibly
- No gradual migration period - clean cut preferred for simplicity

## Out of Scope *(include if relevant)*

- **Report/history write operations**: Explicitly delegated to external systems. Fluxid only provides schema, validation, and read access.
- **Concurrency control for file writes**: External agents must handle concurrent writes if multiple agents run simultaneously. Fluxid only reads.
- **Abort mechanism changes**: Process termination via Ctrl+C (SIGINT/SIGTERM) must remain functional for instant user abort. Signal handling and graceful shutdown are preserved from existing implementation. This refactor only changes report/history interface, not abort behavior.
- **Report/history file cleanup**: Session lifecycle management (when to delete old files) is separate concern.
- **Encryption or access control for report/history files**: Standard file system permissions sufficient.
- **Versioning of schemas**: Initial implementation provides current schema only. Schema evolution strategy is future work.
- **Alternative output formats**: Commands output YAML for reports/history and YAML schema for schemas. No XML, JSON data output, or other formats.
