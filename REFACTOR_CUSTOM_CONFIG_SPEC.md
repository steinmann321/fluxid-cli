# Custom Config Refactoring - Final Specification

**Status:** APPROVED - Ready for Implementation
**Date:** 2025-12-25
**Based on:** REFACTOR_CUSTOM_CONFIG.md + REFACTOR_CUSTOM_CONFIG_CLARIFICATIONS.md

## Core Design Decisions (FINAL)

### 1. Command Paths: Absolute Paths Only (Fail-Safe)

**Decision:** All command file paths MUST be resolved to absolute paths at config load time.

**Rationale:** Eliminates CWD-dependent behavior, makes execution deterministic.

**Implementation:**
```go
// At config load time, convert all relative paths to absolute
func resolveCommandPath(path string, baseDir string) (string, error) {
    if filepath.IsAbs(path) {
        // Already absolute, use as-is
        return path, nil
    }

    // Relative path: resolve to absolute relative to baseDir
    absPath := filepath.Join(baseDir, path)
    absPath, err := filepath.Abs(absPath)
    if err != nil {
        return "", fmt.Errorf("failed to resolve path %s: %w", path, err)
    }

    return absPath, nil
}
```

**Path Resolution Rules:**

| Config Source | Relative Path Base Directory | Example |
|---------------|------------------------------|---------|
| Project config | `<project-root>/.fluxid/` | `.fluxid/config.yaml` → paths relative to `.fluxid/` |
| User config | `~/.fluxid/` | `~/.fluxid/config.yaml` → paths relative to `~/.fluxid/` |
| Custom config | Directory containing config file | `--config=/tmp/cfg.yaml` → paths relative to `/tmp/` |
| CLI flags | Current working directory (CWD) | `--implement-command prompts/impl.md` → CWD |

**After Resolution:**
- All paths stored as absolute in ResolvedConfig
- No CWD dependency during workflow execution
- File existence NOT checked at config load (checked at execution time)

**Test Coverage:**
```go
TestPathResolution_RelativeToAbsolute_ProjectConfig
TestPathResolution_RelativeToAbsolute_UserConfig
TestPathResolution_RelativeToAbsolute_CustomConfig
TestPathResolution_RelativeToAbsolute_CLIFlag
TestPathResolution_AbsolutePathPreserved
TestPathResolution_ParentDirectoryNavigation  // ../prompts/impl.md
```

---

### 2. Default Config Requirement: At Least One Must Exist

**Decision:** Workflow ABORTS if NO default config file exists.

**Load Order (First Found Wins):**
```
1. Project config: <project-root>/.fluxid/config.yaml
2. User config: ~/.fluxid/config.yaml

If NEITHER exists → ERROR, abort
```

**Error Behavior:**
```go
// If no default config found
Error: "no configuration file found. Please create:
  - Project config: .fluxid/config.yaml (recommended)
  - OR User config: ~/.fluxid/config.yaml

Example config:
  agent: claude
  iterations: 20
  implement_retries: 3
  commit_enabled: false
  commands:
    implement: prompts/implement.md
    review: prompts/review.md
    commit: prompts/commit.md"

Exit Code: 1
```

**Rationale:**
- Forces explicit configuration
- Prevents silent fallback to potentially wrong defaults
- Makes configuration visible and version-controllable

**Implementation:**
```go
func LoadDefaultConfig() (*Config, error) {
    // Try project config first
    projectCfg, err := LoadProjectConfig()
    if err == nil && projectCfg != nil {
        return projectCfg, nil
    }

    // Try user config
    userCfg, err := LoadUserConfig()
    if err == nil && userCfg != nil {
        return userCfg, nil
    }

    // Neither found - ERROR
    return nil, ErrNoDefaultConfig
}
```

**Test Coverage:**
```go
TestDefaultConfig_ProjectExists_UserMissing  // Use project
TestDefaultConfig_ProjectMissing_UserExists  // Use user
TestDefaultConfig_BothExist                   // Project wins
TestDefaultConfig_NeitherExists               // ERROR
TestDefaultConfig_ProjectInvalid_UserValid    // Fallback to user? Or error?
```

