# Data Model: Config-Driven Workflow System

**Feature**: 002-config-driven-workflow
**Date**: 2026-01-18
**Phase**: 1 (Design & Contracts)

## Overview

This document defines the data entities, relationships, validation rules, and state transitions for the config-driven workflow system. The data model supports dynamic workflow configuration while maintaining strict validation and sequential execution guarantees.

---

## Core Entities

### 1. WorkflowConfig (YAML Configuration)

**Purpose**: Represents the workflow section in config.yaml as parsed from YAML

**Location**: `internal/config/config.go` (new types)

**Attributes**:

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `steps` | `[]WorkflowStepConfig` | Yes | N/A | Min length: 1, Max length: 10 |
| `review` | `ReviewStepConfig` | Yes | N/A | Must be present |

**Relationships**:
- **Contains** 1..10 `WorkflowStepConfig` objects
- **Contains** exactly 1 `ReviewStepConfig` object

**YAML Example**:
```yaml
workflow:
  steps:
    - name: design
      command: commands/design.md
      retries: 2
    - name: implement
      command: commands/implement.md
      retries: 3
    - name: test
      command: commands/test.md
      retries: 1
  review:
    command: commands/review.md
    retries: 1
```

**Validation Rules**:
- `workflow` section MUST exist in config.yaml (FR-010)
- `workflow.steps` MUST contain at least 1 custom step (FR-011)
- `workflow.steps` MUST contain at most 10 steps (FR-014, A-004)
- `workflow.review` section MUST exist (FR-018)

**Go Struct**:
```go
type WorkflowConfig struct {
    Steps  []WorkflowStepConfig `yaml:"steps"`
    Review ReviewStepConfig     `yaml:"review"`
}
```

---

### 2. WorkflowStepConfig (YAML Step Definition)

**Purpose**: Represents a single workflow step definition in config.yaml

**Location**: `internal/config/config.go` (new type)

**Attributes**:

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `name` | `string` | Yes | N/A | Non-empty, unique within workflow, no whitespace-only |
| `command` | `string` | Yes | N/A | Valid file path (absolute or relative), file must exist and be readable |
| `retries` | `int` | No | 1 | Non-negative integer (0 = infinite retries) |

**Relationships**:
- **Belongs to** one `WorkflowConfig`
- **Maps to** one runtime `WorkflowStep` during initialization

**YAML Example**:
```yaml
- name: implement
  command: commands/implement.md
  retries: 3
```

**Validation Rules**:
- `name` MUST NOT be empty or whitespace-only (FR-011)
- `name` MUST be unique within workflow (FR-011, A-006)
- `name` is case-sensitive (A-006)
- `command` path MUST exist and be readable before workflow execution (FR-015)
- `command` supports both absolute and relative paths (FR-015, A-005)
- Relative paths resolved from config.yaml directory (FR-015, A-005)
- `retries` MUST be non-negative (FR-023)
- `retries = 0` means infinite retries until PASS (FR-002, FR-113)
- `retries >= 1` means limited retry attempts (FR-002, FR-113)

**Go Struct**:
```go
type WorkflowStepConfig struct {
    Name    string `yaml:"name"`
    Command string `yaml:"command"`
    Retries int    `yaml:"retries,omitempty"` // Default: 1
}
```

---

### 3. ReviewStepConfig (YAML Review Definition)

**Purpose**: Represents the mandatory review step configuration in config.yaml

**Location**: `internal/config/config.go` (new type)

**Attributes**:

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `command` | `string` | Yes | N/A | Valid file path (absolute or relative), file must exist and be readable |
| `retries` | `int` | No | 1 | Non-negative integer (0 = infinite retries) |

**Relationships**:
- **Belongs to** one `WorkflowConfig`
- **Maps to** one runtime `WorkflowStep` with special exit gate semantics

**YAML Example**:
```yaml
review:
  command: commands/review.md
  retries: 1
```

