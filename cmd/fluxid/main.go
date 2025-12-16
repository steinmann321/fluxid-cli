// Package main implements the fluxid CLI workflow controller for coding agents.
package main

import (
	"fluxid-loop/internal/config"
	"fluxid-loop/internal/ipc"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

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

//nolint:gochecknoglobals // Global state needed for signal handler cleanup in tests
var (
	signalCleanups []func()
	cleanupMutex   sync.Mutex
)

type Config struct {
	Agent               string
	AgentArgs           []string
	SessionID           string
	MaxReviewCycles     int
	MaxImplementRetries int
	CommitEnabled       bool
	DryRun              bool
	CommandFiles        *config.ResolvedCommandFiles
	OutputFormat        OutputFormat
	Sources             map[string]string
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
	// Check for IPC command first (before loading config)
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

func loadAndResolveConfig() (Config, int) {
	emptyConfig := Config{
		Agent:               "",
		AgentArgs:           nil,
		SessionID:           "",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		CommitEnabled:       false,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        OutputFormatText,
		Sources:             nil,
	}

	// Load all configuration sources
	homeConfig, projectConfig, envConfig, exitCode := loadAllConfigs()
	if exitCode != 0 {
		return emptyConfig, exitCode
	}

	// Parse command-line arguments
	args, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return emptyConfig, 1
	}

	// Resolve configuration with precedence: CLI > env > project > home > defaults
	resolved := config.Resolve(
		projectConfig, homeConfig, envConfig,
		args.cliAgent,
		args.cliIterations, args.cliImplementRetries, args.cliCommitEnabled,
	)

	// Resolve and validate command files if configured
	commandFiles, err := config.ResolveCommandFiles(projectConfig, homeConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving command files: %v\n", err)
		return emptyConfig, 1
	}
	resolved.CommandFiles = commandFiles

	// Validate agent
	if exitCode := validateAgent(resolved.Agent); exitCode != 0 {
		return emptyConfig, exitCode
	}

	// Build final configuration
	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return emptyConfig, 1
	}

	return cfg, 0
}

func loadAllConfigs() (*config.HomeConfig, *config.ProjectConfig, *config.EnvConfig, int) {
	homeConfig, err := config.LoadHomeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading home configuration: %v\n", err)
		return nil, nil, nil, 1
	}

	projectConfig, err := config.LoadProjectConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project configuration: %v\n", err)
		return nil, nil, nil, 1
	}

	envConfig, err := config.LoadEnvConfig(osEnv{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading environment configuration: %v\n", err)
		return nil, nil, nil, 1
	}

	return homeConfig, projectConfig, envConfig, 0
}

func validateAgent(agent string) int {
	// Validate that the resolved agent is supported
	if err := config.ValidateAgent(agent); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Validate agent binary is available in PATH
	agentPath, err := exec.LookPath(agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: agent '%s' not found in PATH\n", agent)
		fmt.Fprintf(os.Stderr, "\nPlease ensure the agent is installed and available in your PATH.\n")
		fmt.Fprintf(os.Stderr, "You can verify with: which %s\n", agent)
		return 1
	}

	// Verify the binary is executable
	if info, err := os.Stat(agentPath); err != nil || info.Mode()&0o111 == 0 {
		fmt.Fprintf(os.Stderr, "Error: agent binary at '%s' is not executable\n", agentPath)
		fmt.Fprintf(os.Stderr, "\nYou may need to run: chmod +x %s\n", agentPath)
		return 1
	}

	return 0
}

