// Package command implements the CLI definition layer for fluxid.
package command

import (
	"errors"
	"fluxid-loop/internal/ipc"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	errSessionFlagMissingValue = errors.New("--session flag requires a value")
	errSessionIDNotProvided    = errors.New(
		"session ID not provided: use --session flag or set FLUXID_SESSION_ID environment variable",
	)
)

// handleGetReportSchema processes the get-report-schema command.
func handleGetReportSchema(args []string) int {
	// Check for --help flag
	for _, arg := range args {
		if arg == flagHelp || arg == "-h" {
			helpText := `Usage: fluxid ipc get-report-schema

Description:
  Prints the YAML schema for workflow reports to stdout.
  The schema defines required fields, types, and constraints
  for valid IPC reports used in the fluxid workflow.

Example:
  fluxid ipc get-report-schema > schema.yaml
  fluxid ipc get-report-schema | yq eval
`
			printHelp(os.Stdout, helpText)
			return 0
		}
	}

	// Output the schema to stdout
	if err := ipc.WriteReportSchema(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write report schema: %v\n", err)
		return 1
	}

	return 0
}

// handleWriteReport processes the write-report command.
func handleWriteReport(args []string) int {
	// Check for --help flag
	for _, arg := range args {
		if arg == flagHelp || arg == "-h" {
			helpText := `Usage: fluxid ipc write-report [--session ID]

Description:
  Reads a YAML report from stdin, validates it against the schema,
  and stores it for the current session. Reports are stored in-memory
  and scoped by session ID.

Options:
  --session ID    Session ID (overrides FLUXID_SESSION_ID env var)

Example:
  fluxid ipc write-report < report.yaml
  cat report.yaml | fluxid ipc write-report --session my-session
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

	// Read report from stdin
	reportBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read stdin: %v\n", err)
		return 1
	}

	reportYAML := string(reportBytes)

	// Validate report
	if err := ipc.ValidateReport(reportYAML); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Store report
	if err := ipc.WriteReport(sessionID, reportYAML); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write report: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Report stored successfully for session: %s\n", sessionID)
	return 0
}

// handleReadReport processes the read-report command.
func handleReadReport(args []string) int {
	// Check for --help flag
	for _, arg := range args {
		if arg == flagHelp || arg == "-h" {
			helpText := `Usage: fluxid ipc read-report [--session ID]

Description:
  Reads the stored report for the current session and outputs it
  as YAML to stdout. Returns exit code 0 even if no report exists.

Options:
  --session ID    Session ID (overrides FLUXID_SESSION_ID env var)

Example:
  fluxid ipc read-report
  fluxid ipc read-report --session my-session > stored-report.yaml
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

	// Read report
	reportYAML, err := ipc.ReadReport(sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read report: %v\n", err)
		return 1
	}

	// Output report if it exists
	if reportYAML != "" {
		_, _ = os.Stdout.WriteString(strings.TrimSpace(reportYAML))
		_, _ = os.Stdout.WriteString("\n")
	} else {
		fmt.Fprintf(os.Stderr, "No report found for session: %s\n", sessionID)
	}

	return 0
}

// parseSessionFlag parses --session flag from args and returns session ID.
// Falls back to FLUXID_SESSION_ID environment variable if flag not provided.
func parseSessionFlag(args []string) (string, error) {
	// Look for --session flag
	for i := 0; i < len(args); i++ {
		if args[i] == "--session" {
			if i+1 >= len(args) {
				return "", errSessionFlagMissingValue
			}
			return args[i+1], nil
		}
	}

	// Fall back to environment variable
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		return "", errSessionIDNotProvided
	}

	return sessionID, nil
}
