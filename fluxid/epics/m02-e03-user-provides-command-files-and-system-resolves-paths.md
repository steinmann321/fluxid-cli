---
id: m02-e03
title: User provides command files and system resolves paths
milestone: m02
status: pending
patterns: []
---

# Epic: User provides command files and system resolves paths

## Overview
User supplies required command markdown files (`implement`, `review`, `commit`) under project or home `.fluxid/commands/` and references them in config. On startup, the system resolves each command file path using project-first then home fallback, validates existence and readability, and surfaces absolute paths in the initialization status.

## Scope
- User actions: Place markdown command files in `./.fluxid/commands/` or `~/.fluxid/commands/`; set command filenames in config; run `fluxid`.
- System responses: Resolve file locations with project-first lookup; validate presence and readability; on failure, abort with clear error; on success, print absolute paths for each command file in initialization output.
- Screens/states: Initialization status with a "Command Files" section listing absolute paths or a startup error message if missing.

## Success Criteria
- [ ] Requires `commands.implement`, `commands.review`, `commands.commit` in config [Test: missing any key causes startup failure with key name in message]
- [ ] Resolves command files with project-over-home precedence [Test: provide files in both locations; project wins and absolute path matches]
- [ ] Prints absolute resolved paths in status, not file contents [Test: assert output paths are absolute and point to existing files]
- [ ] Fails fast if any referenced file missing/unreadable [Test: remove file or set wrong name; expect clear error listing which file cannot be read]

## Dependencies
m02-e01, m02-e02

## E2E Test Mapping
**Test File**: `m02_e03_t01_command-file-resolution.yaml`

**Test Flow**:
1. Provide required command files in home
2. Reference them in home config
3. Add project overrides for one command file and verify precedence
4. Run `fluxid` and confirm absolute paths are listed

**Key Assertions**:
- All three command files resolved
- Precedence respects project over home
- Startup aborts with clear error if a required command file is missing

## Completion Checklist
- [ ] All tasks completed and tested
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Ensure path normalization and cross-platform home expansion. Do not read file contents; only validate existence and permissions.