func buildFinalConfig(resolved *config.ResolvedConfig, args *cliArgs) (Config, error) {
	// Generate or use provided UUID v4 session ID
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Determine dry-run mode
	dryRun := args.cliDryRun != nil && *args.cliDryRun

	// Determine output format
	outputFormat := OutputFormatText
	if args.cliOutputFormat != nil {
		if err := ValidateOutputFormat(*args.cliOutputFormat); err != nil {
			return Config{}, err
		}
		outputFormat = OutputFormat(*args.cliOutputFormat)
	}

	return Config{
		Agent:               resolved.Agent,
		AgentArgs:           args.agentArgs,
		SessionID:           sessionID,
		MaxReviewCycles:     resolved.Iterations,
		MaxImplementRetries: resolved.ImplementRetries,
		CommitEnabled:       resolved.CommitEnabled,
		DryRun:              dryRun,
		CommandFiles:        resolved.CommandFiles,
		OutputFormat:        outputFormat,
		Sources:             resolved.Sources,
	}, nil
}

func executeWorkflow(cfg Config) int {
	// Set up signal handler for graceful abort (skip in dry-run mode)
	if !cfg.DryRun {
		_ = setupSignalHandler(cfg.SessionID) // Cleanup not needed in production since program exits
	}

	// Print dry-run header if in simulation mode and text format
	if cfg.DryRun && cfg.OutputFormat == OutputFormatText {
		log.Println("=== DRY RUN MODE - Simulation Only ===")
		log.Println()
	}

	// Display initialization status
	switch cfg.OutputFormat {
	case OutputFormatJSON:
		if err := PrintInitializationStatusJSON(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing initialization status: %v\n", err)
			return 1
		}
	case OutputFormatYAML:
		if err := PrintInitializationStatusYAML(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing initialization status: %v\n", err)
			return 1
		}
	case OutputFormatText:
		PrintInitializationStatusText(cfg)
	default:
		PrintInitializationStatusText(cfg)
	}

	// Run simulation or actual workflow based on dry-run flag
	var workflowExitCode int
	var workflowErr error
	if cfg.DryRun {
		workflowExitCode = runSimulation(cfg)
		workflowErr = nil
	} else {
		workflowExitCode, workflowErr = runWorkflow(cfg)
	}

	if workflowErr != nil {
		printWorkflowError(workflowErr, workflowExitCode, cfg.SessionID)
		return workflowExitCode
	}

	printWorkflowSuccess(cfg.SessionID)
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

//nolint:gochecknoglobals // Signal handling requires global state for goroutine coordination
var signalCount atomic.Int32

// setupSignalHandler installs a signal handler that sets the abort flag on SIGINT/SIGTERM.
// On the first signal, it sets the abort flag for graceful shutdown.
// On the second signal, it forces immediate exit.
func setupSignalHandler(sessionID string) func() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range sigChan {
			count := signalCount.Add(1)

			if count == 1 {
				// First signal: set abort flag for graceful shutdown
				log.Printf("\nReceived signal %v - requesting graceful abort...", sig)
				log.Println("Workflow will exit after current phase completes.")
				log.Println("Press Ctrl+C again to force immediate exit.")

				if err := ipc.SetAbortFlag(sessionID); err != nil {
					log.Printf("Warning: failed to set abort flag: %v", err)
				}
			} else {
				// Second signal: force immediate exit
				log.Printf("\nReceived signal %v again - forcing immediate exit", sig)
				os.Exit(130)
			}
		}
	}()

	// Return cleanup function to stop signal handling and close channel
	// Use sync.Once to ensure cleanup only runs once even if called multiple times
	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			signal.Stop(sigChan)
			close(sigChan)
		})
	}

	// Track cleanup for tests (store the idempotent cleanup function)
	cleanupMutex.Lock()
	signalCleanups = append(signalCleanups, cleanup)
	cleanupMutex.Unlock()

	return cleanup
}

// cleanupAllSignalHandlers cleans up all signal handlers. Used in tests to prevent goroutine leaks.
// Each cleanup function uses sync.Once internally, so it's safe to call multiple times.
func cleanupAllSignalHandlers() {
	cleanupMutex.Lock()
	defer cleanupMutex.Unlock()

	// Call all cleanup functions (they're idempotent thanks to sync.Once)
	for _, cleanup := range signalCleanups {
		cleanup()
	}
	signalCleanups = nil
}
