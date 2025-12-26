// Package workflow implements the core business logic for fluxid workflows.
package workflow

import (
	"context"
	"errors"
	"fluxid-loop/internal/ipc"
	"fluxid-loop/internal/types"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	errWorkflowAborted      = errors.New("workflow aborted by user request")
	errImplementPhaseFailed = errors.New("implement phase failed")
	errCommitPhaseFailed    = errors.New("commit phase failed")
	errReviewPhaseFailed    = errors.New("review phase failed")
)

const (
	statusPass = "PASS"
	statusFail = "FAIL"

	implementPrompt = "Implement the required changes based on the epic requirements."
	commitPrompt    = "Create a git commit with all changes."
	reviewPrompt    = "Review the implementation and report status."

	exitCodeInterrupted = 130 // Exit code for SIGINT/SIGTERM user interrupt
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
	// Outer loop: Review cycles (1-N)
	for reviewCycle := 1; reviewCycle <= cfg.MaxReviewCycles; reviewCycle++ {
		log.Printf("--- Review Cycle %d/%d ---", reviewCycle, cfg.MaxReviewCycles)

		// Check for abort before starting implement phase
		aborted, err := ipc.CheckAbortFlag(cfg.SessionID)
		if err != nil {
			log.Printf("Warning: failed to check abort flag: %v", err)
		}
		if aborted {
			log.Println("Abort requested - exiting workflow gracefully")
			return exitCodeInterrupted, errWorkflowAborted
		}

		// Run implement phase with retries
		if exitCode, err := runImplementPhase(cfg); err != nil {
			return exitCode, err
		}

		// Check for abort before review phase
		aborted, err = ipc.CheckAbortFlag(cfg.SessionID)
		if err != nil {
			log.Printf("Warning: failed to check abort flag: %v", err)
		}
		if aborted {
			log.Println("Abort requested - exiting workflow gracefully")
			return exitCodeInterrupted, errWorkflowAborted
		}

		// Run review phase
		status, exitCode, err := runReviewPhase(cfg)
		if err != nil {
			return exitCode, err
		}

		if status == statusPass {
			log.Println("Workflow completed successfully.")
			break
		}

		// status == "FAIL": continue to next review cycle
		log.Printf("Review report status is FAIL, continuing to next cycle... (%d/%d)", reviewCycle+1, cfg.MaxReviewCycles)
	}

	return 0, nil
}

// runImplementPhase executes the implement phase with retries until PASS or max retries reached.
func runImplementPhase(cfg types.Config) (int, error) {
	for retry := 1; retry <= cfg.MaxImplementRetries; retry++ {
		log.Printf("Implement attempt %d/%d...", retry, cfg.MaxImplementRetries)

		// Check for abort before implement attempt
		if exitCode, err := checkAbortBeforeImplement(cfg.SessionID); err != nil {
			return exitCode, err
		}

		// Run implement phase
		if exitCode, err := executeImplementPhase(cfg, retry); err != nil {
			if exitCode != 0 {
				return exitCode, err
			}
			continue
		}

		// Check implement report status IMMEDIATELY after implement phase
		// CRITICAL: This must happen BEFORE executeCommit(), otherwise the commit phase
		// will overwrite the implement report with a commit report
		if exitCode, err := checkImplementReportStatus(cfg.SessionID, retry); err != nil {
			return exitCode, err
		} else if exitCode == 0 {
			// Implement phase succeeded, run commit and return success
			if exitCode, err := executeCommit(cfg); err != nil {
				return exitCode, err
			}
			return 0, nil
		}

		// status == statusFail: continue to next retry
		log.Printf("Implement report status is FAIL, retrying... (%d/%d)", retry+1, cfg.MaxImplementRetries)
	}

	// All retries exhausted, but continue to commit/review phases
	retries := cfg.MaxImplementRetries
	log.Printf("Implement phase retries exhausted (%d/%d), continuing to next phase", retries, retries)

	// Run commit phase even when all implement retries failed
	// This ensures the commit phase executes regardless of implement status
	if exitCode, err := executeCommit(cfg); err != nil {
		return exitCode, err
	}

	return 0, nil
}

