// Package workflow implements the core business logic for fluxid workflows.
package workflow

import (
	"fluxid-cli/internal/types"
	"fmt"
	"log"
	"math"
)

// ExecuteStepWithRetry executes a workflow step with retry logic.
// It runs the step up to maxRetries times, checking the report status after each attempt.
// Returns nil on PASS status, or nil after exhausting retries (continues to next step).
func ExecuteStepWithRetry(cfg types.Config, step types.WorkflowStep, iteration int, logger *Logger) error {
	maxRetries := step.Retries
	if maxRetries == 0 {
		maxRetries = math.MaxInt // Infinite retries
	}

	for retry := 1; retry <= maxRetries; retry++ {
		logger.LogStepStart(step.Name, iteration, retry, maxRetries)

		// Get command file path for this step
		commandPrompt := fmt.Sprintf("Run %s command file for task: ${FLUXID_TASK_FILE}", step.Name)

		// Execute step (agent invocation)
		_, execErr := runPhase(cfg, step.Name, commandPrompt)
		if execErr != nil {
			log.Printf("Agent invocation failed for step '%s' (retry %d/%d): %v",
				step.Name, retry, maxRetries, execErr)
			// Treat as FAIL, continue to retry logic
			logger.LogStepComplete(step.Name, statusFail, retry, maxRetries)

			// Check if we should retry
			if retry < maxRetries {
				continue // Try again
			}
			// Max retries exhausted with agent failure
			logger.LogRetriesExhausted(step.Name, maxRetries)
			return nil // Continue to next step (FR-006)
		}

		// Check report status
		status, err := waitForValidReport(cfg.SessionID, cfg.SessionRoot, step.Name)
		if err != nil {
			log.Printf("Failed to read %s report (treating as FAIL): %v", step.Name, err)
			status = statusFail
		}

		logger.LogStepComplete(step.Name, status, retry, maxRetries)

		if status == statusPass {
			return nil // Success - move to next step
		}

		// Status == FAIL: check if we should retry
		if retry >= maxRetries {
			// Max retries exhausted with FAIL status
			logger.LogRetriesExhausted(step.Name, maxRetries)
			return nil // Continue to next step (FR-006)
		}

		// Continue to next retry
	}

	// Should not reach here, but if we do, continue to next step
	return nil
}
