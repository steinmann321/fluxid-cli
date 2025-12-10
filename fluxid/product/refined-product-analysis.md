# Business Understanding: AI Agent Workflow Loop Wrapper with IPC Communication

## Comprehensive Request Summary

This request describes a command-line interface tool called **fluxid** that wraps AI agent CLI tools (Claude, Codex, or OpenCode) to automate iterative development workflows through nested loops of implementation, commitment, and review phases. The tool provides structured data exchange via YAML-based IPC commands for reports and history, while streaming the agent's output directly to users in real-time.

**Key aspects:**

- **Primary purpose**: Wrap AI agent CLI tools (Claude, Codex, or OpenCode) in a configurable nested loop structure (iterations with implement retries and commits), enabling systematic iterative development workflows while maintaining real-time visibility into agent output
- **Target users**: Development teams and individuals who use AI agent CLIs (Claude, Codex, or OpenCode) for automated development tasks and need systematic quality control through repeated implement-commit-review iterations with structured feedback
- **Core capabilities**:
  - **Multi-agent support**: Works with Claude, Codex, or OpenCode agent CLIs selected via explicit flags (`--claude`, `--codex`, `--opencode`)
  - **Nested loop orchestration**: Two-tier loop structure where iterations (default: 20) execute review phases and implement retries (default: 3) execute implement+commit phases
  - **Phase-specific prompts**: Configurable prompts for IMPLEMENT, COMMIT, and REVIEW phases via YAML config files with home/local override support
  - **Real-time streaming**: Direct passthrough of agent's stdin/stdout/stderr for transparent, unmodified output streaming
  - **IPC interface**: YAML-based commands for external processes to write/read reports, write/query history, and retrieve schemas
  - **Session isolation**: Each wrapper invocation creates a unique session ID, preventing data mixing between concurrent or sequential runs
  - **Optional history export**: Flag to write complete session history to disk on completion (for audit/analysis)
  - **Schema-validated communication**: All reports and history entries validated against predefined YAML schemas before storage
- **Expected business outcomes**:
  - Reduced manual overhead in iterative development processes (no manual re-invocation of implement-review iterations)
  - Consistent quality control through systematic review iterations (up to 20 review attempts)
  - Structured feedback mechanism enabling external tools to parse workflow results programmatically
  - Transparent operation through unmodified agent output streaming
  - Flexible workflow customization through config-based prompt management
  - Agent flexibility enabling teams to choose between different AI assistants (Claude, Codex, OpenCode) based on preference or availability
- **Context**: Go CLI tool that enhances AI agent CLI usage (Claude, Codex, or OpenCode) by adding automated loop orchestration and structured IPC communication capabilities

**Scope Clarity:**
- **IN SCOPE**: Multi-agent support (Claude/Codex/OpenCode), loop wrapper, IPC interface for reports/history, session management, command file configuration, agent output streaming
- **OUT OF SCOPE**: Progress tracking (progress.yaml), epic/milestone management, file path conventions, task decomposition - these are handled by external orchestration systems that *call* fluxid

---

## Configuration and Command Structure

### Configuration Files

Fluxid uses a hierarchical configuration system with two-level overrides:

**1. Home Configuration** (`~/.fluxid/config.yaml`):
- User-wide default settings
- Applied to all fluxid invocations for this user
- Lowest precedence level

**2. Project Configuration** (`./.fluxid/config.yaml`):
- Project-specific settings in project root directory
- Overrides home configuration where values overlap
- Enables per-project customization

**3. Command-Line Flags**:
- Highest precedence - always override config files
- Enables per-invocation customization

**Precedence Order** (low → high):
```
Built-in defaults → Home config → Project config → CLI flags
```

### Configuration Schema

The `config.yaml` file structure:

```yaml
# Agent selection (can be overridden by --claude, --codex, --opencode flags)
agent: claude  # Default: claude | codex | opencode

# Loop configuration
implement_retries: 3   # Number of implement retries per iteration (default: 3)
iterations: 20         # Number of iterations (default: 20)

# Phase toggles
commit_enabled: true   # Enable commit phase (default: true)
# Note: Review phase is always enabled and cannot be disabled

# Mandatory: Command file references (relative to .fluxid/commands/)
commands:
  implement: fluxid.implement.md  # REQUIRED
  commit: fluxid.commit.md        # REQUIRED
  review: fluxid.review.md        # REQUIRED
```

**Required Configuration**:
- The three command file references are **mandatory** in config.yaml
- Fluxid will fail startup if any command file is missing or unreadable
- Command files must be markdown files

### Command Files

**Purpose**: Command files are markdown documents that define the prompts given to the AI agent for each workflow phase.

**Location**: `.fluxid/commands/` directory
- Home: `~/.fluxid/commands/`
- Project: `./.fluxid/commands/` (overrides home)

**Overwrite Pattern**:
- If a command file exists in project `.fluxid/commands/`, use it
- Otherwise, fall back to home `~/.fluxid/commands/`
- This enables project-specific command customization

**Three Mandatory Commands**:

1. **Implement Command** (e.g., `fluxid.implement.md`):
   - Defines the role and task for the implementation phase
   - Instructs the agent to write code, update tests, generate reports
   - Must instruct agent to write YAML report using IPC commands
   - Used during implement retries

2. **Commit Command** (e.g., `fluxid.commit.md`):
   - Defines the role and task for the commit phase
   - Instructs the agent to stage changes, commit, handle pre-commit hooks
   - Ensures clean repository state before proceeding
   - Used after each implement phase

3. **Review Command** (e.g., `fluxid.review.md`):
   - Defines the role and task for the review phase
   - Instructs the agent to execute tests, analyze results, diagnose failures
   - Must instruct agent to write YAML report with PASS/FAIL status
   - Used after all implement retries in each iteration

**Command File Structure**:
- Markdown format with structured sections
- Typical sections: Role, Task, Input/Output, Process, Rules
- Full prompt text is passed to the agent
- Files should instruct agent on IPC command usage for reports/history

**Example Command File Resolution**:
```
Config says: commands.implement = "fluxid.implement.md"

Resolution order:
1. Check: ./.fluxid/commands/fluxid.implement.md  (project)
2. Check: ~/.fluxid/commands/fluxid.implement.md  (home)
3. Fail startup if neither exists
```

### Environment Variables

All configuration can be set via environment variables (overridden by CLI flags):

**Agent Selection**:
- `FLUXID_AGENT` - Default agent (claude|codex|opencode)

**Loop Configuration**:
- `FLUXID_IMPLEMENT_RETRIES` - Number of implement retries (default: 3)
- `FLUXID_ITERATIONS` - Number of iterations (default: 20)

**Phase Toggles**:
- `FLUXID_COMMIT_ENABLED` - Enable commit phase (bool, default: true)
# Note: Review phase cannot be disabled

### CLI Flags

All configuration can be overridden per-invocation:

**Agent Selection**:
- `--claude` - Use Claude agent
- `--codex` - Use Codex agent
- `--opencode` - Use OpenCode agent

