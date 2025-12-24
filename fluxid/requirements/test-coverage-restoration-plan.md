# Test Coverage Analysis & Restoration Plan

**Date**: 2025-12-22
**Current Coverage**: 80.0%
**Target Coverage**: 90.0%

## Executive Summary

The refactoring from `cmd/fluxid/` to `internal/` packages resulted in significant test coverage loss. While some packages meet the 90% target (config, ipc, output), critical areas like `cmd/fluxid` (0%), `internal/command` (70.8%), and `internal/workflow` (72.5%) are below target. Additionally, the test distribution heavily favors happy paths (78%) over error paths (22%), inverting the desired 1:3 ratio.

**Key Findings**:
- 13 test files deleted (not moved) during refactoring
- 83 functions below 90% coverage threshold
- 5 critical functions with 0% coverage
- Need ~107 additional test functions to meet goals
- Error path test coverage critically deficient

---

## Current State

### Coverage by Package

| Package | Coverage | Status | Gap to 90% |
|---------|----------|--------|------------|
| `internal/config` | 90.4% | ✅ Meets goal | - |
| `internal/ipc` | 90.2% | ✅ Meets goal | - |
| `internal/output` | 90.3% | ✅ Meets goal | - |
| `e2e-tests` | 81.2% | ❌ Below target | +18.8% (target: 100%) |
| `internal/command` | 70.8% | ❌ Below target | +19.2% |
| `internal/workflow` | 72.5% | ❌ Below target | +17.5% |
| `cmd/fluxid` | 0.0% | ❌ Critical | +90.0% |
| **Total** | **80.0%** | **❌ Below target** | **+10.0%** |

### Test Distribution Analysis

| Category | Current Count | Current % | Target % | Target Count | Gap |
|----------|---------------|-----------|----------|--------------|-----|
| Happy path tests | ~70 | ~33% | 33% | 107 | +37 |
| Unhappy path tests | ~47 | ~22% | 67% | 214 | +167 |
| **Total tests** | **214** | **100%** | **100%** | **321** | **+107** |

**Critical Issue**: Current ratio is approximately 3:1 happy to unhappy, when target is 1:2 happy to unhappy.

---

## Deleted Test Files

The following test files were **deleted** (not moved) during the refactoring:

### Critical Deletions
1. **`coverage_error_paths_test.go`** - All centralized error path tests
2. **`main_test.go`** - Main entry point tests (explains 0% cmd/fluxid coverage)

### High-Priority Deletions
3. `coverage_boost_config_test.go` - Config edge cases and boundary tests
4. `coverage_boost_ipc_test.go` - IPC edge cases and error scenarios
5. `coverage_boost_workflow_test.go` - Workflow edge cases and error paths
6. `ipc_history_integration_test.go` - History integration and persistence tests
7. `main_ipc_integration_test.go` - Main-to-IPC integration tests

### Medium-Priority Deletions
8. `coverage_boost_output_test.go` - Output format edge cases
9. `output_json_advanced_test.go` - Advanced JSON output scenarios
10. `output_json_basic_test.go` - Basic JSON output tests
11. `output_text_test.go` - Text output formatting tests
12. `output_validation_test.go` - Output validation tests
13. `output_yaml_test.go` - YAML output tests

### Support Files
14. `test_helpers_workflow.go` - Test helper utilities and fixtures
15. `coverage_boost_test_constants.go` - Test constants and shared data

---

## Functions with Critical Coverage Gaps

### 0% Coverage (Priority: CRITICAL)

| File | Function | Impact |
|------|----------|--------|
| `cmd/fluxid/main.go` | `main()` | Main entry point - no coverage |
| `internal/command/ipc_history.go` | `handleWriteHistory()` | History writing completely untested |
| `internal/command/ipc_history.go` | `handleIPCWriteHistory()` | IPC history writing untested |
| `internal/command/ipc_history.go` | `handleViewHistory()` | History viewing untested |
| `internal/workflow/workflow.go` | `Run()` | Core workflow orchestration untested |

### Below 70% Coverage (Priority: HIGH)