**Validation Rules**:
- `review` section MUST exist in workflow config (FR-018)
- `command` MUST be specified (FR-017)
- `command` path MUST exist and be readable (FR-017)
- `command` supports both absolute and relative paths (FR-015, A-005)
- Relative paths resolved from config.yaml directory (A-005)
- `retries` MUST be non-negative (FR-023)
- `retries = 0` means infinite retries until PASS (FR-002)
- `retries >= 1` means limited retry attempts (FR-002)

**Go Struct**:
```go
type ReviewStepConfig struct {
    Command string `yaml:"command"`
    Retries int    `yaml:"retries,omitempty"` // Default: 1
}
```

---

### 4. WorkflowStep (Runtime Execution)

**Purpose**: Represents a workflow step during runtime execution (after config parsing and validation)

**Location**: `internal/types/types.go` (new type)

**Attributes**:

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Step name (e.g., "implement", "design", "test") |
| `CommandFilePath` | `string` | Resolved absolute file path to command file |
| `Retries` | `int` | Maximum retry attempts (0 = infinite) |
| `IsReview` | `bool` | True if this is the review step (exit gate) |
| `Order` | `int` | Execution order (0-based index in steps array) |

**Relationships**:
- **Belongs to** one `Workflow`
- **Executes** 1..N times per development iteration (based on retry logic)
- **Produces** one `Report` per execution attempt

**State Transitions**:
```
[Created] → [Executing] → [Checking Report] → [PASS] → [Next Step]
                                             ↓ [FAIL]
                                             ↓
                                   [Retry Available?]
                                        ↓ YES → [Executing]
                                        ↓ NO  → [Next Step]
```

**Validation Rules**:
- `Name` MUST match original WorkflowStepConfig.name (case-sensitive)
- `CommandFilePath` MUST be absolute path (resolved during initialization)
- `CommandFilePath` MUST have been validated at startup (FR-015)
- `Retries` MUST be non-negative
- `Order` MUST reflect array position (0-based)
- If `IsReview == true`, step MUST be last in workflow sequence

**Go Struct**:
```go
type WorkflowStep struct {
    Name            string
    CommandFilePath string
    Retries         int
    IsReview        bool
    Order           int
}
```

---

### 5. Workflow (Runtime Orchestrator)

**Purpose**: Represents the complete workflow at runtime (after config parsing, validation, and initialization)

**Location**: `internal/types/types.go` (new type)

**Attributes**:

| Field | Type | Description |
|-------|------|-------------|
| `Steps` | `[]WorkflowStep` | Ordered list of all workflow steps (custom + review) |
| `MaxIterations` | `int` | Maximum development iterations (0 = infinite) |
| `CurrentIteration` | `int` | Current iteration number (1-based) |

**Relationships**:
- **Contains** 2..11 `WorkflowStep` objects (1..10 custom + 1 review)
- **Executes** steps sequentially in array order (FR-003)
- **Loops** up to `MaxIterations` times (FR-009)

**State Transitions**:
```
[Initialized] → [Iteration Start] → [Execute Step[0]]
                                           ↓
                                    [Execute Step[1]]
                                           ↓
                                          ...
                                           ↓
                                    [Execute Step[N]]
                                           ↓
                                    [Execute Review Step]
                                           ↓
                                  [Review Status Check]
                                           ↓
                                    [PASS] → [Workflow Complete]
                                    [FAIL] → [Iteration++] → [Iteration Start]
                                                           ↓
                                          [MaxIterations Reached?]
                                                 ↓ YES → [Workflow Complete]
                                                 ↓ NO  → [Continue Loop]
```

**Validation Rules**:
- `Steps` MUST have at least 2 elements (1 custom + 1 review)
- `Steps` MUST have at most 11 elements (10 custom + 1 review)
- Last step in `Steps` MUST have `IsReview == true`
- `MaxIterations` MUST be non-negative (FR-019)
- `MaxIterations = 0` means infinite iterations (FR-009)
- `CurrentIteration` MUST be <= `MaxIterations` (unless MaxIterations = 0)

