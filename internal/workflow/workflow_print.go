// Package workflow implements the core business logic for fluxid workflows.
package workflow

import (
	"fluxid-cli/internal/storage"
	"log"
	"strings"
)

// printImplementAttemptStatus prints a formatted status message for an implement attempt.
func printImplementAttemptStatus(sessionID, sessionRoot string, retry, maxRetries int, status string) {
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

	// Print formatted output
	separator := strings.Repeat("━", separatorWidth)
	log.Println(separator)
	log.Printf(" IMPLEMENT PHASE - Attempt %d/%d", retry, maxRetries)
	log.Println(separator)
	log.Println()
	log.Printf("Report Status: %s", status)
	log.Printf("Report File: %s", reportURL)
	log.Printf("History File: %s", historyURL)
	log.Println()

	// Print action
	switch {
	case status == statusPass:
		log.Println("Action: Proceeding to commit phase")
	case retry < maxRetries:
		log.Println("Action: Continuing implementation with feedback from report...")
	}
	// Last attempt with FAIL - no action message, will print summary next
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

	// Print formatted output
	separator := strings.Repeat("━", separatorWidth)
	log.Println(separator)
	log.Printf(" COMMIT PHASE - Attempt %d/%d", retry, maxRetries)
	log.Println(separator)
	log.Println()
	log.Printf("Report Status: %s", status)
	log.Printf("Report File: %s", reportURL)
	log.Printf("History File: %s", historyURL)
	log.Println()

	// Print action
	switch {
	case status == statusPass:
		log.Println("Action: Proceeding to review phase")
	case retry < maxRetries:
		log.Println("Action: Continuing commit with feedback from report...")
	}
	// Last attempt with FAIL - no action message, will print summary next
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
func printReviewStatus(sessionID, sessionRoot string, status string, currentIteration, maxIterations int) {
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

	// Print formatted output
	separator := strings.Repeat("━", separatorWidth)
	log.Println(separator)
	log.Printf(" REVIEW CYCLE %d/%d", currentIteration, maxIterations)
	log.Println(separator)
	log.Println()
	log.Printf("Report Status: %s", status)
	log.Printf("Report File: %s", reportURL)
	log.Printf("History File: %s", historyURL)
	log.Println()

	// Print action
	switch {
	case status == statusPass:
		log.Println("Action: Workflow completed successfully")
	case currentIteration < maxIterations:
		log.Println("Action: Starting next development iteration for improvements...")
	}
	// Last iteration with FAIL - no action message, will print summary next
}

// printIterationsExhausted prints a warning when all development iterations fail.
func printIterationsExhausted(maxIterations int) {
	separator := strings.Repeat("━", separatorWidth)
	log.Println(separator)
	log.Printf("WARNING: All %d development iterations resulted in FAIL reports", maxIterations)
	log.Println("Action: Workflow completed with unresolved issues")
	log.Println(separator)
}