| File | Function | Coverage | Gap to 90% |
|------|----------|----------|------------|
| `internal/command/ipc.go` | `handleIPCCommand()` | 68.8% | +21.2% |
| `internal/workflow/workflow.go` | `waitForValidReport()` | 63.2% | +26.8% |
| `internal/ipc/storage.go` | `acquireFileLock()` | 57.1% | +32.9% |

### 70-89% Coverage (Priority: MEDIUM)

| File | Function | Coverage | Gap to 90% |
|------|----------|----------|------------|
| `internal/command/ipc_abort.go` | `handleAbort()` | 86.7% | +3.3% |
| `internal/command/ipc_handlers.go` | `handleReadReport()` | 88.9% | +1.1% |
| `internal/command/ipc_handlers.go` | `handleWriteReport()` | 81.8% | +8.2% |
| `internal/command/ipc_handlers.go` | `handleGetReportSchema()` | 77.8% | +12.2% |
| `internal/output/output.go` | `PrintJSON()` | 80.0% | +10.0% |
| `internal/output/output.go` | `PrintYAML()` | 71.4% | +18.6% |
| `internal/workflow/workflow.go` | `runImplementPhase()` | 80.0% | +10.0% |
| `internal/workflow/workflow.go` | `runCommitPhase()` | 85.7% | +4.3% |
| `internal/workflow/workflow.go` | `runReviewPhase()` | 80.0% | +10.0% |

**Total**: 83 functions below 90% threshold

---

## Restoration Plan

### Phase 1: Restore Main Entry Point Testing
**Priority**: CRITICAL
**Target**: `cmd/fluxid` from 0% → 90%
**Estimated Effort**: 1-2 days

#### Deliverables

**1.1. Create `cmd/fluxid/main_test.go`**

Test scenarios:
- Happy path: successful execution with exit code 0
- Error path: command.Execute() returns non-zero exit code
- Integration: verify os.Exit() is called with correct code
- Edge cases: nil handling, panic recovery (if applicable)

Expected test count:
- 1 happy path test
- 3-5 unhappy path tests

---

### Phase 2: Restore Error Path Testing
**Priority**: CRITICAL
**Target**: Comprehensive error path coverage across all packages
**Estimated Effort**: 5-7 days

This phase rebuilds the deleted `coverage_error_paths_test.go` functionality, distributed appropriately across packages.

#### Deliverables

**2.1. Create `internal/command/error_paths_test.go`**

Test scenarios:
- Invalid CLI arguments (missing, malformed, conflicting)
- Missing required flags
- Invalid flag values (negative numbers, invalid enums)
- Malformed config files (invalid YAML, missing keys)
- Permission errors (unreadable config, unwritable output)
- File I/O errors (missing command files, directory as file)
- Process spawn failures (agent binary not found, not executable)
- Signal handling edge cases (SIGTERM, SIGINT, multiple signals)
- Environment variable validation errors

Expected test count:
- 3-5 happy path tests
- 15-20 unhappy path tests

**2.2. Create `internal/workflow/error_paths_test.go`**

Test scenarios:
- Phase execution failures (implement/review/commit)
- Agent process crashes (non-zero exit, killed by signal)
- Report validation failures (malformed, incomplete, invalid status)
- Timeout scenarios (report wait timeout, phase timeout)
- Lock acquisition failures (concurrent access, stale locks)
- Invalid state transitions (phase order violations)
- Resource exhaustion (too many retries, too many cycles)
- Recovery scenarios (retry after failure, resume from checkpoint)

Expected test count:
- 5 happy path tests (success scenarios)
- 20-25 unhappy path tests

**2.3. Create `internal/ipc/error_paths_test.go`**

Test scenarios:
- Malformed IPC commands (invalid syntax, unknown commands)
- Invalid session IDs (empty, malformed, non-existent)
- Concurrent access conflicts (write-write, read-write)
- Storage corruption (invalid JSON, truncated files)
- Invalid report formats (wrong schema version, missing fields)
- File system errors (full disk, read-only, permission denied)
- History overflow (FIFO eviction edge cases)
- Command injection attempts (validate sanitization)

Expected test count:
- 5 happy path tests
- 20-25 unhappy path tests

