# Refactoring Plan: Custom Config File Support

**Date:** 2025-12-25
**Objective:** Add `--config` flag support with Spring Boot-style configuration precedence
**Approach:** Test-Driven Development (TDD) - Red → Green → Refactor

## Executive Summary

Add support for custom configuration files via `--config` flag, remove environment variable configuration, and add CLI flags for command file paths. Configuration follows clear precedence: Defaults → Custom Config → CLI Flags.

## Current State (Before)

### Configuration Precedence
```
1. Built-in defaults
2. Home config (~/.fluxid/config.yaml)
3. Project config (./.fluxid/config.yaml) - overrides home
4. Environment variables (FLUXID_ITERATIONS, etc.) - overrides files
5. CLI flags - highest precedence
```

### Available CLI Flags
- `--claude / --codex / --opencode` (agent selection)
- `--fluxid-iterations <n>`
- `--fluxid-implement-retries <n>`
- `--fluxid-commit-enabled / --fluxid-no-commit`
- `--fluxid-dry-run / --dry-run`
- `--fluxid-output / --output <format>`

### Config File Structure
```yaml
agent: claude
iterations: 20
implement_retries: 3
commit_enabled: false
commands:
  implement: path/to/implement.md
  review: path/to/review.md
  commit: path/to/commit.md
```

### Environment Variables (TO BE REMOVED)
- `FLUXID_ITERATIONS`
- `FLUXID_IMPLEMENT_RETRIES`
- `FLUXID_COMMIT_ENABLED`
- `FLUXID_SESSION_ID` (keep - used for IPC, not config)

## Target State (After)

### Configuration Precedence (CHANGED)
```
1. Built-in defaults
2. Home config (~/.fluxid/config.yaml)
3. Project config (./.fluxid/config.yaml) - overrides home
4. Custom config (--config flag) - NEW, overrides project
5. CLI flags - highest precedence, ALL settings overridable
```

### New CLI Flags
- `--config <path>` - Load custom config file
- `--implement-command <path>` - Override implement command file
- `--review-command <path>` - Override review command file
- `--commit-command <path>` - Override commit command file

### Key Design Decisions
1. **Partial configs allowed:** Custom config can specify only what it wants to override
2. **Environment variables removed:** Cleaner precedence, explicit configuration
3. **Relative paths in custom config:** Resolved relative to config file's directory
4. **Backward compatibility:** Existing configs continue to work unchanged

## Test-Driven Development Strategy

### Phase 1: RED - Write Failing Tests
Identify and update all affected tests, write new test scenarios (all failing initially)

### Phase 2: GREEN - Minimal Implementation
Implement just enough to make tests pass

### Phase 3: REFACTOR - Clean Up
Refactor implementation while keeping tests green

## Affected Tests Analysis

### A. Unit Tests - Config Resolution

#### File: `internal/config/config_test.go`

**Existing Tests to Update:**
1. `TestLoadHomeConfig` - No changes needed
2. `TestLoadProjectConfig` - No changes needed
3. `TestResolveConfig` - **MAJOR CHANGES** - Remove env var tests
4. Tests using `t.Setenv()` - **REMOVE ALL** env var test scenarios

**New Tests Required:**
1. `TestLoadCustomConfig` - Load config from arbitrary path
   - Valid config file
   - Missing file (error)
   - Invalid YAML (error)
   - Partial config (valid)
   - Complete config (valid)

2. `TestResolveConfigWithCustomConfig` - Precedence testing
   - Custom config overrides project config
   - Custom config overrides home config
   - CLI flags override custom config
   - All precedence combinations

3. `TestRelativePathResolution` - Command file path resolution
   - Paths in custom config resolved relative to config file location
   - Paths in home config resolved relative to home directory
   - Paths in project config resolved relative to project directory

4. `TestCustomConfigValidation` - Edge cases
   - Empty custom config file
   - Custom config with only commands section
   - Custom config with only scalar values
   - Custom config with invalid agent name
   - Custom config with negative/zero values

**Tests to Remove:**
- All environment variable tests in `TestResolveConfig`
- Any tests using `FLUXID_ITERATIONS`, `FLUXID_IMPLEMENT_RETRIES`, `FLUXID_COMMIT_ENABLED`

