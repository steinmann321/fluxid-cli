package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"fluxid-loop/internal/config"

	"github.com/google/uuid"
)

const (
	defaultMaxReviewCycles     = 20
	defaultMaxImplementRetries = 3
	implementPrompt            = "Implement the required changes based on the epic requirements."
	commitPrompt               = "Create a git commit with all changes."
	reviewPrompt               = "Review the implementation and report status."
)

type Config struct {
	Agent               string
	ClaudeArgs          []string
	SessionID           string
	MaxReviewCycles     int
	MaxImplementRetries int
}

func main() {
	exitCode := run()
	os.Exit(exitCode)
}

func run() int {
	// Load home configuration
	homeConfig, err := config.LoadHomeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading home configuration: %v\n", err)
		return 1
	}

	// Parse command-line arguments manually to support arbitrary Claude args
	var claudeFlag bool
	var claudeArgs []string
	var cliIterations *int
	var cliImplementRetries *int

	// Manual argument parsing to allow passthrough of unknown flags
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if arg == "--claude" {
			claudeFlag = true
			// Continue parsing to find fluxid-specific flags after --claude
			continue
		}

		if arg == "--fluxid-iterations" {
			if i+1 >= len(os.Args) {
				fmt.Fprintf(os.Stderr, "Error: --fluxid-iterations requires a value\n")
				return 1
			}
			val, err := parsePositiveInt(os.Args[i+1], "--fluxid-iterations")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return 1
			}
			cliIterations = &val
			i++ // skip the value
			continue
		}

		if arg == "--fluxid-implement-retries" {
			if i+1 >= len(os.Args) {
				fmt.Fprintf(os.Stderr, "Error: --fluxid-implement-retries requires a value\n")
				return 1
			}
			val, err := parsePositiveInt(os.Args[i+1], "--fluxid-implement-retries")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return 1
			}
			cliImplementRetries = &val
			i++ // skip the value
			continue
		}

		// After --claude, collect remaining args for passthrough
		if claudeFlag {
			claudeArgs = append(claudeArgs, arg)
		}
	}

	if !claudeFlag {
		fmt.Fprintf(os.Stderr, "Usage: fluxid --claude [--fluxid-iterations N] [--fluxid-implement-retries R] [claude-args]\n")
		return 1
	}

	// Resolve configuration from home config, CLI args, and defaults
	resolved := config.Resolve(homeConfig, cliIterations, cliImplementRetries)

	// Validate Claude CLI is available
	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: claude command not found in PATH\n")
		fmt.Fprintf(os.Stderr, "Please install Claude CLI: https://github.com/anthropics/claude-cli\n")
		return 1
	}

	// Generate UUID v4 session ID
	sessionID := uuid.New().String()

	cfg := Config{
		Agent:               resolved.Agent,
		ClaudeArgs:          claudeArgs,
		SessionID:           sessionID,
		MaxReviewCycles:     resolved.Iterations,
		MaxImplementRetries: resolved.ImplementRetries,
	}

	// Display initialization status with source information
	fmt.Println("=== fluxid Workflow Initialization ===")
	fmt.Printf("Agent: %s (source: %s)\n", cfg.Agent, resolved.Sources["agent"])
	fmt.Printf("Session ID: %s\n", cfg.SessionID)
	fmt.Printf("Max Review Cycles: %d (source: %s)\n", cfg.MaxReviewCycles, resolved.Sources["iterations"])
	fmt.Printf("Max Implement Retries: %d (source: %s)\n", cfg.MaxImplementRetries, resolved.Sources["implement_retries"])
	fmt.Printf("Commit Enabled: %v (source: %s)\n", resolved.CommitEnabled, resolved.Sources["commit_enabled"])
	if len(cfg.ClaudeArgs) > 0 {
		fmt.Printf("Claude Args: %v\n", cfg.ClaudeArgs)
	}
	fmt.Println("======================================")
	fmt.Println()

	// Run nested loops: review cycles -> implement retries
	exitCode, err := runWorkflow(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n=== Workflow Aborted ===\n")
		fmt.Fprintf(os.Stderr, "Agent execution failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Exit code: %d\n", exitCode)
		fmt.Fprintf(os.Stderr, "\nNext steps:\n")
		fmt.Fprintf(os.Stderr, "1. Check agent output above for error details\n")
		fmt.Fprintf(os.Stderr, "2. Fix the issue and re-run fluxid\n")
		fmt.Fprintf(os.Stderr, "3. Review logs for Session ID: %s\n", cfg.SessionID)
		fmt.Fprintf(os.Stderr, "========================\n")
		return exitCode
	}

	// Display completion summary
	fmt.Println()
	fmt.Println("=== Workflow Completion Summary ===")
	fmt.Printf("Session ID: %s\n", cfg.SessionID)
	fmt.Println("Status: SUCCESS")
	fmt.Println("All workflow loops completed.")
	fmt.Println("===================================")
	return 0
}

