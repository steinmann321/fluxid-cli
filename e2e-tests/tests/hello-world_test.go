// Package tests contains E2E tests for the fluxid workflow system.
package tests

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
	return "", errors.New("project root with go.mod not found")
}

// TestHelloWorldE2E builds and runs the hello binary and checks its output.
func TestHelloWorldE2E(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	root, err := findProjectRoot(wd)
	if err != nil {
		t.Fatalf("find project root failed: %v", err)
	}

	// Ensure binary is built in project root
	build := exec.CommandContext(t.Context(), "make", "build")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		msg := string(out)
		t.Fatalf("build failed: %v\n%s", err, msg)
	}

	// Execute the binary from project root
	binPath := filepath.Join(root, "bin", "hello")
	run := exec.CommandContext(t.Context(), binPath)
	stdout, err := run.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout: %v", err)
	}
	if err := run.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}
	if err := run.Wait(); err != nil {
		t.Fatalf("binary exited with error: %v", err)
	}

	output := strings.Join(lines, "\n")
	if strings.TrimSpace(output) != "hello world" {
		t.Fatalf("unexpected output: %q", output)
	}
}