func checkAbortBeforeImplement(sessionID string) (int, error) {
	aborted, err := ipc.CheckAbortFlag(sessionID)
	if err != nil {
		log.Printf("Warning: failed to check abort flag: %v", err)
	}
	if aborted {
		log.Println("Abort requested - exiting workflow gracefully")
		return exitCodeInterrupted, errWorkflowAborted
	}
	return 0, nil
}

func executeImplementPhase(cfg types.Config, retry int) (int, error) {
	exitCode, err := runPhase(cfg, "implement", implementPrompt)
	if err != nil {
		// runPhase always returns non-zero exit code on error
		log.Printf(
			"Implement phase failed (retry %d/%d) with exit code %d: %v",
			retry, cfg.MaxImplementRetries, exitCode, err,
		)
		return exitCode, fmt.Errorf("implement phase failed with exit code %d: %w", exitCode, errImplementPhaseFailed)
	}
	return 0, nil
}

func executeCommit(cfg types.Config) (int, error) {
	return runCommitPhase(cfg)
}

func checkImplementReportStatus(sessionID string, _ int) (int, error) {
	status, err := waitForValidReport(sessionID, "implement")
	if err != nil {
		var abortErr *AbortError
		if errors.As(err, &abortErr) {
			return abortErr.ExitCode, err
		}
		return 1, fmt.Errorf("failed to get implement report: %w", err)
	}

	if status == statusPass {
		return 0, nil
	}

	return -1, nil // Signal to continue retry
}

// runCommitPhase executes the commit phase.
func runCommitPhase(cfg types.Config) (int, error) {
	log.Println("Running commit phase...")
	exitCode, err := runPhase(cfg, "commit", commitPrompt)
	if err != nil {
		// runPhase always returns non-zero exit code on error
		return exitCode, fmt.Errorf("commit phase failed with exit code %d: %w", exitCode, errCommitPhaseFailed)
	}
	return 0, nil
}

// runReviewPhase executes the review phase and returns the status.
func runReviewPhase(cfg types.Config) (string, int, error) {
	log.Println("Running review phase...")
	exitCode, err := runPhase(cfg, "review", reviewPrompt)
	if err != nil {
		// runPhase always returns non-zero exit code on error
		return "", exitCode, fmt.Errorf("review phase failed with exit code %d: %w", exitCode, errReviewPhaseFailed)
	}

	// Wait for valid review report and check status
	status, err := waitForValidReport(cfg.SessionID, "review")
	if err != nil {
		var abortErr *AbortError
		if errors.As(err, &abortErr) {
			return "", abortErr.ExitCode, err
		}
		return "", 1, fmt.Errorf("failed to get review report: %w", err)
	}

	return status, 0, nil
}

func runPhase(config types.Config, phase string, prompt string) (int, error) {
	timestamp := time.Now().Format("15:04:05")
	log.Printf("[%s] Starting phase: %s", timestamp, phase)

	// Build Claude command with correct API
	cmd := buildClaudeCommand(config, prompt)

	// Set environment variable for session tracking
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+config.SessionID)

	// Pipe stdout/stderr/stdin for real-time streaming
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Execute Claude CLI
	if err := cmd.Run(); err != nil {
		// Extract exit code from error
		exitCode := 1 // Default to 1 if we can't determine the actual code
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return exitCode, fmt.Errorf("phase %s failed: %w", phase, err)
	}

	timestamp = time.Now().Format("15:04:05")
	log.Printf("[%s] Phase %s completed successfully", timestamp, phase)

	return 0, nil
}

func buildClaudeCommand(config types.Config, prompt string) *exec.Cmd {
	// Build args: --print flag first, then user args, then prompt as positional arg
	args := []string{
		"--print", // Non-interactive mode for automation
	}
	args = append(args, config.AgentArgs...)
	args = append(args, prompt)

	// #nosec G204 - Agent name comes from validated config file, not user input
	return exec.CommandContext(context.Background(), config.Agent, args...)
}

