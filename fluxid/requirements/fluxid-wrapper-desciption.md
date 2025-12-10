Here is the updated concept with the new flag integrated.

---

## 1. Purpose

`fluxid` is a Go CLI wrapper around `claude` that:

- Orchestrates looped workflows: IMPLEMENT → COMMIT → REVIEW.
- Exposes a YAML-based IPC interface for:
  - Report schema + report storage.
  - History schema + history storage/query.
- Streams Claude output directly to the terminal.
- Uses per-run sessions identified by `FLUXID_SESSION_ID`.
- Supports configurable prompts for IMPLEMENT/COMMIT/REVIEW.
- Can optionally write the **full session history** to disk on completion.

---

## 2. Modes

- **Wrapper mode** (default):
  - `fluxid [fluxid-flags...] [claude-args...]`
- **IPC mode**:
  - `fluxid ipc <subcommand> [flags...]`

`ipc` as first arg → IPC mode; otherwise wrapper mode.

---

## 3. Session management

- Each wrapper run creates one **session ID** (opaque string, e.g. UUID).
- On wrapper startup:
  - Generate `session-id`.
  - Set `FLUXID_SESSION_ID=<session-id>` in the environment for all child processes.
- IPC session resolution (for all `fluxid ipc ...` commands):

  1. If `--session <id>` given → use that.
  2. Else if `FLUXID_SESSION_ID` set → use that.
  3. Else → print error + usage to stderr, exit non-zero.

- Sessions are **not persisted** beyond the lifetime of the wrapper process.

---

## 4. Wrapper mode

### 4.1 Fluxid flags and Claude args

Usage:

```bash
fluxid [fluxid-flags...] [claude-args...]
```

Fluxid-specific flags (non-exhaustive):

- `--fluxid-inner-loops N` (int, default 3)
- `--fluxid-outer-loops N` (int, default 20)
- `--fluxid-no-commit` (disable COMMIT phase)
- `--fluxid-no-review` (disable REVIEW phase)
- `--fluxid-claude-bin PATH` (default `"claude"`)
- `--fluxid-output {json|yaml}` (init status format)
- `--write-history <path-to-folder>`  
  When set, on wrapper completion, fluxid writes the **full session history** to a file in the given folder.

All args that are **not** `--fluxid-*` (and not the above) are treated as Claude arguments, subject to prompt handling below.

### 4.2 Environment and precedence

Defaults:

- `innerLoops = 3`
- `outerLoops = 20`
- `commitEnabled = true`
- `reviewEnabled = true`

Env vars:

- `FLUXID_INNER_LOOPS` (int)
- `FLUXID_OUTER_LOOPS` (int)
- `FLUXID_DISABLE_COMMIT` (bool-like)
- `FLUXID_DISABLE_REVIEW` (bool-like)
- `FLUXID_CLAUDE_BIN` (path override for `claude`)

Precedence (low → high):

1. Built-in default.
2. Environment variable.
3. CLI flag.

### 4.3 Loop behavior

Effective configuration:

- `innerLoops`, `outerLoops`
- `commitEnabled`, `reviewEnabled`

Execution:

```text
for outer in 1..outerLoops:
    for inner in 1..innerLoops:
        run claude IMPLEMENT
        if commitEnabled:
            run claude COMMIT
    if reviewEnabled:
        run claude REVIEW
```

Each run:

- Calls `claude` with phase-specific prompt.
- Uses user’s Claude arguments, adjusted for prompt handling.

### 4.4 Prompt configuration and handling

Config files:

- Default: `$HOME/.fluxid/config.yaml`
- Optional local override: `./.fluxid/config.yaml` (overrides home config where overlapping).

Config must define prompts for:

- `implement.prompt`
- `commit.prompt`
- `review.prompt`

Each value can be:

- A literal string prompt.
- A path to a file whose contents are the prompt.

Prompt resolution:

- For each phase (IMPLEMENT/COMMIT/REVIEW), fluxid resolves a concrete prompt string by loading config (and file if path).

Claude’s `-p` / `--prompt`:

- For all three phases, **fluxid must control the prompt**.
- For each Claude invocation:
  - Strip any user-provided `-p` or `--prompt` from the forwarded args.
  - Inject the phase-specific prompt using Claude’s standard prompt flag (e.g. `--prompt "..."`).

User `-p/--prompt` is effectively ignored during wrapper-managed runs; fluxid’s prompts dominate.

### 4.5 Initialization status output

On startup (before loops), fluxid:

- Resolves:
  - Session ID.
  - Effective loop counts.
  - Commit/review enabled flags.
  - Config sources and prompts (at least paths).
  - Claude binary path.

Then prints an initialization status:

- If `--fluxid-output json`:
  - Print a JSON object (session ID, inner/outer loops, flags, config sources, etc.).
- If `--fluxid-output yaml`:
  - Same content as YAML.
- If no `--fluxid-output`:
  - Print simple human-readable log lines.

Output like a header before the loop starts. This serves as a initialization status report

