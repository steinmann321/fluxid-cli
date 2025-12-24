package command

import (
	"fluxid-loop/internal/output"
	"fluxid-loop/internal/types"
	"fmt"
	"log"
	"os"
)

func buildInitializationStatus(cfg types.Config) output.InitializationStatus {
	status := output.InitializationStatus{
		SessionID:              cfg.SessionID,
		Agent:                  cfg.Agent,
		AgentSource:            cfg.Sources["agent"],
		MaxReviewCycles:        cfg.MaxReviewCycles,
		ReviewCyclesSource:     cfg.Sources["iterations"],
		MaxImplementRetries:    cfg.MaxImplementRetries,
		ImplementRetriesSource: cfg.Sources["implement_retries"],
		CommitEnabled:          cfg.CommitEnabled,
		CommitEnabledSource:    cfg.Sources["commit_enabled"],
		CommandFiles:           nil,
		AgentArgs:              nil,
	}

	if cfg.CommandFiles != nil {
		status.CommandFiles = &output.CommandFilesJSON{
			Implement: cfg.CommandFiles.ImplementPath,
			Review:    cfg.CommandFiles.ReviewPath,
			Commit:    cfg.CommandFiles.CommitPath,
		}
	}

	if len(cfg.AgentArgs) > 0 {
		status.AgentArgs = cfg.AgentArgs
	}

	return status
}

func printDryRunHeader(cfg types.Config) {
	if cfg.DryRun && cfg.OutputFormat == output.FormatText {
		log.Println("=== DRY RUN MODE - Simulation Only ===")
		log.Println()
	}
}

func printInitializationStatus(status output.InitializationStatus, format output.Format) int {
	switch format {
	case output.FormatJSON:
		if err := output.PrintJSON(status); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing initialization status: %v\n", err)
			return 1
		}
	case output.FormatYAML:
		if err := output.PrintYAML(status); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing initialization status: %v\n", err)
			return 1
		}
	case output.FormatText:
		output.PrintText(status)
	default:
		output.PrintText(status)
	}
	return 0
}

func printWorkflowError(err error, exitCode int, sessionID string) {
	fmt.Fprintf(os.Stderr, "\n=== Workflow Aborted ===\n")
	fmt.Fprintf(os.Stderr, "Agent execution failed: %v\n", err)
	fmt.Fprintf(os.Stderr, "Exit code: %d\n", exitCode)
	fmt.Fprintf(os.Stderr, "\nNext steps:\n")
	fmt.Fprintf(os.Stderr, "1. Check agent output above for error details\n")
	fmt.Fprintf(os.Stderr, "2. Fix the issue and re-run fluxid\n")
	fmt.Fprintf(os.Stderr, "3. Review logs for Session ID: %s\n", sessionID)
	fmt.Fprintf(os.Stderr, "========================\n")
}

func printWorkflowSuccess(sessionID string) {
	log.Println()
	log.Println("=== Workflow Completion Summary ===")
	log.Printf("Session ID: %s", sessionID)
	log.Println("Status: SUCCESS")
	log.Println("All workflow loops completed.")
	log.Println("===================================")
}
