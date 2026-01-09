// Package workflow implements the core business logic for fluxid workflows.
package workflow

import (
	"context"
	"errors"
	"fluxid-cli/internal/storage"
	"fluxid-cli/internal/stream"
	"fluxid-cli/internal/types"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

func runPhase(config types.Config, phase string, prompt string) (int, error) {
	timestamp := time.Now().Format("15:04:05")
	log.Printf("[%s] Starting phase: %s", timestamp, phase)

	// Compose prompt with command and task file paths
	finalPrompt := composePrompt(config, phase, prompt)

	// Build agent command with correct API for each agent
	cmd := buildAgentCommand(config, finalPrompt)

	// Set environment variables for session tracking and task file
	cmd.Env = append(os.Environ(), "FLUXID_SESSION_ID="+config.SessionID)
	cmd.Env = append(cmd.Env, "FLUXID_TASK_FILE="+config.TaskFilePath)
	// Pass FLUXID_SESSION_ROOT to agent if set in config
	if config.SessionRoot != "" {
		cmd.Env = append(cmd.Env, "FLUXID_SESSION_ROOT="+config.SessionRoot)
	}

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return 1, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return 1, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Set stdin to os.Stdin for interactive input if needed
	cmd.Stdin = os.Stdin

	// Start the command
	if err := cmd.Start(); err != nil {
		return 1, fmt.Errorf("failed to start command: %w", err)
	}

	// Start copying stderr in the background
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(os.Stderr, stderrPipe)
		done <- err
	}()

	// Parse and format the JSON stream from stdout (blocking)
	parser := stream.NewParser(stdoutPipe, os.Stdout, config.Agent)
	parseErr := parser.Parse()
	if parseErr != nil {
		log.Printf("Warning: stream parsing error: %v", parseErr)
	}

	// Wait for command to complete
	waitErr := cmd.Wait()

	// Wait for stderr copying to complete
	<-done

	// Handle command exit error
	if waitErr != nil {
		// Extract exit code from error
		exitCode := 1 // Default to 1 if we can't determine the actual code
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return exitCode, fmt.Errorf("phase %s failed: %w", phase, waitErr)
	}

	// Phase completed - status will be shown in unified transition box
	return 0, nil
}

// buildAgentCommand builds the appropriate command for each agent type.
// Each agent has a different CLI interface that must be respected.
func buildAgentCommand(config types.Config, prompt string) *exec.Cmd {
	var args []string

	switch config.Agent {
	case "claude":
		// Claude: claude --dangerously-skip-permissions --output-format stream-json --verbose -p "prompt"
		args = []string{
			"--dangerously-skip-permissions",
			"--output-format",
			"stream-json",
			"--verbose",
			"-p",
			prompt,
		}

	case "codex":
		// Codex: codex exec --json "prompt"
		args = []string{
			"exec",
			"--json",
			prompt,
		}

	case "opencode":
		// Opencode: opencode run --format json "prompt"
		args = []string{
			"run",
			"--format",
			"json",
			prompt,
		}

	case "gemini":
		// Gemini: gemini --output-format stream-json "prompt"
		args = []string{
			"--output-format",
			"stream-json",
			prompt,
		}

	default:
		// Fallback to Claude-style for unknown agents
		args = []string{
			"--dangerously-skip-permissions",
			"--output-format",
			"stream-json",
			"--verbose",
			"-p",
			prompt,
		}
	}

	// Append additional agent-specific arguments
	args = append(args, config.AgentArgs...)

	// #nosec G204 - Agent name comes from validated config file, not user input
	return exec.CommandContext(context.Background(), config.Agent, args...)
}

// composePrompt builds the phase prompt including command and task file context.
func composePrompt(cfg types.Config, phase string, basePrompt string) string {
	cmdFile := getCommandFilePath(cfg, phase)

	// Read command file content if it's a real file (not builtInPrompt)
	var commandContent string
	if cmdFile != builtInPrompt {
		// #nosec G304 -- cmdFile comes from validated config paths
		content, err := os.ReadFile(cmdFile)
		if err != nil {
			log.Printf("Warning: failed to read command file %s: %v", cmdFile, err)
			return fmt.Sprintf("%s\nCommand file: %s\nTask file: %s", basePrompt, cmdFile, cfg.TaskFilePath)
		}
		commandContent = string(content)
	}

	// If we have command content, include it in the prompt
	if commandContent != "" {
		return fmt.Sprintf("%s\n\nCommand File Content:\n%s\n\nTask File: %s",
			basePrompt, commandContent, cfg.TaskFilePath)
	}

	// Fallback to path reference
	displayPath := cmdFile
	if cmdFile == builtInPrompt {
		displayPath = "<built-in>"
	}
	return fmt.Sprintf("%s\nCommand file: %s\nTask file: %s", basePrompt, displayPath, cfg.TaskFilePath)
}

// checkReportStatus checks for a valid report immediately after agent exits.
// Since the agent writes reports directly to files via 'fluxid report --get-file',
// the report either exists or it doesn't when the agent process completes.
// Returns the status ("PASS" or "FAIL") or treats missing/invalid reports as FAIL.
//
//nolint:unparam // Error return maintained for interface consistency
func waitForValidReport(sessionID string, sessionRoot string, phase string) (string, error) {
	// Abort flag functionality removed per 001-report-history-refactor
	// Abort mechanism is out of scope and requires separate evaluation

	// Read report from file-based storage
	report, err := storage.ReadReport(sessionID, sessionRoot)
	if err != nil {
		// If report doesn't exist or is invalid, treat as FAIL
		// Agent either didn't write one, or the file is malformed
		log.Printf("Failed to read %s report (treating as FAIL): %v", phase, err)
		return statusFail, nil
	}

	// Valid report found - the storage.ReadReport() function already validates
	// the report structure, so if we reach here, the report is valid.
	// Status will be shown in unified transition box.
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
	log.Printf("  Task file: %s", cfg.TaskFilePath)
	log.Printf("  Command file: %s", commandFile)
	log.Println()

	// Simulate commit phase
	commandFile = getCommandFilePath(cfg, "commit")
	log.Printf("Would execute: Iteration %d, Retry %d, Phase: commit", reviewCycle, retry)
	log.Printf("  Task file: %s", cfg.TaskFilePath)
	log.Printf("  Command file: %s", commandFile)
	log.Println()

	// Simulate synthetic PASS report for implement
	log.Printf("Synthetic implement report: PASS")
	log.Println()

	// Simulate review phase
	commandFile = getCommandFilePath(cfg, "review")
	log.Printf("Would execute: Iteration %d, Retry 1, Phase: review", reviewCycle)
	log.Printf("  Task file: %s", cfg.TaskFilePath)
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

// getCommandFilePath returns the command file path for a phase, or builtInPrompt if not configured.
func getCommandFilePath(cfg types.Config, phase string) string {
	if cfg.CommandFiles == nil {
		return builtInPrompt
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
	return builtInPrompt
}