#### File: `internal/config/env_test.go`

**Action:** **DELETE ENTIRE FILE** - No environment variable config support

### B. Unit Tests - CLI Argument Parsing

#### File: `internal/command/args_test.go` (if exists) or create new

**New Tests Required:**
1. `TestParseArgs_ConfigFlag`
   - `--config /path/to/config.yaml`
   - `--config=/path/to/config.yaml` (equals syntax)
   - Missing value (error)
   - Multiple --config flags (last one wins? or error?)

2. `TestParseArgs_CommandFlags`
   - `--implement-command path/to/impl.md`
   - `--review-command path/to/review.md`
   - `--commit-command path/to/commit.md`
   - All three together
   - Missing values (error)

3. `TestParseArgs_CommandFlagPrecedence`
   - Command flags override config file values
   - Command flags work without --config flag

### C. Unit Tests - Command Execution

#### File: `internal/command/root_test.go` (if exists)

**Tests to Update:**
- Any tests that set environment variables - **REMOVE** env var usage
- Any tests checking config source tracking - Update to remove "env" source

**New Tests Required:**
1. `TestExecuteWithCustomConfig`
   - Execute with --config flag
   - Verify correct config values are used
   - Verify command files are loaded from correct paths

2. `TestExecuteWithMixedConfig`
   - Custom config + CLI flag overrides
   - Verify precedence is correct

### D. E2E Tests - Configuration Scenarios

#### File: `e2e-tests/tests/m08-e01-custom-config-basic_test.go` (NEW)

**Test Scenarios:**

1. `TestCustomConfig_CompleteConfig`
   ```go
   // Create custom config with all settings
   // Run: fluxid --config custom.yaml
   // Verify: All settings from custom config are used
   ```

2. `TestCustomConfig_PartialConfig`
   ```go
   // Create custom config with only iterations and commands
   // Run: fluxid --config custom.yaml
   // Verify: Custom values used, other values from defaults
   ```

3. `TestCustomConfig_WithCLIOverride`
   ```go
   // Create custom config with iterations=10
   // Run: fluxid --config custom.yaml --fluxid-iterations 20
   // Verify: iterations=20 (CLI wins)
   ```

4. `TestCustomConfig_CommandFileOverride`
   ```go
   // Create custom config with implement command
   // Run: fluxid --config custom.yaml --implement-command other.md
   // Verify: other.md is used, not config file value
   ```

5. `TestCustomConfig_RelativePaths`
   ```go
   // Create config in subdir with relative command paths
   // Run: fluxid --config subdir/config.yaml
   // Verify: Command files resolved relative to subdir
   ```

6. `TestCustomConfig_MissingFile`
   ```go
   // Run: fluxid --config /nonexistent/config.yaml
   // Verify: Error message, non-zero exit code
   ```

7. `TestCustomConfig_InvalidYAML`
   ```go
   // Create config with invalid YAML syntax
   // Run: fluxid --config invalid.yaml
   // Verify: Clear error message about YAML parsing
   ```

#### File: `e2e-tests/tests/m08-e02-custom-config-precedence_test.go` (NEW)

**Test Scenarios:**

1. `TestConfigPrecedence_DefaultOnly`
   ```go
   // Setup: Only ~/.fluxid/config.yaml
   // Run: fluxid
   // Verify: Home config values used
   ```

2. `TestConfigPrecedence_ProjectOverridesHome`
   ```go
   // Setup: Home config + Project config
   // Run: fluxid
   // Verify: Project config wins for overlapping keys
   ```

3. `TestConfigPrecedence_CustomOverridesProject`
   ```go
   // Setup: Home + Project + Custom config
   // Run: fluxid --config custom.yaml
   // Verify: Custom wins for overlapping keys
   ```

4. `TestConfigPrecedence_CLIOverridesAll`
   ```go
   // Setup: Home + Project + Custom config
   // Run: fluxid --config custom.yaml --fluxid-iterations 99
   // Verify: CLI value (99) used, custom for other values
   ```

5. `TestConfigPrecedence_FullChain`
   ```go
   // Setup: All config sources with different values
   // Verify: Correct precedence for each setting:
   //   - iterations: CLI (highest)
   //   - retries: custom config
   //   - commit: project config
   //   - agent: home config
   //   - output: default (lowest)
   ```

