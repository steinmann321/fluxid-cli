// Package main implements the fluxid CLI workflow controller for coding agents.
package main

import (
	"fluxid-loop/internal/ipc"
	"fmt"
	"io"
	"os"
	"strings"
)

// printHelp prints help text to the given writer, ignoring errors.
func printHelp(w io.Writer, text string) {
	_, _ = fmt.Fprint(w, text)
}

// printUsage prints the usage information for the fluxid CLI.
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  fluxid --claude [--fluxid-iterations N] [--fluxid-implement-retries R]\n")
	fmt.Fprintf(os.Stderr, "         [--fluxid-no-commit] [claude-args]\n")
	fmt.Fprintf(os.Stderr, "  fluxid ipc get-report-schema [--help]\n")
	fmt.Fprintf(os.Stderr, "  fluxid ipc write-report [--session ID] [--help]\n")
	fmt.Fprintf(os.Stderr, "  fluxid ipc read-report [--session ID] [--help]\n")
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  (default)                Run workflow controller (requires --claude)\n")
	fmt.Fprintf(os.Stderr, "  ipc get-report-schema    Print YAML schema for workflow reports\n")
	fmt.Fprintf(os.Stderr, "  ipc write-report         Write and validate a report from stdin\n")
	fmt.Fprintf(os.Stderr, "  ipc read-report          Read stored report for session\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --claude                 Enable workflow mode (required for workflow)\n")
	fmt.Fprintf(os.Stderr, "  --fluxid-iterations N    Set max review cycles (default: 20)\n")
	fmt.Fprintf(os.Stderr, "  --fluxid-implement-retries R  Set max implement retries (default: 3)\n")
	fmt.Fprintf(os.Stderr, "  --fluxid-no-commit       Disable commit phase\n")
	fmt.Fprintf(os.Stderr, "  --session ID             Specify session ID (overrides FLUXID_SESSION_ID)\n")
	fmt.Fprintf(os.Stderr, "  --help                   Show help information\n")
}

// handleIPCCommand processes IPC subcommands.
func handleIPCCommand(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: 'ipc' requires a subcommand\n")
		fmt.Fprintf(os.Stderr, "Usage: fluxid ipc {get-report-schema|write-report|read-report} [--help]\n")
		return 1
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "get-report-schema":
		return handleGetReportSchema(subArgs)
	case "write-report":
		return handleWriteReport(subArgs)
	case "read-report":
		return handleReadReport(subArgs)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown ipc subcommand: %s\n", subcommand)
		fmt.Fprintf(os.Stderr, "Available subcommands: get-report-schema, write-report, read-report\n")
		return 1
	}
}

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

// parseSessionFlag parses --session flag from args and returns session ID.
// Falls back to FLUXID_SESSION_ID environment variable if flag not provided.
func parseSessionFlag(args []string) (string, error) {
	// Look for --session flag
	for i := 0; i < len(args); i++ {
		if args[i] == "--session" {
			if i+1 >= len(args) {
				return "", fmt.Errorf("--session flag requires a value")
			}
			return args[i+1], nil
		}
	}

	// Fall back to environment variable
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		return "", fmt.Errorf("session ID not provided: use --session flag or set FLUXID_SESSION_ID environment variable")
	}

	return sessionID, nil
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