**Decision Needed:** What if project config exists but is invalid YAML?
- **Option A:** Try user config as fallback
- **Option B:** Error immediately (fail fast)
- **Recommendation:** Option B (fail fast) - invalid config is a bug, don't hide it

---

### 3. Default Values: Hardcoded Defaults Except Commands

**Decision:**
- All scalar settings (iterations, retries, agent, etc.) have **hardcoded defaults**
- Commands section has **NO defaults** - must come from config file
- Full merge strategy (deep merge)

**Hardcoded Defaults:**
```go
const (
    DefaultAgent            = "claude"
    DefaultIterations       = 20
    DefaultImplementRetries = 3
    DefaultCommitEnabled    = false
    DefaultOutputFormat     = "text"
    DefaultDryRun           = false
)

// Commands: NO DEFAULTS
// commands.implement: REQUIRED in config
// commands.review:    REQUIRED in config
// commands.commit:    REQUIRED in config
```

**Config Validation:**
```go
// After merging all configs, validate required fields
func ValidateResolvedConfig(cfg *ResolvedConfig) error {
    if cfg.CommandFiles == nil {
        return errors.New("config: commands section is required")
    }

    if cfg.CommandFiles.ImplementPath == "" {
        return errors.New("config: commands.implement is required")
    }

    if cfg.CommandFiles.ReviewPath == "" {
        return errors.New("config: commands.review is required")
    }

    if cfg.CommandFiles.CommitPath == "" {
        return errors.New("config: commands.commit is required")
    }

    return nil
}
```

**Merge Strategy (Full/Deep Merge):**
```yaml
# Example: Project config (default)
agent: claude
iterations: 20
commands:
  implement: prompts/impl.md
  review: prompts/review.md
  commit: prompts/commit.md

# Custom config (--config)
iterations: 30
commands:
  implement: custom/impl.md
  # review NOT specified
  # commit NOT specified

# RESULT after merge (deep merge):
agent: claude              # From project (inherited)
iterations: 30             # From custom (overridden)
commands:
  implement: custom/impl.md  # From custom (overridden)
  review: prompts/review.md  # From project (inherited)
  commit: prompts/commit.md  # From project (inherited)
```

**Test Coverage:**
```go
TestDefaults_AllScalarSettings        // Verify hardcoded defaults
TestDefaults_CommandsNotProvided      // ERROR if no commands in any config
TestMerge_PartialCommandsSection      // Inherit missing command paths
TestMerge_ScalarOverride              // iterations override
TestMerge_FullDeepMerge               // Complete merge scenario
TestValidation_MissingImplement       // ERROR
TestValidation_MissingReview          // ERROR
TestValidation_MissingCommit          // ERROR
```

---

### 4. Centralized Error Logging

**Decision:** Use centralized error logging with consistent format.

**Error Format Standard:**
```
error: <component>: <description>

Components:
  - config   (configuration loading/validation)
  - args     (CLI argument parsing)
  - workflow (workflow execution)
  - ipc      (inter-process communication)

Examples:
  error: config: no configuration file found
  error: config: commands.implement is required
  error: config: invalid YAML in .fluxid/config.yaml: line 5: unexpected token
  error: args: --config requires a value
  error: args: multiple --config flags specified
  error: workflow: implement phase failed after 3 retries
```

**Implementation:**
```go
// Package errors (new or existing)
package errors

import "fmt"

type ComponentError struct {
    Component string
    Err       error
}

func (e *ComponentError) Error() string {
    return fmt.Sprintf("error: %s: %s", e.Component, e.Err.Error())
}

func NewConfigError(msg string, args ...interface{}) error {
    return &ComponentError{
        Component: "config",
        Err:       fmt.Errorf(msg, args...),
    }
}

func NewArgsError(msg string, args ...interface{}) error {
    return &ComponentError{
        Component: "args",
        Err:       fmt.Errorf(msg, args...),
    }
}

// Usage:
return errors.NewConfigError("no configuration file found")
return errors.NewConfigError("invalid YAML in %s: %w", path, err)
return errors.NewArgsError("--config requires a value")
```

**Centralized Logging:**
```go
// Use standard log package or structured logger
import "log"

// All user-facing errors go through this
func LogError(err error) {
    log.Println(err.Error())
}

// Fatal errors (exit immediately)
func LogFatal(err error) {
    log.Println(err.Error())
    os.Exit(1)
}
```

