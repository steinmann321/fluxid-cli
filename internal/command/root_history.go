package command

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
		// Use ErrorWriter to properly classify errors and determine exit code
		// Per FR-006, FR-007, FR-040: Different error types have different exit codes
		errorWriter := NewErrorWriter()
		return errorWriter.WriteError(err, "")
	}

	return ExitSuccess
}
