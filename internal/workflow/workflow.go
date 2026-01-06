// Package workflow implements the core business logic for fluxid workflows.
package workflow

import (
	"errors"
	"fluxid-cli/internal/types"
	"fmt"
	"log"
)

var (
	errImplementPhaseFailed = errors.New("implement phase failed")
	errCommitPhaseFailed    = errors.New("commit phase failed")
	errReviewPhaseFailed    = errors.New("review phase failed")
)

const (
	statusPass    = "PASS"
	statusFail    = "FAIL"
	builtInPrompt = "built-in prompt"

	implementPrompt = "Run implement command file for task: ${FLUXID_TASK_FILE}"
	commitPrompt    = "Execute commit command file to create git commit"
	reviewPrompt    = "Run review command file for task: ${FLUXID_TASK_FILE}"
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
	// Outer loop: Review cycles (1-N)
	for reviewCycle := 1; reviewCycle <= cfg.MaxReviewCycles; reviewCycle++ {
		log.Printf("--- Review Cycle %d/%d ---", reviewCycle, cfg.MaxReviewCycles)

		// Run implement phase with retries
		if exitCode, err := runImplementPhase(cfg); err != nil {
			return exitCode, err
		}

		// Run review phase
		status, exitCode, err := runReviewPhase(cfg)
		if err != nil {
			return exitCode, err
		}

		if status == statusPass {
			log.Println("Workflow completed successfully.")
			break
		}

		// status == "FAIL": continue to next review cycle
		log.Printf("Review report status is FAIL, continuing to next cycle... (%d/%d)", reviewCycle+1, cfg.MaxReviewCycles)
	}

	return 0, nil
}

// runImplementPhase executes the implement phase with retries until PASS or max retries reached.
func runImplementPhase(cfg types.Config) (int, error) {
	for retry := 1; retry <= cfg.MaxImplementRetries; retry++ {
		log.Printf("Implement attempt %d/%d...", retry, cfg.MaxImplementRetries)

		// Check for abort before implement attempt
		if exitCode, err := checkAbortBeforeImplement(cfg.SessionID); err != nil {
			return exitCode, err
		}

		// Run implement phase
		if exitCode, err := executeImplementPhase(cfg, retry); err != nil {
			if exitCode != 0 {
				return exitCode, err
			}
			continue
		}

		// Check implement report status IMMEDIATELY after implement phase
		// CRITICAL: This must happen BEFORE executeCommit(), otherwise the commit phase
		// will overwrite the implement report with a commit report
		if exitCode, err := checkImplementReportStatus(cfg.SessionID, cfg.SessionRoot, retry); err != nil {
			return exitCode, err
		} else if exitCode == 0 {
			// Implement phase succeeded, run commit and return success
			if exitCode, err := executeCommit(cfg); err != nil {
				return exitCode, err
			}
			return 0, nil
		}

		// status == statusFail: continue to next retry
		log.Printf("Implement report status is FAIL, retrying... (%d/%d)", retry+1, cfg.MaxImplementRetries)
	}

	// All retries exhausted, but continue to commit/review phases
	retries := cfg.MaxImplementRetries
	log.Printf("Implement phase retries exhausted (%d/%d), continuing to next phase", retries, retries)

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

func executeImplementPhase(cfg types.Config, retry int) (int, error) {
	exitCode, err := runPhase(cfg, "implement", implementPrompt)
	if err != nil {
		// runPhase always returns non-zero exit code on error
		log.Printf(
			"Implement phase failed (retry %d/%d) with exit code %d: %v",
			retry, cfg.MaxImplementRetries, exitCode, err,
		)
		return exitCode, fmt.Errorf("implement phase failed with exit code %d: %w", exitCode, errImplementPhaseFailed)
	}
	return 0, nil
}

func executeCommit(cfg types.Config) (int, error) {
	return runCommitPhaseWithRetry(cfg)
}

func checkImplementReportStatus(sessionID string, sessionRoot string, _ int) (int, error) {
	status, err := waitForValidReport(sessionID, sessionRoot, "implement")
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

// runCommitPhaseWithRetry executes the commit phase with retries until PASS or max retries reached.
func runCommitPhaseWithRetry(cfg types.Config) (int, error) {
	for retry := 1; retry <= cfg.MaxCommitRetries; retry++ {
		log.Printf("Commit attempt %d/%d...", retry, cfg.MaxCommitRetries)

		// Check for abort before commit attempt
		if exitCode, err := checkAbortBeforeCommit(cfg.SessionID); err != nil {
			return exitCode, err
		}

		// Run commit phase
		if exitCode, err := executeCommitPhase(cfg, retry); err != nil {
			if exitCode != 0 {
				return exitCode, err
			}
			continue
		}

		// Check commit report status IMMEDIATELY after commit phase
		if exitCode, err := checkCommitReportStatus(cfg.SessionID, cfg.SessionRoot, retry); err != nil {
			return exitCode, err
		} else if exitCode == 0 {
			// Commit phase succeeded (REPORT says PASS)
			log.Println("Commit phase completed successfully with PASS report")
			return 0, nil
		}

		// status == statusFail: continue to next retry
		log.Printf("Commit report status is FAIL, retrying... (%d/%d)", retry+1, cfg.MaxCommitRetries)
	}

	// All retries exhausted
	retries := cfg.MaxCommitRetries
	log.Printf("Commit phase retries exhausted (%d/%d), workflow cannot continue", retries, retries)
	return 1, fmt.Errorf("commit phase failed after %d retries: %w", retries, errCommitPhaseFailed)
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
	log.Println("Running review phase...")
	exitCode, err := runPhase(cfg, "review", reviewPrompt)
	if err != nil {
		// runPhase always returns non-zero exit code on error
		return "", exitCode, fmt.Errorf("review phase failed with exit code %d: %w", exitCode, errReviewPhaseFailed)
	}

	// Wait for valid review report and check status
	status, err := waitForValidReport(cfg.SessionID, cfg.SessionRoot, "review")
	if err != nil {
		// Note: AbortError handling removed per 001-report-history-refactor
		// If abort functionality is needed in future, restore error type checking here
		return "", 1, fmt.Errorf("failed to get review report: %w", err)
	}

	return status, 0, nil
}