**Test Coverage:**
```go
TestErrorFormat_Config       // Verify "error: config: ..." format
TestErrorFormat_Args         // Verify "error: args: ..." format
TestErrorLogging_Centralized // Verify all errors go through logger
```

---

### 5. Single --config Flag Only

**Decision:** Only ONE --config flag allowed per invocation.

**Behavior:**
```bash
# Valid
fluxid --config custom.yaml

# INVALID - Error
fluxid --config a.yaml --config b.yaml
```

**Error Message:**
```
error: args: multiple --config flags specified. Use only one --config flag.
```

**Exit Code:** 1

**Implementation:**
```go
func ParseArgs() (*CLIArgs, error) {
    args := &CLIArgs{}
    configCount := 0

    for i := 1; i < len(os.Args); i++ {
        arg := os.Args[i]

        if arg == "--config" || strings.HasPrefix(arg, "--config=") {
            configCount++
            if configCount > 1 {
                return nil, NewArgsError("multiple --config flags specified. Use only one --config flag")
            }

            // Parse --config value
            // ...
        }
    }

    return args, nil
}
```

**Test Coverage:**
```go
TestCLIArgs_SingleConfig_Valid
TestCLIArgs_MultipleConfig_Error
TestCLIArgs_ConfigEqualsAndSpace  // --config=a vs --config a (both valid)
```

---

### 6. Agent Precedence: CLI Flag Same as Config Field

**Decision:** Agent selection via CLI flag follows exact same precedence as any other setting.

**Clarification:** The spec's merge/override behavior applies uniformly to ALL settings including agent.

**Agent Selection Mapping:**
```go
// CLI flags map to config field
--claude   → agent: "claude"
--codex    → agent: "codex"
--opencode → agent: "opencode"

// Then normal precedence applies (same as iterations, retries, etc.)
```

**Precedence (same as all other settings):**
```
1. CLI flag: --claude                    (highest)
2. Custom config: agent: "codex"
3. Project config: agent: "claude"
4. User config: agent: "opencode"
5. Hardcoded default: "claude"           (lowest)
```

**Example:**
```yaml
# User config
agent: opencode

# Project config
agent: codex

# Run with CLI flag
$ fluxid --claude

# Result: agent = "claude" (CLI wins, same as any setting)
```

**Test Coverage:**
```go
TestAgentPrecedence_CLIOverridesConfig
// Config: agent: "codex", CLI: --claude
// Result: agent = "claude"

TestAgentPrecedence_CustomConfigOverridesProject
// Project: agent: "claude", Custom: agent: "codex"
// Result: agent = "codex"

TestAgentPrecedence_DefaultUsed
// No config specified, no CLI flag
// Result: agent = "claude" (hardcoded default)
```

**Important:** Agent flag is NOT special - it follows the same precedence rules as all other settings.

---

### 7. Test Data Organization: Use .tmp/ Folder

**Decision:** All test config files and temporary test data stored in `.tmp/` directory in project root.

**Directory Structure:**
```
<project-root>/
  .tmp/                          # Created by tests, gitignored
    configs/                     # Config files for tests
      valid-complete.yaml
      valid-partial.yaml
      invalid-yaml.yaml
      custom-feature.yaml
    prompts/                     # Test command files
      test-implement.md
      test-review.md
      test-commit.md
    sessions/                    # Test session data
      test-session-123/
```

**Test Implementation:**
```go
func TestCustomConfig_ValidFile(t *testing.T) {
    // Setup: Create test config in .tmp/
    testDir := filepath.Join(projectRoot, ".tmp", "configs")
    os.MkdirAll(testDir, 0755)

    configPath := filepath.Join(testDir, "test-config.yaml")
    configContent := `
agent: claude
iterations: 10
commands:
  implement: prompts/impl.md
  review: prompts/review.md
  commit: prompts/commit.md
`
    os.WriteFile(configPath, []byte(configContent), 0644)
    defer os.Remove(configPath) // Cleanup

    // Test logic...
}
```