### 4.6 Claude process wiring (streaming)

For each Claude run (IMPLEMENT/COMMIT/REVIEW):

- `stdin` → `os.Stdin`
- `stdout` → `os.Stdout`
- `stderr` → `os.Stderr`

No buffering or modification; Claude output is streamed directly to the terminal.

### 4.7 Error handling in wrapper

- If a Claude process exits non-zero:
  - Wrapper aborts immediately and exits with the same exit code.
- If fluxid internally calls any IPC command and it fails:
  - Do **not** abort the wrapper.
  - Print error + brief usage for that IPC command.
  - Continue running loops.

### 4.8 `--write-history <folder>` behavior

- If `--write-history <folder>` is **not** specified:
  - No export of history to disk is required (internal storage only).

- If `--write-history <folder>` **is** specified:
  - At wrapper completion (whether loops end normally or due to a Claude error), fluxid must:
    1. Ensure `<folder>` exists (or create it).  
       - If `<folder>` cannot be created/written:
         - Print a clear error to stderr.
         - Continue normal exit; do not change wrap-up behavior.
    2. Retrieve the **full history** for the session.
    3. Write it to a file in `<folder>`, using:
       - A deterministic filename, e.g.  
         `fluxid-history-<session-id>.yaml`
       - Contents: a YAML sequence of all history entries, ordered **newest first** (consistent with IPC `view-history`).
    4. Exit with:
       - The wrapper’s normal exit code (Claude’s exit or 0), **not** overridden by write-history success/failure.

- `--write-history` does not change IPC behavior; it is just a post-run export.

---

## 5. IPC mode

Commands are called as:

```bash
fluxid ipc <subcommand> [flags...]
```

Session resolution: see section 3.

All inputs/outputs are YAML except errors, which are human-readable.

### 5.1 Report schema and report

#### `fluxid ipc get-report-schema`

- Out: YAML schema for report, to stdout.
- Exit: 0 on success.

#### `fluxid ipc write-report`

- In: report YAML on stdin.
- Behavior:
  - Resolve session.
  - Parse YAML; validate against report schema.
- On success:
  - Set as current report for that session (replacing any previous).
  - Exit 0.
- On validation/parse error:
  - Do not change the report.
  - Print a **human-readable error** to stderr, containing:
    - Exact validation/parse error messages and/or error codes.
  - Exit non-zero.

#### `fluxid ipc read-report`

- Input: none.
- Behavior:
  - Resolve session.
  - Check for current report.
- If report exists:
  - Output report as YAML to stdout.
  - Exit 0.
- If no report exists:
  - Print a clear message to stderr (e.g. “No report available for this session.”).
  - Exit 0 (this is a valid state).

### 5.2 History schema and history

#### `fluxid ipc get-history-schema`

- Out: YAML schema for a **single history entry**.
- Exit: 0 on success.

#### `fluxid ipc write-history`

- In: history entry YAML on stdin.
- Behavior:
  - Resolve session.
  - Parse YAML.
  - Validate against history-entry schema.
- On success:
  - Append a new history entry:
    - Include **only one automatic field** managed by fluxid:
      - A `timestamp` (name must be fixed in the schema).
      - This is never set by external input.
  - Exit 0.
- On validation/parse error:
  - Do not modify history.
  - Print a human-readable error to stderr including exact error messages/codes.
  - Exit non-zero.

#### `fluxid ipc view-history [--last N | --all]`

- Flags:
  - `--last N` (N > 0): default is `--last 40` if neither `--last` nor `--all` is given.
  - `--all`: overrides `--last`, returns all entries.

- Behavior:
  - Resolve session.
  - Retrieve history entries for that session.
  - Sort **by timestamp, newest first**.
  - Select:
    - `--all`: all entries.
    - `--last N`: the first N entries (most recent N).

- Output:
  - YAML **sequence** of entries (simple list), ordered newest → oldest.
- Edge case:
  - If there are no entries:
    - Valid state.
    - Output an empty YAML sequence (e.g. `[]`) or nothing (but consistent).
    - May print a note to stderr.
    - Exit 0.

---

## 6. Storage model

- Per session:
  - At most one current report.
  - Zero or more history entries (each with a fluxid-generated timestamp).
- Sessions isolated by ID.
- Data only needs to live for the process lifetime:
  - No required persistence beyond the wrapper’s lifetime.
- Implementation (memory vs temp files) is internal and does not change behavior.

---

## 7. PATH and discoverability

- `fluxid` must be on `PATH` (via brew/choco).
- Claude calls `fluxid ipc ...` and inherits `FLUXID_SESSION_ID`.
- Fluxid resolves `claude` via:
  - `--fluxid-claude-bin` flag,
  - or `FLUXID_CLAUDE_BIN` env,
  - or default `claude` on PATH.

---

## 8. Schemas/Templates to use
- History: fluxid/requirements/history-schema.yaml
- Report: fluxid/requirements/report-schema.yaml
