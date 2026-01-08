// Package workflow implements the core business logic for fluxid workflows.
package workflow

import (
	"errors"
	"fluxid-cli/internal/types"
	"fmt"
	"log"
	"strings"
)

var (
	errImplementPhaseFailed = errors.New("implement phase failed")
	errCommitPhaseFailed    = errors.New("commit phase failed")
)

const (
	statusPass    = "PASS"
	statusFail    = "FAIL"
	builtInPrompt = "built-in prompt"

	implementPrompt = "Run implement command file for task: ${FLUXID_TASK_FILE}"
	commitPrompt    = "Execute commit command file to create git commit"
	reviewPrompt    = "Run review command file for task: ${FLUXID_TASK_FILE}"

	// Error messages for path resolution failures.
	unableToResolveReportPath  = "<unable to resolve report path>"
	unableToResolveHistoryPath = "<unable to resolve history path>"

	// UI formatting.
	separatorWidth = 70
)

// AbortError represents a workflow abort with specific exit code.
type AbortError struct {
	ExitCode int
	Message  string
}

func (e *AbortError) Error() string {
	return e.Message
}

// Run executes the main workflow loop.
func Run(cfg types.Config) (int, error) {
	// Outer loop: Development iterations (1-N)
	for iteration := 1; iteration <= cfg.MaxReviewCycles; iteration++ {
		// Print development iteration header.
		separator := strings.Repeat("━", separatorWidth)
		log.Println(separator)
		log.Printf(" DEVELOPMENT ITERATION %d/%d", iteration, cfg.MaxReviewCycles)
		log.Println(separator)

		// Run implement phase with retries
		if exitCode, err := runImplementPhase(cfg); err != nil {
			return exitCode, err
		}

		// Run review phase
		status, exitCode, err := runReviewPhase(cfg)
		if err != nil {
			return exitCode, err
		}

		// Print review status with formatted output
		printReviewStatus(cfg.SessionID, cfg.SessionRoot, status, iteration, cfg.MaxReviewCycles)

		if status == statusPass {
			return 0, nil
		}

		// status == "FAIL": continue to next development iteration (already printed in printReviewStatus)
	}

	// All iterations exhausted
	printIterationsExhausted(cfg.MaxReviewCycles)
	return 0, nil
}

// runImplementPhase executes the implement phase with retries until PASS or max retries reached.
func runImplementPhase(cfg types.Config) (int, error) {
	for retry := 1; retry <= cfg.MaxImplementRetries; retry++ {
		// Check for abort before implement attempt
		if exitCode, err := checkAbortBeforeImplement(cfg.SessionID); err != nil {
			return exitCode, err
		}

		// Run implement phase (agent execution) - errors are expected, report status matters
		_ = executeImplementPhaseQuietly(cfg, retry)

		// Check implement report status IMMEDIATELY after implement phase
		// CRITICAL: This must happen BEFORE executeCommit(), otherwise the commit phase
		// will overwrite the implement report with a commit report
		status, exitCode, err := checkImplementReportStatusWithStatus(cfg.SessionID, cfg.SessionRoot, retry)
		if err != nil {
			return exitCode, err
		}

		// Print formatted status with report file link
		printImplementAttemptStatus(cfg.SessionID, cfg.SessionRoot, retry, cfg.MaxImplementRetries, status)

		if status == statusPass {
			// Implement phase succeeded, run commit and return success
			if exitCode, err := executeCommit(cfg); err != nil {
				return exitCode, err
			}
			return 0, nil
		}

		// status == statusFail: continue to next retry (already printed status above)
	}

	// All retries exhausted, but continue to commit/review phases
	printImplementExhausted(cfg.MaxImplementRetries)

	// Run commit phase even when all implement retries failed
	// This ensures the commit phase executes regardless of implement status
	if exitCode, err := executeCommit(cfg); err != nil {
		return exitCode, err
	}

	return 0, nil
}

func checkAbortBeforeImplement(_ string) (int, error) {
	// Abort mechanism removed per 001-report-history-refactor
	// This function is retained for API consistency but always returns success
	return 0, nil
}

func executeImplementPhase(cfg types.Config) (int, error) {
	exitCode, err := runPhase(cfg, "implement", implementPrompt)
	if err != nil {
		// runPhase always returns non-zero exit code on error
		return exitCode, fmt.Errorf("implement phase failed with exit code %d: %w", exitCode, errImplementPhaseFailed)
	}
	return 0, nil
}

// executeImplementPhaseQuietly runs the implement phase without detailed error logging.
// Used during retry loops where report status is what matters, not agent exit codes.
func executeImplementPhaseQuietly(cfg types.Config, _ int) error {
	_, err := runPhase(cfg, "implement", implementPrompt)
	// Don't log errors - report status is what matters
	return err
}

func executeCommit(cfg types.Config) (int, error) {
	return runCommitPhaseWithRetry(cfg)
}

func checkImplementReportStatus(sessionID string, _ int) (int, error) {
	status, err := waitForValidReport(sessionID, "", "implement")
	if err != nil {
		// Note: AbortError handling removed per 001-report-history-refactor
		// If abort functionality is needed in future, restore error type checking here
		return 1, fmt.Errorf("failed to get implement report: %w", err)
	}

	if status == statusPass {
		return 0, nil
	}

	return -1, nil // Signal to continue retry
}

