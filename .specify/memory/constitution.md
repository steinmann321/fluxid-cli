<!--
Sync Impact Report:
Version Change: 1.0.0 → Initial creation
Modified Principles: N/A (initial creation)
Added Sections: All sections (initial creation)
Removed Sections: None
Templates Requiring Updates:
  ✅ .specify/templates/plan-template.md - Constitution Check section aligns
  ✅ .specify/templates/spec-template.md - Requirements structure compatible
  ✅ .specify/templates/tasks-template.md - Task structure aligns with sequential workflow
Follow-up TODOs:
  - RATIFICATION_DATE set to today (2026-01-05) as initial adoption
-->

# fluxid Constitution

## Core Principles

### I. Test-Driven Development (NON-NEGOTIABLE)

Test-Driven Development is MANDATORY for all code changes. Tests MUST be written before implementation begins. The Red-Green-Refactor cycle is strictly enforced:

1. Write failing tests (RED)
2. Implement minimal code to pass tests (GREEN)
3. Refactor while keeping tests green (REFACTOR)

**Rationale**: TDD ensures correctness, prevents regressions, and documents expected behavior through executable specifications. This is non-negotiable because workflow orchestration requires predictable behavior under all conditions.

### II. Full E2E Test Coverage (NON-NEGOTIABLE)

Every workflow path MUST have comprehensive end-to-end test coverage. E2E tests MUST verify:

- Complete user workflows from CLI invocation to completion
- Agent integration and IPC communication
- Error handling and graceful degradation
- State persistence and recovery scenarios
- Configuration resolution across all precedence layers

**Rationale**: As a workflow controller, fluxid's correctness depends on validating entire execution paths. Unit tests alone cannot catch integration issues between CLI, config, IPC, and agent subsystems.

### III. Strictly Sequential Workflow (NON-NEGOTIABLE)

The workflow MUST execute in a fully sequential manner with NO parallelism. All operations within the implement → review loop MUST complete before the next iteration begins:

- One agent invocation at a time
- Sequential report validation
- Ordered history writes
- No concurrent state modifications
- No async operations requiring synchronization

**Rationale**: Sequential execution eliminates race conditions, timing dependencies, and the need for complex synchronization. This prevents flaky tests and ensures deterministic behavior that can be reliably validated.

### IV. Strict Code Quality Enforcement (NON-NEGOTIABLE)

Code quality gates are enforced through pre-commit hooks that MUST pass before any commit. These hooks are strictly read-only regarding repository contents and MUST NOT:

- Write, modify, or stage files (except formatting of already-staged files)
- Be loosened, weakened, bypassed, or disabled
- Reduce strictness of validation (formatting, linting, security, coverage)

The following code quality principles MUST be followed:

- **SOLID**: Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
- **DRY**: Don't Repeat Yourself - eliminate code duplication
- **KISS**: Keep It Simple, Stupid - favor simplicity over cleverness
- **YAGNI**: You Aren't Gonna Need It - implement only what's required now

**Rationale**: Consistent code quality prevents technical debt accumulation, improves maintainability, and ensures the codebase remains comprehensible as it scales. Pre-commit enforcement catches issues before they enter the repository.

### V. Pure Go Implementation

The final system MUST be a pure Go implementation. Shell scripts present in the repository are examples, templates, or build helpers only - they MUST NOT be runtime dependencies. All workflow logic, agent integration, and IPC mechanisms MUST be implemented in Go.

**Rationale**: Pure Go ensures cross-platform portability, type safety, predictable performance, and maintainable code. Shell script dependencies introduce platform-specific behavior and fragile execution environments.

### VI. Explicit Interfaces Over Implicit Behavior

All interfaces MUST be explicit and documented. The system MUST NOT rely on:

- Implicit conventions or "magic" behavior
- Undocumented side effects
- Hidden global state
- Reflection-based "convention over configuration"

Every component interface, IPC protocol, and agent contract MUST have clear, documented specifications.

**Rationale**: Explicit interfaces make the system testable, debuggable, and comprehensible. Workflow orchestration requires precise contracts between components - implicit behavior leads to integration failures and maintenance burden.

