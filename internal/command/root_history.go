package command

import (
	"fmt"
	"os"
)

// handleHistoryCommand handles the `fluxid history` command.
//
// This function parses command-line arguments and delegates to the appropriate handler:
// - --get-file: Get history file path
// - --validate: Validate history file
// - --get-schema: Get history schema
//
// Per User Story 4 (FR-004, FR-005, FR-006): Enable file-based history interface for agents.
func handleHistoryCommand(args []string) int {
	cmd := NewHistoryCommand()

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