**2.4. Create `internal/output/error_paths_test.go`**

Test scenarios:
- Marshal errors (JSON: circular refs, YAML: unsupported types)
- Write failures (stdout closed, pipe broken)
- Invalid format strings (unknown format, empty format)
- Nil pointer handling (nil status object, nil fields)
- Character encoding errors (invalid UTF-8)
- Large payload handling (memory limits, truncation)
- Concurrent output (multiple goroutines writing)

Expected test count:
- 3 happy path tests
- 12-15 unhappy path tests

---

### Phase 3: Restore Missing Unit Tests
**Priority**: HIGH
**Target**: All packages to 90%+ coverage
**Estimated Effort**: 5-7 days

#### Deliverables

**3.1. Create `internal/command/ipc_history_test.go`**

Fill 0% coverage gaps:
- `handleWriteHistory()`: all code paths, error scenarios
- `handleIPCWriteHistory()`: all code paths, validation errors
- `handleViewHistory()`: all code paths, empty history, malformed entries

Test scenarios:
- Valid history writes (single, multiple, UTF-8)
- Invalid writes (missing session, missing message, invalid format)
- History reads (empty, single, multiple, pagination if applicable)
- Session isolation (verify history per-session)
- Concurrent access (parallel writes, read during write)
- Error recovery (partial write, corrupted storage)

Expected test count:
- 8-10 happy path tests
- 20-25 unhappy path tests

**3.2. Create `internal/workflow/workflow_run_test.go`**

Fill `Run()` 0% coverage gap:
- Full workflow orchestration (implement → review → commit)
- Phase toggling (commit disabled, review-only)
- Retry logic (implement retries, review cycles)
- Error propagation (phase failure stops workflow)
- Success scenarios (all phases pass, loop completion)

Test scenarios:
- Happy path: complete workflow with all phases
- Partial workflow: commit disabled via flag
- Retry exhaustion: max retries reached, workflow fails
- Early termination: FAIL report stops workflow
- Signal handling: graceful abort during workflow

Expected test count:
- 5-8 happy path tests
- 15-20 unhappy path tests

**3.3. Expand existing `internal/command/*_test.go` coverage**

Target functions below 90%:
- `handleIPCCommand()`: 68.8% → 90% (+21.2%)
- `handleAbort()`: 86.7% → 90% (+3.3%)
- `handleGetReportSchema()`: 77.8% → 90% (+12.2%)
- `handleWriteReport()`: 81.8% → 90% (+8.2%)
- `handleReadReport()`: 88.9% → 90% (+1.1%)

Add test scenarios:
- Boundary conditions (empty strings, zero values, max values)
- Nil checks (nil arguments, nil context)
- Invalid inputs (malformed data, wrong types)
- Error propagation (underlying function failures)
- Edge cases specific to each function

Expected additional test count:
- 5-8 happy path tests
- 15-20 unhappy path tests

---

### Phase 4: Restore Advanced Output Testing
**Priority**: MEDIUM
**Target**: Comprehensive output format validation
**Estimated Effort**: 3-4 days

#### Deliverables

**4.1. Create `internal/output/json_advanced_test.go`**

Test scenarios:
- Complex nested structures (deep nesting, arrays of objects)
- Unicode/UTF-8 handling (emoji, CJK characters, RTL text)
- Large payloads (MB-sized outputs, thousands of fields)
- Special characters escaping (quotes, backslashes, control chars)
- Pretty vs compact formatting (indentation, whitespace)
- Null vs omitted fields (JSON null vs absent)
- Number precision (float handling, large integers)

Expected test count:
- 5-7 happy path tests
- 10-12 unhappy path tests

**4.2. Create `internal/output/yaml_advanced_test.go`**

Test scenarios:
- Multi-line strings (literal, folded styles)
- Nested structures (deep maps, mixed types)
- Anchors and references (YAML-specific features)
- Special YAML features (tags, directives)
- Unicode handling (same as JSON)
- Boolean variants (true/false, yes/no, on/off)
- Null handling (null, ~, empty)

Expected test count:
- 5-7 happy path tests
- 10-12 unhappy path tests

**4.3. Create `internal/output/text_formatting_test.go`**