**Cleanup Strategy:**
```go
// Option 1: Per-test cleanup
defer os.RemoveAll(testConfigPath)

// Option 2: Test suite cleanup (TestMain)
func TestMain(m *testing.M) {
    code := m.Run()

    // Cleanup all test data
    os.RemoveAll(filepath.Join(projectRoot, ".tmp"))

    os.Exit(code)
}
```

**Rationale:**
- Centralized test data location
- Easy to clean up (single directory)
- Mirrors production structure (.fluxid/ vs .tmp/)
- Gitignored by default (`.tmp/` in .gitignore)

**Update .gitignore:**
```gitignore
# Add to .gitignore
.tmp/
```

**Test Coverage:**
```go
// No specific tests needed for this (implementation detail)
// All E2E tests will use .tmp/ for config files
```

---

### 8. Backward Compatibility: None Required

**Decision:** NO backward compatibility support. Remove all obsolete tests and code.

**Breaking Changes (Acceptable):**
1. Environment variables removed (FLUXID_ITERATIONS, etc.)
2. Default config now REQUIRED (cannot run without config file)
3. Commands must be in config (no defaults)

**Migration Path for Users:**
```
Old (pre-refactoring):
  $ FLUXID_ITERATIONS=10 fluxid
  → Uses env var, no config file needed

New (post-refactoring):
  $ fluxid --fluxid-iterations 10
  → ERROR: no configuration file found

  User must create:
  $ mkdir -p .fluxid
  $ cat > .fluxid/config.yaml <<EOF
  agent: claude
  iterations: 20
  implement_retries: 3
  commit_enabled: false
  commands:
    implement: prompts/implement.md
    review: prompts/review.md
    commit: prompts/commit.md
  EOF

  $ fluxid --fluxid-iterations 10
  → Works (config + CLI override)
```

**Code Removal (Aggressive Cleanup):**
```
DELETE:
  ✗ internal/config/env.go              (entire file)
  ✗ internal/config/env_test.go         (entire file)
  ✗ All env var loading code
  ✗ All env var tests
  ✗ Optional config loading (make required)
  ✗ Any default command paths (if they exist)

KEEP:
  ✓ FLUXID_SESSION_ID (used for IPC, not config)
```

**Test Removal:**
```go
// DELETE these test scenarios:
TestConfig_EnvironmentVariables*        // All env var tests
TestConfig_OptionalConfig*              // Config now required
TestConfig_DefaultCommands*             // Commands now required
TestBackwardCompatibility_*             // No backward compat
```

**Rationale:**
- Clean slate approach
- Simpler codebase (less maintenance)
- Forces explicit configuration (better UX long-term)
- Removes ambiguous behavior (env vars vs config)

**Migration Notice (README/CHANGELOG):**
```markdown
## Breaking Changes in v2.0

### Environment Variables Removed
Environment variables (FLUXID_ITERATIONS, etc.) are no longer supported.
Use CLI flags or config files instead.

**Before:**
```bash
FLUXID_ITERATIONS=10 fluxid
```

**After:**
```bash
fluxid --fluxid-iterations 10
# OR
fluxid --config custom-config.yaml
```

### Config File Now Required
You must have a config file (project or user) before running fluxid.

Create `.fluxid/config.yaml` in your project:
```yaml
agent: claude
iterations: 20
implement_retries: 3
commit_enabled: false
commands:
  implement: prompts/implement.md
  review: prompts/review.md
  commit: prompts/commit.md
```

See [Configuration Guide](docs/configuration.md) for details.
```

**Test Coverage:**
```go
// REMOVE backward compatibility tests
// ADD breaking change tests (verify old behavior fails properly)

TestNoConfigFile_Error
// No .fluxid/config.yaml, no ~/.fluxid/config.yaml
// Expect: Clear error message
// Expect: Exit code 1

TestEnvVarsIgnored_VerifyRemoval
// Set FLUXID_ITERATIONS=99
// Config has iterations: 10
// CLI has no override
// Expect: iterations = 10 (env var has no effect)
```

**Important:** This is a MAJOR version bump (v1.x → v2.0) due to breaking changes.

---

## Updated Configuration Precedence

### Final Precedence Chain (Lowest to Highest)

