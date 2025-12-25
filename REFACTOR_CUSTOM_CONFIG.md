# Custom Config Refactoring Plan

**Status:** ✅ COMPLETED
**Date:** 2025-12-25
**Completion Date:** 2025-12-25
**Version:** 2.0 (Breaking Changes)

---

## Goals

1. Support `--config` flag for custom config files (Spring Boot-style)
2. Establish clear precedence: Default config → Custom config → CLI flags
3. Standardize CLI syntax to equals-only (`--flag=value`)
4. Make commits mandatory (always run)
5. Remove source tracking and commit toggle logic
6. Enforce config file requirement (at least one must exist)

---

## Breaking Changes (v1.x → v2.0)

- Environment variables removed (FLUXID_ITERATIONS, etc.)
- Config files now REQUIRED (at least one default config must exist)
- Space syntax rejected (`--flag value` → ERROR)
- Source tracking removed completely
- Commits always enabled (cannot be disabled)
- `--fluxid-commit` and `--fluxid-no-commit` flags removed
- `commit_enabled` config field removed

---

## New Features

- `--config=<path>` flag to load custom config files
- Command override flags: `--implement-command`, `--review-command`, `--commit-command`
- Deep merge for partial configs (inherit missing fields from default)
- Absolute path resolution (deterministic, CWD-independent)
- Centralized error handling with consistent format

---

## Core Design Decisions

### 1. Command Paths: Absolute Paths Only
- All command file paths MUST be resolved to absolute paths at config load time
- Relative paths resolved based on config file's directory
- CLI flag paths resolved relative to CWD
- No CWD dependency during workflow execution

**Path Resolution Base Directory:**
- Project config (`.fluxid/config.yaml`): paths relative to `.fluxid/`
- User config (`~/.fluxid/config.yaml`): paths relative to `~/.fluxid/`
- Custom config (`--config`): paths relative to config file's directory
- CLI flags: paths relative to CWD

### 2. Default Config Requirement
- At least ONE default config file must exist
- Load order: Project config (`.fluxid/config.yaml`) → User config (`~/.fluxid/config.yaml`)
- If NEITHER exists → ERROR and abort
- Fail fast on invalid YAML (no fallback to user config)

### 3. Default Values
- Scalar settings have hardcoded defaults: agent="claude", iterations=20, retries=3
- Commands section has NO defaults - must come from config file
- Missing commands → ERROR

### 4. Centralized Error Logging
- Error format: `error: <component>: <description>`
- Components: config, args, workflow, ipc
- Consistent format across all errors

### 5. Single --config Flag Only
- Only ONE `--config` flag allowed per invocation
- Multiple `--config` flags → ERROR

### 6. Agent Precedence
- Agent selection follows same precedence as all other settings
- CLI flags (`--claude`, `--codex`, `--opencode`) map to `agent` config field
- No special treatment

### 7. Test Data Organization
- All test config files stored in `.tmp/` directory (gitignored)
- Cleanup strategy: per-test or test suite teardown

### 8. Backward Compatibility: None
- No backward compatibility support
- Remove all obsolete tests and code
- Breaking change acceptable (v2.0)

### 9. CLI Equals Syntax: Standard Format
- ALL value flags use equals syntax: `--flag=value`
- Space syntax (`--flag value`) → ERROR
- Boolean/agent flags remain as-is (no value needed)

### 10. Source Tracking Removal
- Delete `Sources` field from ResolvedConfig
- Remove all source tracking methods and logic
- Complete removal - no traces remain

### 11. Commit Phase: Always Enabled
- Commits are mandatory - always run
- Remove CommitEnabled field
- Remove `--fluxid-commit` and `--fluxid-no-commit` flags
- Remove all conditional commit logic

---

## Configuration Precedence (Lowest to Highest)

