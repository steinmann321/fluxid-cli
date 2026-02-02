# fluxid Loop

## CRITICAL REMINDER (2025-12-12T10:51:42.483Z)
This is a reimplementation project: the final system must be a pure Go implementation. Any shell scripts present are examples/templates of a prior hacky solution and must not be used as runtime dependencies.

## Overview
- The app fluxid is a thin CLI wrapper around coding agents (claude, codex, opencode)
- The app is a workflow controller that enables coding agents to break the context window with configurable workflow steps
- Supports both config-driven workflows (1..N custom steps) and legacy hardcoded workflow (implement → commit → review)
- The workflow steps allow coding agents to stop working and provides a structured way to hand over past, current and future task states

## Tech Stack

**CLI**: Go
**Data**: YAML (reports, schemas, templates)
**Shell**: Bash (build/helper scripts)

## Architecture

**CLI Layer**
- Command parser (commands, flags, subcommands)
- Workflow orchestration (config-driven custom steps OR legacy hardcoded workflow)
- Agent delegation and session management
- File path resolution and context management

**Workflow System** (002-config-driven-workflow)
- **Config-Driven Workflow**: Users define 1..N custom workflow steps in config.yaml with individual retry limits
- **Legacy Workflow**: Fallback to hardcoded 3-step workflow (implement → commit → review) for backward compatibility
- **Review Exit Gate**: Review step always executes last and serves as the only valid exit point (PASS exits, FAIL triggers next iteration)
- **Sequential Execution**: All steps execute sequentially in order (no parallelism)
- **Retry Logic**: Each step has configurable retry limit (0 = infinite retries, N = limited retries)
- **Startup Validation**: Fail-fast validation of workflow config before execution (missing files, duplicate names, invalid paths)

**Core Logic**
- Epic/task execution
- Report generation and validation
- History tracking
- Progress state management

**Integration**
- Coding agent adapters (Claude, OpenCode, etc.)
- Test runner integration
- Git workflow integration
- Hook system (pre-commit, validation gates)

**Design Principles**
- Explicit interfaces over implicit behavior
- Command-line first, scriptable always
- Fail fast with clear diagnostics
- Stateless commands, stateful sessions

## Hooks Policy
- The pre-commit hook configuration is strictly read-only regarding repository contents. It must never write, modify, or stage files outside of formatting already present in staged files.
- It is never permitted under any circumstances to loosen, weaken, bypass, or disable the pre-commit checks. NEVER EVER.
- Any change that reduces the strictness of validation gates (formatting, linting, security, coverage) is forbidden.
- Agents working in this repository must uphold this policy and refuse changes that relax pre-commit enforcement.

## Active Technologies
- Go 1.25 (001-report-history-refactor)
- File-based YAML storage in session-specific directories (report.yaml, history.yaml) (001-report-history-refactor)
- Go 1.25 + gopkg.in/yaml.v3 (YAML parsing), github.com/google/uuid (Session ID generation) (002-config-driven-workflow)
- File-based YAML storage (report.yaml, history.yaml) in session-specific directories (002-config-driven-workflow)

## Recent Changes
- 002-config-driven-workflow: Config-driven workflow system
  - Replace hardcoded 3-step workflow with configurable workflow steps
  - Support 1..N custom steps with individual retry limits per step
  - Mandatory review step as exit gate (PASS exits successfully, FAIL triggers next iteration)
  - Backward compatible: Legacy workflow still available when workflow section not configured
  - Startup validation: Fail-fast on invalid configs (missing files, duplicate names, negative retries)
  - File structure: workflow.steps[] + workflow.review in config.yaml
  - New E2E tests: workflow_config_driven_test.go, workflow_startup_validation_test.go
  - Implementation: internal/workflow/workflow.go (runConfigDrivenWorkflow + runLegacyWorkflow)
- 001-report-history-refactor: Added Go 1.25