```
1. Hardcoded Defaults (scalar values only)
   ├─ agent: "claude"
   ├─ iterations: 20
   ├─ implement_retries: 3
   ├─ commit_enabled: false
   └─ Commands: NO DEFAULTS (must come from config)

2. Default Config (REQUIRED - at least one must exist)
   ├─ Try: <project-root>/.fluxid/config.yaml
   └─ Fallback: ~/.fluxid/config.yaml
   └─ If neither: ERROR (abort)

3. Custom Config (OPTIONAL)
   └─ --config <path>

4. CLI Flags (highest precedence)
   ├─ --claude / --codex / --opencode
   ├─ --fluxid-iterations <n>
   ├─ --fluxid-implement-retries <n>
   ├─ --fluxid-commit-enabled / --fluxid-no-commit
   ├─ --implement-command <path>
   ├─ --review-command <path>
   └─ --commit-command <path>
```

**Note:** Environment variables REMOVED completely (except FLUXID_SESSION_ID for IPC).

---

## Configuration Flow Diagram

```
┌─────────────────────────────────────────────────────┐
│ 1. Load Default Config (REQUIRED)                  │
│    ├─ Try: .fluxid/config.yaml                     │
│    └─ Fallback: ~/.fluxid/config.yaml              │
│    └─ Neither exists? → ERROR (abort)              │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│ 2. Apply Hardcoded Defaults (for missing scalars)  │
│    - agent, iterations, retries, commit_enabled    │
│    - Commands: NO defaults (must be in config)     │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│ 3. Validate Required Fields                        │
│    - commands.implement: REQUIRED                   │
│    - commands.review: REQUIRED                      │
│    - commands.commit: REQUIRED                      │
│    - Missing? → ERROR (abort)                       │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│ 4. Resolve All Paths to Absolute                   │
│    - commands.implement → absolute path             │
│    - commands.review → absolute path                │
│    - commands.commit → absolute path                │
│    - Base: config file's directory                  │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│ 5. Load Custom Config (if --config provided)       │
│    - Parse YAML                                     │
│    - Validate structure                             │
│    - Resolve paths to absolute                      │
│    - Deep merge with default config                 │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│ 6. Apply CLI Flags (highest precedence)            │
│    - Override any config values                     │
│    - Resolve command paths to absolute (CWD base)   │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│ 7. Final Validation                                │
│    - All required fields present?                   │
│    - All values valid? (positive numbers, etc.)     │
│    - Ready for workflow execution                   │
└─────────────────────────────────────────────────────┘
```

---

## Updated Test Scenarios

### Unit Tests - Config Loading

```go
// Default config requirement
TestDefaultConfig_ProjectOnly_Success
TestDefaultConfig_UserOnly_Success
TestDefaultConfig_ProjectPriority              // Both exist, project wins
TestDefaultConfig_NeitherExists_Error          // ABORT
TestDefaultConfig_ProjectInvalid_Error         // Fail fast (don't try user)

// Path resolution
TestPathResolution_ProjectConfig_RelativeToAbsolute
TestPathResolution_UserConfig_RelativeToAbsolute
TestPathResolution_CustomConfig_RelativeToAbsolute
TestPathResolution_CLIFlag_RelativeToAbsolute
TestPathResolution_AbsolutePath_NoChange
TestPathResolution_ParentDirectory             // ../prompts/impl.md

// Validation
TestValidation_MissingCommands_Error
TestValidation_MissingImplement_Error
TestValidation_MissingReview_Error
TestValidation_MissingCommit_Error
TestValidation_AllPresent_Success

// Merge strategy
TestMerge_PartialCommands_DeepMerge
TestMerge_ScalarOverride
TestMerge_CommandsOverride
TestMerge_FullChain                            // Default → Custom → CLI

// Hardcoded defaults
TestDefaults_Agent
TestDefaults_Iterations
TestDefaults_ImplementRetries
TestDefaults_CommitEnabled
TestDefaults_NoCommandDefaults                 // Commands NOT defaulted
```

### Unit Tests - CLI Args

```go
TestCLIArgs_ConfigFlag_Valid
TestCLIArgs_ConfigFlag_EqualsFormat           // --config=path
TestCLIArgs_ConfigFlag_SpaceFormat            // --config path
TestCLIArgs_ConfigFlag_MissingValue_Error
TestCLIArgs_ConfigFlag_Multiple_Error         // Only one allowed

TestCLIArgs_CommandFlags_All
TestCLIArgs_ImplementCommand
TestCLIArgs_ReviewCommand
TestCLIArgs_CommitCommand
```

