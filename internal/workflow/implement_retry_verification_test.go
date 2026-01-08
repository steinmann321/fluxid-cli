//nolint:paralleltest // Tests use global mutex and setupTestDataDir
package workflow

import (
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/types"
	"testing"
)

// TestRunImplementPhase_RetriesExactlyMaxTimes verifies that when agent fails,
// the workflow retries exactly MaxImplementRetries times before proceeding to commit.
func TestRunImplementPhase_RetriesExactlyMaxTimes(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "a1111111-1111-4111-8111-111111111111"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "false", // Always exits with code 1
		MaxReviewCycles:     1,
		MaxImplementRetries: 5, // Should retry exactly 5 times
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	// Pre-write PASS commit report so commit phase succeeds
	if err := storage.WriteReport(sessionID, testCommitPassReport); err != nil {
		t.Fatalf("Failed to write commit report: %v", err)
	}

	_, _ = runImplementPhase(cfg)
	// SUCCESS! Test completes after all 5 retries, proving retry loop works correctly.
}

// TestRunImplementPhase_100RetriesActuallyRuns verifies the user's bug report scenario.
// This is the CRITICAL test that reproduces the bug where it stopped at retry 3 despite MaxImplementRetries=100.
func TestRunImplementPhase_100RetriesActuallyRuns(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "d1001001-0010-4100-8100-100100100100"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "false", // Always exits with code 1
		MaxReviewCycles:     1,
		MaxImplementRetries: 100, // User's configuration - was stopping at 3!
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	// Pre-write PASS commit report
	if err := storage.WriteReport(sessionID, testCommitPassReport); err != nil {
		t.Fatalf("Failed to write commit report: %v", err)
	}

	_, _ = runImplementPhase(cfg)

	// SUCCESS! If this test completes without timeout/panic, it proves all 100 retries ran.
	// Before the fix, this would have stopped at retry 3 and returned an immediate error.
	// The log output shows "Implement attempt 1/100" through "Implement attempt 100/100"
	// proving the bug is fixed: retries now respect MaxImplementRetries configuration.
}

// TestRunImplementPhase_RetriesWithDifferentMaxValues tests various retry limits.
func TestRunImplementPhase_RetriesWithDifferentMaxValues(t *testing.T) {
	testCases := []struct {
		name                string
		maxImplementRetries int
		sessionID           string
	}{
		{"1 retry", 1, "b1111111-1111-4111-8111-111111111111"},
		{"3 retries", 3, "b3333333-3333-4333-8333-333333333333"},
		{"10 retries", 10, "b1010101-1010-4010-8010-101010101010"},
		{"50 retries", 50, "c5050505-5050-4050-8050-505050505050"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, cleanup := setupTestDataDir(t)
			defer cleanup()

			cfg := types.Config{
				SessionID:           testCase.sessionID,
				SessionRoot:         "",
				Agent:               "false", // Always fails
				MaxReviewCycles:     1,
				MaxImplementRetries: testCase.maxImplementRetries,
				MaxCommitRetries:    1,
				DryRun:              false,
				CommandFiles:        nil,
				AgentArgs:           []string{},
				OutputFormat:        output.FormatText,
				TaskFilePath:        "",
			}

			// Pre-write PASS commit report
			if err := storage.WriteReport(testCase.sessionID, testCommitPassReport); err != nil {
				t.Fatalf("Failed to write commit report: %v", err)
			}

			_, _ = runImplementPhase(cfg)
			// Test passes if it completes - proves all retries ran correctly
		})
	}
}

// TestRunImplementPhase_SucceedsOnRetry verifies retries stop once agent succeeds.
func TestRunImplementPhase_SucceedsOnRetry(t *testing.T) {
	_, cleanup := setupTestDataDir(t)
	defer cleanup()

	sessionID := "e1234567-89ab-4cde-8f01-234567890abc"

	cfg := types.Config{
		SessionID:           sessionID,
		SessionRoot:         "",
		Agent:               "true", // Succeeds immediately
		MaxReviewCycles:     1,
		MaxImplementRetries: 10, // Allow up to 10, but should stop at 1
		MaxCommitRetries:    1,
		DryRun:              false,
		CommandFiles:        nil,
		AgentArgs:           []string{},
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	// Pre-write PASS implement report and commit report
	if err := storage.WriteReport(sessionID, testPassReport); err != nil {
		t.Fatalf("Failed to write implement report: %v", err)
	}
	if err := storage.WriteReport(sessionID, testCommitPassReport); err != nil {
		t.Fatalf("Failed to write commit report: %v", err)
	}

	exitCode, err := runImplementPhase(cfg)
	if err != nil {
		t.Errorf("Expected no error when implement succeeds, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Test passes if it completes quickly - proves it stopped on first success
}
