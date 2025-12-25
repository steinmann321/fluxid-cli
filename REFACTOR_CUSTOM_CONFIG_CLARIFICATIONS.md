# Refactoring Plan Clarifications

**Purpose:** Resolve ambiguities and unclear points in `REFACTOR_CUSTOM_CONFIG.md` to ensure correct implementation.

## Critical Ambiguities Requiring Resolution

### 1. Path Resolution - Complete Specification Needed

**Problem:** The plan states "relative paths in custom config resolved relative to config file's directory" but lacks complete specification.

**Missing Details:**
- How are relative paths in HOME config resolved?
- How are relative paths in PROJECT config resolved?
- How are relative paths in CLI flags resolved?
- What happens with absolute paths in any config?

**Proposed Resolution:**
```
Path Resolution Rules (COMPLETE):

1. Absolute paths (start with / or C:\ etc.):
   - Used as-is, no resolution needed

2. Relative paths in HOME config (~/.fluxid/config.yaml):
   - Resolved relative to: ~/.fluxid/
   - Example: "prompts/impl.md" → "~/.fluxid/prompts/impl.md"

3. Relative paths in PROJECT config (./.fluxid/config.yaml):
   - Resolved relative to: ./.fluxid/
   - Example: "prompts/impl.md" → "./.fluxid/prompts/impl.md"

4. Relative paths in CUSTOM config (--config /path/to/custom.yaml):
   - Resolved relative to: directory containing custom.yaml
   - Example:
     --config=/home/user/configs/feature.yaml
     commands.implement: "../prompts/impl.md"
     Resolves to: /home/user/prompts/impl.md

5. Relative paths in CLI flags:
   - Resolved relative to: current working directory (CWD)
   - Example: --implement-command prompts/impl.md
   - Resolves to: $(pwd)/prompts/impl.md
```

**Test Scenarios to Add:**
```go
TestPathResolution_AbsolutePaths
TestPathResolution_HomeConfigRelative
TestPathResolution_ProjectConfigRelative
TestPathResolution_CustomConfigRelative
TestPathResolution_CLIFlagRelative
TestPathResolution_CustomConfigWithRelativePathToParent  // ../
TestPathResolution_CustomConfigLocation                   // config at different CWD
```

**Update Required:** Section "New Tests Required" → Add complete path resolution test matrix

---

### 2. Config Merging Logic - Deep vs Shallow

**Problem:** "Partial configs allowed" but merging strategy not specified.

**Ambiguous Scenarios:**
```yaml
# Home config
commands:
  implement: home-impl.md
  review: home-review.md

# Custom config (partial)
commands:
  implement: custom-impl.md
  # review is NOT specified
```

**Question:** Does final config have:
- A) `implement: custom-impl.md, review: home-review.md` (deep merge)
- B) `implement: custom-impl.md, review: <missing>` (shallow replace)

**Proposed Resolution:**
```
Config Merging Strategy: DEEP MERGE

Rules:
1. Scalar values (string, int, bool): Higher precedence replaces lower
2. Objects (commands): Merge fields, higher precedence wins per-field
3. nil/missing values: Inherit from lower precedence

Example:
  Home:    {implement: "a.md", review: "b.md", commit: "c.md"}
  Custom:  {implement: "x.md"}
  Result:  {implement: "x.md", review: "b.md", commit: "c.md"}
```

**Test Scenarios to Add:**
```go
TestConfigMerge_PartialCommandsObject     // Only some command fields
TestConfigMerge_ScalarOverride            // iterations override
TestConfigMerge_DeepMergeCommands         // Verify deep merge behavior
TestConfigMerge_NilValuesInherit          // nil in custom → use default
```

**Update Required:** Section "New Tests Required" → Add merge strategy tests

---

### 3. CLI Flag Equals Syntax - Consistency

**Problem:** Plan shows `--config=/path` but doesn't specify if this applies to all flags.

**Inconsistency:**
- `--config /path` vs `--config=/path` ✓ (mentioned)
- `--implement-command path` vs `--implement-command=path` ? (not mentioned)
- `--fluxid-iterations 10` vs `--fluxid-iterations=10` ? (not mentioned)

**Proposed Resolution:**
```
ALL string/path flags support BOTH syntaxes:
  --config /path          ✓
  --config=/path          ✓
  --implement-command path ✓
  --implement-command=path ✓

Integer flags also support both:
  --fluxid-iterations 10  ✓
  --fluxid-iterations=10  ✓
```

**Test Scenarios to Add:**
```go
TestCLIFlags_EqualsSyntax_Config
TestCLIFlags_EqualsSyntax_ImplementCommand
TestCLIFlags_EqualsSyntax_Iterations
```