### E2E Tests - Updated Scenarios

```go
// m08-e01: Basic custom config
TestCustomConfig_Complete
TestCustomConfig_Partial_DeepMerge
TestCustomConfig_AbsolutePaths
TestCustomConfig_RelativePaths
TestCustomConfig_MissingFile_Error
TestCustomConfig_InvalidYAML_Error

// m08-e02: Precedence
TestPrecedence_ProjectOnly
TestPrecedence_UserOnly
TestPrecedence_CustomOverridesProject
TestPrecedence_CLIOverridesCustom
TestPrecedence_FullChain

// m08-e03: Default config requirement
TestDefaultConfig_Required_NeitherExists_Error
TestDefaultConfig_ProjectExists
TestDefaultConfig_UserExists
TestDefaultConfig_BothExist_ProjectWins

// m08-e04: Error handling
TestError_MultipleConfigFlags
TestError_NoDefaultConfig
TestError_MissingCommands
TestError_InvalidYAML
TestError_Format_Consistent                    // Verify error: component: msg

// m08-e05: Backward compatibility
TestBackwardCompatibility_ProjectConfigOnly
TestBackwardCompatibility_UserConfigOnly
TestBackwardCompatibility_ExistingWorkflows
```

---

## Implementation Checklist (Updated)

### Phase 1: Centralized Error Handling (NEW FIRST STEP)

**Step 1: Add Error Package (RED)**
- [ ] Create `internal/errors/errors.go`
- [ ] Write tests for ComponentError type
- [ ] Write tests for error format

**Step 2: Implement Error Package (GREEN)**
- [ ] Implement ComponentError
- [ ] Implement NewConfigError, NewArgsError, etc.
- [ ] Add LogError, LogFatal functions

### Phase 2: Default Config Requirement

**Step 3: Update Default Config Loading (RED)**
- [ ] Write test: `TestDefaultConfig_NeitherExists_Error`
- [ ] Write test: `TestDefaultConfig_ProjectPriority`
- [ ] Expect failures (current code loads optionally)

**Step 4: Implement Required Default Config (GREEN)**
- [ ] Update `LoadDefaultConfig()` to require at least one config
- [ ] Return error if neither project nor user config exists
- [ ] Tests pass

### Phase 3: Path Resolution to Absolute

**Step 5: Path Resolution (RED)**
- [ ] Write tests for relative → absolute conversion
- [ ] Write tests for all config sources
- [ ] Expect failures

**Step 6: Implement Path Resolution (GREEN)**
- [ ] Add `resolveCommandPath()` function
- [ ] Update config loading to resolve paths
- [ ] Store absolute paths in ResolvedConfig
- [ ] Tests pass

### Phase 4: Commands Validation (No Defaults)

**Step 7: Commands Validation (RED)**
- [ ] Write tests for missing commands section
- [ ] Write tests for missing individual commands
- [ ] Expect failures

**Step 8: Implement Commands Validation (GREEN)**
- [ ] Remove any command defaults (if they exist)
- [ ] Add validation in ResolveConfig
- [ ] Tests pass

### Phase 5: CLI Flags

**Step 9: Custom Config Flag (RED)**
- [ ] Write tests for --config parsing
- [ ] Write tests for multiple --config error
- [ ] Expect failures

**Step 10: Implement Custom Config Flag (GREEN)**
- [ ] Add --config to CLI parser
- [ ] Add multiple flag check
- [ ] Tests pass

**Step 11: Command Override Flags (RED)**
- [ ] Write tests for --implement-command, --review-command, --commit-command
- [ ] Expect failures

**Step 12: Implement Command Override Flags (GREEN)**
- [ ] Add command flags to CLI parser
- [ ] Resolve paths to absolute (CWD base)
- [ ] Tests pass

### Phase 6: Config Merging

**Step 13: Deep Merge (RED)**
- [ ] Write tests for partial config merge
- [ ] Write tests for full precedence chain
- [ ] Expect failures