// checkImplementReportStatusWithStatus checks the implement report and returns the status string.
// Returns: (status string, exitCode, error).
func checkImplementReportStatusWithStatus(sessionID string, sessionRoot string, _ int) (string, int, error) {
	status, err := waitForValidReport(sessionID, sessionRoot, "implement")
	if err != nil {
		// Note: AbortError handling removed per 001-report-history-refactor
		// If abort functionality is needed in future, restore error type checking here
		return "", 1, fmt.Errorf("failed to get implement report: %w", err)
	}

	if status == statusPass {
		return statusPass, 0, nil
	}

	return statusFail, -1, nil // Signal to continue retry
}

// runCommitPhaseWithRetry executes the commit phase with retries until PASS or max retries reached.
func runCommitPhaseWithRetry(cfg types.Config) (int, error) {
	for retry := 1; retry <= cfg.MaxCommitRetries; retry++ {
		// Check for abort before commit attempt
		if exitCode, err := checkAbortBeforeCommit(cfg.SessionID); err != nil {
			return exitCode, err
		}

		// Run commit phase (agent execution) - errors are expected, report status matters
		_ = executeCommitPhaseQuietly(cfg, retry)

		// Check commit report status IMMEDIATELY after commit phase
		status, exitCode, err := checkCommitReportStatusWithStatus(cfg.SessionID, cfg.SessionRoot, retry)
		if err != nil {
			return exitCode, err
		}

		// Print formatted status with report file link
		printCommitAttemptStatus(cfg.SessionID, cfg.SessionRoot, retry, cfg.MaxCommitRetries, status)

		if status == statusPass {
			// Commit phase succeeded, return success
			return 0, nil
		}

		// status == statusFail: continue to next retry (already printed status above)
	}

	// All retries exhausted - workflow cannot continue
	printCommitExhausted(cfg.MaxCommitRetries)
	return 1, fmt.Errorf("commit phase failed after %d retries: %w", cfg.MaxCommitRetries, errCommitPhaseFailed)
}

func checkAbortBeforeCommit(_ string) (int, error) {
	// Abort mechanism removed per 001-report-history-refactor
	// This function is retained for API consistency but always returns success
	return 0, nil
}

func executeCommitPhase(cfg types.Config, retry int) (int, error) {
	exitCode, err := runPhase(cfg, "commit", commitPrompt)
	if err != nil {
		// runPhase always returns non-zero exit code on error
		log.Printf(
			"Commit phase failed (retry %d/%d) with exit code %d: %v",
			retry, cfg.MaxCommitRetries, exitCode, err,
		)
		return exitCode, fmt.Errorf("commit phase failed with exit code %d: %w", exitCode, errCommitPhaseFailed)
	}
	return 0, nil
}

// executeCommitPhaseQuietly runs the commit phase without detailed error logging.
// Used during retry loops where report status is what matters, not agent exit codes.
func executeCommitPhaseQuietly(cfg types.Config, _ int) error {
	_, err := runPhase(cfg, "commit", commitPrompt)
	// Don't log errors - report status is what matters
	return err
}

func checkCommitReportStatus(sessionID string, sessionRoot string, _ int) (int, error) {
	status, err := waitForValidReport(sessionID, sessionRoot, "commit")
	if err != nil {
		// Note: AbortError handling removed per 001-report-history-refactor
		// If abort functionality is needed in future, restore error type checking here
		return 1, fmt.Errorf("failed to get commit report: %w", err)
	}

	if status == statusPass {
		return 0, nil
	}

	return -1, nil // Signal to continue retry
}

// checkCommitReportStatusWithStatus checks the commit report and returns the status string.
// Returns: (status string, exitCode, error).
func checkCommitReportStatusWithStatus(sessionID string, sessionRoot string, _ int) (string, int, error) {
	status, err := waitForValidReport(sessionID, sessionRoot, "commit")
	if err != nil {
		// Note: AbortError handling removed per 001-report-history-refactor
		// If abort functionality is needed in future, restore error type checking here
		return "", 1, fmt.Errorf("failed to get commit report: %w", err)
	}

	if status == statusPass {
		return statusPass, 0, nil
	}

	return statusFail, -1, nil // Signal to continue retry
}

// runCommitPhase executes the commit phase once without retry logic.
// This is used by tests that want to test a single commit execution.
func runCommitPhase(cfg types.Config) (int, error) {
	log.Println("Running commit phase...")
	exitCode, err := runPhase(cfg, "commit", commitPrompt)
	if err != nil {
		// runPhase always returns non-zero exit code on error
		return exitCode, fmt.Errorf("commit phase failed with exit code %d: %w", exitCode, errCommitPhaseFailed)
	}
	return 0, nil
}

// runReviewPhase executes the review phase and returns the status.
func runReviewPhase(cfg types.Config) (string, int, error) {
	// Run review phase (agent execution) - errors are expected, report status matters
	_, _ = runPhase(cfg, "review", reviewPrompt)

	// Wait for valid review report and check status
	status, err := waitForValidReport(cfg.SessionID, cfg.SessionRoot, "review")
	if err != nil {
		// Note: AbortError handling removed per 001-report-history-refactor
		// If abort functionality is needed in future, restore error type checking here
		return "", 1, fmt.Errorf("failed to get review report: %w", err)
	}

	return status, 0, nil
}