#### File: `e2e-tests/tests/m08-e03-custom-config-workflows_test.go` (NEW)

**Test Scenarios:**

1. `TestMultiStageWorkflow_DifferentConfigs`
   ```go
   // Create 3 config files: feature.yaml, docs.yaml, tests.yaml
   // Run workflow:
   //   1. fluxid --config feature.yaml
   //   2. fluxid --config docs.yaml
   //   3. fluxid --config tests.yaml
   // Verify: Each stage uses correct config
   ```

2. `TestMultiStageWorkflow_SharedBaseCustomOverride`
   ```go
   // Setup: Project config as base
   // Run workflow:
   //   1. fluxid --implement-command prompts/phase1.md
   //   2. fluxid --implement-command prompts/phase2.md
   // Verify: Same base config, different commands per phase
   ```

3. `TestCustomConfig_CommandFileResolution`
   ```go
   // Create config at: workflows/feature/config.yaml
   // Config contains: commands.implement: "../prompts/impl.md"
   // Verify: Path resolved correctly relative to config file
   ```

#### File: `e2e-tests/tests/m08-e04-env-vars-removed_test.go` (NEW)

**Test Scenarios:**

1. `TestEnvVarsIgnored_Iterations`
   ```go
   // Setup: Project config with iterations=10
   // Set: FLUXID_ITERATIONS=99 (should be ignored)
   // Run: fluxid
   // Verify: iterations=10 (not 99), env var has no effect
   ```

2. `TestEnvVarsIgnored_AllSettings`
   ```go
   // Set all old env vars: FLUXID_ITERATIONS, FLUXID_IMPLEMENT_RETRIES, etc.
   // Run: fluxid
   // Verify: All settings use config/defaults, env vars ignored
   ```

3. `TestSessionIDEnvStillWorks`
   ```go
   // Set: FLUXID_SESSION_ID=test-session-123
   // Run: fluxid
   // Verify: Session ID is used (this env var is for IPC, not config)
   ```

### E. E2E Tests - Existing Tests to Update

#### File: `e2e-tests/tests/m01-e01-basic-run_test.go` (and similar)

**Changes Required:**
- Remove any `t.Setenv()` calls for config env vars
- If tests relied on env vars, update to use config files or CLI flags

#### File: `e2e-tests/tests/m07-e01-implement-retries-exhausted-continues-to-review_test.go`

**Changes Required:**
- Review if any env var usage exists
- Update if necessary

### F. Integration Tests - Config Resolution Flow

#### File: `internal/config/integration_test.go` (NEW)

**Test Scenarios:**

1. `TestFullConfigResolution_AllSources`
   ```go
   // Setup: Home, Project, Custom configs with different values
   // Add: CLI flags for some values
   // Verify: Final resolved config has correct precedence
   // Verify: Source tracking shows where each value came from
   ```

2. `TestConfigResolution_MissingCustomConfig`
   ```go
   // Run with --config pointing to missing file
   // Verify: Appropriate error, no panic
   ```

3. `TestConfigResolution_CommandFilePathResolution`
   ```go
   // Test all path resolution scenarios:
   //   - Absolute paths (used as-is)
   //   - Relative in home config (relative to home dir)
   //   - Relative in project config (relative to project dir)
   //   - Relative in custom config (relative to config file dir)
   ```

## Test Coverage Requirements

### Unit Test Coverage
- **Target:** 90%+ coverage for all modified files
- **Critical paths:**
  - Config loading (all sources)
  - Config merging (precedence rules)
  - CLI flag parsing (new flags)
  - Path resolution (relative paths)

### E2E Test Coverage
- **Minimum:** 15 new E2E tests covering:
  - Basic custom config usage (5 tests)
  - Precedence scenarios (5 tests)
  - Multi-stage workflows (3 tests)
  - Env var removal verification (3 tests)

### Edge Cases to Cover
1. Empty config files
2. Config files with only some fields
3. Invalid YAML syntax
4. Missing config files
5. Invalid file paths
6. Circular path references (if relative paths used)
7. Absolute vs relative path mixing
8. Unicode/special characters in paths
9. Very long config values
10. Concurrent config file reads (if applicable)

