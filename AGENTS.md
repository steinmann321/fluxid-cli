# fluxid Loop

## CRITICAL REMINDER (2025-12-12T10:51:42.483Z)
This is a reimplementation project: the final system must be a pure Go implementation. Any shell scripts present are examples/templates of a prior hacky solution and must not be used as runtime dependencies.

## Overview
- The app fluxid is a thin CLI wrapper around coding agents (claude, codex, opencode)
- The app is a workflow controller that enables coding agents to break the context window with a IMPLEMENT - REVIEW loop
- The loop steps allow the coding agents to stop working and provides a structured way to hand over past, current and future task states

## Tech Stack

**CLI**: Go
**Data**: YAML (reports, schemas, templates)
**Shell**: Bash (build/helper scripts)

## Architecture

**CLI Layer**
- Command parser (commands, flags, subcommands)
- Workflow orchestration (implement → review → validate loops)
- Agent delegation and session management
- File path resolution and context management

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