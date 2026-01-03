# Task File Workflow Implementation Plan (What Only)

- Add required `--file=PATH` CLI flag for workflow runs.
- Store the task file path in runtime configuration.
- Validate `--file` as an absolute, existing, readable path.
- Continue using resolved command files (implement/review/commit) from config; do not read their contents in Go.
- Implement phase prompt: instruct agent to run the implement command file for the provided task file.
- Commit phase prompt: instruct agent to read and execute the commit command file to create the commit.
- Review phase prompt: instruct agent to run the review command file for the provided task file.
- Export `FLUXID_TASK_FILE` to the agent process environment.
- Include the task file path in initialization status (text/json/yaml).
- Show task file and command file paths in dry‑run simulation output per phase.
- Update CLI usage/help to document `--file` requirement in workflow mode.
- Add tests verifying (Test-Driven):
  - Happy path (1/3): workflow runs with valid `--file`, prompts include task and command file paths, and success outputs include the task file.
  - Unhappy path (2/3): missing `--file`, non-absolute path, nonexistent file, unreadable file, and invalid run mode produce clear errors.
  - Coverage: ensure ≥90% branch coverage on new logic (args parsing, validation, prompt composition, init status, dry-run display).
- Keep runtime pure Go; do not introduce shell runtime dependencies.
