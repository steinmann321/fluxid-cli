package command

import (
	"fmt"
	"os"
)

// handleReportCommand handles the `fluxid report` command.
//
// This function parses command-line arguments and delegates to the appropriate handler:
// - --get-file: Get report file path
// - --validate: Validate report file
// - --get-schema: Get report schema
//
// Per User Story 1 (FR-001, FR-002, FR-003): Enable file-based report interface for agents.
func handleReportCommand(args []string) int {
	cmd := NewReportCommand()

	// Set args for cobra command
	cmd.SetArgs(args)

	// Execute command
	if err := cmd.Execute(); err != nil {
		// Error handling is done within individual handlers
		// They call os.Exit with appropriate exit codes
		// If we reach here, it's an unexpected error
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return ExitInternalError
	}

	return ExitSuccess
}
