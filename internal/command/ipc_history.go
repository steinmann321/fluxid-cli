// Package command implements the CLI definition layer for fluxid.
package command

import (
	"fluxid-loop/internal/ipc"
	"fmt"
	"os"
	"strings"
)

// handleWriteHistory processes the --write-history flag command.
func handleWriteHistory(args []string) int {
	// Check for --help flag
	for _, arg := range args {
		if arg == flagHelp || arg == "-h" {
			helpText := `Usage: fluxid --write-history <message>

Description:
  Appends a timestamped history entry to the current session's history.
  The entry is prefixed with an ISO 8601 timestamp in the format
  [YYYY-MM-DDTHH:MM:SSZ]. History is stored in-memory and scoped
  per session. Requires FLUXID_SESSION_ID environment variable.

Example:
  fluxid --write-history "Implemented user authentication"
  fluxid --write-history "Fixed bug in login flow"
`
			printHelp(os.Stdout, helpText)
			return 0
		}
	}

	// Get session ID from environment
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		fmt.Fprintf(os.Stderr, "Error: FLUXID_SESSION_ID environment variable not set\n")
		fmt.Fprintf(os.Stderr, "History entries require an active session context.\n")
		return 1
	}

	// Get message from arguments
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: --write-history requires a message argument\n")
		fmt.Fprintf(os.Stderr, "Usage: fluxid --write-history <message>\n")
		return 1
	}

	message := strings.Join(args, " ")

	// Write history entry
	if err := ipc.WriteHistoryEntry(sessionID, message); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write history entry: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "History entry recorded for session: %s\n", sessionID)
	return 0
}

// handleIPCWriteHistory processes the ipc write-history subcommand.
func handleIPCWriteHistory(args []string) int {
	// Check for --help flag
	for _, arg := range args {
		if arg == flagHelp || arg == "-h" {
			helpText := `Usage: fluxid ipc write-history <message> [--session ID]

Description:
  Appends a timestamped history entry to the session's history.
  The entry is prefixed with an ISO 8601 timestamp in the format
  [YYYY-MM-DDTHH:MM:SSZ]. History is stored in-memory and scoped
  per session.

Options:
  --session ID    Session ID (overrides FLUXID_SESSION_ID env var)

Example:
  fluxid ipc write-history "Decision: adopt FIFO eviction"
  fluxid ipc write-history "Implemented feature X" --session my-session
`
			printHelp(os.Stdout, helpText)
			return 0
		}
	}

	// Parse session ID (supports both --session flag and FLUXID_SESSION_ID)
	sessionID, err := parseSessionFlag(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Filter out --session flag and its value to get message args
	var messageArgs []string
	for argIndex := 0; argIndex < len(args); argIndex++ {
		if args[argIndex] == "--session" {
			// Skip --session and its value
			argIndex++
			continue
		}
		messageArgs = append(messageArgs, args[argIndex])
	}

	// Get message from arguments
	if len(messageArgs) == 0 {
		fmt.Fprintf(os.Stderr, "Error: write-history requires a message argument\n")
		fmt.Fprintf(os.Stderr, "Usage: fluxid ipc write-history <message> [--session ID]\n")
		return 1
	}

	message := strings.Join(messageArgs, " ")

	// Write history entry
	if err := ipc.WriteHistoryEntry(sessionID, message); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write history entry: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "History entry recorded for session: %s\n", sessionID)
	return 0
}

// handleViewHistory processes the ipc view-history subcommand.
func handleViewHistory(args []string) int {
	// Check for --help flag
	for _, arg := range args {
		if arg == flagHelp || arg == "-h" {
			helpText := `Usage: fluxid ipc view-history [--session ID]

Description:
  Retrieves and displays all history entries for the session in chronological
  order. Each entry is displayed on a single line with an ISO 8601 timestamp
  prefix in the format [YYYY-MM-DDTHH:MM:SSZ] message.

Options:
  --session ID    Session ID (overrides FLUXID_SESSION_ID env var)

Example:
  fluxid ipc view-history
  fluxid ipc view-history --session my-session
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

	// Read history
	history, err := ipc.ReadHistory(sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read history: %v\n", err)
		return 1
	}

	// Output history entries to stdout (plain text, no extra formatting)
	if history != "" {
		_, _ = os.Stdout.WriteString(strings.TrimSpace(history))
		_, _ = os.Stdout.WriteString("\n")
	}

	return 0
}
