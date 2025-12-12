# FluxID Loop

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