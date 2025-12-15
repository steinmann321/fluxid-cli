// Package main implements the fluxid CLI workflow controller for coding agents.
package main

import (
	"context"
	"errors"
	"fluxid-loop/internal/ipc"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	statusPass = "PASS"
	statusFail = "FAIL"
)

func runWorkflow(cfg Config) (int, error) {
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
			return 130, fmt.Errorf("workflow aborted by user request")
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
			return 130, fmt.Errorf("workflow aborted by user request")
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
func runImplementPhase(cfg Config) (int, error) {
	for retry := 1; retry <= cfg.MaxImplementRetries; retry++ {
		log.Printf("Implement attempt %d/%d...", retry, cfg.MaxImplementRetries)

		// Check for abort before implement attempt
		aborted, err := ipc.CheckAbortFlag(cfg.SessionID)
		if err != nil {
			log.Printf("Warning: failed to check abort flag: %v", err)
		}
		if aborted {
			log.Println("Abort requested - exiting workflow gracefully")
			return 130, fmt.Errorf("workflow aborted by user request")
		}

		// Phase 1: Implement
		exitCode, err := runPhase(cfg, "implement", implementPrompt)
		if err != nil {
			if exitCode != 0 {
				// Non-zero exit code from Claude: abort immediately
				return exitCode, fmt.Errorf("implement phase failed with exit code %d", exitCode)
			}
			log.Printf("Implement phase failed (retry %d/%d): %v", retry, cfg.MaxImplementRetries, err)
			continue
		}

		// Phase 2: Commit (only if enabled)
		if cfg.CommitEnabled {
			exitCode, err := runCommitPhase(cfg)
			if err != nil {
				return exitCode, err
			}
		}

		// Wait for valid implement report and check status
		status, err := waitForValidReport(cfg.SessionID, "implement")
		if err != nil {
			return 1, fmt.Errorf("failed to get implement report: %w", err)
		}

		if status == statusPass {
			return 0, nil
		}

		// status == statusFail: continue to next retry
		log.Printf("Implement report status is FAIL, retrying... (%d/%d)", retry+1, cfg.MaxImplementRetries)
	}

	return 1, fmt.Errorf("implement phase failed after %d retries", cfg.MaxImplementRetries)
}

// runCommitPhase executes the commit phase.
func runCommitPhase(cfg Config) (int, error) {
	log.Println("Running commit phase...")
	exitCode, err := runPhase(cfg, "commit", commitPrompt)
	if err != nil {
		if exitCode != 0 {
			// Non-zero exit code from Claude: abort immediately
			return exitCode, fmt.Errorf("commit phase failed with exit code %d", exitCode)
		}
		return 1, fmt.Errorf("commit phase failed: %w", err)
	}
	return 0, nil
}

// runReviewPhase executes the review phase and returns the status.
func runReviewPhase(cfg Config) (string, int, error) {
	log.Println("Running review phase...")
	exitCode, err := runPhase(cfg, "review", reviewPrompt)
	if err != nil {
		if exitCode != 0 {
			// Non-zero exit code from Claude: abort immediately
			return "", exitCode, fmt.Errorf("review phase failed with exit code %d", exitCode)
		}
		return "", 1, fmt.Errorf("review phase failed: %w", err)
	}

	// Wait for valid review report and check status
	status, err := waitForValidReport(cfg.SessionID, "review")
	if err != nil {
		return "", 1, fmt.Errorf("failed to get review report: %w", err)
	}

	return status, 0, nil
}

func runPhase(config Config, phase string, prompt string) (int, error) {
	timestamp := time.Now().Format("15:04:05")
	log.Printf("[%s] Starting phase: %s", timestamp, phase)

	// Build Claude command with correct API
	cmd := buildClaudeCommand(config, prompt)

	// Set environment variable for session tracking
	cmd.Env = append(os.Environ(), fmt.Sprintf("FLUXID_SESSION_ID=%s", config.SessionID))

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

func buildClaudeCommand(config Config, prompt string) *exec.Cmd {
	// Build args: --print flag first, then user args, then prompt as positional arg
	args := []string{
		"--print", // Non-interactive mode for automation
	}
	args = append(args, config.AgentArgs...)
	args = append(args, prompt)

	// #nosec G204 - Agent name comes from validated config file, not user input
	return exec.CommandContext(context.Background(), config.Agent, args...)
}

// waitForValidReport waits indefinitely for a valid report with PASS or FAIL status.
// Returns the status ("PASS" or "FAIL") once a valid report is found.
// Retries every 2 seconds if report is missing or invalid.
func waitForValidReport(sessionID string, phase string) (string, error) {
	for {
		// Read report from IPC storage
		reportYAML, err := ipc.ReadReport(sessionID)
		if err != nil {
			return "", fmt.Errorf("failed to read report: %w", err)
		}

		// If no report exists, wait and retry
		if reportYAML == "" {
			log.Printf("Waiting for %s report... (no report found yet)", phase)
			time.Sleep(2 * time.Second)
			continue
		}

		// Validate report
		if err := ipc.ValidateReport(reportYAML); err != nil {
			log.Printf("Invalid %s report (retrying): %v", phase, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Parse report to extract status
		var report ipc.Report
		if err := yaml.Unmarshal([]byte(reportYAML), &report); err != nil {
			log.Printf("Failed to parse %s report (retrying): %v", phase, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Valid report found
		log.Printf("Valid %s report received with status: %s", phase, report.Status)
		return report.Status, nil
	}
}

// runSimulation simulates the workflow execution without spawning agent processes.
// It prints the execution plan showing all iterations, retries, and phases that would be executed,
// using synthetic PASS reports to drive loop progression.
func runSimulation(cfg Config) int {
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

	// Simulate commit phase if enabled
	if cfg.CommitEnabled {
		commandFile := getCommandFilePath(cfg, "commit")
		log.Printf("Would execute: Iteration %d, Retry %d, Phase: commit", reviewCycle, retry)
		log.Printf("  Command file: %s", commandFile)
		log.Println()
	}

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
func getCommandFilePath(cfg Config, phase string) string {
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