**Update Required:** Section "New Tests Required" in args_test.go

---

### 4. Multiple --config Flags - Explicit Handling

**Problem:** Plan asks question but doesn't clearly state the decision in implementation spec.

**Recommendation says:** "Error on multiple --config flags"

**Missing:**
- Error message text not specified
- No test scenario for this
- Not in acceptance criteria

**Proposed Resolution:**
```
Behavior: ERROR on multiple --config flags

Error Message:
  "multiple --config flags specified. Please use only one --config flag."

Exit Code: 1 (invalid usage)
```

**Test Scenarios to Add:**
```go
TestCLIFlags_MultipleConfigFlags_Error
// Run: fluxid --config a.yaml --config b.yaml
// Expect: Error with clear message, exit code 1
```

**Update Required:**
- Section "Test Scenarios" → Add to m08-e01
- Section "Acceptance Criteria" → Add clear error for multiple --config

---

### 5. Empty/Invalid Config Values - Edge Cases

**Problem:** Not specified what happens with empty or invalid values.

**Ambiguous Cases:**
```yaml
# Case 1: Empty strings
commands:
  implement: ""
  review: ""

# Case 2: Whitespace only
agent: "   "

# Case 3: Invalid values
iterations: -5
implement_retries: 0

# Case 4: Wrong types
iterations: "not a number"
commit_enabled: "yes"  # should be boolean
```

**Proposed Resolution:**
```
Empty String Handling:
  - Empty command paths ("") → Error: "command path cannot be empty"
  - Empty agent ("" or whitespace) → Error: "agent cannot be empty"

Invalid Value Handling:
  - iterations ≤ 0 → Error: "iterations must be positive integer"
  - implement_retries ≤ 0 → Error: "implement_retries must be positive integer"
  - Invalid types → Error: "field X: expected Y, got Z"

Validation occurs at config loading time, before merging.
```

**Test Scenarios to Add:**
```go
TestConfigValidation_EmptyCommandPaths
TestConfigValidation_EmptyAgent
TestConfigValidation_NegativeIterations
TestConfigValidation_ZeroRetries
TestConfigValidation_InvalidTypes
TestConfigValidation_WhitespaceOnly
```

**Update Required:** Section "New Tests Required" → TestCustomConfigValidation

---

### 6. Session ID Environment Variable - Documentation Gap

**Problem:** Plan says keep FLUXID_SESSION_ID but doesn't document clearly.

**Confusion:**
- "Remove environment variables" sounds like ALL env vars
- FLUXID_SESSION_ID is actually kept but not clearly separated

**Proposed Resolution:**
```
Environment Variables (AFTER REFACTORING):

REMOVED (no longer used for config):
  ❌ FLUXID_ITERATIONS
  ❌ FLUXID_IMPLEMENT_RETRIES
  ❌ FLUXID_COMMIT_ENABLED

KEPT (used for IPC, not config):
  ✓ FLUXID_SESSION_ID  (workflow session identifier)

Rationale: FLUXID_SESSION_ID is not a configuration setting,
it's a runtime IPC mechanism for agent communication.
```

**Test Scenarios to Add:**
```go
TestEnvVars_SessionIDStillWorks
// Set FLUXID_SESSION_ID, verify it's used for IPC
// Set FLUXID_ITERATIONS, verify it's IGNORED

TestEnvVars_ConfigEnvVarsIgnoredCompletely
// Set all old config env vars, verify none affect config
```

**Update Required:**
- Section "Target State" → Clarify env var distinction
- Section "E2E Tests" m08-e04 → Add SessionID test

---

### 7. Config Source Tracking - Specification Needed

**Problem:** Plan mentions "source tracking" but doesn't specify format after changes.

**Current (from code):**
```go
type ResolvedConfig struct {
    // ...
    Sources map[string]string  // e.g., {"iterations": "cli", "agent": "home"}
}
```

**Question:** What are the source values after refactoring?

**Proposed Resolution:**
```
Source Tracking Values:

"default"  - Built-in default value
"home"     - ~/.fluxid/config.yaml
"project"  - ./.fluxid/config.yaml
"custom"   - --config file (value: absolute path to file)
"cli"      - CLI flag

Examples:
  {
    "iterations": "cli",                           // --fluxid-iterations
    "implement_retries": "custom:/path/to/cfg",    // --config file
    "agent": "project",                            // ./.fluxid/config.yaml
    "commit_enabled": "home",                      // ~/.fluxid/config.yaml
    "output_format": "default"                     // built-in default
  }
```

