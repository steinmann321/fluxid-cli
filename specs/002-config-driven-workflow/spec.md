# Feature Specification: Config-Driven Workflow System

**Feature Branch**: `002-config-driven-workflow`
**Created**: 2026-01-18
**Status**: Draft
**Input**: User description: "the current workflow needs to be reworked to a pluggable system. currently 3 steps are defined. but all 3 steps do the same, use the same config and are only different in relation to their individual settings. only exception is a mandatory 'review' step that does a final exit assessment and is the only valid exit gate out of the loop besides implement retry exhaustion. i need a new workflow system purely based on the config yaml: a config section configures a workflow of 1..n steps (name, command, iterations, order) and creates a sequential order of steps finished by a mandatory (configurable) review step finally. the workflow controller (fluxid) always checks the report status (PASS?FAIL?) as usual and the loop limit"

## Clarifications

### Session 2026-01-18

- Q: When a workflow step's command file path is invalid or doesn't exist during startup validation (FR-015), how should the system respond? → A: Fail startup immediately with a clear error message listing all invalid paths; prevent workflow execution
- Q: When duplicate workflow step names are detected during startup validation (A-006 states names must be unique), how should the system respond? → A: Fail startup immediately with error listing all duplicate step names
- Q: When a workflow configuration has zero custom steps (only the mandatory review step would execute), how should the system respond? → A: Fail startup with error stating at least one custom workflow step is required before review
- Q: The current implementation uses a single report.yaml file that each phase overwrites. Should the new config-driven workflow maintain this single-report behavior or create separate per-step reports for history tracking? → A: Keep single report.yaml that each step overwrites (matches current behavior)
- Q: The current implementation has asymmetric retry behavior: implement has 3 retries, commit has 100 retries, but review has ZERO retry logic (always exactly 1 attempt per iteration). Should the config-driven review step support configurable retries like other steps? → A: Review step can have configurable retries like other steps, but defaults to 1
- Q: FR-015 states command file paths must be validated at startup and fail if invalid. However, the current implementation allows missing command files and falls back to built-in prompts. Should the config-driven system maintain this fallback behavior or strictly require all command files to exist? → A: Strictly require all command file paths to exist (fail startup if missing)
- Q: After a workflow step exhausts all its configured retries with final status FAIL, should the workflow continue to the next step, skip to review, or abort entirely? → A: After retry exhaustion with FAIL, continue to next step (current behavior)
- Q: Should the system provide backward compatibility by auto-generating a default workflow when no workflow section is present in config.yaml, or require users to explicitly define workflow steps? → A: No backward compatibility - workflow section is required; users must explicitly configure workflow steps including the mandatory review step
- Q: Should each workflow step have independent retry logic within its own execution block, or should there be a single generic step execution function that handles retries for any workflow step? → A: Single generic step execution function that handles retries for any workflow step (DRY principle)
- Q: How should the mandatory review step be configured in the workflow YAML structure? → A: Review step is mandatory in config.yaml under workflow.review section; has default values for command and retries if not specified; fully customizable by users
- Q: When max_iterations (development cycles) is set to 0 or a negative number in config.yaml, how should the system respond? → A: 0 means infinite loops allowed
- Q: When a workflow step name is empty or contains only whitespace, how should the system respond? → A: Not allowed; fail at startup
- Q: When a workflow step references a command file that exists but has syntax errors (e.g., malformed YAML/Markdown metadata), how should the system respond? → A: Fail at startup validation with syntax error details before any workflow execution begins
- Q: FR-015 now requires syntax validation of command files at startup. What specific syntax elements should be validated before workflow execution? → A: No check its a text file
- Q: How should the system handle retry count values that are unreasonably large (e.g., 1000000)? → A: Accept any positive integer value without upper limits; trust users to configure appropriately
- Q: What should the system log/output during workflow step execution? → A: Detailed logging: log step transitions, retry attempts, report status checks, command file paths being executed, and iteration cycle boundaries
- Q: Given that steps execute in array order, how should the system handle circular dependencies or infinite loops in step configurations? → A: Not applicable - steps execute in sequential array order with no dependency specification, making circular dependencies structurally impossible
- Q: When the system logs detailed workflow execution information (FR-020: step transitions, retry attempts, status checks, command file paths, iteration boundaries), what should the logging output format and destination be? → A: use default human readable to stdout, if --out=json deliver json
- Q: When executing a workflow step's command file (invoking the coding agent), the system currently checks the report.yaml for PASS/FAIL status. However, what should happen if the agent invocation itself fails catastrophically (e.g., command execution error, agent process crash, unable to write report.yaml)? → A: Treat as FAIL status and retry according to step's retry limit (transient errors get retried, permanent failures eventually continue to next step)
- Q: When executing a workflow step's command file (agent invocation), should there be a maximum execution time limit (timeout) to prevent workflows from hanging indefinitely if an agent becomes unresponsive? → A: No timeout enforcement - agents can run indefinitely until completion or user interruption
- Q: Command file paths require validation at startup (FR-015). Should the system support only absolute paths, or also relative paths? If relative paths are supported, what should be the resolution base directory? → A: Support both absolute and relative paths; relative paths resolved from config.yaml directory
- Q: FR-012 requires replacing the hardcoded 3-step workflow with the config-driven system. How should this integrate with existing fluxid CLI commands (e.g., `fluxid run`)? Should there be new commands, changed parameters, or maintained command interface? → A: Existing command interface remains unchanged (e.g., `fluxid run`); internal implementation fully switches from hardcoded to config-driven workflow; no backward compatibility for missing workflow config
- Q: FR-017 states that if the review step command path is not specified, the system uses "a default review command file". What is the location and path of this default file? → A: Review command is mandatory; no default fallback - exit with error if not specified or not readable
- Q: What are the exact YAML field names and nesting structure for workflow configuration? → A: workflow: {steps: [{name, command, retries}], review: {command, retries}} - flat structure with simple field names
- Q: FR-002 specifies that the `retries` field is optional with default 1. What is the behavior when `retries` is explicitly set to 0? → A: retries=0 means infinite retry attempts until PASS status is achieved (consistent with max_iterations=0 behavior)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Custom Workflow Steps (Priority: P1)

