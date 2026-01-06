package command

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
		// Use ErrorWriter to properly classify errors and determine exit code
		// Per FR-006, FR-007, FR-040: Different error types have different exit codes
		errorWriter := NewErrorWriter()
		return errorWriter.WriteError(err, "")
	}

	return ExitSuccess
}
