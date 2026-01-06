//nolint:paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"os"
	"path/filepath"
	"testing"
)

func TestRunImplementPhase_WithAbort(t *testing.T) {
	t.Skip("Abort mechanism removed in 001-report-history-refactor - out of scope")
	// Test that implement phase checks abort flag
	sessionID := "c1d2e3f4-5a6b-7c8d-9e0f-1a2b3c4d5e6f"
	tmpDir, cleanup := setupTestDataDir(t)
	defer cleanup()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	// Set abort flag
	// SKIP: Abort removed in 001-refactor
	/*if err := ipc.SetAbortFlag(sessionID); err != nil {
		t.Fatalf("Failed to set abort flag: %v", err)
	}*/

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "echo",
		MaxReviewCycles:     1,
		MaxImplementRetries: 3,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error for aborted implement phase")
	}
	if exitCode != 130 {
		t.Errorf("Expected exit code 130 for abort, got %d", exitCode)
	}
}
