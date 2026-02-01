// Package workflow implements the core business logic for fluxid workflows.
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/types"
	"log"
)

const (
	statusPass    = "PASS"
	statusFail    = "FAIL"
	builtInPrompt = "built-in prompt"

	implementPrompt = "Run implement command file for task: ${FLUXID_TASK_FILE}"
	commitPrompt    = "Execute commit command file to create git commit"
	reviewPrompt    = "Run review command file for task: ${FLUXID_TASK_FILE}"

	// Error messages for path resolution failures.
	unableToResolveReportPath  = "<unable to resolve report path>"
	unableToResolveHistoryPath = "<unable to resolve history path>"

	// UI formatting.
	separatorWidth = 70
)

// AbortError represents a workflow abort with specific exit code.
type AbortError struct {
	ExitCode int
	Message  string
}

func (e *AbortError) Error() string {
	return e.Message
}

// Run executes the main workflow loop.
func Run(cfg types.Config) (int, error) {
	// Check if config-driven workflow is available
	if cfg.Workflow != nil && len(cfg.Workflow.Steps) > 0 {
		return runConfigDrivenWorkflow(cfg)
	}

	// Fallback to old hardcoded workflow for backward compatibility (if workflow not configured)
	return runLegacyWorkflow(cfg)
}

// runConfigDrivenWorkflow executes the config-driven workflow with custom steps.
func runConfigDrivenWorkflow(cfg types.Config) (int, error) {
	logger := &Logger{OutputFormat: cfg.OutputFormat}
	maxIterations := getMaxIterations(cfg)

	// Outer loop: Development iterations
	for iteration := 1; iteration <= maxIterations; iteration++ {
		cfg.Workflow.CurrentIteration = iteration
		logger.LogIterationStart(iteration, maxIterations)

		// Execute all workflow steps and check for completion
		exitCode, shouldReturn, err := executeWorkflowIteration(cfg, iteration, maxIterations, logger)
		if shouldReturn {
			return exitCode, err
		}
	}

	return 0, nil
}

// getMaxIterations returns the maximum number of iterations for the workflow.
func getMaxIterations(cfg types.Config) int {
	maxIterations := cfg.Workflow.MaxIterations
	if maxIterations == 0 {
		maxIterations = cfg.MaxReviewCycles // Fallback to old config if not set
	}
	return maxIterations
}

// executeWorkflowIteration executes all steps in a single workflow iteration.
// Returns (exitCode, shouldReturn, error).
func executeWorkflowIteration(cfg types.Config, iteration int, maxIterations int, logger *Logger) (int, bool, error) {
	var reviewStatus string

	for _, step := range cfg.Workflow.Steps {
		updatedCfg := prepareConfigForStep(cfg, step)

		// Execute step with retry logic
		if err := ExecuteStepWithRetry(updatedCfg, step, iteration, logger); err != nil {
			return 1, true, err
		}

		// Check review step status for exit gate
		if step.IsReview {
			reviewStatus = getReviewStatus(cfg)
			logger.LogIterationComplete(iteration, reviewStatus)

			if reviewStatus == statusPass {
				logger.LogWorkflowComplete(true, iteration)
				return 0, true, nil // Workflow succeeded
			}
		}
	}

	// Check iteration exhaustion
	if iteration >= maxIterations {
		logger.LogWorkflowComplete(false, iteration)
		return 0, true, nil // Iterations exhausted
	}

	return 0, false, nil // Continue to next iteration
}

// prepareConfigForStep prepares a config with the command file path for a specific step.
func prepareConfigForStep(cfg types.Config, step types.WorkflowStep) types.Config {
	updatedCfg := cfg
	if updatedCfg.CommandFiles == nil {
		updatedCfg.CommandFiles = &config.ResolvedCommandFiles{
			ImplementPath: "",
			ReviewPath:    "",
			CommitPath:    "",
		}
	}

	// Map step name to command file for runPhase compatibility
	switch step.Name {
	case "implement":
		updatedCfg.CommandFiles.ImplementPath = step.CommandFilePath
	case "commit":
		updatedCfg.CommandFiles.CommitPath = step.CommandFilePath
	case "review":
		updatedCfg.CommandFiles.ReviewPath = step.CommandFilePath
	default:
		// For custom step names, temporarily override implement path
		updatedCfg.CommandFiles.ImplementPath = step.CommandFilePath
	}

	return updatedCfg
}

// getReviewStatus reads and returns the review report status.
func getReviewStatus(cfg types.Config) string {
	report, err := storage.ReadReport(cfg.SessionID, cfg.SessionRoot)
	if err != nil {
		log.Printf("Failed to read review report: %v (treating as FAIL)", err)
		return statusFail
	}
	return report.Status
}
