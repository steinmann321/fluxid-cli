// Package command implements the CLI definition layer for fluxid.
package command

import (
	"fluxid-cli/internal/types"
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
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-h" {
			printHelp()
			return 0, true
		}
	}

	// Check for init command first (before loading config)
	// Init creates the config, so it must run before config loading
	if len(os.Args) > 1 && os.Args[1] == "init" {
		return handleInit(os.Args[2:]), true
	}

	// Check for report command (file-based interface for agents)
	// Per User Story 1 (FR-001): Enable agents to write reports via file-based interface
	if len(os.Args) > 1 && os.Args[1] == "report" {
		return handleReportCommand(os.Args[2:]), true
	}

	// Check for history command (file-based interface for agents)
	// Per User Story 4 (FR-004): Enable agents to write history via file-based interface
	if len(os.Args) > 1 && os.Args[1] == "history" {
		return handleHistoryCommand(os.Args[2:]), true
	}

	return 0, false
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
	helpText := `fluxid - AI-powered workflow automation tool

USAGE:
    fluxid [OPTIONS] --<agent> --file=<task-file>
    fluxid init [OPTIONS]
    fluxid report <write|read|validate> [OPTIONS]
    fluxid history <write|read> [OPTIONS]

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
    init                Initialize fluxid configuration
    report write        Write a report file
    report read         Read a report file
    report validate     Validate a report file
    history write       Write to history
    history read        Read history entries

EXAMPLES:
    # Run workflow with Claude agent
    fluxid --claude --file=/path/to/task.txt

    # Run with custom iterations and dry-run mode
    fluxid --claude --fluxid-iterations=10 --fluxid-dry-run --file=/path/to/task.txt

    # Initialize configuration
    fluxid init

For more information, visit: https://github.com/fluxid/fluxid-cli
`
	_, _ = fmt.Fprint(os.Stdout, helpText)
}
