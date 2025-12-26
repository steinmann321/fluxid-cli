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
	// Check for init command first (before loading config)
	// Init creates the config, so it must run before config loading
	if len(os.Args) > 1 && os.Args[1] == "init" {
		return handleInit(os.Args[2:]), true
	}

	// Check for IPC command (before loading config)
	if len(os.Args) > 1 && os.Args[1] == "ipc" {
		return handleIPCCommand(os.Args[2:]), true
	}

	// Check for --write-history flag (standalone operation)
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--write-history" {
			return handleWriteHistory(os.Args[i+1:]), true
		}
	}

	// Check for --help flag
	for _, arg := range os.Args[1:] {
		if arg == flagHelp || arg == "-h" {
			printUsage()
			return 0, true
		}
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