A developer wants to define a custom workflow with multiple steps (e.g., "design", "implement", "test") that execute sequentially, where each step can have its own command file and retry limit.

**Why this priority**: This is the core functionality that enables workflow customization. Without this, users cannot configure their own workflows, which is the primary goal of this feature.

**Independent Test**: Can be fully tested by configuring a workflow with 2-3 custom steps in the YAML config, running the workflow, and verifying that each step executes in order with its own command file. Delivers immediate value by allowing workflow customization.

**Acceptance Scenarios**:

1. **Given** a config.yaml with a workflow section defining steps [{"name": "design", "command": "path/to/design.md", "retries": 2}, {"name": "implement", "command": "path/to/implement.md", "retries": 3}], **When** the user runs fluxid, **Then** the workflow executes "design" step first (up to 2 retries on FAIL), followed by "implement" step (up to 3 retries on FAIL)
2. **Given** a workflow configuration with 4 custom steps, **When** step 2 completes with PASS status, **Then** the workflow immediately proceeds to step 3 without retrying step 2
3. **Given** a workflow with a step that exhausts all its retries with FAIL status, **When** all retries are exhausted, **Then** the workflow continues to the next step (does not abort the entire workflow)

---

### User Story 2 - Mandatory Review Exit Gate (Priority: P1)

A developer needs the review step to always execute as the final step in each development iteration, serving as the only valid exit gate where PASS status allows workflow completion.

**Why this priority**: This is critical for maintaining workflow integrity and ensuring proper validation before completion. Without this, workflows could exit prematurely without proper review.

**Independent Test**: Can be tested by configuring any workflow (with or without custom steps), running it to completion, and verifying that the review step always executes last and that PASS status exits the workflow while FAIL status triggers the next development iteration (until max iterations reached).

**Acceptance Scenarios**:

1. **Given** a workflow with 3 custom steps and a review step, **When** all custom steps complete and the review step returns PASS status, **Then** the workflow exits successfully with exit code 0
2. **Given** a workflow where review step returns FAIL status and development iterations remain, **When** review completes, **Then** the workflow starts a new development iteration from the first step
3. **Given** a workflow where review step returns FAIL status and max iterations are exhausted, **When** review completes, **Then** the workflow exits with a message indicating all iterations exhausted
4. **Given** any workflow configuration, **When** the workflow runs, **Then** the review step cannot be removed or skipped - it must always be present as the final step

---

### User Story 3 - Flexible Step Configuration (Priority: P3)

A developer wants to configure workflow steps with minimal required fields (name and command) while optionally specifying retries and other step-specific settings.

**Why this priority**: This improves usability by making step configuration simple while allowing advanced customization when needed.

**Independent Test**: Can be tested by creating workflow configs with varying levels of detail (minimal vs. fully specified) and verifying that defaults are applied correctly for omitted fields.

**Acceptance Scenarios**:

1. **Given** a workflow step definition with only "name" and "command" fields, **When** the workflow runs, **Then** the step uses default values for retries (1) and other settings
2. **Given** a workflow step with "retries: 5" specified, **When** the step executes and fails, **Then** the step retries up to 5 times before proceeding to the next step
3. **Given** a workflow step with "retries: 0" specified, **When** the step executes and fails, **Then** the step retries infinitely until PASS status is achieved (never proceeds to next step while status is FAIL)
4. **Given** a workflow with steps in array order [step1, step2, step3], **When** the workflow runs, **Then** steps execute in that exact order