Test scenarios:
- Column alignment (left, right, center)
- Truncation behavior (long strings, overflow)
- Color codes (if applicable - ANSI escape sequences)
- Wide character handling (CJK characters, emoji width)
- Line wrapping (long lines, word breaks)
- Table formatting (borders, headers, footers)

Expected test count:
- 5-7 happy path tests
- 8-10 unhappy path tests

**4.4. Create `internal/output/validation_test.go`**

Test scenarios:
- Format validation (valid formats, invalid formats)
- Schema validation (required fields, type checking)
- Cross-format consistency (same data in JSON/YAML/text)
- Output equality (verify formats represent same data)
- Roundtrip testing (serialize → deserialize → compare)

Expected test count:
- 5-7 happy path tests
- 8-10 unhappy path tests

---

### Phase 5: Expand E2E Coverage
**Priority**: MEDIUM
**Target**: E2E from 81.2% → 100%
**Estimated Effort**: 4-5 days

#### Deliverables

**5.1. Analyze E2E Coverage Gaps**

Steps:
1. Run e2e tests with coverage profile
2. Identify uncovered code paths in e2e flow
3. Map uncovered paths to missing test scenarios
4. Prioritize by risk (critical paths first)

**5.2. Add Missing E2E Scenarios**

Potential gaps (to be confirmed by coverage analysis):
- Edge case workflows (0 retries, max retries, infinite loops prevention)
- Multi-session concurrency (parallel workflows, resource contention)
- Agent variations (all supported agents: claude, codex, opencode)
- Config precedence edge cases (all 4 sources combined)
- Command file variations (missing, partial, all present)
- Output format combinations (all formats × all commands)
- History edge cases (FIFO at exactly limit, limit+1, limit×2)
- Signal handling combinations (signal at each phase)
- Report validation edge cases (all invalid report types)

Expected test count:
- 10-15 new e2e test files
- 30-40 new test functions
- Focus: 2:1 unhappy to happy ratio

**5.3. Add Chaos/Stress Tests**

Test scenarios:
- Concurrent sessions (10+ parallel workflows)
- Resource exhaustion (disk full, memory pressure)
- Network issues (if applicable - agent download, etc.)
- Filesystem race conditions (concurrent file access)
- Long-running workflows (hours, not minutes)
- Interrupted workflows (kill -9, power loss simulation)

Expected test count:
- 5-8 chaos test scenarios

---

### Phase 6: Restore Integration Tests
**Priority**: LOW
**Target**: Cross-component integration validation
**Estimated Effort**: 3-4 days

#### Deliverables

**6.1. Create `internal/command/ipc_history_integration_test.go`**

Test scenarios:
- Full history lifecycle (write → read → evict)
- Multi-session scenarios (session A/B isolation)
- Persistence validation (restart fluxid, history persists)
- Concurrent access patterns (10+ concurrent writers)
- History size limits (exactly at limit, overflow behavior)
- Corruption recovery (detect and handle corrupted history)

Expected test count:
- 5-7 happy path tests
- 10-12 unhappy path tests

**6.2. Create `internal/command/main_ipc_integration_test.go`**

Test scenarios:
- Full IPC command workflow (CLI → IPC handler → storage → output)
- IPC → workflow → report → history chain (full flow)
- Error propagation across components (each layer handles errors)
- State consistency (verify state across components)
- Transaction-like behavior (all-or-nothing operations)

Expected test count:
- 5-7 happy path tests
- 10-12 unhappy path tests

**6.3. Create `internal/workflow/workflow_integration_test.go`**

Test scenarios:
- Full workflow integration (config → command → IPC → storage)
- Agent process lifecycle (spawn → execute → terminate)
- Report generation and validation (full cycle)
- History tracking during workflow (verify all entries)
- Resource cleanup (temp files, processes, locks)

Expected test count:
- 5-7 happy path tests
- 10-12 unhappy path tests

---

## Implementation Timeline

### Week 1: Critical Foundations
- **Days 1-2**: Phase 1 - Main entry point testing
- **Days 3-5**: Phase 2.1-2.2 - Command and workflow error paths