// checkReportStatus checks for a valid report immediately after agent exits.
// Since the agent writes reports synchronously via 'fluxid ipc write-report',
// the report either exists or it doesn't when the agent process completes.
// Returns the status ("PASS" or "FAIL") or treats missing/invalid reports as FAIL.
func waitForValidReport(sessionID string, phase string) (string, error) {
	// Check for abort flag
	aborted, err := ipc.CheckAbortFlag(sessionID)
	if err != nil {
		log.Printf("Warning: failed to check abort flag while checking %s report: %v", phase, err)
	}
	if aborted {
		return "", &AbortError{
			ExitCode: exitCodeInterrupted,
			Message:  fmt.Sprintf("workflow aborted while checking %s report", phase),
		}
	}

	// Read report from IPC storage
	reportYAML, err := ipc.ReadReport(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to read report: %w", err)
	}

	// If no report exists, treat as FAIL
	// Agent either didn't write one, or the IPC write command failed
	if reportYAML == "" {
		log.Printf("No %s report found - agent did not write report", phase)
		return statusFail, nil
	}

	// Validate report
	if err := ipc.ValidateReport(reportYAML); err != nil {
		log.Printf("Invalid %s report (treating as FAIL): %v", phase, err)
		return statusFail, nil
	}

	// Parse report to extract status
	var report ipc.Report
	if err := yaml.Unmarshal([]byte(reportYAML), &report); err != nil {
		log.Printf("Failed to parse %s report (treating as FAIL): %v", phase, err)
		return statusFail, nil
	}

	// Valid report found
	log.Printf("Valid %s report received with status: %s", phase, report.Status)
	return report.Status, nil
}

// RunSimulation simulates the workflow execution without spawning agent processes.
// It prints the execution plan showing all iterations, retries, and phases that would be executed,
// using synthetic PASS reports to drive loop progression.
func RunSimulation(cfg types.Config) int {
	log.Println("=== Simulation Plan ===")
	log.Println()

	// Simulate a single successful review cycle (happy path)
	reviewCycle := 1
	retry := 1

	log.Printf("--- Review Cycle %d/%d ---", reviewCycle, cfg.MaxReviewCycles)
	log.Println()

	log.Printf("Implement attempt %d/%d...", retry, cfg.MaxImplementRetries)

	// Simulate implement phase
	commandFile := getCommandFilePath(cfg, "implement")
	log.Printf("Would execute: Iteration %d, Retry %d, Phase: implement", reviewCycle, retry)
	log.Printf("  Command file: %s", commandFile)
	log.Println()

	// Simulate commit phase
	commandFile = getCommandFilePath(cfg, "commit")
	log.Printf("Would execute: Iteration %d, Retry %d, Phase: commit", reviewCycle, retry)
	log.Printf("  Command file: %s", commandFile)
	log.Println()

	// Simulate synthetic PASS report for implement
	log.Printf("Synthetic implement report: PASS")
	log.Println()

	// Simulate review phase
	commandFile = getCommandFilePath(cfg, "review")
	log.Printf("Would execute: Iteration %d, Retry 1, Phase: review", reviewCycle)
	log.Printf("  Command file: %s", commandFile)
	log.Println()

	// Simulate synthetic PASS report for review
	log.Printf("Synthetic review report: PASS")
	log.Println()

	// Review PASS: workflow completes successfully
	log.Printf("Simulated workflow completed successfully after %d review cycle(s)", reviewCycle)

	log.Println("=== End Simulation ===")
	return 0
}

// getCommandFilePath returns the command file path for a phase, or "built-in prompt" if not configured.
func getCommandFilePath(cfg types.Config, phase string) string {
	if cfg.CommandFiles == nil {
		return "built-in prompt"
	}

	switch phase {
	case "implement":
		if cfg.CommandFiles.ImplementPath != "" {
			return cfg.CommandFiles.ImplementPath
		}
	case "review":
		if cfg.CommandFiles.ReviewPath != "" {
			return cfg.CommandFiles.ReviewPath
		}
	case "commit":
		if cfg.CommandFiles.CommitPath != "" {
			return cfg.CommandFiles.CommitPath
		}
	}
	return "built-in prompt"
}
