//nolint:funlen // Test helper: setup code justifies length
package tests

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var errProjectRootNotFound = errors.New("project root with go.mod not found")

// findProjectRoot walks up from the starting directory to find the project root.
func findProjectRoot(start string) (string, error) {
	cur := start
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return "", errProjectRootNotFound
}

// getProjectRoot returns the absolute path to the project root directory.
func getProjectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	root, err := findProjectRoot(wd)
	if err != nil {
		t.Fatalf("find project root failed: %v", err)
	}

	return root
}

// buildFluxid builds the fluxid binary for testing.
func buildFluxid(t *testing.T, root string) {
	t.Helper()

	// Build fluxid binary
	build := exec.CommandContext(context.Background(), "go", "build", "-o", "bin/fluxid", "./cmd/fluxid")
	build.Dir = root

	var stderr bytes.Buffer
	build.Stderr = &stderr

	if err := build.Run(); err != nil {
		t.Fatalf("build failed: %v\nStderr: %s", err, stderr.String())
	}
}

// stubOnce ensures createStubClaude runs exactly once across all parallel tests.
// This prevents race conditions where multiple tests try to write to the same
// stub binary files simultaneously, causing "exec format error" failures.
//
//nolint:gochecknoglobals // Global required for sync.Once to work across all parallel tests
var stubOnce sync.Once

// createStubClaude creates stub agent binaries for testing.
// Uses sync.Once to ensure it runs exactly once, preventing race conditions
// when multiple parallel tests call this function simultaneously.
//
// RACE CONDITION FIX:
// Previously, 62+ parallel tests all called createStubClaude(), causing concurrent
// writes to bin/claude, bin/opencode, etc. This resulted in intermittent
// "fork/exec: exec format error" failures when a test tried to execute a stub
// while another test was writing to it.
func createStubClaude(t *testing.T, root string) {
	t.Helper()

	stubOnce.Do(func() {
		stubScript := `#!/bin/bash
# Stub agent CLI for testing

# Echo all arguments to demonstrate passthrough
echo "Claude stub invoked with args: $@"

# Echo environment variables for validation
echo "FLUXID_SESSION_ID=$FLUXID_SESSION_ID"

# Detect phase from prompt argument (last argument)
PROMPT="${@: -1}"
COMMAND="test"

# Determine phase-specific command based on prompt keywords
if [[ "$PROMPT" == *"Implement the required changes"* ]]; then
  COMMAND="fluxid.implement"
elif [[ "$PROMPT" == *"Create a git commit"* ]]; then
  COMMAND="fluxid.commit"
elif [[ "$PROMPT" == *"Review the implementation"* ]]; then
  COMMAND="fluxid.review"
fi

# Write a valid PASS report so workflow can proceed
# Using new file-based interface
FLUXID_BIN="$(dirname "$0")/fluxid"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$("$FLUXID_BIN" report --get-file)
cat > "$REPORT_FILE" <<REPORT_EOF
command: $COMMAND
artifact: stub-test
timestamp: $TIMESTAMP
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
REPORT_EOF

# Simulate successful execution
exit 0
`

		// Create bin directory if it doesn't exist
		binDir := filepath.Join(root, "bin")
		const dirPerms = 0o755  // rwxr-xr-x: owner can read/write/execute, others can read/execute
		const filePerms = 0o755 // rwxr-xr-x: executable scripts

		if err := os.MkdirAll(binDir, dirPerms); err != nil {
			t.Fatalf("failed to create bin directory: %v", err)
		}

		// Create stubs for all agents used in tests
		agents := []string{"claude", "opencode", "codex", "project-agent"}
		for _, agent := range agents {
			agentPath := filepath.Join(binDir, agent)
			if err := os.WriteFile(agentPath, []byte(stubScript), filePerms); err != nil {
				t.Fatalf("failed to create stub %s: %v", agent, err)
			}
		}
	})
}
