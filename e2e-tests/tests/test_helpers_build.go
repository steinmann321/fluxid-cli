package tests

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
