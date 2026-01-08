# fluxid v0.1.0 - First Stable Release

## Overview

First stable release of fluxid, a CLI workflow orchestration tool that enables coding agents to break through context window limitations using an Implement-Review-Commit loop.

## Key Features

### Core Functionality
- **Pure Go Implementation**: Complete rewrite in Go 1.25 for better performance and maintainability
- **Implement-Review-Commit Loop**: Structured workflow phases with configurable retry limits
- **Native CLI Commands**: Built-in `fluxid report` and `fluxid history` commands for workflow state management
- **Multi-Agent Support**: Compatible with Claude, OpenCode, and other coding agents

### Critical Bug Fixes
- **Fixed Retry Logic Bug** ([#issue](https://github.com/steinmann321/fluxid-cli/commit/6ba1044)): `MaxImplementRetries` now correctly respects configured values (previously stopped at 3 regardless of configuration)
- Workflow now properly exhausts all retry attempts before proceeding to next phase

### Quality & Testing
- **90.3% Test Coverage**: Comprehensive unit and e2e tests
- **Multi-Platform Support**: Binaries for Darwin, Linux, and Windows on both amd64 and arm64
- **Pre-commit Hooks**: Automated formatting, linting, security scanning

## Installation

### Package Managers (Recommended)

**Homebrew (macOS/Linux)**:
```bash
brew tap steinmann321/tap
brew install fluxid
```

**Chocolatey (Windows)**:
```powershell
choco install fluxid
```

### Download Pre-built Binaries

Alternatively, download the appropriate archive for your platform:

- **macOS (Apple Silicon)**: `fluxid_darwin_arm64_v8.0.tar.gz`
- **macOS (Intel)**: `fluxid_darwin_amd64_v1.tar.gz`
- **Linux (amd64)**: `fluxid_linux_amd64_v1.tar.gz`
- **Linux (arm64)**: `fluxid_linux_arm64_v8.0.tar.gz`
- **Windows (amd64)**: `fluxid_windows_amd64_v1.tar.gz`
- **Windows (arm64)**: `fluxid_windows_arm64_v8.0.tar.gz`

### Verify Checksums

```bash
# Download checksums.txt and verify your download
shasum -a 256 -c checksums.txt
```

### Extract and Install

```bash
# Extract (macOS/Linux)
tar xzf fluxid_*.tar.gz

# Move to your PATH
mv fluxid /usr/local/bin/  # or ~/go/bin/ or any directory in your PATH

# Verify installation
fluxid version
```

## Configuration

Create `~/.fluxid/config.yaml`:

```yaml
agent:
  command: claude  # or your preferred agent
  args: []

workflow:
  max_review_cycles: 3
  max_implement_retries: 3  # Now works correctly!
  max_commit_retries: 1

commands:
  implement: ~/.fluxid/commands/implement.md
  review: ~/.fluxid/commands/review.md
  commit: ~/.fluxid/commands/commit.md
```

## Usage

```bash
# Start a workflow with a task file
fluxid --claude --file=task.md

# Configure retry limits
fluxid --claude --fluxid-max-implement-retries=10 --file=task.md

# Check version
fluxid version
```

## What's Changed

### Full Changelog
- Initial stable release v0.1.0
- Fix: implement phase now respects MaxImplementRetries configuration
- Chore: clean up unused dependencies

**Full Commit History**: https://github.com/steinmann321/fluxid-cli/commits/v0.1.0

## Contributors

This release was made possible by [@steinmann321](https://github.com/steinmann321) with AI assistance from Claude.

---

**Note**: This is the first stable release. Please report any issues at https://github.com/steinmann321/fluxid-cli/issues
