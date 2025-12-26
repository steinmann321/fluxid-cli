// Package command implements the CLI definition layer for fluxid.
package command

import (
	"fluxid-cli/internal/ipc"
	"fmt"
	"os"
)

// handleAbort processes the abort command.
func handleAbort(args []string) int {
	// Check for --help flag
	for _, arg := range args {
		if arg == flagHelp || arg == "-h" {
			helpText := `Usage: fluxid ipc abort [--session ID]

Description:
  Requests a graceful abort of the running workflow for the current session.
  Sets an abort flag that the workflow checks between phases. The workflow
  will complete the current agent invocation and then exit cleanly.

Options:
  --session ID    Session ID (overrides FLUXID_SESSION_ID env var)

Example:
  fluxid ipc abort
  fluxid ipc abort --session my-session
`
			printHelp(os.Stdout, helpText)
			return 0
		}
	}

	// Parse session ID
	sessionID, err := parseSessionFlag(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Set abort flag
	if err := ipc.SetAbortFlag(sessionID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to set abort flag: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Abort requested for session: %s\n", sessionID)
	fmt.Fprintf(os.Stderr, "Workflow will exit gracefully after current phase completes.\n")
	return 0
}