**Test Scenarios to Add:**
```go
TestSourceTracking_AllSources
// Verify source tracking shows correct origin for each value

TestSourceTracking_CustomConfigPath
// Verify custom config source includes file path
```

**Update Required:** Section "Integration Tests" → Add source tracking test

---

### 8. Config File Not Found - Error Handling

**Problem:** Not clear what error message to show when --config file missing.

**Proposed Resolution:**
```
Behavior: ERROR immediately, before any workflow execution

Error Message Format:
  "failed to load config file: /path/to/config.yaml: file does not exist"

Exit Code: 1

Verification:
  - Path in error message is the exact path user provided
  - Helpful hint: "ensure the path is correct and file exists"
```

**Test Scenarios to Add:**
```go
TestCustomConfig_MissingFile
// --config /nonexistent/config.yaml
// Expect: Clear error, exit 1, before any workflow execution

TestCustomConfig_RelativePathNotFound
// --config configs/missing.yaml (relative)
// Expect: Error shows resolved path
```

**Update Required:** Already in plan, but clarify error message format

---

### 9. Config File Permission Errors - Missing

**Problem:** No specification for file permission/IO errors.

**Proposed Resolution:**
```
I/O Error Handling:

1. Permission Denied:
   "failed to load config file: /path/to/config.yaml: permission denied"

2. Is Directory (not file):
   "failed to load config file: /path/to/config.yaml: is a directory"

3. Other I/O errors:
   "failed to load config file: /path/to/config.yaml: <system error>"

All result in exit code 1, no workflow execution.
```

**Test Scenarios to Add:**
```go
TestCustomConfig_PermissionDenied
// Create config with 000 permissions
// Expect: Clear permission error

TestCustomConfig_IsDirectory
// --config /path/to/directory/
// Expect: Clear "is a directory" error
```

**Update Required:** Section "Edge Cases to Cover" → Add I/O errors

---

### 10. Agent Flag vs Agent Config - Precedence Clarity

**Problem:** CLI has `--claude` but config has `agent: claude`. Precedence not crystal clear.

**Proposed Resolution:**
```
Agent Selection Precedence (same as other settings):

1. CLI flag (highest):     --claude / --codex / --opencode
2. Custom config:          agent: "claude"
3. Project config:         agent: "claude"
4. Home config:            agent: "claude"
5. Default (lowest):       agent: "claude"

Conversion:
  --claude   → agent: "claude"
  --codex    → agent: "codex"
  --opencode → agent: "opencode"

Then normal precedence rules apply.
```

**Test Scenarios to Add:**
```go
TestAgentPrecedence_CLIOverridesConfig
// Config: agent: "claude", CLI: --codex
// Expect: agent = "codex"

TestAgentPrecedence_ConfigOverridesDefault
// No CLI flag, config: agent: "codex"
// Expect: agent = "codex"
```

**Update Required:** Section "Config Precedence Integration" → Clarify agent handling

---

### 11. Backward Compatibility - Missing E2E Test

**Problem:** Plan mentions backward compatibility but no dedicated E2E test.

**Proposed Resolution:**
Add explicit backward compatibility E2E test suite.

**Test Scenarios to Add:**
```go
// New file: m08-e05-backward-compatibility_test.go

TestBackwardCompatibility_ProjectConfigOnly
// Setup: Only ./.fluxid/config.yaml (no --config, no CLI flags)
// Verify: Works exactly as before refactoring

TestBackwardCompatibility_HomeConfigOnly
// Setup: Only ~/.fluxid/config.yaml
// Verify: Works exactly as before

TestBackwardCompatibility_ProjectOverridesHome
// Setup: Both home and project configs
// Verify: Project wins, same as before

TestBackwardCompatibility_CLIOverridesAll
// Setup: Home + project + CLI flags (no --config)
// Verify: CLI wins, same as before
```

**Update Required:** Add new E2E file m08-e05-backward-compatibility_test.go

---

### 12. Symlinks in Config Path - Edge Case

**Problem:** Not specified how to handle symlinks.

**Proposed Resolution:**
```
Symlink Handling:

1. Config file is symlink:
   - Follow symlink, load target file
   - Relative paths resolved relative to SYMLINK location (not target)

2. Config directory contains symlinks:
   - Follow symlinks normally

3. Relative command paths with symlinks:
   - Resolved relative to config file's actual location (after following symlink)

Rationale: Matches common CLI tool behavior (e.g., Docker, kubectl)
```

**Test Scenarios to Add:**
```go
TestPathResolution_SymlinkedConfigFile
// --config points to symlink
// Verify: Config loaded from target, paths resolved correctly

TestPathResolution_RelativePathWithSymlink
// Config contains relative paths, config file is symlink
// Verify: Paths resolved relative to symlink location
```