**Go Struct**:
```go
type Workflow struct {
    Steps            []WorkflowStep
    MaxIterations    int
    CurrentIteration int
}
```

---

### 6. Report (Execution Result)

**Purpose**: Represents the result of a single step execution (PASS or FAIL status)

**Location**: `internal/storage/report.go` (existing type, unchanged)

**Attributes**:

| Field | Type | Description |
|-------|------|-------------|
| `Command` | `string` | Command file path that was executed |
| `Artifact` | `string` | Artifact identifier |
| `Timestamp` | `string` | Execution timestamp |
| `Status` | `string` | "PASS" or "FAIL" |
| `Issues` | `Issues` | Issues detected during execution |
| `NextSteps` | `[]string` | Recommended next steps |
| `Summary` | `string` | Execution summary |

**Relationships**:
- **Produced by** one `WorkflowStep` execution
- **Read by** workflow orchestrator to determine retry/continue logic

**Validation Rules**:
- `Status` MUST be either "PASS" or "FAIL" (case-sensitive)
- If report read fails, treat as FAIL status (FR-022)
- Single report.yaml per session (overwritten by each step) (A-008)

**State Transitions**:
```
[Agent Invocation] → [Report Written] → [Report Read] → [Status Evaluated]
                                                              ↓
                                                    [PASS] or [FAIL]
```

**Go Struct** (existing, no changes):
```go
type Report struct {
    Command   string
    Artifact  string
    Timestamp string
    Status    string
    Issues    Issues
    NextSteps []string
    Summary   string
}
```

---

## Entity Relationship Diagram

```text
┌─────────────────────┐
│   WorkflowConfig    │ (YAML Parsing)
│ ─────────────────── │
│ + steps: []Config   │
│ + review: Config    │
└──────────┬──────────┘
           │ contains 1..10
           ├──────────────────────────────────┐
           │                                  │
           ↓                                  ↓
┌──────────────────────┐         ┌────────────────────────┐
│ WorkflowStepConfig   │         │   ReviewStepConfig     │
│ ──────────────────── │         │ ────────────────────── │
│ + name: string       │         │ + command: string      │
│ + command: string    │         │ + retries: int         │
│ + retries: int       │         └────────────────────────┘
└──────────┬───────────┘                     │
           │ maps to                         │ maps to
           ↓                                  ↓
┌──────────────────────┐         ┌────────────────────────┐
│   WorkflowStep       │         │   WorkflowStep         │
│ (Runtime)            │         │ (Runtime, IsReview=T)  │
│ ──────────────────── │         │ ────────────────────── │
│ + Name: string       │         │ + Name: "review"       │
│ + CommandFilePath    │         │ + CommandFilePath      │
│ + Retries: int       │         │ + Retries: int         │
│ + IsReview: false    │         │ + IsReview: true       │
│ + Order: 0..N        │         │ + Order: N+1           │
└──────────┬───────────┘         └────────┬───────────────┘
           │ belongs to                   │ belongs to
           └──────────┬───────────────────┘
                      ↓
           ┌──────────────────────┐
           │      Workflow        │ (Runtime Orchestrator)
           │ ──────────────────── │
           │ + Steps: []Step      │
           │ + MaxIterations: int │
           │ + CurrentIteration   │
           └──────────┬───────────┘
                      │ executes sequentially
                      ↓
           ┌──────────────────────┐
           │  Step Execution      │
           │ ──────────────────── │
           │ 1. Invoke agent      │
           │ 2. Write report.yaml │
           │ 3. Read report.yaml  │
           │ 4. Check status      │
           └──────────┬───────────┘
                      │ produces
                      ↓
           ┌──────────────────────┐
           │       Report         │
           │ ──────────────────── │
           │ + Status: PASS/FAIL  │
           │ + Command: string    │
           │ + Timestamp: string  │
           │ + Issues: []Issue    │
           └──────────────────────┘
```

