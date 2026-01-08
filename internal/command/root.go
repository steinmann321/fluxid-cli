// Package command implements the CLI definition layer for fluxid.
package command

import (
	"fluxid-cli/internal/types"
	"fluxid-cli/internal/version"
	"fluxid-cli/internal/workflow"
	"fmt"
	"os"
)

const (
	defaultMaxReviewCycles     = 20
	defaultMaxImplementRetries = 3
	flagHelp                   = "--help"
)

// Execute is the main entry point for the fluxid CLI.
func Execute() int {
	// Handle special commands first
	if exitCode, handled := handleSpecialCommands(); handled {
		return exitCode
	}

	// Load and resolve configuration
	cfg, exitCode := loadAndResolveConfig()
	if exitCode != 0 {
		return exitCode
	}

	// Execute workflow or simulation
	return executeWorkflow(cfg)
}

func handleSpecialCommands() (int, bool) {
	// Check for help flags (must be handled before config loading to avoid validation errors)
	if checkHelpFlag() {
		printHelp()
		return 0, true
	}

	// Check for subcommands that don't require config
	if len(os.Args) > 1 {
		return handleSubcommand(os.Args[1], os.Args[2:])
	}

	return 0, false
}

func checkHelpFlag() bool {
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func handleSubcommand(cmd string, args []string) (int, bool) {
	switch cmd {
	case "init":
		return handleInit(args), true
	case "version", "--version":
		_, _ = fmt.Fprintln(os.Stdout, version.Full())
		return 0, true
	case "report":
		return handleReportCommand(args), true
	case "history":
		return handleHistoryCommand(args), true
	default:
		return 0, false
	}
}

func executeWorkflow(cfg types.Config) int {
	// Set up signal handler for graceful abort (skip in dry-run mode)
	if !cfg.DryRun {
		_ = setupSignalHandler(cfg.SessionID) // Cleanup not needed in production since program exits
	}

	// Print dry-run header if in simulation mode and text format
	printDryRunHeader(cfg)

	// Build and display initialization status
	status := buildInitializationStatus(cfg)
	if exitCode := printInitializationStatus(status, cfg.OutputFormat); exitCode != 0 {
		return exitCode
	}

	// Run workflow
	workflowExitCode, workflowErr := runWorkflow(cfg)

	if workflowErr != nil {
		printWorkflowError(workflowErr, workflowExitCode, cfg.SessionID)
		return workflowExitCode
	}

	printWorkflowSuccess(cfg.SessionID)
	return 0
}

func runWorkflow(cfg types.Config) (int, error) {
	if cfg.DryRun {
		return workflow.RunSimulation(cfg), nil
	}
	exitCode, err := workflow.Run(cfg)
	if err != nil {
		return exitCode, fmt.Errorf("workflow execution failed: %w", err)
	}
	return exitCode, nil
}

func printHelp() {
	helpText := fmt.Sprintf(`fluxid %s - AI-powered workflow automation tool

USAGE:
    fluxid [OPTIONS] --<agent> --file=<task-file>
    fluxid init [OPTIONS]
    fluxid version
    fluxid report [--get-file|--validate|--get-schema]
    fluxid history [--get-file|--validate|--get-schema]

AGENTS:
    --claude            Use Claude agent
    --codex             Use Codex agent
    --opencode          Use OpenCode agent

OPTIONS:
    --file=PATH                    Path to task file (required for workflow execution)
    --fluxid-iterations=N          Max review cycles (default: 20)
    --fluxid-implement-retries=N   Max implement retries (default: 3)
    --fluxid-dry-run               Run simulation without executing agents
    --fluxid-output=FORMAT         Output format: text|yaml|json (default: text)
    --config=PATH                  Path to configuration file
    --implement-command=PATH       Override implement command file
    --review-command=PATH          Override review command file
    --commit-command=PATH          Override commit command file
    -h, --help                     Show this help message

COMMANDS:
    init                           Initialize fluxid configuration
    version                        Show version information

    report --get-file              Get absolute path to report file for current session
    report --validate              Validate existing report file against schema
    report --get-schema            Output report schema in YAML format

    history --get-file             Get absolute path to history file for current session
    history --validate             Validate existing history file against schema
    history --get-schema           Output history schema in YAML format

EXAMPLES:
    # Run workflow with Claude agent
    fluxid --claude --file=/path/to/task.txt

    # Run with custom iterations and dry-run mode
    fluxid --claude --fluxid-iterations=10 --fluxid-dry-run --file=/path/to/task.txt

    # Initialize configuration
    fluxid init

    # Show version
    fluxid version

    # Get report file path (for external agents)
    fluxid report --get-file

    # Validate report before workflow reads it
    fluxid report --validate

For more information, visit: https://github.com/fluxid/fluxid-cli
`, version.Get())
	_, _ = fmt.Fprint(os.Stdout, helpText)
}