### VII. Fail Fast with Clear Diagnostics

The system MUST fail fast when encountering invalid states, configuration errors, or constraint violations. Error messages MUST:

- Clearly identify what failed
- Explain why it failed
- Provide actionable remediation steps
- Include relevant context (file paths, session IDs, config values)

No silent failures, best-effort workarounds, or error suppression.

**Rationale**: Fast failure prevents cascading errors and makes debugging straightforward. Clear diagnostics reduce user frustration and support burden. Workflow controllers must be reliable - unclear failures undermine trust.

### VIII. Command-Line First, Scriptable Always

All functionality MUST be accessible via CLI with:

- Parseable output formats (text, JSON, YAML)
- Exit codes indicating success/failure
- Stdin/stdout/stderr adherence (data in/out via stdin/stdout, errors to stderr)
- Environment variable support where appropriate
- No required interactive prompts (all inputs via flags or files)

**Rationale**: CLI-first design ensures automation compatibility, pipeline integration, and programmatic control. Users need to script workflows, integrate with CI/CD, and build tooling on top of fluxid.

## Development Workflow

### Code Review Requirements

All code changes MUST undergo review before merging. Reviews MUST verify:

1. **Constitution Compliance**: Changes adhere to all core principles
2. **Test Coverage**: New code has comprehensive E2E tests, tests written first
3. **Sequential Execution**: No parallelism introduced, no race conditions possible
4. **Code Quality**: SOLID, DRY, KISS, YAGNI principles followed
5. **Error Handling**: Fail-fast behavior with clear diagnostics
6. **Documentation**: Interfaces documented, behavior specified

### Pre-Commit Quality Gates

The following gates MUST pass before any commit:

- **Formatting**: Code formatted per Go standards (gofmt, goimports)
- **Linting**: No linter warnings (golangci-lint with strict configuration)
- **Security**: No security vulnerabilities detected (gosec)
- **Coverage**: Test coverage maintains or improves project threshold (currently 90%)

These gates MUST NOT be bypassed, disabled, or weakened.

### Testing Discipline

Tests MUST be:

- **Written First**: TDD cycle enforced - failing tests before implementation
- **Comprehensive**: E2E coverage for all workflow paths
- **Deterministic**: No flaky tests, no timing dependencies
- **Fast**: Test suite executes quickly to support rapid iteration
- **Isolated**: Tests independent, no shared state between tests

Flaky tests MUST be fixed or removed immediately - they MUST NOT be merged.

### Commit Discipline

Commits MUST:

- Be atomic (single logical change)
- Have clear, descriptive messages
- Pass all pre-commit hooks
- Include tests for new functionality
- Reference related issues/specs where applicable

## Governance

### Constitution Authority

This constitution supersedes all other practices, conventions, and ad-hoc decisions. When conflicts arise:

1. Constitution principles take precedence
2. Technical feasibility concerns must be raised explicitly
3. Principle violations require documented justification and approval
4. Temporary exceptions must include remediation timeline

### Amendment Process

Constitution amendments require:

1. **Documentation**: Proposed change with rationale and impact analysis
2. **Review**: Technical review for implications on architecture and existing code
3. **Approval**: Explicit approval from project maintainers
4. **Migration Plan**: If changes affect existing code, migration plan required
5. **Version Update**: Semantic versioning (MAJOR.MINOR.PATCH) applied

### Versioning Policy

- **MAJOR**: Backward-incompatible governance/principle removals or redefinitions
- **MINOR**: New principle/section added or materially expanded guidance
- **PATCH**: Clarifications, wording refinements, typo fixes

### Compliance Reviews

All pull requests MUST verify constitution compliance before merge. Regular constitution reviews MUST occur:

- After major feature additions
- When complexity trends upward
- When principles are challenged repeatedly
- At least quarterly

### Runtime Development Guidance

For runtime development guidance, refer to `/CLAUDE.md` which provides context-specific implementation notes. The constitution defines **what** principles govern the project; CLAUDE.md provides **how** to apply them during development.

**Version**: 1.0.0 | **Ratified**: 2026-01-05 | **Last Amended**: 2026-01-05