---

## State Machine: Workflow Execution

### Workflow Lifecycle

```text
┌─────────────────┐
│  Config Loaded  │
└────────┬────────┘
         │
         ↓ Validate
┌─────────────────┐     Validation Failed
│   Validated     │────────────────────────┐
└────────┬────────┘                        │
         │                                 ↓
         ↓ Initialize                ┌─────────────┐
┌─────────────────┐                 │ Exit Error  │
│  Initialized    │                 └─────────────┘
└────────┬────────┘
         │
         ↓ Start Iteration Loop
┌─────────────────┐
│ Iteration Start │◄───────────────┐
└────────┬────────┘                │
         │                          │
         ↓ Execute Steps            │
┌─────────────────┐                │
│ Steps Running   │                │
└────────┬────────┘                │
         │                          │
         ↓ Review Complete          │
┌─────────────────┐                │
│  Review Status  │                │
└────────┬────────┘                │
         │                          │
         ├─[PASS]──────────────────┼──────┐
         │                          │      │
         └─[FAIL]──────────────────┤      │
                   │                │      │
                   ↓                │      │
           [Iterations Left?]       │      │
                   ↓ YES────────────┘      │
                   ↓ NO                    │
                   │                       │
                   ↓                       ↓
           ┌─────────────┐         ┌─────────────┐
           │Exit Complete│         │Exit Success │
           └─────────────┘         └─────────────┘
```

### Step Execution State Machine

```text
┌─────────────────┐
│  Step Created   │
└────────┬────────┘
         │
         ↓ Start Execution (retry=1)
┌─────────────────┐
│   Executing     │
└────────┬────────┘
         │
         ↓ Agent Invocation
┌─────────────────┐     Agent Failed
│  Agent Running  │────────────────┐
└────────┬────────┘                │
         │                          │
         ↓ Agent Complete           │
┌─────────────────┐                │
│ Reading Report  │                │
└────────┬────────┘                │
         │                          │
         ├─[Report Read Success]   │
         │         ↓                │
         │  ┌─────────────┐        │
         │  │Check Status │        │
         │  └──────┬──────┘        │
         │         ├─[PASS]────────┼────────────┐
         │         └─[FAIL]────────┤            │
         │                          │            │
         └─[Report Read Failed]────┘            │
                   │                             │
                   ↓ Treat as FAIL              │
           ┌───────────────┐                    │
           │ [Retries Left?]                    │
           └───────┬───────┘                    │
                   ├─[YES]──────────────┐       │
                   │                     │       │
                   └─[NO]───────────────┼───────┼────┐
                                        │       │    │
                                        ↓       │    │
                              ┌─────────────┐  │    │
                              │ Retry++     │  │    │
                              └──────┬──────┘  │    │
                                     │         │    │
                                     ↓         │    │
                              ┌─────────────┐ │    │
                              │ Executing   │ │    │
                              └─────────────┘ │    │
                                              │    │
                                              ↓    ↓
                                        ┌──────────────┐
                                        │Step Complete │
                                        └──────────────┘
```

---

## Validation Rules Summary

### Startup Validation (Before Workflow Execution)

All validation occurs in `internal/config/validate_workflow.go` (new file):

| Rule ID | Validation | Error Message | FR Reference |
|---------|-----------|---------------|--------------|
| V001 | Workflow section exists | "workflow section is required in config.yaml" | FR-010 |
| V002 | Review section exists | "workflow.review section is required" | FR-018 |
| V003 | Review command specified | "workflow.review.command is required" | FR-017 |
| V004 | At least 1 custom step | "at least one custom workflow step is required before review" | FR-011 |
| V005 | Step name non-empty | "step name cannot be empty or whitespace-only" | FR-011 |
| V006 | Step names unique | "duplicate step name: {name}" | FR-011 |
| V007 | Command file exists | "command file not found: {path}" | FR-015 |
| V008 | Command file readable | "command file not readable: {path}" | FR-015 |
| V009 | Retries non-negative | "retries cannot be negative" | FR-023 |
| V010 | Max iterations non-negative | "max_iterations cannot be negative" | FR-019 |

