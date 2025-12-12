package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
)

const (
	maxReviewCycles     = 20
	maxImplementRetries = 3
	implementPrompt     = "Implement the required changes based on the epic requirements."
	commitPrompt        = "Create a git commit with all changes."
	reviewPrompt        = "Review the implementation and report status."
)

type Config struct {
	Agent      string
	ClaudeArgs []string
	SessionID  string
}

func main() {
	// Parse command-line arguments manually to support arbitrary Claude args
	var claudeFlag bool
	var claudeArgs []string

	// Manual argument parsing to allow passthrough of unknown flags
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--claude" {
			claudeFlag = true
			if i+1 < len(os.Args) {
				claudeArgs = os.Args[i+1:]
			}
			break
		}
	}

	if !claudeFlag {
		fmt.Fprintf(os.Stderr, "Usage: fluxid --claude [claude-args]\n")
		os.Exit(1)
	}

	// Validate Claude CLI is available
	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: claude command not found in PATH\n")
		fmt.Fprintf(os.Stderr, "Please install Claude CLI: https://github.com/anthropics/claude-cli\n")
		os.Exit(1)
	}

	// Generate UUID v4 session ID
	sessionID := uuid.New().String()

	config := Config{
		Agent:      "claude",
		ClaudeArgs: claudeArgs,
		SessionID:  sessionID,
	}

	// Display initialization status
	fmt.Println("=== fluxid Workflow Initialization ===")
	fmt.Printf("Agent: %s\n", config.Agent)
	fmt.Printf("Session ID: %s\n", config.SessionID)
	fmt.Printf("Max Review Cycles: %d\n", maxReviewCycles)
	fmt.Printf("Max Implement Retries: %d\n", maxImplementRetries)
	if len(config.ClaudeArgs) > 0 {
		fmt.Printf("Claude Args: %v\n", config.ClaudeArgs)
	}
	fmt.Println("======================================")
	fmt.Println()

	// Run nested loops: review cycles -> implement retries
	if err := runWorkflow(config); err != nil {
		fmt.Fprintf(os.Stderr, "Workflow failed: %v\n", err)
		os.Exit(1)
	}

	// Display completion summary
	fmt.Println()
	fmt.Println("=== Workflow Completion Summary ===")
	fmt.Printf("Session ID: %s\n", config.SessionID)
	fmt.Println("Status: SUCCESS")
	fmt.Println("All workflow loops completed.")
	fmt.Println("===================================")
}

func runWorkflow(config Config) error {
	// Outer loop: Review cycles (1-20)
	for reviewCycle := 1; reviewCycle <= maxReviewCycles; reviewCycle++ {
		fmt.Printf("--- Review Cycle %d/%d ---\n", reviewCycle, maxReviewCycles)

		// Inner loop: Implement retries (1-3)
		var implementSuccess bool
		for retry := 1; retry <= maxImplementRetries; retry++ {
			fmt.Printf("Implement attempt %d/%d...\n", retry, maxImplementRetries)

			// Phase 1: Implement
			if err := runPhase(config, "implement", implementPrompt); err != nil {
				log.Printf("Implement phase failed (retry %d/%d): %v", retry, maxImplementRetries, err)
				continue
			}

			// Phase 2: Commit
			fmt.Println("Running commit phase...")
			if err := runPhase(config, "commit", commitPrompt); err != nil {
				return fmt.Errorf("commit phase failed: %w", err)
			}

			implementSuccess = true
			break
		}

		if !implementSuccess {
			return fmt.Errorf("implement phase failed after %d retries", maxImplementRetries)
		}

		// Phase 3: Review
		fmt.Println("Running review phase...")
		if err := runPhase(config, "review", reviewPrompt); err != nil {
			return fmt.Errorf("review phase failed: %w", err)
		}

		// TODO: Parse review report to determine if workflow should complete early
		// For now, complete after first successful cycle
		fmt.Println("Workflow completed successfully.")
		break
	}

	return nil
}

func runPhase(config Config, phase string, prompt string) error {
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
		return fmt.Errorf("phase %s failed: %w", phase, err)
	}

	timestamp = time.Now().Format("15:04:05")
	fmt.Printf("[%s] Phase %s completed successfully\n", timestamp, phase)

	return nil
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