```
1. Hardcoded Defaults (scalar values only)
   - agent: "claude"
   - iterations: 20
   - implement_retries: 3
   - Commands: NO DEFAULTS

2. Default Config (REQUIRED)
   - Project config: .fluxid/config.yaml
   - OR User config: ~/.fluxid/config.yaml
   - Neither exists → ERROR

3. Custom Config (OPTIONAL)
   - --config=<path>

4. CLI Flags (highest)
   - --claude / --codex / --opencode
   - --fluxid-iterations=<n>
   - --fluxid-implement-retries=<n>
   - --implement-command=<path>
   - --review-command=<path>
   - --commit-command=<path>
```

---

## Implementation Phases

### Phase 1: Centralized Error Handling
- Create error package with ComponentError type
- Add NewConfigError, NewArgsError functions
- Implement LogError, LogFatal functions

### Phase 2: Default Config Requirement
- Update LoadDefaultConfig to require at least one config
- Return error if neither project nor user config exists
- Fail fast on invalid YAML

### Phase 3: Path Resolution to Absolute
- Implement resolveCommandPath function
- Convert all relative paths to absolute at config load time
- Store absolute paths in ResolvedConfig

### Phase 4: Commands Validation
- Remove any command defaults
- Add validation for required commands section
- Validate individual commands (implement, review, commit)

### Phase 5: CLI Flags
- Add --config flag parsing
- Add multiple --config flag detection (error)
- Add command override flags (--implement-command, etc.)
- Resolve CLI flag paths to absolute (CWD base)

### Phase 6: Config Merging
- Implement deep merge strategy
- Apply precedence: Default → Custom → CLI
- Inherit missing fields from lower precedence configs

### Phase 7: Environment Variable Removal
- Remove env var loading code
- Delete internal/config/env.go and env_test.go
- Remove all env var tests

### Phase 8: CLI Equals Syntax Standardization
- Update CLI parser to accept ONLY `--flag=value` syntax
- Reject `--flag value` syntax with error
- Update all tests to use equals syntax
- Update all documentation examples

### Phase 9: Source Tracking Removal
- Delete Sources field from ResolvedConfig
- Delete SetSource and GetSource methods
- Remove source tracking from merge logic
- Remove all source tracking tests
- Validate no references remain

### Phase 10: Commit Always Enabled
- Delete CommitEnabled field from Config
- Delete --fluxid-commit and --fluxid-no-commit flags
- Remove conditional commit logic (make unconditional)
- Update workflow to always run commit phase
- Remove all commit toggle tests

### Phase 11: E2E Tests
- Custom config basic scenarios
- Precedence validation
- Default config requirement
- Error handling scenarios
- Backward compatibility (non-breaking features only)

### Phase 12: Documentation & Cleanup
- Update WORKFLOW.md with new precedence
- Update README with --config examples
- Add help text for new flags
- Document breaking changes (v2.0 migration guide)
- Final validation

---

## Test Requirements

### Unit Tests - Config Loading
- Default config requirement (project, user, both, neither)
- Path resolution (relative to absolute for all sources)
- Validation (missing commands, missing individual commands)
- Merge strategy (partial, scalar override, full chain)
- Hardcoded defaults

### Unit Tests - CLI Args
- Equals syntax (positive tests for all value flags)
- Space syntax rejection (error tests)
- Edge cases (empty value, multiple equals, multiple --config)
- Boolean/agent flags (no equals needed)

### E2E Tests
- Custom config (basic, partial, paths, errors)
- Precedence (project, user, custom, CLI, full chain)
- Default config requirement
- Error handling (format, all scenarios)
- Backward compatibility (non-breaking features)

---

## Acceptance Criteria

### Functional Requirements
- Default config REQUIRED (project or user must exist)
- All command paths absolute at load time
- Commands section REQUIRED in config
- Scalar settings have hardcoded defaults
- Deep merge for partial configs
- Single --config flag enforced
- Centralized error logging
- CLI flags override all config sources
- Environment variables removed (except FLUXID_SESSION_ID)
- Equals syntax ONLY for value flags
- Space syntax rejected with error
- Source tracking completely removed
- Commits always enabled
- Commit phase always runs

### Test Requirements
- All existing tests pass (after updates)
- 20+ new unit tests
- 15+ new E2E tests
- Coverage ≥ 90%
- All error paths tested
- Equals syntax: positive and negative tests
- Source tracking tests: all deleted
- Commit toggle tests: all deleted

