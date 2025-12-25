//nolint:paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunImplementPhase_WithAbort(t *testing.T) {
	// Test that implement phase checks abort flag
	sessionID := "test-implement-abort-session-" + time.Now().Format("20060102150405.000000")
	tmpDir, cleanup := setupTestDataDir(t)
	defer cleanup()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Set abort flag
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "echo",
		MaxReviewCycles:     1,
		MaxImplementRetries: 3,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		Sources:             map[string]string{},
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error for aborted implement phase")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}
