// Package command implements the CLI definition layer for fluxid.
package command

import (
	"fmt"
	"io"
	"os"
)

// printHelp prints help text to the given writer, ignoring errors.
func printHelp(w io.Writer, text string) {
	_, _ = fmt.Fprint(w, text)
}

// printUsage prints the usage information for the fluxid CLI.
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  fluxid init [path]\n")
	fmt.Fprintf(os.Stderr, "  fluxid --claude [--fluxid-iterations N] [--fluxid-implement-retries R] [claude-args]\n")
	fmt.Fprintf(os.Stderr, "  fluxid --write-history <message> [--help]\n")
	fmt.Fprintf(os.Stderr, "  fluxid ipc get-report-schema [--help]\n")
	fmt.Fprintf(os.Stderr, "  fluxid ipc write-report [--session ID] [--help]\n")
	fmt.Fprintf(os.Stderr, "  fluxid ipc read-report [--session ID] [--help]\n")
	fmt.Fprintf(os.Stderr, "  fluxid ipc abort [--session ID] [--help]\n")
	fmt.Fprintf(os.Stderr, "  fluxid ipc write-history <message> [--session ID] [--help]\n")
	fmt.Fprintf(os.Stderr, "  fluxid ipc view-history [--session ID] [--help]\n")
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  init [path]              Initialize fluxid configuration (global or project)\n")
	fmt.Fprintf(os.Stderr, "  (default)                Run workflow controller (requires --claude)\n")
	fmt.Fprintf(os.Stderr, "  --write-history          Append timestamped history entry to session\n")
	fmt.Fprintf(os.Stderr, "  ipc get-report-schema    Print YAML schema for workflow reports\n")
	fmt.Fprintf(os.Stderr, "  ipc write-report         Write and validate a report from stdin\n")
	fmt.Fprintf(os.Stderr, "  ipc read-report          Read stored report for session\n")
	fmt.Fprintf(os.Stderr, "  ipc abort                Request graceful abort of running workflow\n")
	fmt.Fprintf(os.Stderr, "  ipc write-history        Append timestamped history entry via IPC\n")
	fmt.Fprintf(os.Stderr, "  ipc view-history         Display session history in chronological order\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  --claude                      Enable workflow mode (required for workflow)\n")
	fmt.Fprintf(os.Stderr, "  --fluxid-iterations N         Set max review cycles (default: 20)\n")
	fmt.Fprintf(os.Stderr, "  --fluxid-implement-retries R  Set max implement retries (default: 3)\n")
	fmt.Fprintf(os.Stderr, "  --fluxid-dry-run              Run simulation without executing agent\n")
	fmt.Fprintf(os.Stderr, "  --fluxid-output {text|json|yaml}  Set initialization output format\n")
	fmt.Fprintf(os.Stderr, "  --session ID                  Specify session ID (overrides FLUXID_SESSION_ID)\n")
	fmt.Fprintf(os.Stderr, "  --help                        Show help information\n")
}

// handleIPCCommand processes IPC subcommands.
func handleIPCCommand(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: 'ipc' requires a subcommand\n")
		fmt.Fprintf(
			os.Stderr,
			"Usage: fluxid ipc {get-report-schema|write-report|read-report|abort|write-history|view-history} [--help]\n",
		)
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
	case "abort":
		return handleAbort(subArgs)
	case "write-history":
		return handleIPCWriteHistory(subArgs)
	case "view-history":
		return handleViewHistory(subArgs)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown ipc subcommand: %s\n", subcommand)
		fmt.Fprintf(
			os.Stderr,
			"Available subcommands: get-report-schema, write-report, read-report, "+
				"abort, write-history, view-history\n",
		)
		return 1
	}
}