**Update Required:** Section "Edge Cases to Cover" → Add symlink handling

---

### 13. Relative Path Escaping - Security Consideration

**Problem:** What if config has `../../../etc/passwd`?

**Proposed Resolution:**
```
Relative Path Security:

1. Allow .. in relative paths (legitimate use case)
2. Resolve to absolute path using filepath.Clean()
3. NO restriction on path escaping (user provides config, they control it)
4. Command execution will fail naturally if file doesn't exist

Example:
  Config at: /home/user/configs/feature.yaml
  Command: ../../../tmp/prompts/impl.md
  Resolves to: /tmp/prompts/impl.md (valid)

Security: Config files are trusted (user creates them), so path escaping is acceptable.
```

**Test Scenarios to Add:**
```go
TestPathResolution_RelativePathEscaping
// Config: commands.implement: "../../../tmp/test.md"
// Verify: Resolved correctly, no error

TestPathResolution_RelativePathWithDots
// Config: commands.implement: "./prompts/../other/test.md"
// Verify: Cleaned and resolved correctly
```

**Update Required:** Section "Edge Cases to Cover" → Document path escaping policy

---

### 14. Help Text - Missing Specification

**Problem:** "Help text includes new flags" but no specification of content.

**Proposed Resolution:**
```
Help Text Requirements:

1. New flags section:
   Configuration:
     --config PATH              Load custom configuration file
     --implement-command PATH   Override implement command file
     --review-command PATH      Override review command file
     --commit-command PATH      Override commit command file

2. Config precedence explanation:
   "Configuration precedence (highest to lowest):
    1. CLI flags
    2. Custom config (--config)
    3. Project config (./.fluxid/config.yaml)
    4. Home config (~/.fluxid/config.yaml)
    5. Built-in defaults"

3. Examples section:
   "Examples:
    fluxid --config workflows/feature.yaml
    fluxid --implement-command prompts/custom.md
    fluxid --config base.yaml --fluxid-iterations 30"
```

**Test Scenarios to Add:**
```go
TestHelpText_IncludesNewFlags
// Run: fluxid --help
// Verify: Contains --config, --implement-command, etc.

TestHelpText_IncludesPrecedenceInfo
// Verify: Help text explains config precedence
```

**Update Required:** Section "Documentation Requirements" → Add help text spec

---

### 15. Test Data Organization - Missing Specification

**Problem:** E2E tests need config files - storage strategy not specified.

**Proposed Resolution:**
```
Test Data Organization:

Location: e2e-tests/testdata/configs/

Structure:
  e2e-tests/
    testdata/
      configs/
        valid-complete.yaml       # All fields populated
        valid-partial.yaml        # Only some fields
        invalid-yaml.yaml         # Malformed YAML
        empty.yaml                # Empty file
        commands-only.yaml        # Only commands section
        relative-paths.yaml       # Relative command paths
      prompts/
        test-implement.md
        test-review.md
        test-commit.md

Usage in tests:
  configPath := "testdata/configs/valid-complete.yaml"
  fluxid --config=<absolute-path-to-config>
```

**Update Required:** Section "Files to Create" → Add testdata structure

---

### 16. Error Message Format - Consistency Needed

**Problem:** No specification for consistent error message format.

**Proposed Resolution:**
```
Error Message Format:

Pattern: "error: <component>: <description>"

Examples:
  ✓ "error: config: failed to load /path/to/config.yaml: file does not exist"
  ✓ "error: config: invalid YAML in /path/to/config.yaml: line 5: unexpected token"
  ✓ "error: config: iterations must be positive integer, got -5"
  ✓ "error: args: --config requires a value"
  ✓ "error: args: multiple --config flags specified"

Exit Codes:
  1 = Configuration/usage error
  130 = User abort (Ctrl+C)
  Other = Workflow failures
```

**Update Required:** Section "Acceptance Criteria" → Add error format requirement

---

### 17. Dry Run Behavior - Clarification Needed

**Problem:** Should --config be validated during dry run?

**Proposed Resolution:**
```
Dry Run Behavior with --config:

1. Config file IS loaded and validated
2. Config file paths ARE resolved
3. Config merging IS performed
4. Command files ARE checked for existence
5. NO workflow execution occurs

Rationale: Dry run should catch config errors, not just workflow errors.

Example:
  fluxid --config bad.yaml --dry-run
  → Error: Invalid config (same as normal run)

  fluxid --config good.yaml --dry-run
  → Shows what would execute, no actual execution
```

