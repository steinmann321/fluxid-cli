// Package main implements the fluxid CLI workflow controller for coding agents.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func runWorkflow(cfg Config) (int, error) {
	// Outer loop: Review cycles (1-N)
	//nolint:staticcheck // SA4008: loop exits via break, maxReviewCycles is intentionally constant
	for reviewCycle := 1; reviewCycle <= cfg.MaxReviewCycles; reviewCycle++ {
		log.Printf("--- Review Cycle %d/%d ---", reviewCycle, cfg.MaxReviewCycles)

		// Inner loop: Implement retries (1-R)
		var implementSuccess bool
		for retry := 1; retry <= cfg.MaxImplementRetries; retry++ {
			log.Printf("Implement attempt %d/%d...", retry, cfg.MaxImplementRetries)

			// Phase 1: Implement
			if exitCode, err := runPhase(cfg, "implement", implementPrompt); err != nil {
				if exitCode != 0 {
					// Non-zero exit code from Claude: abort immediately
					return exitCode, fmt.Errorf("implement phase failed with exit code %d", exitCode)
				}
				log.Printf("Implement phase failed (retry %d/%d): %v", retry, cfg.MaxImplementRetries, err)
				continue
			}

			// Phase 2: Commit (only if enabled)
			if cfg.CommitEnabled {
				log.Println("Running commit phase...")
				if exitCode, err := runPhase(cfg, "commit", commitPrompt); err != nil {
					if exitCode != 0 {
						// Non-zero exit code from Claude: abort immediately
						return exitCode, fmt.Errorf("commit phase failed with exit code %d", exitCode)
					}
					return 1, fmt.Errorf("commit phase failed: %w", err)
				}
			}

			implementSuccess = true
			break
		}

		if !implementSuccess {
			return 1, fmt.Errorf("implement phase failed after %d retries", cfg.MaxImplementRetries)
		}

		// Phase 3: Review
		log.Println("Running review phase...")
		if exitCode, err := runPhase(cfg, "review", reviewPrompt); err != nil {
			if exitCode != 0 {
				// Non-zero exit code from Claude: abort immediately
				return exitCode, fmt.Errorf("review phase failed with exit code %d", exitCode)
			}
			return 1, fmt.Errorf("review phase failed: %w", err)
		}

		// Placeholder: review phase currently completes after first successful cycle
		log.Println("Workflow completed successfully.")
		break
	}

	return 0, nil
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
	args = append(args, config.ClaudeArgs...)
	args = append(args, prompt)

	// #nosec G204 - Agent name comes from validated config file, not user input
	return exec.CommandContext(context.Background(), config.Agent, args...)
}