---

### Edge Cases

- **Missing workflow section in config.yaml**: System fails at startup with clear error message instructing user to configure workflow steps
- **Missing workflow.review section**: System fails at startup with clear error message stating review section is mandatory
- **Missing review command path in workflow.review section**: System fails at startup with clear error message stating review command is mandatory
- **Invalid command file paths**: System fails at startup with clear error message listing all invalid paths; workflow execution is prevented (no fallback to built-in prompts)
- **Duplicate workflow step names**: System fails at startup with error listing all duplicate step names
- **Zero custom steps (only review)**: System fails at startup with error stating at least one custom workflow step is required before review
- **Review command file path specified but doesn't exist or not readable**: System fails at startup (same strict validation as custom steps)
- **Empty or whitespace-only workflow step names**: System fails at startup with error "step name cannot be empty or whitespace-only"
- **Command file exists but is not a readable text file**: System fails at startup with error showing file path; no syntax validation of file content is performed (only verify it's a text file)
- **Circular dependencies in step configurations**: Not applicable - steps execute in sequential array order with no dependency specification, making circular dependencies structurally impossible
- **max_iterations set to 0**: System allows infinite iterations (workflow runs until review returns PASS)
- **max_iterations set to negative number**: System fails at startup with error "max_iterations cannot be negative"
- **Step retries set to 0**: System allows infinite retry attempts for that step until PASS status is achieved (consistent with max_iterations=0 behavior)
- **Step retries set to negative number**: System fails at startup with error "retries cannot be negative"
- **Extremely large retry counts**: System accepts any positive integer value without upper limits (e.g., 1000000 is valid); users are trusted to configure appropriately
- **Agent invocation failures**: When agent command execution fails (process crash, unable to write report.yaml, command error), system treats as FAIL status and retries according to step's retry limit; does not abort workflow

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support a `workflow` section in config.yaml that defines an ordered list of workflow steps; the workflow section uses exact field names: `steps` (array) and `review` (object)
- **FR-002**: System MUST allow each workflow step to specify exact field names: `name` (required, string), `command` (required, string - file path), `retries` (optional, integer, default: 1); when retries is set to 0, the step retries infinitely until PASS status is achieved (consistent with max_iterations=0 behavior)
- **FR-003**: System MUST execute workflow steps in the order they are defined in the configuration array
- **FR-004**: System MUST append a mandatory review step as the final step in every workflow, regardless of custom step configuration
- **FR-005**: System MUST check report status (PASS/FAIL) after each step execution and respect retry limits
- **FR-006**: System MUST continue to the next step when a step exhausts all retries with FAIL status (not abort workflow); this allows all steps in an iteration to execute and collect diagnostics regardless of individual step failures; this only applies to steps with retries > 0 (steps with retries=0 retry infinitely until PASS)
- **FR-007**: System MUST exit the workflow successfully when the review step returns PASS status
- **FR-008**: System MUST start a new development iteration (from first step) when the review step returns FAIL status and iterations remain
- **FR-009**: System MUST exit the workflow when max development iterations (review cycles) are exhausted; if max_iterations is set to 0, system MUST allow infinite iterations (workflow continues until review returns PASS)
- **FR-019**: System MUST treat negative max_iterations values as errors and fail at startup with clear error message
- **FR-023**: System MUST treat negative retries values as errors and fail at startup with clear error message; only non-negative integers (0 or positive) are valid retry values
- **FR-010**: System MUST require a workflow section in config.yaml; if the workflow section is missing, system MUST fail startup with clear error message instructing user to configure workflow steps
- **FR-011**: System MUST validate workflow configuration at startup (before executing any steps) and report clear errors for invalid configurations; validation MUST include checking for duplicate step names and MUST fail immediately with error listing all duplicates if found; validation MUST also ensure at least one custom workflow step exists before the review step and MUST fail with error if zero custom steps are configured; validation MUST also check that all step names are non-empty and contain at least one non-whitespace character
- **FR-012**: System MUST replace the hardcoded 3-step workflow (implement, commit, review) with the config-driven workflow engine; existing command interface (e.g., `fluxid run`) remains unchanged while internal implementation switches to config-driven approach
- **FR-013**: System MUST allow workflow steps to have unique retry limits independent of other steps; retry count can be any non-negative integer value without upper bounds; retries=0 enables infinite retry attempts until PASS, while positive values limit retry attempts to that number
- **FR-014**: System MUST support 1 to N custom workflow steps (where N is at least 10)
- **FR-015**: System MUST validate that all command file paths in workflow steps exist and are readable text files before workflow execution; system MUST support both absolute and relative paths, resolving relative paths from the config.yaml directory location; if any path is missing, unreadable, or invalid, system MUST fail startup immediately with a clear error message listing all invalid paths and prevent workflow execution (NOTE: This removes the current fallback to built-in prompts for missing command files); system MUST NOT perform syntax validation of command file content - only verify files exist and are readable as text
- **FR-016**: System MUST use the same report status checking mechanism (PASS/FAIL) for all workflow steps, whether custom or review
- **FR-017**: System MUST require a `workflow.review` section in config.yaml with a mandatory command path field; if command path is not specified or not readable, system MUST fail startup with clear error message; retries field is optional and defaults to 1 if not specified
- **FR-018**: System MUST validate that the `workflow.review` section exists in config.yaml; if missing, system MUST fail startup with clear error message
- **FR-020**: System MUST provide detailed logging during workflow execution including: step transitions (when each step starts/ends), retry attempts (current attempt number and remaining retries), report status checks (PASS/FAIL results), command file paths being executed, and iteration cycle boundaries (when new development iterations begin); logging MUST output to stdout in human-readable plain text format by default
- **FR-021**: System MUST support a `--out=json` flag that changes logging output format from human-readable plain text to structured JSON format (written to stdout); this enables programmatic parsing of workflow execution logs
- **FR-022**: System MUST treat agent invocation failures (command execution errors, agent process crashes, missing report.yaml) as FAIL status and apply the step's retry logic; catastrophic execution failures MUST NOT abort the workflow but instead follow the same retry-then-continue behavior as report-based FAIL status

### Key Entities *(include if feature involves data)*

- **WorkflowStep**: Represents a single step in the workflow execution sequence
  - Attributes: name (string), command_file_path (string), max_retries (integer), order (implicit from array position)
  - Relationships: belongs to a Workflow, executed in sequence with other WorkflowSteps

- **Workflow**: Represents the complete workflow configuration
  - Attributes: steps (array of WorkflowStep), max_iterations (integer, from existing iterations config), review_step (WorkflowStep, mandatory)
  - Relationships: contains 1 to N WorkflowSteps, has exactly 1 review step as final step

- **WorkflowConfiguration**: Represents the workflow section in config.yaml
  - Attributes: steps (array of custom step definitions with fields: name, command, retries), review (mandatory review step section with required command field and optional retries field)
  - Relationships: maps to Workflow entity during runtime initialization
  - YAML structure (exact field names):
    ```yaml
    workflow:
      steps:
        - name: "implement"
          command: "commands/implement.md"
          retries: 3
        - name: "test"
          command: "commands/test.md"
          retries: 2
      review:
        command: "commands/review.md"
        retries: 1
    ```

## Dependencies and Assumptions

### Dependencies

- **D-001**: Existing report validation mechanism (PASS/FAIL status checking) remains available and functional
- **D-002**: Current config.yaml file format and parsing infrastructure supports adding new sections
- **D-003**: Command file execution mechanism (agent invocation) continues to work without changes
- **D-004**: Session management and storage (SessionID, SessionRoot) remains unchanged

### Assumptions

- **A-001**: Default retry count for workflow steps is 1 (reasonable default for most steps)
- **A-002**: Workflow steps execute sequentially in array order with no dependency specification or parallel execution support; this design makes circular dependencies structurally impossible
- **A-003**: Review step supports configurable retries like other workflow steps; defaults to 1 retry for backward compatibility
- **A-004**: Maximum workflow step count of 10 is sufficient for anticipated use cases
- **A-005**: Command file paths support both absolute and relative paths; relative paths are resolved from the config.yaml directory location, making workflow configurations portable across environments
- **A-006**: Workflow step names are case-sensitive and must be unique within a workflow
- **A-007**: The existing "iterations" config value maps to max_iterations (development cycles)
- **A-008**: Single report.yaml file is used for all steps; each step execution overwrites the previous report (consistent with current implementation)
- **A-009**: All workflow steps (custom and review) use a unified generic step execution function with retry logic (DRY principle); this replaces the separate `runImplementPhase` and `runCommitPhaseWithRetry` functions
- **A-010**: Workflow step execution has no timeout enforcement; agents can run indefinitely until completion or user interruption (Ctrl+C); this accommodates complex coding tasks that may require extended execution time

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can define and execute custom workflows with 1 to 10 steps without modifying Go code
- **SC-002**: Workflow configuration changes take effect immediately without recompilation (only config.yaml modification required)
- **SC-003**: Missing workflow section in config.yaml is detected at startup and fails with clear error message instructing user to configure workflow
- **SC-004**: Workflow configuration errors are detected and reported before any step execution begins (validation at startup)
- **SC-005**: Each workflow step can independently retry up to its configured retry limit without affecting other steps
- **SC-006**: Review step successfully acts as the only exit gate - workflows only complete when review returns PASS or max iterations exhausted
