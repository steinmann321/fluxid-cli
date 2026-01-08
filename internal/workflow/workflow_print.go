// Package workflow implements the core business logic for fluxid workflows.
package workflow

import (
	"fluxid-cli/internal/storage"
	"fmt"
	"log"
	"strings"
)

// printPhaseTransition prints a unified transition box showing completed phase and next phase.
func printPhaseTransition(sessionID, sessionRoot, completedPhase, status, nextPhase string) {
	// Get report file path
	reportPath, err := storage.ResolveSessionPath(sessionID, "report.yaml", sessionRoot)
	if err != nil {
		reportPath = unableToResolveReportPath
	}

	// Get history file path
	historyPath, err := storage.ResolveSessionPath(sessionID, "history.yaml", sessionRoot)
	if err != nil {
		historyPath = unableToResolveHistoryPath
	}

	// Convert to file:// URLs for clickability in terminals
	reportURL := "file://" + reportPath
	historyURL := "file://" + historyPath

	// Print unified transition box
	separator := strings.Repeat("━", separatorWidth)
	log.Println(separator)
	log.Println()
	log.Printf(" PHASE COMPLETED: %s", completedPhase)
	log.Println()
	log.Printf("Session ID: %s", sessionID)
	log.Println()
	log.Printf("Status:  %s", status)
	log.Printf("Report:  %s", reportURL)
	log.Printf("History: %s", historyURL)
	log.Println()
	if nextPhase != "" {
		log.Printf("Next Phase: %s", nextPhase)
		log.Println()
	}
	log.Println(separator)
}

// printImplementAttemptStatus prints a formatted status message for an implement attempt.
func printImplementAttemptStatus(
	sessionID, sessionRoot string,
	retry, maxRetries int,
	status string,
	maxCommitRetries int,
) {
	completedPhase := formatImplementPhase(retry, maxRetries)
	nextPhase := determineNextPhaseAfterImplement(status, retry, maxRetries, maxCommitRetries)
	printPhaseTransition(sessionID, sessionRoot, completedPhase, status, nextPhase)
}

// formatImplementPhase formats the implement phase label.
func formatImplementPhase(retry, maxRetries int) string {
	return fmt.Sprintf("IMPLEMENT - Attempt %d/%d", retry, maxRetries)
}

// determineNextPhaseAfterImplement determines what phase comes after implement.
func determineNextPhaseAfterImplement(status string, retry, maxRetries, maxCommitRetries int) string {
	switch {
	case status == statusPass:
		return fmt.Sprintf("COMMIT - Attempt 1/%d", maxCommitRetries)
	case retry < maxRetries:
		return formatImplementPhase(retry+1, maxRetries)
	default:
		// Last retry failed, continuing to commit anyway
		return fmt.Sprintf("COMMIT - Attempt 1/%d", maxCommitRetries)
	}
}

// printImplementExhausted prints a warning when all implement attempts fail.
func printImplementExhausted(maxRetries int) {
	separator := strings.Repeat("━", separatorWidth)
	log.Println(separator)
	log.Printf("WARNING: All %d implement attempts resulted in FAIL reports", maxRetries)
	log.Println("Action: Proceeding to commit phase")
	log.Println(separator)
}

// printCommitAttemptStatus prints a formatted status message for a commit attempt.
func printCommitAttemptStatus(sessionID, sessionRoot string, retry, maxRetries int, status string) {
	completedPhase := formatCommitPhase(retry, maxRetries)
	nextPhase := determineNextPhaseAfterCommit(status, retry, maxRetries)
	printPhaseTransition(sessionID, sessionRoot, completedPhase, status, nextPhase)
}

// formatCommitPhase formats the commit phase label.
func formatCommitPhase(retry, maxRetries int) string {
	return fmt.Sprintf("COMMIT - Attempt %d/%d", retry, maxRetries)
}

// determineNextPhaseAfterCommit determines what phase comes after commit.
func determineNextPhaseAfterCommit(status string, retry, maxRetries int) string {
	switch {
	case status == statusPass:
		return "REVIEW"
	case retry < maxRetries:
		return formatCommitPhase(retry+1, maxRetries)
	default:
		// Last retry failed, workflow terminates
		return ""
	}
}

// printCommitExhausted prints an error when all commit attempts fail.
func printCommitExhausted(maxRetries int) {
	separator := strings.Repeat("━", separatorWidth)
	log.Println(separator)
	log.Printf("ERROR: All %d commit attempts resulted in FAIL reports", maxRetries)
	log.Println("Action: Workflow cannot continue")
	log.Println(separator)
}

// printReviewStatus prints a formatted status message for the review phase.
func printReviewStatus(
	sessionID, sessionRoot string,
	status string,
	currentIteration, maxIterations int,
	maxImplementRetries int,
) {
	completedPhase := formatReviewPhase(currentIteration, maxIterations)
	nextPhase := determineNextPhaseAfterReview(status, currentIteration, maxIterations, maxImplementRetries)
	printPhaseTransition(sessionID, sessionRoot, completedPhase, status, nextPhase)
}

// formatReviewPhase formats the review phase label.
func formatReviewPhase(currentIteration, maxIterations int) string {
	return fmt.Sprintf("REVIEW CYCLE %d/%d", currentIteration, maxIterations)
}

// determineNextPhaseAfterReview determines what phase comes after review.
func determineNextPhaseAfterReview(status string, currentIteration, maxIterations, maxImplementRetries int) string {
	switch {
	case status == statusPass:
		return "" // Workflow complete
	case currentIteration < maxIterations:
		return fmt.Sprintf("IMPLEMENT - Attempt 1/%d", maxImplementRetries)
	default:
		// Last iteration failed, workflow complete with issues
		return ""
	}
}

// printIterationsExhausted prints a warning when all development iterations fail.
func printIterationsExhausted(maxIterations int) {
	separator := strings.Repeat("━", separatorWidth)
	log.Println(separator)
	log.Printf("WARNING: All %d development iterations resulted in FAIL reports", maxIterations)
	log.Println("Action: Workflow completed with unresolved issues")
	log.Println(separator)
}
