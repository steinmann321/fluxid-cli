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

// getWorkingStubScript returns the script content for a working (PASS) agent stub.
// This stub writes valid PASS reports so workflows can proceed.
func getWorkingStubScript() string {
	return `#!/bin/bash
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
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
REPORT_FILE=$(fluxid report --get-file)
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
}

// createStubAgentsInDir creates stub agent binaries in the specified directory.
// Returns the directory path for convenience.
func createStubAgentsInDir(t *testing.T, dir string, stubScript string) string {
	t.Helper()

	const dirPerms = 0o755  // rwxr-xr-x: owner can read/write/execute, others can read/execute
	const filePerms = 0o755 // rwxr-xr-x: executable scripts

	if err := os.MkdirAll(dir, dirPerms); err != nil {
		t.Fatalf("failed to create stub directory: %v", err)
	}

	// Create stubs for all agents used in tests
	agents := []string{"claude", "opencode", "codex", "gemini", "project-agent"}
	for _, agent := range agents {
		agentPath := filepath.Join(dir, agent)
		if err := os.WriteFile(agentPath, []byte(stubScript), filePerms); err != nil {
			t.Fatalf("failed to create stub %s: %v", agent, err)
		}
	}

	return dir
}

// stubMutex protects stub creation from race conditions during parallel test execution.
// Unlike sync.Once, this allows stubs to be recreated when needed (for -count=N).
//
//nolint:gochecknoglobals // Global mutex required for cross-test synchronization
var stubMutex sync.Mutex

// createStubClaude creates stub agent binaries for testing in the shared bin/ directory.
// Uses a mutex to prevent race conditions when multiple parallel tests call this function,
// while detecting when fluxid binary changes to force stub recreation.
//
// Deprecated: Tests that need custom stubs (failing, conditional, etc.) should use
// createStubAgentsInDir with a test-specific temp directory to avoid race conditions.
//
// FLAKINESS FIX (2026-01-06):
// Replaced sync.Once with mutex + timestamp comparison. sync.Once prevented stub recreation
// across -count iterations, causing 90% test failures. The mutex prevents race conditions
// while timestamp comparison detects when fluxid binary changes, forcing stub recreation.
//
// RACE CONDITION FIX:
// Previously, 62+ parallel tests all called createStubClaude(), causing concurrent
// writes to bin/claude, bin/opencode, etc. This resulted in intermittent
// "fork/exec: exec format error" failures when a test tried to execute a stub
// while another test was writing to it.
func createStubClaude(t *testing.T, root string) {
	t.Helper()

	stubMutex.Lock()
	defer stubMutex.Unlock()

	binDir := filepath.Join(root, "bin")
	stubPath := filepath.Join(binDir, "claude")
	fluxidPath := filepath.Join(binDir, "fluxid")

	// Check if stubs need recreation by comparing timestamps
	// If fluxid is newer than stubs, recreate them
	stubInfo, stubErr := os.Stat(stubPath)
	fluxidInfo, fluxidErr := os.Stat(fluxidPath)

	needsRecreation := stubErr != nil || // Stub doesn't exist
		fluxidErr != nil || // Fluxid doesn't exist (create stubs anyway)
		stubInfo.ModTime().Before(fluxidInfo.ModTime()) || // Stub older than fluxid
		stubInfo.Mode()&0o111 == 0 // Stub not executable

	if !needsRecreation {
		// Stubs are fresh and valid
		return
	}

	// Create fresh stubs
	stubScript := getWorkingStubScript()
	createStubAgentsInDir(t, binDir, stubScript)
}