func runWorkflow(config Config) (int, error) {
	// Outer loop: Review cycles (1-N)
	for reviewCycle := 1; reviewCycle <= config.MaxReviewCycles; reviewCycle++ {
		fmt.Printf("--- Review Cycle %d/%d ---\n", reviewCycle, config.MaxReviewCycles)

		// Inner loop: Implement retries (1-R)
		var implementSuccess bool
		for retry := 1; retry <= config.MaxImplementRetries; retry++ {
			fmt.Printf("Implement attempt %d/%d...\n", retry, config.MaxImplementRetries)

			// Phase 1: Implement
			if exitCode, err := runPhase(config, "implement", implementPrompt); err != nil {
				if exitCode != 0 {
					// Non-zero exit code from Claude: abort immediately
					return exitCode, fmt.Errorf("implement phase failed with exit code %d", exitCode)
				}
				log.Printf("Implement phase failed (retry %d/%d): %v", retry, config.MaxImplementRetries, err)
				continue
			}

			// Phase 2: Commit
			fmt.Println("Running commit phase...")
			if exitCode, err := runPhase(config, "commit", commitPrompt); err != nil {
				if exitCode != 0 {
					// Non-zero exit code from Claude: abort immediately
					return exitCode, fmt.Errorf("commit phase failed with exit code %d", exitCode)
				}
				return 1, fmt.Errorf("commit phase failed: %w", err)
			}

			implementSuccess = true
			break
		}

		if !implementSuccess {
			return 1, fmt.Errorf("implement phase failed after %d retries", config.MaxImplementRetries)
		}

		// Phase 3: Review
		fmt.Println("Running review phase...")
		if exitCode, err := runPhase(config, "review", reviewPrompt); err != nil {
			if exitCode != 0 {
				// Non-zero exit code from Claude: abort immediately
				return exitCode, fmt.Errorf("review phase failed with exit code %d", exitCode)
			}
			return 1, fmt.Errorf("review phase failed: %w", err)
		}

		// TODO: Parse review report to determine if workflow should complete early
		// For now, complete after first successful cycle
		fmt.Println("Workflow completed successfully.")
		break
	}

	return 0, nil
}

func runPhase(config Config, phase string, prompt string) (int, error) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] Starting phase: %s\n", timestamp, phase)

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
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return exitCode, fmt.Errorf("phase %s failed: %w", phase, err)
	}

	timestamp = time.Now().Format("15:04:05")
	fmt.Printf("[%s] Phase %s completed successfully\n", timestamp, phase)

	return 0, nil
}

func buildClaudeCommand(config Config, prompt string) *exec.Cmd {
	// Build args: --print flag first, then user args, then prompt as positional arg
	args := []string{
		"--print", // Non-interactive mode for automation
	}
	args = append(args, config.ClaudeArgs...)
	args = append(args, prompt)

	return exec.Command("claude", args...)
}

func getDefaultPrompt(phase string) string {
	switch phase {
	case "implement":
		return implementPrompt
	case "commit":
		return commitPrompt
	case "review":
		return reviewPrompt
	default:
		return ""
	}
}

func parsePositiveInt(value string, flagName string) (int, error) {
	var n int
	_, err := fmt.Sscanf(value, "%d", &n)
	if err != nil {
		return 0, fmt.Errorf("Error: %s requires a valid integer, got: %s", flagName, value)
	}
	if n < 1 {
		return 0, fmt.Errorf("Error: %s must be a positive integer (≥1), got: %d", flagName, n)
	}
	return n, nil
}
