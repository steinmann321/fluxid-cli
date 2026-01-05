//nolint:paralleltest // Tests use global mutex, cannot run in parallel
package workflow

import (
	"fluxid-cli/internal/ipc"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunImplementPhase_MaxRetries(t *testing.T) {
	t.Parallel()
	// Test that implement phase fails immediately when agent command doesn't exist
	cfg := types.Config{
		SessionID:           "test-retries-session",
		Agent:               "nonexistent-agent-xyz",
		MaxReviewCycles:     1,
		MaxImplementRetries: 2,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	exitCode, err := runImplementPhase(cfg)
	if err == nil {
		t.Error("Expected error when agent command doesn't exist")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	// Command not found should abort immediately with exit code error message
	if !strings.Contains(err.Error(), "failed with exit code") {
		t.Errorf("Expected exit code error message, got: %v", err)
	}
}

func TestRunImplementPhase_NonZeroExitCode(t *testing.T) {
	t.Parallel()
	// Test that implement phase aborts on non-zero exit code
	cfg := types.Config{
		SessionID:           "test-nonzero-exit",
		Agent:               "false",
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
		t.Error("Expected error for non-zero exit code")
	}
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	if !strings.Contains(err.Error(), "failed with exit code") {
		t.Errorf("Expected exit code error message, got: %v", err)
	}
}

func TestRunImplementPhase_FailRetryThenPass(t *testing.T) {
	// Test implement phase succeeds with pre-written PASS report
	sessionID := "test-implement-retry-pass-" + time.Now().Format("20060102150405.000000")
	tmpDir, cleanup := setupTestDataDir(t)
	defer cleanup()
	storageDir := filepath.Join(tmpDir, ".fluxid")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}

	cfg := types.Config{
		SessionID:           sessionID,
		Agent:               "true",
		MaxReviewCycles:     1,
		MaxImplementRetries: 3,
		MaxCommitRetries:    100,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	// Pre-write PASS report before workflow starts
	// No timing dependencies - completely deterministic
	if err := ipc.WriteReport(sessionID, testPassReport); err != nil {
		t.Fatalf("Failed to write report: %v", err)
	}

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
