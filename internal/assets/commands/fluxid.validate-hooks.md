# Task
Validate pre-commit hooks match actual project structure.

## Ultimate Goal
Ensure every test file and application code file in the project is caught by appropriate pre-commit hook patterns. No test or coverage check should be skipped due to restrictive path patterns.

## Process

1. **Discover project** - Identify tech stack from project files (package.json, pyproject.toml, build configs, etc.). Find ALL test files using tech-specific conventions, runner scripts and/or build systems. Record locations and types. Also identify ALL application code directories (not test/config/migration files).

2. **Validate test patterns** - Read `.pre-commit-config.yaml`. Check if `files:` patterns for test-related hooks match ALL discovered test locations.

3. **Validate coverage patterns** - Check coverage enforcement scripts (e.g., `scripts/backend_coverage_enforce.sh`, `.fluxid/hooks/flutter/flutter_coverage_enforce.sh`). Ensure their grep/glob patterns match ALL discovered application code directories. Update hardcoded directory patterns to match actual project structure.

4. **Update restrictive patterns** - Replace hardcoded directories with dynamic regex matching all test and app code locations.

5. **Verify scripts** - Confirm referenced hook scripts exist. Create missing ones for discovered tech stack using `.fluxid/templates/` as reference.

6. **Verify setup** - Check pre-commit installed and hooks active in `.git/hooks/`.

## Output
Report changes or "All hooks validated successfully". Stage modified files.
