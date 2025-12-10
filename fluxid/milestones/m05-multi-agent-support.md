---
id: m05
title: Users can select between Claude, Codex, and OpenCode agents
status: pending
---

# Milestone: Users can select between Claude, Codex, and OpenCode agents

## Deliverable
Users can choose which AI agent CLI to use for workflow execution via explicit agent selection flags (`--claude`, `--codex`, `--opencode`) or configuration (`agent: claude` in config.yaml, `FLUXID_AGENT` environment variable). The tool validates that exactly one agent is selected, resolves the agent binary from PATH, and invokes it with the same process orchestration and IPC mechanisms used for Claude. This enables teams to use their preferred agent without changing workflow automation infrastructure.

**What users can do:**
- Select agent via CLI flags: `fluxid --claude [args]`, `fluxid --codex [args]`, `fluxid --opencode [args]`
- Set default agent in config.yaml: `agent: codex`
- Use environment variable: `FLUXID_AGENT=opencode`
- Switch between agents without modifying command files or configuration (agent-agnostic workflow)
- See initialization status showing which agent was selected and source of selection
- Get clear error if multiple agent flags specified or no agent available on PATH

## Success Criteria
- [ ] Agent selection flags implemented:
  - [ ] `--claude` - use Claude agent CLI
  - [ ] `--codex` - use Codex agent CLI
  - [ ] `--opencode` - use OpenCode agent CLI
- [ ] Configuration-based agent selection:
  - [ ] `agent: claude|codex|opencode` in config.yaml (default: claude)
  - [ ] `FLUXID_AGENT` environment variable
  - [ ] Precedence: CLI flags > env vars > config files > built-in default (claude)
- [ ] Agent validation:
  - [ ] Exactly one agent must be selected (error if multiple flags specified)
  - [ ] Agent binary must exist on PATH (error if not found)
  - [ ] Clear error messages: "Multiple agents specified" or "Agent 'codex' not found on PATH"
- [ ] Agent resolution:
  - [ ] Look up agent binary on PATH (e.g., `claude`, `codex`, `opencode`)
  - [ ] Verify binary is executable
  - [ ] No hardcoded paths (rely on PATH environment)
- [ ] Process orchestration works identically for all agents:
  - [ ] Same loop control (implement → commit → review)
  - [ ] Same stdin/stdout/stderr streaming
  - [ ] Same IPC mechanisms (reports and history)
  - [ ] Same command file loading
  - [ ] Same session management
- [ ] Initialization status shows selected agent and selection source
- [ ] All existing features (m01-m04) work with any agent
- [ ] Complete UI: agent selection in initialization status, clear error messages
- [ ] Full backend: agent resolution logic, PATH lookup, multi-agent process spawning
- [ ] Can be deployed independently: works with all previous milestones
- [ ] Requires no additional milestones: full multi-agent support

## Validation Questions
**Before marking this milestone complete, answer:**
- [ ] Can a real user perform complete workflows with any agent with only this milestone? **YES** - they can select Claude, Codex, or OpenCode and execute identical workflows
- [ ] Is it polished enough to ship publicly? **YES** - clear agent selection, helpful error messages, consistent behavior
- [ ] Does it solve a real problem end-to-end? **YES** - enables organizational flexibility and avoids vendor lock-in
- [ ] Does it include both complete UI and functional backend integration? **YES** - agent selection flags (UI) + resolution/spawning (backend)
- [ ] Can it run independently without waiting for other milestones? **YES** - builds on m01-m04, adds agent abstraction
- [ ] Would you personally use this if it were released today? **YES** - critical for teams using different agents across projects

## Vertical Slice - All Layers Included
This milestone includes:
- **CLI Flag Parsing**: Agent selection flag handling, mutual exclusion validation
- **Configuration Schema**: Agent field in config.yaml, validation
- **Environment Variables**: FLUXID_AGENT parsing and precedence handling
- **Agent Resolution**: PATH lookup, binary existence check, executable verification
- **Precedence Logic**: Resolving agent selection from multiple sources
- **Process Spawning**: Agent-agnostic subprocess invocation
- **Error Handling**: Clear messages for missing agents, conflicting selections
- **Initialization Status**: Agent selection display with source indication

## Notes
- This milestone makes fluxid **agent-agnostic** - same workflow automation works with different agents
- All three agents (Claude, Codex, OpenCode) are **third-party dependencies** (not installed by fluxid)
- Users must ensure chosen agent is installed and available on PATH
- Command files work identically regardless of agent (they all support prompt-based invocation)
- Default agent is Claude (most common usage based on bash implementation)
- This completes the core feature set - future milestones can add conveniences (dry-run, output formats)