**Step 14: Implement Deep Merge (GREEN)**
- [ ] Update ResolveConfig to use deep merge
- [ ] Apply precedence: Default → Custom → CLI
- [ ] Tests pass

### Phase 7: Environment Variable Removal

**Step 15: Remove Env Vars (RED)**
- [ ] Comment out env var tests
- [ ] Expect failures if env vars still loaded

**Step 16: Remove Env Var Support (GREEN)**
- [ ] Remove env var loading code
- [ ] Delete `internal/config/env.go` and `env_test.go`
- [ ] Tests pass

### Phase 8: E2E Tests

**Step 17: E2E Tests (RED)**
- [ ] Write all E2E scenarios (m08-e01 through m08-e05)
- [ ] Expect failures

**Step 18: E2E Tests (GREEN)**
- [ ] Fix integration issues
- [ ] All E2E tests pass

### Phase 9: Documentation & Cleanup

**Step 19: Documentation**
- [ ] Update WORKFLOW.md with new precedence
- [ ] Update README with --config examples
- [ ] Add help text for new flags

**Step 20: Final Validation**
- [ ] All tests pass (unit + E2E)
- [ ] Coverage ≥ 90%
- [ ] No linter warnings
- [ ] Backward compatibility verified

---

## Acceptance Criteria (Updated)

### Functional Requirements
- [ ] Default config REQUIRED (project or user, at least one must exist)
- [ ] All command paths resolved to absolute at load time
- [ ] Commands section REQUIRED in config (no defaults)
- [ ] Scalar settings have hardcoded defaults
- [ ] Deep merge strategy for partial configs
- [ ] Single --config flag enforced (error on multiple)
- [ ] Centralized error logging with consistent format
- [ ] CLI flags override all config sources
- [ ] Environment variables removed (except FLUXID_SESSION_ID)

### Test Requirements
- [ ] All existing tests pass (after updates)
- [ ] 20+ new unit tests
- [ ] 15+ new E2E tests
- [ ] Coverage ≥ 90%
- [ ] All error paths tested
- [ ] Backward compatibility tests pass

### Error Handling
- [ ] All errors use centralized logging
- [ ] Error format: `error: component: description`
- [ ] Clear error messages for all failure modes
- [ ] Exit code 1 for config/usage errors

### Documentation
- [ ] WORKFLOW.md updated with new precedence
- [ ] Help text includes all new flags
- [ ] Example configs provided
- [ ] Migration guide for env var removal

---

## Files to Create/Modify

### New Files
```
internal/errors/errors.go              # Centralized error handling
internal/errors/errors_test.go         # Error tests
e2e-tests/tests/m08-e01-custom-config-basic_test.go
e2e-tests/tests/m08-e02-custom-config-precedence_test.go
e2e-tests/tests/m08-e03-default-config-required_test.go
e2e-tests/tests/m08-e04-error-handling_test.go
.tmp/                                  # Test data directory (gitignored)
```

**Note:** Test config files created dynamically in `.tmp/` during tests (not committed to repo).

### Modified Files
```
internal/config/config.go              # Default config requirement, path resolution
internal/config/config_test.go         # Updated tests
internal/command/args.go               # New CLI flags
internal/command/args_test.go          # CLI flag tests
internal/command/root.go               # Integration
WORKFLOW.md                            # Documentation
.gitignore                             # Add .tmp/ directory
```

### Deleted Files
```
internal/config/env.go                 # Environment variable support
internal/config/env_test.go            # Env var tests
```

---

## Questions Resolved

1. ✅ **Path resolution:** Absolute paths only (fail-safe)
2. ✅ **Default config:** REQUIRED (at least one must exist)
3. ✅ **Default values:** Hardcoded for scalars, NO defaults for commands
4. ✅ **Error handling:** Centralized with consistent format
5. ✅ **Multiple --config:** Single flag only (error on multiple)

## Questions Deferred (Later)

- Symlink handling (edge case, can handle during implementation)
- Help text exact content (documentation phase)
- Dry run behavior with custom config (implementation detail)
- Config file encoding (assume UTF-8, standard YAML)

---

**Status:** READY FOR IMPLEMENTATION

This specification is complete and unambiguous for the approved decisions.
Deferred questions can be addressed during implementation or in future iterations.