## Implementation Checklist (TDD Order)

### Step 1: Environment Variable Removal (RED)
- [ ] Identify all env var tests
- [ ] Mark env var tests as failing (comment out env var setup)
- [ ] Delete `internal/config/env_test.go`
- [ ] Update any E2E tests using env vars
- [ ] Run tests - expect failures

### Step 2: Environment Variable Removal (GREEN)
- [ ] Remove env var loading from `internal/config/env.go`
- [ ] Remove env var precedence from config resolution
- [ ] Update source tracking (remove "env" source)
- [ ] Run tests - env var tests should be gone, others pass

### Step 3: Custom Config Flag - CLI Parsing (RED)
- [ ] Write tests for `--config` flag parsing
- [ ] Write tests for command file flags parsing
- [ ] Run tests - expect failures (flags don't exist)

### Step 4: Custom Config Flag - CLI Parsing (GREEN)
- [ ] Add `--config` to CLI flag parser
- [ ] Add `--implement-command`, `--review-command`, `--commit-command`
- [ ] Update `CLIArgs` struct
- [ ] Run tests - parser tests pass

### Step 5: Custom Config Loading (RED)
- [ ] Write `TestLoadCustomConfig` scenarios
- [ ] Write path resolution tests
- [ ] Run tests - expect failures (loader doesn't exist)

### Step 6: Custom Config Loading (GREEN)
- [ ] Implement `LoadCustomConfig(path)` function
- [ ] Implement relative path resolution for config file directory
- [ ] Run tests - loader tests pass

### Step 7: Config Precedence Integration (RED)
- [ ] Write `TestResolveConfigWithCustomConfig` scenarios
- [ ] Write full precedence chain tests
- [ ] Run tests - expect failures

### Step 8: Config Precedence Integration (GREEN)
- [ ] Update `ResolveConfig()` to include custom config step
- [ ] Update precedence logic: Defaults → Home → Project → Custom → CLI
- [ ] Update source tracking for custom config
- [ ] Run tests - precedence tests pass

### Step 9: E2E Tests (RED)
- [ ] Write all E2E test scenarios (m08-e01, m08-e02, m08-e03, m08-e04)
- [ ] Run tests - expect failures

### Step 10: E2E Tests (GREEN)
- [ ] Fix any integration issues discovered by E2E tests
- [ ] Ensure all E2E scenarios pass
- [ ] Run full test suite - all tests pass

### Step 11: Refactor
- [ ] Clean up code duplication
- [ ] Improve error messages
- [ ] Add documentation comments
- [ ] Update WORKFLOW.md with new config precedence
- [ ] Run tests - all still pass

## Acceptance Criteria

### Functional Requirements
- [ ] `--config` flag loads custom config file
- [ ] Custom config can be partial (only override some fields)
- [ ] Precedence chain works correctly (Defaults → Home → Project → Custom → CLI)
- [ ] Command file paths in custom config resolve relative to config file location
- [ ] All CLI flags can override any config setting
- [ ] Environment variables no longer affect configuration (except FLUXID_SESSION_ID)
- [ ] Clear error messages for invalid config files
- [ ] Backward compatibility: existing configs work unchanged

### Test Requirements
- [ ] All existing tests pass (after updates)
- [ ] 15+ new E2E tests covering custom config scenarios
- [ ] 90%+ unit test coverage for modified files
- [ ] No regressions in existing functionality
- [ ] All edge cases covered by tests

### Documentation Requirements
- [ ] WORKFLOW.md updated with new config precedence
- [ ] README examples showing --config usage
- [ ] Config file examples in docs/examples/ directory
- [ ] Help text (`--help`) includes new flags
- [ ] Error messages are clear and actionable

## Files to Create

### New Test Files
```
e2e-tests/tests/m08-e01-custom-config-basic_test.go
e2e-tests/tests/m08-e02-custom-config-precedence_test.go
e2e-tests/tests/m08-e03-custom-config-workflows_test.go
e2e-tests/tests/m08-e04-env-vars-removed_test.go
internal/config/integration_test.go
internal/config/custom_config_test.go (if splitting tests)
```

### New Documentation
```
docs/examples/custom-config-basic.yaml
docs/examples/custom-config-feature.yaml
docs/examples/custom-config-docs.yaml
docs/examples/multi-stage-workflow.sh
```

## Files to Modify

### Implementation Files (discovered via test failures)
```
internal/config/config.go        - Add LoadCustomConfig, update precedence
internal/config/env.go           - Delete or remove env var loading
internal/command/args.go         - Add new CLI flags
internal/command/root.go         - Integrate custom config loading
internal/types/config.go         - Add CustomConfigPath field if needed
WORKFLOW.md                      - Update config precedence section
```

### Test Files to Update
```
internal/config/config_test.go   - Remove env var tests, add custom config tests
e2e-tests/tests/m01-e01-*.go     - Remove env var usage if any
e2e-tests/tests/m07-e01-*.go     - Remove env var usage if any
```

### Files to Delete
```
internal/config/env_test.go      - Complete removal
internal/config/env.go           - If only contains env var loading
```

## Risk Analysis

### High Risk Areas
1. **Config precedence logic** - Complex merging, easy to get wrong
2. **Path resolution** - Relative paths can be tricky, platform differences
3. **Backward compatibility** - Must not break existing configs
4. **Error handling** - Invalid configs should fail gracefully

### Mitigation Strategies
1. **Comprehensive test coverage** - Especially precedence and edge cases
2. **Property-based testing** - Consider using fuzzing for config parsing
3. **Integration tests** - Test full workflow end-to-end
4. **Gradual rollout** - Feature flag for custom config initially?

## Success Metrics

### Before Refactoring
- Test count: ~77 tests
- Coverage: ~91.3%
- Config sources: 5 (defaults, home, project, env, CLI)
- CLI flags: 7

### After Refactoring (Target)
- Test count: ~95+ tests (18+ new tests)
- Coverage: 90%+ (maintained or improved)
- Config sources: 4 (defaults, home, project, custom, CLI) - env removed
- CLI flags: 11 (added: --config, --implement-command, --review-command, --commit-command)

### Quality Gates
- [ ] All tests pass
- [ ] No decrease in code coverage
- [ ] No new linter warnings
- [ ] No new security issues (gosec, govulncheck)
- [ ] Documentation updated
- [ ] Example configs provided

## Timeline Estimate

Using TDD approach with full test coverage:

1. **Phase 1 - Env Var Removal:** 2-3 hours
   - Update/delete tests
   - Remove env var loading code

2. **Phase 2 - CLI Flags:** 2-3 hours
   - Add new flags
   - Update parser tests

3. **Phase 3 - Custom Config Loading:** 3-4 hours
   - Implement loader
   - Path resolution logic
   - Unit tests

4. **Phase 4 - Precedence Integration:** 3-4 hours
   - Update ResolveConfig
   - Integration tests

5. **Phase 5 - E2E Tests:** 4-6 hours
   - Write 15+ E2E scenarios
   - Debug integration issues

6. **Phase 6 - Documentation & Cleanup:** 2-3 hours
   - Update docs
   - Refactor code
   - Final review

**Total:** 16-23 hours of focused development

## Getting Started (Next Session)

1. Read this plan thoroughly
2. Run existing tests to establish baseline: `go test ./...`
3. Start with **Step 1: Environment Variable Removal (RED)**
4. Follow TDD cycle strictly: Red → Green → Refactor
5. Commit after each GREEN phase
6. Keep test coverage high throughout

## Questions to Resolve Before Starting

1. Should `--config` flag error if file doesn't exist, or silently ignore?
   - **Recommendation:** Error with clear message (explicit is better)

2. Should multiple `--config` flags be allowed? If so, which wins?
   - **Recommendation:** Error on multiple --config flags (avoid confusion)

3. Should we validate that custom config contains valid paths before execution?
   - **Recommendation:** Yes, fail early with helpful error messages

4. Path resolution strategy for absolute paths in config files?
   - **Recommendation:** Use as-is (don't modify absolute paths)

5. Should we add a `--config-help` flag to print config file schema?
   - **Recommendation:** Nice-to-have, not required for MVP

---

**Ready to start TDD implementation in next session!**
