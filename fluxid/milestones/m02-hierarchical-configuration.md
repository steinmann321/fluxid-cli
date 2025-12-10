---
id: m02
title: Users can customize workflow behavior through hierarchical configuration
status: pending
---

# Milestone: Users can customize workflow behavior through hierarchical configuration

## Deliverable
Users can customize fluxid's behavior (loop counts, phase toggles, command prompts) through a hierarchical configuration system with home config (`~/.fluxid/config.yaml`), project config (`./.fluxid/config.yaml`), environment variables, and CLI flags, with clear precedence rules (defaults → home → project → env → CLI). Users create markdown command files defining agent prompts for each phase and configure fluxid to use them.

**What users can do:**
- Create user-wide configuration in `~/.fluxid/config.yaml` for default settings
- Override with project-specific configuration in `./.fluxid/config.yaml`
- Set configuration via environment variables (e.g., `FLUXID_IMPLEMENT_RETRIES=5`)
- Override everything with CLI flags (highest precedence)
- Write custom command files (markdown) with phase-specific prompts in `.fluxid/commands/`
- Configure which command files to use via config.yaml
- See initialization status showing resolved configuration with source precedence
- Disable commit phase via `--fluxid-no-commit` flag (review remains mandatory)

## Success Criteria
- [ ] Users can create `~/.fluxid/config.yaml` with configuration (loop counts, agent selection, command file references)
- [ ] Users can create `./.fluxid/config.yaml` to override home config for specific projects
- [ ] Config schema validated on startup:
  - [ ] `agent: claude` (default)
  - [ ] `implement_retries: 3` (default, must be ≥1)
  - [ ] `iterations: 20` (default, must be ≥1)
  - [ ] `commit_enabled: true` (default)
  - [ ] `commands.implement: fluxid.implement.md` (required reference)
  - [ ] `commands.commit: fluxid.commit.md` (required reference)
  - [ ] `commands.review: fluxid.review.md` (required reference)
- [ ] Command files resolve with home/project override pattern:
  - [ ] Check `./.fluxid/commands/` first, fall back to `~/.fluxid/commands/`
  - [ ] Abort at startup if any mandatory command file missing/unreadable
- [ ] Environment variables override config files:
  - [ ] `FLUXID_AGENT` sets default agent
  - [ ] `FLUXID_IMPLEMENT_RETRIES` sets retry count
  - [ ] `FLUXID_ITERATIONS` sets iteration count
  - [ ] `FLUXID_COMMIT_ENABLED` enables/disables commit phase
- [ ] CLI flags override everything:
  - [ ] `--fluxid-implement-retries N`
  - [ ] `--fluxid-iterations N`
  - [ ] `--fluxid-no-commit` disables commit phase
- [ ] Initialization status displays resolved configuration with clear precedence indication
- [ ] Initialization status shows command file paths (absolute paths, not content)
- [ ] Invalid config values (negative loop counts, missing files) cause immediate startup failure with clear error messages
- [ ] Complete UI: clear config validation error messages, initialization status display
- [ ] Full backend: YAML parsing, file resolution, precedence logic, validation
- [ ] Can be deployed independently: works with existing m01 loop orchestration
- [ ] Requires no additional milestones: full configuration customization capability

## Validation Questions
**Before marking this milestone complete, answer:**
- [ ] Can a real user perform complete configuration workflows with only this milestone? **YES** - they can customize all behavior through config files, env vars, and CLI flags
- [ ] Is it polished enough to ship publicly? **YES** - clear validation errors, helpful messages about precedence
- [ ] Does it solve a real problem end-to-end? **YES** - enables team-specific workflow customization without code changes
- [ ] Does it include both complete UI and functional backend integration? **YES** - validation messages (UI) + config system (backend)
- [ ] Can it run independently without waiting for other milestones? **YES** - builds on m01, adds configuration layer
- [ ] Would you personally use this if it were released today? **YES** - essential for adapting tool to different team processes

## Vertical Slice - All Layers Included
This milestone includes:
- **File System Operations**: YAML file reading, markdown command file loading
- **Configuration Parsing**: YAML schema validation, structure parsing
- **Precedence Resolution**: Merging config from multiple sources (defaults → home → project → env → CLI)
- **File Resolution**: Command file path resolution with home/project override pattern
- **Validation Logic**: Loop count validation (≥1), file existence checks, schema compliance
- **Error Reporting**: Clear error messages for missing files, invalid values, schema violations
- **Status Display**: Initialization status showing resolved config and source precedence
- **Phase Control**: Commit phase toggle implementation (`--fluxid-no-commit`)

## Notes
- This milestone makes the tool **configurable for different teams and projects** without code changes
- Command files are markdown documents that serve as prompts for the agent
- Built-in defaults from m01 remain as fallback when no config provided
- Future milestones (m03) will enhance command files to use IPC instead of placeholder mechanisms
- Review phase cannot be disabled (business rule)
- Configuration files use YAML for human readability and version control compatibility