**Loop Configuration**:
- `--fluxid-implement-retries N` - Set number of implement retries
- `--fluxid-iterations N` - Set number of iterations

**Phase Toggles**:
- `--fluxid-no-commit` - Disable commit phase
# Note: Review phase cannot be disabled

**History Management**:
- `--write-history <message>` - Write message to in-memory session history log

**Dry-Run**:
- `--fluxid-dry-run` - Simulate workflow execution without calling agent

**Output Format**:
- `--fluxid-output {json|yaml}` - Initialization status format

### Configuration Resolution Example

Given:
```yaml
# ~/.fluxid/config.yaml
agent: claude
implement_retries: 5
commands:
  implement: fluxid.implement.md
  commit: fluxid.commit.md
  review: fluxid.review.md

# ./.fluxid/config.yaml
implement_retries: 2
iterations: 15
```

And environment:
```bash
export FLUXID_AGENT=codex
```

And CLI:
```bash
fluxid --opencode --fluxid-iterations 10 [claude-args]
```

**Resolved configuration**:
- Agent: `opencode` (CLI overrides env)
- Implement retries: `2` (project config overrides home config)
- Iterations: `10` (CLI overrides project config)
- Commands: merged from both configs (project takes precedence if overlap)

---

## Business Context

**Business problem being solved:**

Development workflows using AI agents often require repetitive iterations of implementation, validation, and review. Managing these iterations manually is time-consuming, error-prone, and inconsistent. Teams need a way to automate these repetitive patterns while maintaining visibility and control over the process. Additionally, teams using different AI agent tools (Claude, Codex, OpenCode) need a unified workflow wrapper that works consistently across agents without requiring separate tooling for each.

**Opportunity being captured:**

- Enable automated, consistent execution of multi-phase development workflows across different AI agents
- Provide agent-agnostic workflow orchestration, allowing teams to switch agents without changing their automation infrastructure
- Create structured feedback mechanisms that allow external systems to understand workflow state and results regardless of which agent is being used
- Build audit trails of workflow execution for compliance, debugging, and improvement
- Reduce context-switching and manual orchestration overhead for development teams
- Support organizational flexibility where different teams or projects may prefer different AI agents

**Strategic alignment:**

This tool serves as infrastructure for AI-assisted development automation, enabling teams to define and execute consistent workflow patterns regardless of which AI agent they choose. It complements existing development tools by adding orchestration and structured communication capabilities while maintaining agent flexibility.

**User pain points addressed:**

- Manual execution of repetitive implementation-review iterations with AI agents is tedious and time-consuming
- Each AI agent (Claude, Codex, OpenCode) has different CLIs requiring separate wrapper implementations
- Teams locked into specific AI agents due to custom workflow automation tooling
- Lack of structured feedback mechanisms makes it difficult for external tools to understand workflow results from AI agents
- Inconsistent workflow execution leads to quality variations
- Missing audit trails make it hard to understand what happened during automated AI-assisted workflows
- Rigid workflow tools don't adapt to different team needs, process variations, or agent preferences

**Expected business impact:**

- Faster iterations through automation
- Higher quality through consistent review processes
- Better visibility into automated workflows through structured reporting
- Improved compliance through complete history tracking
- Reduced manual effort in workflow orchestration

## Stakeholders

**Primary Users:**

- **Development teams using AI agents**: Need to automate repetitive implementation-review iterations while maintaining quality control and visibility into the automated process, regardless of which AI agent (Claude, Codex, or OpenCode) they choose
- **Individual developers**: Want to leverage automated AI development assistance with systematic quality checks, reducing manual intervention while ensuring consistent results
- **Multi-agent organizations**: Teams that use different AI agents across projects or departments and need unified workflow orchestration infrastructure
- **Automation system integrators**: Need structured data exchange (reports and history) to build workflows that react to execution results, agnostic to which underlying AI agent is being used

**Secondary Stakeholders:**

- **DevOps/Platform teams**: May need to standardize workflow patterns across teams, requiring consistent configuration and execution models
- **Compliance/Audit teams**: May need access to complete execution history for audit trails and compliance verification
- **Tool developers**: May build integrations with this tool, requiring stable data schemas and predictable behavior

## Scope

### Includes

- **Automated workflow orchestration**: Users can execute configurable loops of implementation, commitment, and review phases without manual intervention between phases
- **Structured report management**: External processes can write and read standardized status reports, enabling programmatic decision-making based on workflow results
- **Execution history tracking**: Users can log structured events during workflow execution and query that history for analysis, debugging, or audit purposes
- **Real-time output streaming**: Users see live output from the underlying claude tool as workflow executes, maintaining transparency and interactivity
- **Flexible configuration**: Users can customize workflow behavior through configuration files, environment variables, and command-line flags with clear precedence rules
- **Session-based isolation**: Each workflow execution is isolated with a unique session identifier, preventing data mixing between concurrent or sequential runs
- **Optional history persistence**: Users can optionally export complete session history to disk for long-term storage and analysis
- **Schema-based validation**: All structured data (reports and history) are validated against predefined schemas, ensuring data quality and preventing invalid inputs

### Excludes