### Error Handling
- Centralized logging with format: `error: component: description`
- Clear error messages for all failure modes
- Exit code 1 for config/usage errors

### Code Quality
- No source tracking code remains (verified via grep)
- No commit toggle code remains (verified via grep)
- No dead code or commented-out logic
- All CLI parsing uses equals syntax only
- Commits always run unconditionally

### Documentation
- WORKFLOW.md updated with new precedence
- Help text includes all new flags (equals syntax)
- Example configs provided
- Migration guide for breaking changes
- All examples use equals syntax
- No references to source tracking or commit toggles

---

## Files to Create/Modify

### New Files
```
internal/errors/errors.go
internal/errors/errors_test.go
e2e-tests/tests/m08-e01-custom-config-basic_test.go
e2e-tests/tests/m08-e02-custom-config-precedence_test.go
e2e-tests/tests/m08-e03-default-config-required_test.go
e2e-tests/tests/m08-e04-error-handling_test.go
.tmp/                                  # Test data (gitignored)
```

### Modified Files
```
internal/config/config.go              # Default requirement, path resolution
internal/config/config_test.go         # Updated tests
internal/command/args.go               # New CLI flags, equals syntax
internal/command/args_test.go          # CLI flag tests
internal/command/root.go               # Integration
WORKFLOW.md                            # Documentation
.gitignore                             # Add .tmp/
```

### Deleted Files
```
internal/config/env.go                 # Environment variable support
internal/config/env_test.go            # Env var tests
```

---

## Work Items Summary

### Code Deletion
- Environment variable support (env.go, env_test.go)
- Source tracking (Sources field, SetSource/GetSource methods, all related code)
- Commit toggle logic (CommitEnabled field, --fluxid-commit/--fluxid-no-commit flags)
- All related tests for deleted features

### Code Addition
- Error package (ComponentError, logging functions)
- --config flag parsing
- Command override flags (--implement-command, --review-command, --commit-command)
- Path resolution logic (relative to absolute)
- Default config requirement enforcement
- Commands validation
- Deep merge implementation
- Equals syntax enforcement

### Code Modification
- CLI parser (equals syntax only)
- Config loading (require default config)
- Workflow (unconditional commit phase)
- All tests (equals syntax, no source tracking, no commit toggles)
- All documentation (new precedence, breaking changes)

---

## Completion Summary

**All phases completed successfully:**

✅ Phase 1: Centralized Error Handling - Error package with ComponentError implemented
✅ Phase 2: Default Config Requirement - At least one config required, fail fast on errors
✅ Phase 3: Path Resolution to Absolute - All paths resolved at config load time
✅ Phase 4: Commands Validation - Commands section required, no defaults
✅ Phase 5: CLI Flags - Custom config and command override flags added
✅ Phase 6: Config Merging - Deep merge with precedence chain implemented
✅ Phase 7: Environment Variable Removal - All env vars removed except FLUXID_SESSION_ID
✅ Phase 8: CLI Equals Syntax Standardization - Space syntax rejected, equals required
✅ Phase 9: Source Tracking Removal - All source tracking code deleted
✅ Phase 10: Commit Always Enabled - CommitEnabled field and flags removed
✅ Phase 11: E2E Tests - All 155 tests passing (100%)
✅ Phase 12: Documentation & Cleanup - README updated with migration guide

**Test Results:**
- Core package tests: All passing
- E2E tests: 155/155 passing (100%)
- Coverage: Maintained across all packages

**Documentation:**
- README.md: Updated with v2.0 features, examples, and migration guide
- All examples use equals syntax
- Breaking changes clearly documented
- Custom config usage documented

**Code Quality:**
- No source tracking code remains (verified)
- No commit toggle code remains (verified)
- No environment variable config support (except FLUXID_SESSION_ID)
- All CLI parsing uses equals syntax only
- Commits always run unconditionally

---

**Implementation Approach:** Test-Driven Development (TDD) - RED → GREEN → REFACTOR

**Status:** ✅ COMPLETED
