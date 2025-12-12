// Package main implements the fluxid CLI workflow controller for coding agents.
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"fluxid-loop/internal/config"

	"github.com/google/uuid"
)

const (
	defaultMaxReviewCycles     = 20
	defaultMaxImplementRetries = 3
	implementPrompt            = "Implement the required changes based on the epic requirements."
	commitPrompt               = "Create a git commit with all changes."
	reviewPrompt               = "Review the implementation and report status."
	flagHelp                   = "--help"
)

type Config struct {
	Agent               string
	ClaudeArgs          []string
	SessionID           string
	MaxReviewCycles     int
	MaxImplementRetries int
	CommitEnabled       bool
}

// osEnv implements config.EnvGetter using os.Getenv.
type osEnv struct{}

func (osEnv) Getenv(key string) string {
	return os.Getenv(key)
}

func main() {
	exitCode := run()
	os.Exit(exitCode)
}

func run() int {
	// Check for IPC command first (before loading config)
	if len(os.Args) > 1 && os.Args[1] == "ipc" {
		return handleIPCCommand(os.Args[2:])
	}

	// Check for --help flag
	for _, arg := range os.Args[1:] {
		if arg == flagHelp || arg == "-h" {
			printUsage()
			return 0
		}
	}

	// Load home configuration
	homeConfig, err := config.LoadHomeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading home configuration: %v\n", err)
		return 1
	}

	// Load project configuration
	projectConfig, err := config.LoadProjectConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project configuration: %v\n", err)
		return 1
	}

	// Load environment configuration
	envConfig, err := config.LoadEnvConfig(osEnv{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading environment configuration: %v\n", err)
		return 1
	}

	// Parse command-line arguments
	args, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage()
		return 1
	}

	// Resolve configuration with precedence: CLI > env > project > home > defaults
	resolved := config.Resolve(
		projectConfig, homeConfig, envConfig,
		args.cliIterations, args.cliImplementRetries, args.cliCommitEnabled,
	)

	// Resolve and validate command files if configured
	commandFiles, err := config.ResolveCommandFiles(projectConfig, homeConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving command files: %v\n", err)
		return 1
	}
	resolved.CommandFiles = commandFiles

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
		ClaudeArgs:          args.claudeArgs,
		SessionID:           sessionID,
		MaxReviewCycles:     resolved.Iterations,
		MaxImplementRetries: resolved.ImplementRetries,
		CommitEnabled:       resolved.CommitEnabled,
	}

	// Display initialization status with source information
	log.Println("=== fluxid Workflow Initialization ===")
	log.Printf("Agent: %s (source: %s)", cfg.Agent, resolved.Sources["agent"])
	log.Printf("Session ID: %s", cfg.SessionID)
	log.Printf("Max Review Cycles: %d (source: %s)", cfg.MaxReviewCycles, resolved.Sources["iterations"])
	log.Printf("Max Implement Retries: %d (source: %s)", cfg.MaxImplementRetries, resolved.Sources["implement_retries"])
	log.Printf("Commit Enabled: %v (source: %s)", resolved.CommitEnabled, resolved.Sources["commit_enabled"])

	// Display command file paths if resolved
	if resolved.CommandFiles != nil {
		log.Println()
		log.Println("Command Files:")
		log.Printf("  Implement: %s", resolved.CommandFiles.ImplementPath)
		log.Printf("  Review: %s", resolved.CommandFiles.ReviewPath)
		log.Printf("  Commit: %s", resolved.CommandFiles.CommitPath)
	}

	if len(cfg.ClaudeArgs) > 0 {
		log.Printf("Claude Args: %v", cfg.ClaudeArgs)
	}
	log.Println("======================================")
	log.Println()

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
	log.Println()
	log.Println("=== Workflow Completion Summary ===")
	log.Printf("Session ID: %s", cfg.SessionID)
	log.Println("Status: SUCCESS")
	log.Println("All workflow loops completed.")
	log.Println("===================================")
	return 0
}