**Test Scenarios to Add:**
```go
TestDryRun_ValidatesCustomConfig
// --config with invalid file + --dry-run
// Expect: Error (config validated)

TestDryRun_ShowsCustomConfigValues
// --config + --dry-run
// Verify: Dry run output shows custom config values
```

**Update Required:** Section "E2E Tests" → Add dry-run scenarios

---

### 18. Concurrent Config Access - Consideration

**Problem:** Plan mentions "Concurrent config file reads (if applicable)" but doesn't specify if this is a concern.

**Proposed Resolution:**
```
Concurrent Config Access:

Scenario: User runs multiple fluxid instances simultaneously.

Current Behavior:
  - Each instance reads configs independently
  - No shared state between instances
  - No locking required

After Refactoring:
  - Same behavior (read-only operations)
  - No concurrency issues
  - No special handling needed

Test: Not required (reads are inherently safe)
```

**Update Required:** Section "Edge Cases to Cover" → Remove or clarify concurrency item

---

### 19. Config File Encoding - Assumption

**Problem:** No specification of expected file encoding.

**Proposed Resolution:**
```
Config File Encoding:

Expected: UTF-8 (standard YAML encoding)

Handling:
  - Most text will work (ASCII subset)
  - Non-ASCII characters in paths/values: Supported
  - BOM (Byte Order Mark): Handled by YAML parser
  - Other encodings: May fail, error message from YAML parser

No special encoding detection/conversion needed.
```

**Update Required:** Not critical, document as assumption

---

### 20. Implementation Order - Potential Issue

**Problem:** Step order might create intermediate broken states.

**Current Plan:**
```
Step 1-2: Remove env vars
Step 3-4: Add CLI flags
Step 5-6: Add custom config loading
Step 7-8: Integrate precedence
```

**Potential Issue:**
- After Step 2 (env vars removed), users lose env var functionality
- Custom config not available until Step 8
- Intermediate commits might break workflows

**Proposed Resolution:**
```
Revised Step Order (safer):

Phase 1: Add New Features (doesn't break anything)
  Step 1-2: Add CLI flags (--config, --implement-command, etc.)
  Step 3-4: Add custom config loading
  Step 5-6: Integrate custom config into precedence

Phase 2: Remove Old Features (breaking change)
  Step 7-8: Remove env var support
  Step 9: Update tests

Phase 3: Polish
  Step 10: E2E tests
  Step 11: Refactor

Benefit: New features work before old features removed.
Each commit leaves system in working state.
```

**Update Required:** Section "Implementation Checklist" → Reorder steps

---

## Summary of Critical Clarifications

### Must Resolve Before Implementation (Blockers)

1. ✅ **Path Resolution Complete Spec** - Affects multiple files
2. ✅ **Config Merging Strategy** - Deep vs shallow merge
3. ✅ **Empty/Invalid Value Handling** - Validation rules
4. ✅ **Error Message Format** - Consistency
5. ✅ **Multiple --config Handling** - Error specification

### Should Resolve Before Implementation (Important)

6. ✅ **CLI Equals Syntax** - All flags or just some?
7. ✅ **Source Tracking Format** - After env var removal
8. ✅ **Agent Precedence** - CLI flag vs config field
9. ✅ **Backward Compatibility Tests** - Missing E2E coverage
10. ✅ **Test Data Organization** - Where to store test configs

### Can Resolve During Implementation (Nice-to-have)

11. ✅ **Symlink Handling** - Edge case behavior
12. ✅ **Path Escaping Policy** - Security consideration
13. ✅ **Help Text Content** - Documentation
14. ✅ **Dry Run Behavior** - With custom config
15. ✅ **Implementation Order** - Safer step sequence

## Recommended Actions

### Before Starting Implementation

1. **Update REFACTOR_CUSTOM_CONFIG.md** with all critical clarifications (items 1-5)
2. **Add missing test scenarios** from this document
3. **Review with team** if multi-person project
4. **Document decisions** in plan file

### During Implementation

1. **Reference this document** when unclear behavior encountered
2. **Add tests first** (TDD) - tests encode the decisions
3. **Update this document** if new ambiguities discovered
4. **Keep both documents** in sync

### Acceptance Criteria Addition

Add to plan's acceptance criteria:
- [ ] All ambiguities in this document are resolved
- [ ] All new test scenarios from this document are implemented
- [ ] Error messages follow consistent format
- [ ] Path resolution follows documented rules
- [ ] Config merging follows deep merge strategy

---

**Next Steps:**
1. Review these clarifications
2. Decide on each ambiguity resolution
3. Update REFACTOR_CUSTOM_CONFIG.md with decisions
4. Begin TDD implementation with clear specifications