**Milestone**: `cmd/fluxid` at 90%, error path foundation established

### Week 2: Complete Error Coverage
- **Days 1-2**: Phase 2.3-2.4 - IPC and output error paths
- **Days 3-5**: Phase 3.1-3.2 - IPC history and workflow Run() tests

**Milestone**: All 0% coverage functions tested, error paths comprehensive

### Week 3: Unit Test Completion
- **Days 1-3**: Phase 3.3 - Expand existing command tests to 90%
- **Days 4-5**: Phase 4.1-4.2 - Advanced JSON and YAML output tests

**Milestone**: All packages at 90%+ unit test coverage

### Week 4: Integration & E2E
- **Days 1-2**: Phase 4.3-4.4 - Text formatting and validation tests
- **Days 3-5**: Phase 5 - E2E coverage gaps and chaos tests

**Milestone**: E2E at 100%, chaos tests in place

### Week 5: Polish & Documentation
- **Days 1-2**: Phase 6 - Integration tests
- **Days 3-5**: Test documentation, coverage verification, cleanup

**Milestone**: All goals met, test suite documented

---

## Success Metrics

### Coverage Targets

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Overall line coverage | 80.0% | 90.0% | ❌ |
| `cmd/fluxid` | 0.0% | 90.0% | ❌ |
| `internal/command` | 70.8% | 90.0% | ❌ |
| `internal/workflow` | 72.5% | 90.0% | ❌ |
| `internal/config` | 90.4% | 90.0% | ✅ |
| `internal/ipc` | 90.2% | 90.0% | ✅ |
| `internal/output` | 90.3% | 90.0% | ✅ |
| E2E coverage | 81.2% | 100.0% | ❌ |

### Test Distribution Targets

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Total test functions | 214 | 321 | ❌ (-107) |
| Happy path tests | ~70 (33%) | 107 (33%) | ❌ (-37) |
| Unhappy path tests | ~47 (22%) | 214 (67%) | ❌ (-167) |
| Happy:Unhappy ratio | 3:2 | 1:2 | ❌ |

### Quality Targets

- ✅ All functions have at least 1 test
- ❌ All error paths have dedicated tests
- ❌ All edge cases have dedicated tests
- ✅ All public APIs have happy path tests
- ❌ All public APIs have comprehensive error tests
- ❌ Integration tests cover all cross-component flows
- ❌ E2E tests cover all user workflows

---

## Test Implementation Guidelines

### Happy vs Unhappy Path Distribution

**Rule**: For every 1 happy path test, write 2 unhappy path tests.

**Happy Path Tests**:
- Valid inputs, expected behavior
- Default configurations
- Successful operations end-to-end
- Typical use cases

**Unhappy Path Tests** (2× happy):
- Invalid inputs (malformed, missing, wrong type)
- Boundary conditions (zero, negative, max values)
- Error conditions (file not found, permission denied)
- Edge cases (empty, nil, concurrent access)
- Resource limits (disk full, memory exhausted)
- Timeout scenarios
- Recovery scenarios (retry after failure)

### Test Organization

**File Naming**:
- Unit tests: `{package}_test.go` (e.g., `args_test.go`)
- Error path tests: `error_paths_test.go` or `{component}_error_test.go`
- Integration tests: `{component}_integration_test.go`
- E2E tests: `m{milestone}-e{epic}-{description}_test.go`

**Test Naming**:
- Happy path: `Test{Component}_{Scenario}` (e.g., `TestParseArgs_ValidFlags`)
- Unhappy path: `Test{Component}_{Error}` (e.g., `TestParseArgs_MissingRequiredFlag`)
- Use `t.Run()` for sub-tests to group related scenarios