- Persistence of session data beyond the lifetime of a single workflow execution (unless explicitly exported via history persistence)
- Built-in implementation of the actual work phases (implementation, commit, review logic) - this is delegated to the "claude" tool
- Graphical user interface or web-based interface
- Distributed execution across multiple machines
- Built-in version control integration beyond what the "claude" tool provides
- Automatic workflow decision-making based on report content (the tool provides data exchange mechanisms but doesn't interpret results)
- Database storage of history or reports
- Multi-user collaboration features within a single session

## User Journeys

### Journey 1: Basic Automated Workflow Execution

**User Goal:** Execute an automated development workflow with systematic review iterations

**Steps:**
1. User runs the wrapper command with desired configuration (e.g., `fluxid --fluxid-implement-retries 5 --fluxid-iterations 10 [claude-args]`)
2. System generates a unique session identifier and displays initialization status showing configuration, loop counts, and session details
3. System executes nested loops: for each iteration, run implement retries of (implement → commit), then run review
4. User sees real-time output from each phase as it executes, streaming directly to their terminal
5. System completes all loops and exits
6. Outcome: User has completed multiple implementation-review iterations automatically without manual intervention

**Success Scenario:** All phases complete successfully, user sees complete output, workflow ends normally

**Alternative Scenarios:**
- If any claude execution fails (exits non-zero), system aborts immediately and exits with the same error code, preserving failure visibility
- If user disables commit or review phases via flags, those phases are skipped in the loop structure
- If configuration files are missing, system uses built-in defaults

### Journey 2: Structured Report Exchange During Workflow

**User Goal:** External process writes status reports during workflow execution and reads them later for decision-making

**Steps:**
1. Workflow execution is running (session ID is set in environment)
2. External process (e.g., a review script called by claude) needs to report its findings
3. External process retrieves the report schema to understand required data structure
4. External process creates a report following the schema (status, categorized issues, metadata)
5. External process writes the report via IPC command, passing it on standard input
6. System validates the report against schema
7. If valid, report is stored for this session; if invalid, clear error message explains validation failure
8. Later in the workflow, another external process reads the current report to make decisions
9. Outcome: Structured data flows between workflow phases, enabling programmatic decision-making

**Success Scenario:** Report validates successfully, stored for session, later retrieved accurately

**Alternative Scenarios:**
- If report doesn't match schema, validation fails with detailed error message, no report is stored
- If no report has been written yet, read operation returns cleanly indicating no report exists (not an error)
- If IPC command fails, wrapper continues executing (doesn't abort) but prints error information

### Journey 3: History Tracking During Execution

**User Goal:** Log important events during workflow execution for real-time visibility and debugging

**Steps:**
1. User starts workflow (`fluxid [args]`)
2. System generates session ID and displays initialization status
3. During workflow execution, agents log events by calling `fluxid ipc write-history` with messages
4. Each history entry is stored in-memory with automatic ISO 8601 timestamp
5. Agents or external processes can query history using `fluxid ipc view-history` to see recent events
6. Workflow completes (either successfully or due to error)
7. Session history is automatically cleared from memory
8. Outcome: Runtime visibility into workflow events, useful for understanding agent decisions and debugging

**Success Scenario:** History entries are logged throughout execution, queryable during runtime, cleared automatically at session end

**Alternative Scenarios:**
- If history exceeds 32MB size limit, oldest entries are automatically dropped (FIFO)
- If no history entries were logged, view-history returns empty list
- If wrapper aborts due to agent failure, history is still accessible via IPC until process exits
- User can also write history entries from CLI using `fluxid --write-history "message"` format

### Journey 4: Graceful Abort During Execution

**User Goal:** Stop a runaway workflow that is infinitely retrying due to misconfigured prompts

**Steps:**
1. Workflow is running and stuck in infinite retry loop (missing/invalid reports)
2. User observes repeated validation errors in output
3. **Option A (Interactive):** User presses Ctrl+C to send SIGINT signal
4. **Option B (Programmatic):** External monitoring process calls `fluxid ipc abort` to programmatically trigger abort
5. System catches abort signal/command and sets abort flag for session
6. System allows current agent invocation to complete naturally
7. System performs graceful shutdown (clears partial reports, cleans up session state)
8. System exits with appropriate error code
9. Outcome: Workflow stopped cleanly without leaving corrupt session state

**Success Scenario:** Abort signal received, current phase completes, workflow exits gracefully

**Alternative Scenarios:**
- If abort occurs between phases (not during agent invocation), exit immediately
- If multiple abort signals received, force immediate exit without waiting for agent
- If abort via IPC specifies wrong session ID, clear error message returned

### Journey 5: Configuration-Driven Workflow Customization

**User Goal:** Customize workflow behavior to match team processes without modifying the tool

**Steps:**
1. User creates configuration file in home directory (`~/.fluxid/config.yaml`) defining prompts for implement, commit, and review phases
2. User optionally creates local override configuration in project directory (`./.fluxid/config.yaml`) for project-specific customization
3. User sets environment variables for loop counts and phase toggles (e.g., `FLUXID_IMPLEMENT_RETRIES=2`, `FLUXID_COMMIT_ENABLED=false`)
4. User runs wrapper with command-line flags that override environment and config (e.g., `--fluxid-no-commit`)
5. System resolves effective configuration using precedence: defaults → environment → CLI flags
6. System displays initialization status showing all resolved configuration values
7. Workflow executes using customized prompts and loop structure
8. Outcome: Workflow behavior matches team processes without code changes

**Success Scenario:** Configuration is resolved correctly with proper precedence, workflow uses customized settings

**Alternative Scenarios:**
- If configuration files are missing, system uses built-in defaults and continues
- If prompt configuration points to files, system loads file contents as prompts
- If user provides claude prompt flags (`-p` or `--prompt`), system strips them and uses configured prompts instead
- User can request initialization status as JSON or YAML for programmatic parsing

## Interaction Patterns

### Pattern 1: Wrapper Mode (Default Execution)

**Description:** Users run the tool in wrapper mode where it orchestrates workflow loops and manages claude execution

**User Experience:** User invokes the tool with a mix of fluxid-specific flags and claude arguments. The tool handles all workflow orchestration transparently, streaming output directly to the terminal. User experiences seamless execution of multiple workflow phases without manual intervention.

**Business Rationale:** Provides simple, command-line native experience that integrates naturally with existing development workflows. Users maintain full visibility through streaming output while gaining automation benefits.

### Pattern 2: IPC Mode (Data Exchange)

**Description:** External processes use IPC subcommands to exchange structured data with the running workflow session

**User Experience:** External processes (typically called by claude during workflow phases) invoke IPC commands to write reports, log history, retrieve schemas, and query data. Session context is automatically inherited from environment, requiring no manual session management.

**Business Rationale:** Enables structured communication between workflow phases without requiring complex integration mechanisms. Schema-based validation ensures data quality and prevents integration errors.

### Pattern 3: Session-Based Isolation

**Description:** Each workflow execution operates in an isolated session identified by a unique session ID

**User Experience:** System automatically generates a session ID on startup and propagates it to all child processes via environment variable. External processes automatically operate on the correct session without manual coordination.

**Business Rationale:** Prevents data contamination between concurrent or sequential workflow runs, enabling safe parallel execution and clear separation of concerns. Simplifies troubleshooting by clearly delineating execution boundaries.

### Pattern 4: Streaming Output with Structured Data

**Description:** Real-time output flows directly to user terminal while structured data flows through IPC channels

**User Experience:** Users see live output from all workflow phases in their terminal, maintaining transparency and enabling real-time monitoring. Simultaneously, structured data (reports and history) flows through separate IPC channels for programmatic consumption.

**Business Rationale:** Balances human and machine needs - users get visibility and interactivity, while automation systems get structured, parseable data. Avoids forcing choice between human-friendly and machine-friendly output.

## Business Rules

### 1. Rule Category: Configuration Precedence

- **Built-in defaults are overridden by environment variables**: Ensures users can set persistent preferences across invocations without command-line clutter
- **Environment variables are overridden by CLI flags**: Allows per-invocation customization without changing persistent configuration
- **Home config is overridden by local config where they overlap**: Enables project-specific customization while maintaining user-wide defaults
- **User-provided claude prompts are always stripped in wrapper mode**: Ensures consistent workflow behavior - prompts must be controlled by fluxid configuration to maintain workflow integrity

### 2. Rule Category: Session Lifecycle and Isolation

- **Each wrapper invocation creates exactly one session**: Establishes clear execution boundaries for data isolation and audit trails
- **Session ID is automatically propagated to child processes via environment**: Eliminates manual coordination overhead and potential for session mismatch errors
- **Session data exists only for wrapper process lifetime (unless explicitly exported)**: Prevents unlimited storage growth and clarifies data lifecycle expectations
- **IPC commands without explicit session use environment session ID**: Simplifies usage within workflow context while allowing explicit override when needed

### 3. Rule Category: Error Handling and Resilience

- **Agent process failure aborts wrapper immediately with same exit code**: Preserves error visibility and prevents cascading failures from hidden errors
- **IPC command failure does not abort wrapper**: Treats data exchange failures as non-critical, allowing workflow to continue even if reporting fails. Clear error messages printed to stderr.
- **Missing or invalid report triggers infinite retry of that phase**: Reports are mandatory for workflow control. If implement or review phase produces invalid/missing report, that exact phase is re-executed infinitely until a valid report is produced. Loop counters only increment when valid PASS/FAIL report received.
- **Report validation failure prints clear diagnostic**: When validation fails, error message must explain what failed, why it failed, and how to fix it (schema mismatch details, missing fields, invalid values, etc.)
- **Graceful abort via signal or IPC**: Users can stop infinite retries via Ctrl+C (SIGINT/SIGTERM) for interactive use, or `fluxid ipc abort` for programmatic control. Abort allows current agent invocation to complete, then exits cleanly.

### 4. Rule Category: Data Validation and Schema Compliance

- **All reports must validate against report schema**: Guarantees consistent data structure for consumers, preventing integration errors. Invalid reports trigger infinite phase retry until valid report produced.
- **History entries use simple log format with timestamps**: Format is `[ISO8601 timestamp] user message`. Plain text, one entry per line, no YAML structure required.
- **History timestamp is managed exclusively by system, never by external input**: Prevents timestamp manipulation and ensures accurate chronological ordering
- **History size limited to 32MB per session**: When limit reached, oldest entries are dropped (FIFO) to prevent unbounded memory growth
- **Missing report is unacceptable state**: Unlike empty history (which is valid), missing reports after a phase execution trigger infinite retry of that phase

### 5. Rule Category: Loop and Phase Control

- **Implement retries execute implement→commit sequences**: Creates checkpoints through commit phase after each implementation attempt. Retry counter increments only when valid PASS/FAIL report received from implement phase.
- **Iterations execute review after all implement retries**: Enables periodic comprehensive review after multiple implementation retries. Iteration counter increments only when valid PASS/FAIL report received from review phase.
- **Review phase cannot be disabled**: Only commit phase can be optionally disabled via `--fluxid-no-commit`. Implement and review are mandatory workflow components.
- **Loop counts must be positive integers (≥1)**: Zero or negative values rejected at startup with validation error. Prevents configuration errors.
- **PASS status breaks loops, FAIL continues**: Implement retry breaks on implement PASS (proceeds to review). Iteration exits on review PASS (workflow complete). FAIL status continues to next retry.
- **Invalid/missing reports trigger infinite phase retry**: Phase is re-executed without incrementing loop counter until valid report produced. Only escape via abort mechanism (Ctrl+C or IPC abort).
- **Default loop counts balance retry opportunity with execution time**: Provides reasonable defaults (implement retries: 3, iterations: 20) that work for common scenarios while remaining configurable

## Information Architecture

**Core Entities:**

- **Session**: Represents a single workflow execution from start to finish
  - Attributes: unique identifier, creation time, configuration settings, associated report and history entries
  - Lifecycle: Created on wrapper startup, exists during wrapper process lifetime, optionally exported on completion

- **Report**: Represents the current status and findings from workflow phases
  - Attributes: command that generated it, artifact being validated, timestamp, status (PASS/FAIL), categorized issues (blockers, defects, concerns, observations, enhancements), optional summary and next steps
  - Lifecycle: Written by external processes during workflow phases, replaces any previous report for the session, read by subsequent phases for decision-making

- **History Entry**: Represents a single logged event during workflow execution
  - Attributes: user-provided message, system-generated ISO 8601 timestamp
  - Format: Plain text log entry `[timestamp] message`
  - Lifecycle: Appended to in-memory session history when written, never modified, queryable during workflow execution, cleared automatically when session ends. Oldest entries dropped if total size exceeds 32MB.

- **Configuration**: Represents workflow behavior settings
  - Attributes: loop counts (inner and outer), phase toggles (commit enabled, review enabled), prompts (implement, commit, review)
  - Sources: Built-in defaults, home config file, local config file, environment variables, CLI flags

- **Phase Execution**: Represents a single invocation of implement, commit, or review
  - Attributes: phase type, associated prompt, claude arguments, exit status, output (streamed, not stored)
  - Lifecycle: Created and executed within loop structure, exit status determines workflow continuation

**Relationships:**

- Session contains zero or one Report (one-to-zero-or-one)
- Session contains zero or more History Entries (one-to-many)
- Session uses one Configuration (one-to-one)
- Session executes multiple Phase Executions (one-to-many)
- Report categorizes Issues into multiple severity levels (one-to-many)
- Phase Execution produces output that may trigger Report or History Entry creation (one-to-zero-or-many)

## Success Indicators

**User Success:**

- Users can execute multi-phase workflows without manual intervention between phases
- Users can customize workflow behavior to match their team processes through configuration
- Users receive real-time visibility into workflow execution through streaming output
- External processes can reliably exchange structured data with workflows using IPC commands
- Users can export complete workflow history for analysis and audit purposes

**Business Success:**

- Reduced time spent on manual workflow orchestration
- Increased consistency in development workflow execution across teams
- Improved visibility into automated workflow results through structured reporting
- Better compliance posture through comprehensive audit trails
- Higher adoption of automated development workflows due to flexibility and transparency

**Quality Indicators:**

- All structured data (reports and history) validates against schemas before acceptance
- Schema validation errors provide clear, actionable error messages for correction
- Workflow failures are immediately visible through exit codes and error output
- Session isolation prevents data mixing between concurrent or sequential runs
- Configuration precedence is predictable and well-documented

## Assumptions

1. **One of the supported agent CLIs (Claude, Codex, or OpenCode) is already available and functional in the user's environment** - fluxid wraps these agents but doesn't install or manage them. Users must have at least one agent installed and accessible via PATH.
2. **External processes that use IPC commands are invoked by the agent (Claude/Codex/OpenCode) during workflow phases** - the IPC mechanism is designed for this calling pattern where agents call `fluxid ipc` commands. Agents can also be used by external orchestration systems that inject instructions via IPC.
3. ~~**Users have write access to the directory specified for history export**~~ - CLARIFIED: History is in-memory only, no file persistence. No file system access needed for history.
4. ~~**Session IDs generated by the tool are sufficiently unique to prevent collisions**~~ - CLARIFIED: Using UUID v4 which guarantees collision-free identifiers.
5. **The report schema provided is complete and correct** - schema validation depends on schema accuracy. History no longer uses YAML schema (plain text log format).
6. **Users understand the precedence order for configuration sources** - documentation will need to clearly explain: defaults → home config → project config → env vars → CLI flags
7. **Streaming output to terminal is acceptable for all workflow phases** - no requirement for output filtering or redirection within the tool
8. ~~**Sessions only need to exist during wrapper process lifetime**~~ - CLARIFIED: Confirmed, sessions are in-memory only with no persistence requirement.
9. **IPC commands are invoked with correct schema-compliant data for reports** - while validation catches errors, callers are expected to attempt compliance. History uses simple string messages (no schema).
10. **The tool will be distributed via standard package managers (brew/choco)** - making it available on PATH as described
11. **NEW: Users understand that infinite retry behavior requires manual intervention via Ctrl+C or IPC abort** - misconfigured prompts that never produce valid reports will retry forever unless user aborts
12. **NEW: Command files crafted by users will correctly instruct agents to call IPC commands** - fluxid depends on command files telling agents when and how to use `fluxid ipc write-report` and `fluxid ipc write-history`
13. **NEW: Agents will complete execution within reasonable time when infinite retry occurs** - each agent invocation is assumed to eventually finish (not hang indefinitely), allowing abort signals to be checked between invocations

## Constraints

**Business Constraints:**

- Must maintain backwards compatibility with existing "claude" CLI arguments and behavior (except for prompt override in wrapper mode)
- Must not require database or external service dependencies for core functionality
- Configuration must be file-based and human-readable for transparency and version control compatibility

**User Constraints:**

- Users must have at least one supported agent CLI (Claude, Codex, or OpenCode) available in their environment (either on PATH or specified via configuration)
- ~~Users must have write access to directories where history export is requested~~ - REMOVED: History is in-memory only
- Users running IPC commands must have FLUXID_SESSION_ID in environment or provide explicit --session flag
- Configuration files must be valid YAML and conform to expected structure
- Command files must be valid markdown and instruct agents to use IPC commands for reports
- Loop count configuration values must be positive integers (≥1)

**External Constraints:**

- Must work across different operating systems (implied by brew/choco distribution)
- Must respect standard CLI conventions (exit codes, stdin/stdout/stderr, signal handling including SIGINT/SIGTERM)
- Must not modify or interfere with agent CLI functionality beyond wrapping its execution
- Report schema is externally defined and must be used exactly as specified
- History format is plain text log with ISO 8601 timestamps (no external schema dependency)

## Validation Checklist

### Clarified Decisions

- ✅ **Multi-agent support is required**: Fluxid must support Claude, Codex, and OpenCode agent CLIs with explicit selection flags: `--claude`, `--codex`, `--opencode`. These flags select which agent to use for workflow execution. Only one agent flag should be specified per invocation.
  - **Rationale**: Production bash implementation demonstrates working multi-agent support with explicit agent flags. User confirmed all three agents must be supported with this flag design.
  - **Impact**:
    - Flag design: `--claude`, `--codex`, `--opencode` (not `--fluxid-agent-bin`)
    - Default agent selection (if no flag specified): likely `--claude` as default
    - Environment variable: `FLUXID_AGENT` with values `claude|codex|opencode`
    - Agents must be available on PATH (third-party dependencies)
    - Validation: Error if multiple agent flags specified simultaneously

- ✅ **Configuration structure is hierarchical with command file pattern**: Fluxid uses a two-level configuration system with home (`~/.fluxid/config.yaml`) and project (`./.fluxid/config.yaml`) configs. All flags/settings can be set via config files, environment variables, or CLI flags (precedence: defaults → home config → project config → env vars → CLI flags).
  - **Rationale**: User clarified complete configuration pattern matching bash implementation. Command files live in `.fluxid/commands/` with same home/project override pattern.
  - **Impact**:
    - Config file structure documented with YAML schema
    - Three mandatory command file references in config: `commands.implement`, `commands.commit`, `commands.review`
    - Command files are markdown documents that serve as agent prompts
    - Command file resolution: check project `.fluxid/commands/` first, fall back to home
    - Startup validation: fail if any mandatory command file is missing
    - Created `fluxid.commit.md` command file based on bash script instruction pattern
    - Environment variables for all config values documented
    - CLI flag precedence clearly defined

- ✅ **Phase toggles**: The `--fluxid-no-review` flag does not exist. Implement and commit phases are mandatory; only commit can be optionally disabled via `--fluxid-no-commit`. Review phase is always enabled.
  - **Rationale**: User clarified that review should not be optional - implement phases are mandatory workflow components.
  - **Impact**:
    - Remove `--fluxid-no-review` flag from design
    - Remove `FLUXID_DISABLE_REVIEW` environment variable
    - Remove `review_enabled` from config.yaml schema
    - Validation: tool should never allow skipping review phase
    - Documentation must clarify that review is mandatory

- ✅ **Initialization status output**: Displays only prompt source paths (file locations), not full prompt content. Shows effective configuration values and session ID/output format.
  - **Rationale**: Prompts can be large; showing paths keeps output clean while users can inspect files separately if needed.
  - **Impact**:
    - Initialization message includes: resolved loop counts, phase toggles, selected agent, command file paths (absolute), session UUID, history handling mode
    - Does NOT include: full prompt text, agent binary resolved path
    - Format: Human-readable text with clear labels

- ✅ **Missing command file handling**: If a configured command file path doesn't exist or can't be read, abort with error at startup.
  - **Rationale**: Prevents execution with incomplete configuration. Clear failure is better than surprising behavior.
  - **Impact**:
    - Startup validation must check all three command files exist and are readable
    - Error message must show which file is missing and expected path
    - Exit with non-zero code before any workflow execution

- ✅ **Session ID format**: Use UUID v4 (random) for session identifiers.
  - **Rationale**: Guarantees collision-free IDs across all invocations, simple to implement, industry standard.
  - **Impact**:
    - Generate UUID v4 on wrapper startup
    - Set FLUXID_SESSION_ID environment variable for child processes
    - Include session ID in initialization status output
    - Use session ID in any logging or error messages

- ✅ **IPC error handling**: IPC commands are for communication between commands within a session and can be used by callers to inject instructions into running process. IPC errors should output clear error messages but not be logged to history automatically.
  - **Rationale**: IPC is a communication mechanism, not a logging mechanism. Errors should be visible but not clutter session history.
  - **Impact**:
    - IPC command failures print clear error to stderr
    - IPC errors do NOT auto-append to history
    - Only user-initiated history writes appear in history log
    - Error messages must be actionable (show what failed, why, how to fix)

- ✅ **History mechanism**: History only exists in-memory during workflow execution and is removed when done. No history files are persisted. The `--write-history` flag parameter is treated as a string message to write to the in-memory history log, not a file path.
  - **Rationale**: History is a session log for runtime visibility, not an audit archive. Simplifies implementation and cleanup.
  - **Impact**:
    - No file I/O for history persistence
    - Session history stored in memory only (map of session ID → history entries)
    - `--write-history <message>` adds entry to current session history
    - History cleared automatically when session ends
    - `fluxid ipc view-history` reads from in-memory store
    - Remove any file export logic for history

- ✅ **History output format**: The `view-history` IPC command includes auto-generated timestamps in output along with user-provided message data.
  - **Rationale**: Timestamps enable chronological analysis and debugging. Complete data visibility for consumers.
  - **Impact**:
    - History entries format: `[timestamp] message`
    - Timestamp format: ISO 8601 (YYYY-MM-DDTHH:MM:SSZ)
    - Both timestamp and message returned by view-history command

- ✅ **Loop count validation**: Reject zero or negative values for implement_retries and iterations. Must be positive integers (≥1).
  - **Rationale**: Prevents configuration errors that make workflow meaningless. Zero loops would skip phases entirely.
  - **Impact**:
    - Startup validation: check implement_retries ≥ 1 and iterations ≥ 1
    - Error message: "Loop counts must be positive integers (got implement_retries=X, iterations=Y)"
    - Applies to all configuration sources (config file, env vars, CLI flags)

- ✅ **History size limits**: Limit total history size to 32MB per session. When limit reached, remove oldest entries (FIFO). No per-entry count limit.
  - **Rationale**: Prevents unbounded memory growth in very long workflows while allowing detailed logging.
  - **Impact**:
    - Track cumulative size of all history entries in session
    - When write-history would exceed 32MB, drop oldest entries until space available
    - No limit on number of entries, only total size
    - Consider UTF-8 byte size, not character count

- ✅ **Resume logic**: No resume logic implemented. Each invocation starts fresh from implement phase regardless of previous state. External orchestration handles resume if needed.
  - **Rationale**: Keeps fluxid simple and focused. Resume is orchestration concern, not wrapper concern.
  - **Impact**:
    - No `--fluxid-resume` flag
    - No checking for existing reports on startup
    - Every invocation begins with iteration 0, retry 0, implement phase
    - External systems can implement resume by checking reports before calling fluxid

- ✅ **Session data persistence**: Keep all session data (reports, history) in memory until wrapper exit. No periodic cleanup during execution. Respect 32MB history limit by dropping oldest entries.
  - **Rationale**: Agents should have complete context throughout workflow. Memory limit prevents runaway growth.
  - **Impact**:
    - No cleanup intervals or periodic clearing
    - Reports remain available for entire session
    - History truncation only happens when size limit reached
    - Session cleanup only occurs on normal exit or abort

- ✅ **History format**: History is a simple logfile with datetime stamps. Format: `[ISO8601 timestamp] user message`. Plain text, one entry per line.
  - **Rationale**: Simple, human-readable, easy to parse. No complex structure needed for session logs.
  - **Impact**:
    - Not YAML format, not markdown - plain text log format
    - Each entry: `[2025-12-10T14:30:52Z] Implemented feature X with approach Y`
    - System generates timestamp, user provides message
    - view-history returns log entries in chronological order

- ✅ **Dry-run mode**: Implement `--fluxid-dry-run` flag that simulates workflow execution without calling actual agent.
  - **Rationale**: Useful for testing configuration, loop logic, and IPC behavior without real agent invocations.
  - **Impact**:
    - Add `--fluxid-dry-run` CLI flag
    - In dry-run: print what would execute (phase, loop iteration, command file used)
    - Skip actual agent process spawning
    - Generate synthetic PASS reports to allow loop progression testing
    - Validate configuration and command files normally

- ✅ **Report validation behavior - infinite retry**: Missing or invalid reports are NOT acceptable. If a step (implement/review) outputs an invalid report, redo the step infinitely until a valid report is produced. A missing report is unacceptable and a step HAS to output it.
  - **Rationale**: Reports are critical workflow control mechanism. Steps must produce valid reports or workflow cannot proceed correctly.
  - **Impact**:
    - After implement phase: validate report. If invalid/missing, re-run implement phase (same retry attempt)
    - After review phase: validate report. If invalid/missing, re-run review phase (same iteration)
    - Infinite retry until valid report produced - no maximum attempt limit
    - Loop counters only increment when valid PASS/FAIL report received
    - Validation errors must print clear diagnostic: what failed, why, how to fix

- ✅ **Loop escape mechanism**: Users interrupt runaway infinite retries via standard process signals (Ctrl+C / SIGINT/SIGTERM). Additionally, implement IPC command `abort` to enable automated processes using fluxid to programmatically abort workflows.
  - **Rationale**: Simple manual control for interactive use, programmatic control for automation scenarios.
  - **Impact**:
    - Implement signal handlers for SIGINT/SIGTERM for graceful shutdown
    - Add `fluxid ipc abort [--session <id>]` command
    - Abort command sets abort flag for session, causing wrapper to exit after current agent invocation completes
    - Abort should return session to clean state (no partial reports)

- ✅ **Command files must use IPC**: All command files (implement, commit, review) MUST instruct agents to use IPC commands for communication. The current commands use file.sh script - all command communication MUST be converted to IPC calls.
  - **Rationale**: IPC is the official communication protocol. File-based communication from bash prototype must be replaced.
  - **Impact**:
    - Example/template command files must show `fluxid ipc write-report` usage
    - Example/template command files must show `fluxid ipc write-history` usage
    - Documentation must clearly explain IPC requirement for command authors
    - Bash implementation's file.sh calls must be replaced with IPC equivalents
    - Command files should reference report schema location for validation

### Pending Clarifications

None - all clarifications have been resolved.

---

---

## Analysis of Existing Bash Implementation

**Purpose:** The existing bash implementation (`epic-loop.sh`) demonstrates a working prototype of the fluxid loop concept that has been used in production. Analyzing this implementation reveals critical insights about how the loop mechanics, report validation, and phase transitions should work.

### Implementation Architecture

The bash implementation consists of several modular components:

**Core Loop Script** (`epic-loop.sh` - 583 lines):
- Entry point that orchestrates the complete implement → commit → review iteration
- Manages nested loops (iterations: review iterations up to 20, retries: implementation attempts up to 3)
- Handles agent selection (--claude, --codex, --opencode flags)
- Implements intelligent resume logic based on current state
- Performs systematic cleanup every 5th iteration

**Module Libraries:**
- `workflow-steps.sh` - Encapsulates agent invocations for implement, commit, and review phases
- `progress-tracking.sh` - Progress state management (external to fluxid, but shows integration pattern)
- `report-operations.sh` - Report parsing, validation, and metadata checking
- `status-display.sh` - User-facing status messages and visual separators

**Agent Abstraction:**
- `run-claude-streaming.sh` - Parses Claude's streaming JSON output, displays clean status messages
- `run-codex-streaming.sh` - Equivalent for Codex agent
- `run-opencode-streaming.sh` - Equivalent for OpenCode agent

**Support Scripts:**
- `files.sh` - Deterministic file path resolution for reports, history, test files
- `validate-report.sh` - Python-based YAML schema validation against report-schema.yaml
- `commit.sh` - Delegates to streaming agent with commit-specific instruction

### Nested Loop Mechanics (As Implemented)

**Note**: The bash implementation uses `outer`/`inner` variable names. Semantically, `outer` represents **iterations** and `inner` represents **implement retries**.

```
iteration = 0
while iteration < MAX_ITERATIONS (20):
    retry = 0
    impl_passed = false

    # Implement retries: Implementation attempts
    while retry < 3:
        ┌─ IMPLEMENT PHASE ─┐
        │ 1. Set status to 'implement'
        │ 2. Invoke agent with implement command
        │ 3. Agent writes report to workflow-report.yaml
        │ 4. If agent fails (non-zero exit), abort entire workflow
        └───────────────────┘

        ┌─ COMMIT PHASE ─┐
        │ 1. Invoke agent with commit instruction
        │ 2. Agent commits all changes, fixes pre-commit issues
        │ 3. If commit fails, abort entire workflow
        └────────────────┘

        ┌─ VALIDATE REPORT ─┐
        │ 1. Check report file exists
        │ 2. Validate YAML structure against schema
        │ 3. Validate metadata (command, artifact match expectations)
        │ 4. Parse status field
        │ 5. If status == "PASS":
        │      impl_passed = true
        │      break retry loop
        │ 6. Else: continue to next retry iteration
        └───────────────────┘

        retry++
    end retry loop

    ┌─ REVIEW PHASE ─┐
    │ 1. Clean up screenshots before review
    │ 2. Set status to 'review'
    │ 3. Invoke agent with review command
    │ 4. Agent writes review report
    │ 5. If agent fails, abort entire workflow
    │ 6. Validate review report (same validation steps)
    │ 7. Parse review status
    │ 8. If status == "PASS":
    │      Mark task complete
    │      Exit with success
    │ 9. Else: continue to next iteration
    └───────────────┘

    # Periodic cleanup (every 5th iteration)
    if (iteration + 1) % 5 == 0:
        - Remove all screenshots
        - Remove report file
        - Truncate history file
        - Reset status to 'pending'

    iteration++
end iteration loop

# If we get here, max iterations exhausted
Set status to 'error'
Exit with failure
```

### Critical Insights from Bash Implementation

#### 1. Agent Invocation Pattern
The bash version doesn't call the agent binary directly - it constructs an instruction string and passes it to a streaming wrapper:

```bash
instruction="Read and understand the command in \`$implement_cmd_file\` and execute it
for this epic id \`$epic_id\` and its E2E test file \`$test_file\`.
Use the shared workflow report protocol: write a PURE YAML report to the path from
\`.fluxid/scripts/command/files.sh --report\` following
\`.fluxid/templates/report-schema.yaml\` and validate it with
\`.fluxid/scripts/command/validate-report.sh\`."

"$streaming_script" "$instruction"
```

**Translation to fluxid product**: The prompts configured in fluxid's config.yaml should contain similar instructions that direct Claude to write reports following the schema.

#### 2. Report-Driven Loop Control
The loop continuation logic is entirely driven by report validation:

```bash
# Retry loop breaks on PASS
if [[ "$impl_status" == "PASS" ]]; then
    impl_passed=true
    break  # Exit retry loop, proceed to review
fi

# Iteration loop exits on review PASS
if [[ "$review_status" == "PASS" ]]; then
    mark_task_complete
    exit 0
fi
```

**Translation to fluxid product**: After each phase execution, fluxid must read the report via IPC, validate it, and use the status field to determine loop continuation. This means:
- fluxid itself doesn't write reports - it expects Claude (via configured prompts) to call `fluxid ipc write-report`
- fluxid reads back the report using its internal storage to check status
- PASS status breaks retry loop (implement) or exits iteration loop (review)

#### 3. Hard Failure on Agent Errors
Any agent execution failure immediately aborts the entire workflow:

```bash
if ! "$streaming_script" "$instruction"; then
    error "CRITICAL: Halting workflow loop due to implement step failure."
    exit 1
fi
```

**Translation to fluxid product**: If `claude` process exits non-zero, fluxid must immediately abort and exit with the same code. No retry, no continuation.

#### 4. Validation Before Trust
Every report undergoes multi-stage validation:

```bash
# Stage 1: File exists
if ! report_exists "$REPORT_FILE"; then
    error "Report file missing"
    continue
fi

# Stage 2: YAML structure validates against schema
if ! validate_report_structure "$REPORT_FILE"; then
    error "Report structure validation failed"
    continue
fi

# Stage 3: Metadata matches expectations
if ! validate_report_metadata "$REPORT_FILE" "$EXPECTED_COMMAND" "$EXPECTED_ARTIFACT"; then
    error "Report metadata validation failed"
    continue
fi

# Stage 4: Extract status
impl_status="$(parse_report_status "$REPORT_FILE")"
if [[ -z "$impl_status" ]]; then
    error "Report status field missing"
    continue
fi
```

**Translation to fluxid product**: The `write-report` IPC command must perform equivalent validation. The bash implementation uses Python for schema validation - the Go implementation should use a YAML schema validator library.

#### 5. Resume Logic Implementation
The bash version implements intelligent resume based on current state:

```bash
CURRENT_STATUS=$(get_task_status "$EPIC_ID_TOKEN")  # External progress system
RESUME_FROM="implement"

if [[ "$CURRENT_STATUS" == "review" ]]; then
    # Check if valid PASS report exists for implement phase
    if valid_report_with_pass_status_exists; then
        RESUME_FROM="review"
        # Skip implement retries, go straight to review
    else
        # Invalid/missing report, restart from implement
        RESUME_FROM="implement"
        rm -f "$REPORT_FILE"
    fi
fi
```

**✅ Clarified**: Fluxid does NOT implement resume logic. Each invocation starts fresh from implement phase. This is the responsibility of external orchestration systems.

#### 6. Cleanup Strategy
Every 5th iteration triggers comprehensive cleanup:

```bash
if [[ $(((iteration + 1) % 5)) -eq 0 ]]; then
    # Remove screenshots, report, truncate history
    # This prevents infinite context accumulation
fi
```

**✅ Clarified**: Fluxid does NOT implement periodic cleanup. Session data remains in memory until wrapper exit, respecting only the 32MB history limit.

#### 7. Streaming Output Handling
The bash implementation uses a sophisticated JSON stream parser for Claude's output:

```bash
claude --output-format stream-json --include-partial-messages -p "$PROMPT" 2>&1 |
while IFS= read -r line; do
    # Parse JSON events: system, stream_event, assistant, tool_result, result
    # Display: text deltas, tool names, completion status
done
```

**Translation to fluxid product**: The Go implementation should pipe Claude's output directly without parsing (per requirements: "stdout → os.Stdout"). The bash version's parsing is specific to the bash wrapper's needs, not a core fluxid requirement.

#### 8. Dry-Run Mode Implementation
The bash version includes a `--dry-run` flag for testing:

```bash
if [[ "$DRY_RUN" == true ]]; then
    # Generate synthetic reports based on control file
    # Allows testing loop logic without executing actual agent
fi
```

**Clarification needed**: ❓ Should fluxid support a dry-run mode for testing?

### Schema Validation Details

The bash implementation uses Python for YAML validation with detailed error reporting:

**Report Schema Requirements** (from `validate-report.sh`):
- Required fields: `command`, `artifact`, `timestamp`, `status`, `issues`
- `status` enum: `PASS` or `FAIL`
- `artifact` must be single token (no `/` or `.` characters - prevents paths/filenames)
- `timestamp` must be ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)
- `issues` must contain all 5 categories: `blockers`, `defects`, `concerns`, `observations`, `enhancements`
- Each category must be an array (empty if no issues)
- Each issue requires `message` field, optional: `location`, `code`, `suggestion`, `reference`
- `additionalProperties: false` for issues object (no extra categories allowed)
- Optional fields: `next_steps` (array of strings), `summary` (string)

**Translation to fluxid product**: The `fluxid ipc write-report` command must implement equivalent validation logic, returning human-readable errors on validation failure.

### Agent Abstraction Pattern

The bash implementation supports multiple agents through a unified interface:

```bash
case "$AGENT" in
    claude)
        STREAMING_SCRIPT="$PROJECT_ROOT/.fluxid/scripts/run-claude-streaming.sh"
        ;;
    codex)
        STREAMING_SCRIPT="$PROJECT_ROOT/.fluxid/scripts/run-codex-streaming.sh"
        ;;
    opencode)
        STREAMING_SCRIPT="$PROJECT_ROOT/.fluxid/scripts/run-opencode-streaming.sh"
        ;;
esac
```

**✅ Clarified**: Multi-agent support (Claude, Codex, OpenCode) is a core requirement using explicit agent selection flags: `--claude`, `--codex`, `--opencode`. The bash implementation demonstrates this pattern correctly. The Go implementation should:
- Accept one of three agent flags to select which agent to use
- Default to `--claude` if no agent flag specified
- Error if multiple agent flags provided simultaneously
- Agents must be available on PATH (third-party dependencies)

### Commit Phase Details

The bash implementation shows the commit phase delegates to an agent with specific instructions:

```bash
INSTRUCTION="Commit all open changes, create a fully clean repo state.
Fix all pre-commit issues that may appear."

"$STREAMING_SCRIPT" "$INSTRUCTION"
```

**Translation to fluxid product**: The `commit.prompt` configuration should contain similar instructions. The commit phase is not fundamentally different from implement/review - it's just another claude invocation with a different prompt.

### History File Usage

The bash implementation uses history for:
- Logging decisions and trade-offs during implementation
- Recording which subtasks were delegated and their scope
- Tracking postponed work and scope cuts

The history file path is resolved via: `.fluxid/scripts/command/files.sh --history`

However, history validation is required before report validation succeeds:

```python
# From validate-report.sh
HISTORY_FILE=$("$SCRIPT_DIR/files.sh" --history)
if [[ ! -f "$HISTORY_FILE" ]]; then
    echo "Error: History file not found: $HISTORY_FILE"
    echo "The workflow-loop-history.md file must exist for validation to pass"
    exit 1
fi
```

**✅ Clarified**: Fluxid uses plain text log format for history (not YAML, not markdown). Format: `[ISO8601 timestamp] message`. In-memory only, no file persistence.

### Key Differences: Bash vs. Original Requirements

| Aspect | Bash Implementation | Original Requirements | Resolution Status |
|--------|---------------------|----------------------|-------------------|
| Agent support | Multi-agent (claude/codex/opencode) with `--claude`, `--codex`, `--opencode` flags | Claude-specific in docs | ✅ **CLARIFIED**: Multi-agent required via explicit flags `--claude`, `--codex`, `--opencode` |
| Resume logic | Implemented with state checking | Not mentioned | ✅ **CLARIFIED**: NOT included in fluxid - external orchestration responsibility |
| Periodic cleanup | Every 5th iteration | Not mentioned | ✅ **CLARIFIED**: NOT included in fluxid - respect 32MB history limit only |
| History format | Markdown file (`.md`) | YAML entries | ✅ **CLARIFIED**: Plain text log format `[timestamp] message`, in-memory only |
| Dry-run mode | Implemented for testing | Not mentioned | ✅ **CLARIFIED**: Include via `--fluxid-dry-run` flag |
| Output parsing | JSON stream parsing for display | Direct passthrough | ✅ Requirements clear (passthrough) |
| Progress tracking | External (progress.yaml) | Not mentioned (external) | ✅ Out of scope confirmed |
| Epic/task files | Reads epic files, resolves test paths | Not mentioned | ✅ Out of scope confirmed |

---

## Document Metadata

**Analysis Date:** 2025-12-10
**Updates:**
- Added comprehensive bash implementation analysis (330+ lines)
- Clarified multi-agent support (Claude/Codex/OpenCode) as core requirement
- Agent selection via explicit flags: `--claude`, `--codex`, `--opencode`
- Agents must be available on PATH (third-party dependencies)
- Documented complete configuration structure (home/project config.yaml, command files, env vars, CLI flags)
- Created `fluxid.commit.md` command file based on bash implementation
- Documented command file pattern (`.fluxid/commands/` with home/project override)
- Documented precedence order: defaults → home config → project config → env vars → CLI flags
- **COMPLETED ALL CLARIFICATIONS (20+ decisions)**:
  - Removed `--fluxid-no-review` flag (review is mandatory)
  - History is in-memory only, plain text log format with timestamps, 32MB size limit
  - `--write-history <message>` writes to in-memory log (not file export)
  - Session ID format: UUID v4
  - Initialization status shows config values and paths (not full prompts)
  - Missing command files abort at startup
  - Loop counts validated (must be ≥1)
  - Infinite retry behavior for missing/invalid reports
  - Abort mechanism via Ctrl+C or `fluxid ipc abort` command
  - Dry-run mode via `--fluxid-dry-run` flag
  - Command files MUST use IPC (not file.sh from bash prototype)
  - No resume logic (each invocation starts fresh)
  - No periodic cleanup (respect 32MB history limit only)

**Requirements Sources:**
- fluxid/requirements/fluxid-wrapper-desciption.md (comprehensive specification)
- fluxid/requirements/report-schema.yaml (report data structure)
- fluxid/requirements/history-schema.yaml (history entry data structure)
- fluxid/requirements/fluxid.implement.md (implement command template)
- fluxid/requirements/fluxid.review.md (review command template)
- fluxid/requirements/fluxid.commit.md (commit command template - created)

**Sections Included:**
✅ Comprehensive Request Summary
✅ Configuration and Command Structure (NEW - comprehensive config documentation)
✅ Business Context
✅ Stakeholders
✅ Scope
✅ User Journeys
✅ Interaction Patterns
✅ Business Rules
✅ Information Architecture
✅ Success Indicators
✅ Assumptions
✅ Constraints
✅ Validation Checklist
✅ Analysis of Existing Bash Implementation (NEW - 330+ lines)

**Sections Skipped:**
❌ Interface Understanding (no UI component - this is a CLI tool)
