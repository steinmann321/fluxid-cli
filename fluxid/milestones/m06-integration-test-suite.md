---
id: m06
title: Integration Test Suite
status: pending
---

# Milestone: Integration Test Suite

## Deliverable
Developers can validate cross-component interactions and full system integration by running comprehensive integration tests that verify data flows correctly through all application layers (CLI → command → IPC → storage → output), ensuring components work together correctly with proper error propagation, state consistency, and resource cleanup.

## Success Criteria
- [ ] Developers can run integration tests validating all cross-component interactions
- [ ] All component interaction patterns tested (CLI→IPC, IPC→storage, workflow→report, etc.)
- [ ] Full lifecycle testing validated (write → read → evict flows)
- [ ] Multi-session isolation verified (session A/B don't interfere)
- [ ] Persistence and recovery validated (restart doesn't lose data)
- [ ] Error propagation across layers tested (errors bubble correctly)
- [ ] State consistency validated (all components have coherent view)
- [ ] Resource cleanup verified (no leaks, proper temp file removal)
- [ ] Transaction-like behavior tested (all-or-nothing operations)
- [ ] Test failures identify which component interaction failed
- [ ] Integration tests complement unit tests (test different concerns)

## Vertical Slice Components
**Testing Layer (UI):**
- Integration test files organized by component interaction
- Test output showing component interaction flow
- Failure diagnostics identifying integration point that failed
- Coverage showing integration vs unit test contribution

**Validation Layer (Business Logic):**
- 25-35 new integration test functions
- IPC history lifecycle tests (write → read → evict)
- Main-to-IPC integration tests (CLI → handler → storage → output)
- Workflow integration tests (config → command → IPC → storage)
- Cross-component error propagation tests
- State consistency validation tests

**Quality Gate Layer (Integration):**
- Integration test success required for merge
- Integration coverage tracked separately from unit coverage
- Regression prevention for component interactions

**Data Layer (Test Fixtures):**
- Multi-component test scenarios
- Integration test utilities (setup/teardown helpers)
- Shared state management for integration tests
- Resource tracking and cleanup verification

## Validation Questions
**Before marking this milestone complete, answer:**
- [x] Can a real user (developer) perform complete workflows with only this milestone? **YES** - Developers can validate component interactions work correctly, catching integration bugs
- [x] Is it polished enough to ship publicly? **YES** - Integration tests follow professional standards, comprehensive interaction coverage
- [x] Does it solve a real problem end-to-end? **YES** - Unit tests can't catch integration issues; this validates components work together correctly
- [x] Does it include both complete UI and functional backend integration? **YES** - Test output (UI) + integration validation (backend) + pre-commit hooks (integration)
- [x] Can it run independently without waiting for other milestones? **YES** - Integration tests are self-contained, though they build on unit test foundation
- [x] Would you personally use this if it were released today? **YES** - Catches integration bugs that unit tests miss

## Notes
**Dependencies:**
- Milestones m01-m05 (unit and e2e tests provide context)
- Integration test infrastructure (setup/teardown helpers)
- Multi-component test environment

**Maps to Requirements:**
- Addresses Phase 6 from test coverage restoration plan
- Restores deleted integration test files (ipc_history_integration_test.go, main_ipc_integration_test.go)
- Validates comprehensive cross-component interactions (not just critical flows)
- Validates business rule: "Integration tests must validate cross-component interactions"

**Test Scenarios Included:**

**IPC History Integration (internal/command/ipc_history_integration_test.go):**
- Full history lifecycle (write → read → evict)
- Multi-session scenarios (session A/B isolation verification)
- Persistence validation (restart fluxid, history remains)
- Concurrent access patterns (10+ concurrent writers)
- History size limits (exactly at limit, overflow behavior, FIFO eviction)
- Corruption recovery (detect corrupted history, recover gracefully)
- Integration with IPC command layer (end-to-end history operations)
- Estimated: 5-7 happy path, 10-12 unhappy path

**Main-to-IPC Integration (internal/command/main_ipc_integration_test.go):**
- Full IPC command workflow (CLI → IPC handler → storage → output)
- IPC → workflow → report → history chain (complete flow)
- Error propagation across components (each layer handles errors correctly)
- State consistency (verify state coherent across components)
- Transaction-like behavior (all-or-nothing operations)
- Configuration propagation (config flows correctly through layers)
- Output format integration (IPC data → correct output format)
- Estimated: 5-7 happy path, 10-12 unhappy path

**Workflow Integration (internal/workflow/workflow_integration_test.go):**
- Full workflow integration (config → command → IPC → storage)
- Agent process lifecycle (spawn → execute → monitor → terminate)
- Report generation and validation (full cycle with all components)
- History tracking during workflow (verify all entries recorded)
- Resource cleanup (temp files removed, processes terminated, locks released)
- Configuration precedence (all sources combined correctly)
- Phase transition integration (implement → review → commit flow)
- Estimated: 5-7 happy path, 10-12 unhappy path

**Cross-Component Error Propagation:**
- Errors from storage layer bubble to command layer
- Errors from workflow bubble to main layer
- Errors from IPC bubble to CLI output
- Error context preserved across boundaries
- User-friendly error messages at top level
- Estimated: distributed across integration test files

**State Consistency Validation:**
- All components see same session state
- Configuration changes propagate correctly
- Report status consistent across components
- History entries match workflow events
- Estimated: distributed across integration test files

**Resource Cleanup Verification:**
- Temp files removed after workflow
- Process cleanup on success and failure
- Lock release on normal and error paths
- No resource leaks over multiple workflows
- Estimated: distributed across integration test files

**Estimated Test Count:**
- IPC history integration: 5-7 happy, 10-12 unhappy
- Main-IPC integration: 5-7 happy, 10-12 unhappy
- Workflow integration: 5-7 happy, 10-12 unhappy
- **Total: ~15-21 happy, 30-36 unhappy**

**Integration Test Characteristics:**
- Test component interactions (not individual components)
- Use real implementations where possible (minimal mocking)
- Validate data flow across boundaries
- Verify error propagation paths
- Check state consistency across components
- Confirm resource cleanup end-to-end

**Difference from Unit Tests:**
- Unit tests: Validate individual functions in isolation (with mocks)
- Integration tests: Validate components working together (real implementations)

**Difference from E2E Tests:**
- E2E tests: Validate complete user workflows (black box)
- Integration tests: Validate component interactions (white box, targeted)