**Test Structure**:
```go
func TestComponent_Scenario(t *testing.T) {
    t.Parallel() // if safe

    // Arrange
    input := setupTestData()

    // Act
    result, err := ComponentFunction(input)

    // Assert
    if err != nil {
        t.Errorf("expected no error, got: %v", err)
    }
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Coverage Measurement

**Per-commit**:
```bash
go test -cover ./...
```

**Detailed analysis**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

**CI/CD gate**:
- Fail build if overall coverage < 90%
- Fail build if any package < 90%
- Fail build if new code has < 90% coverage

---

## Risk Assessment

### High Risk Areas

1. **`cmd/fluxid/main.go`** - No tests, main entry point
   - **Risk**: Undetected runtime failures in production
   - **Mitigation**: Phase 1 priority, add comprehensive main tests

2. **Error path testing gap** - Only 22% unhappy path coverage
   - **Risk**: Production errors not caught in testing
   - **Mitigation**: Phase 2 priority, comprehensive error path tests

3. **IPC history functions** - 0% coverage on 3 critical functions
   - **Risk**: History feature completely untested
   - **Mitigation**: Phase 3.1 priority, full IPC history test suite

4. **Workflow Run()** - 0% coverage on main orchestration
   - **Risk**: Core workflow logic untested
   - **Mitigation**: Phase 3.2 priority, comprehensive workflow tests

### Medium Risk Areas

1. **E2E coverage gaps** - 81.2% vs 100% target
   - **Risk**: Some user workflows not validated end-to-end
   - **Mitigation**: Phase 5, analyze and fill coverage gaps

2. **Output format edge cases** - Advanced tests deleted
   - **Risk**: Output corruption in edge cases (Unicode, large payloads)
   - **Mitigation**: Phase 4, restore advanced output tests

### Low Risk Areas

1. **Integration test gaps** - Cross-component flows
   - **Risk**: Component integration issues not caught
   - **Mitigation**: Phase 6, add integration tests for critical flows

2. **Test helper utilities** - `test_helpers_workflow.go` deleted
   - **Risk**: Test code duplication, maintenance burden
   - **Mitigation**: Recreate helpers as needed during test development

---

## Appendix: Deleted Test File Mapping

### Tests Moved Successfully

| Old Location | New Location | Status |
|-------------|--------------|--------|
| `cmd/fluxid/args_test.go` | `internal/command/args_test.go` | ✅ Moved |
| `cmd/fluxid/args_error_test.go` | `internal/command/args_error_test.go` | ✅ Moved |
| `cmd/fluxid/config_test.go` | `internal/command/config_test.go` | ✅ Moved |
| `cmd/fluxid/ipc_*_test.go` | `internal/command/ipc_*_test.go` | ✅ Moved |
| `cmd/fluxid/workflow_test.go` | `internal/workflow/workflow_test.go` | ✅ Moved |
| `cmd/fluxid/*_phase_test.go` | `internal/workflow/*_phase_test.go` | ✅ Moved |

### Tests Deleted (Need Restoration)

| Deleted File | New Location | Priority |
|-------------|--------------|----------|
| `main_test.go` | `cmd/fluxid/main_test.go` | CRITICAL |
| `coverage_error_paths_test.go` | Split across packages | CRITICAL |
| `ipc_history_integration_test.go` | `internal/command/ipc_history_test.go` | HIGH |
| `main_ipc_integration_test.go` | `internal/command/main_ipc_integration_test.go` | MEDIUM |
| `coverage_boost_*_test.go` | Split across packages | MEDIUM |
| `output_*_test.go` | `internal/output/*_test.go` | MEDIUM |

---

## Next Steps

1. **Review and approve this plan** with stakeholders
2. **Allocate resources** (developers, timeline)
3. **Set up tracking** (Jira, GitHub issues, project board)
4. **Begin Phase 1** (main entry point testing)
5. **Establish coverage monitoring** (CI/CD integration)
6. **Weekly progress reviews** (coverage metrics, blockers)

---

## Questions for Stakeholders

1. **Timeline**: Is the 5-week timeline acceptable, or do we need to accelerate?
2. **Resources**: How many developers can be allocated to this effort?
3. **Priorities**: Should we adjust phase priorities based on business needs?
4. **Coverage target**: Is 90% the right target, or should we aim higher (95%)?
5. **Test distribution**: Is 1:2 happy:unhappy ratio appropriate for this project?
6. **CI/CD**: Should we block merges that reduce coverage below 90%?

---

**Document Version**: 1.0
**Author**: Test Coverage Analysis (Automated)
**Last Updated**: 2025-12-22