### Runtime Validation (During Execution)

| Rule ID | Validation | Behavior | FR Reference |
|---------|-----------|----------|--------------|
| R001 | Report status is PASS/FAIL | Treat missing/invalid as FAIL | FR-005, FR-022 |
| R002 | Agent invocation failure | Treat as FAIL, retry per step config | FR-022 |
| R003 | Report read failure | Treat as FAIL, retry per step config | FR-022 |
| R004 | Sequential execution | Steps execute in array order | FR-003 |
| R005 | Review PASS exits workflow | Return exit code 0 | FR-007 |
| R006 | Review FAIL continues iteration | Start next iteration if iterations remain | FR-008 |
| R007 | Max iterations exhausted | Exit workflow | FR-009 |

---

## Data Flow Diagram

```text
[config.yaml]
     ↓ Parse YAML
[WorkflowConfig]
     ↓ Validate (startup)
[Validation Errors] → [Exit with Error]
     ↓ Pass
[Resolve Paths]
     ↓
[WorkflowStep[]] + [ReviewStep]
     ↓
[Workflow Orchestrator]
     ↓ Loop: Iterations
     ├──→ Loop: Steps
     │      ├──→ [Execute Step]
     │      │        ↓
     │      │   [Invoke Agent]
     │      │        ↓
     │      │   [Write report.yaml]
     │      │        ↓
     │      │   [Read report.yaml]
     │      │        ↓
     │      │   [Check Status]
     │      │        ├─[PASS]──→ [Next Step]
     │      │        └─[FAIL]──→ [Retry Logic]
     │      │                        ├─[Retries Left]──→ [Execute Step]
     │      │                        └─[No Retries]───→ [Next Step]
     │      │
     │      └──→ [Review Step]
     │                ↓
     │           [Check Status]
     │                ├─[PASS]──→ [Exit Success (0)]
     │                └─[FAIL]──→ [Check Iterations]
     │                                ├─[Iterations Left]──→ [Next Iteration]
     │                                └─[No Iterations]───→ [Exit Complete (0)]
```

---

## Configuration Precedence

Workflow configuration follows the same precedence as existing fluxid config:

```text
CLI Flags (highest precedence)
    ↓
Project Config (.fluxid/config.yaml in current directory)
    ↓
Home Config (~/.fluxid/config.yaml)
    ↓
Built-in Defaults (lowest precedence)
```

**Workflow-Specific Precedence**:
- `workflow` section can only be defined in project or home config (NOT overridable via CLI)
- `max_iterations` follows existing `--fluxid-iterations=N` CLI flag precedence
- Individual step `retries` values are NOT overridable via CLI (config only)

---

## File System Layout (Runtime)

```text
Project Root/
├── .fluxid/
│   └── config.yaml                # Workflow config defined here
│
├── commands/                       # Command files (referenced in workflow)
│   ├── design.md
│   ├── implement.md
│   ├── test.md
│   └── review.md
│
└── .fluxid-sessions/
    └── {session-id}/
        ├── report.yaml            # Single report, overwritten by each step
        └── history.yaml           # History tracking (unchanged)
```

**Key Points**:
- Command files can live anywhere (absolute or relative to config.yaml)
- Single `report.yaml` per session (A-008)
- Each step overwrites `report.yaml` (maintains current behavior)
- History tracking in `history.yaml` (unchanged by this feature)

---

## Next Steps

Data model design is complete. Proceed to contracts generation (YAML schemas for validation).
