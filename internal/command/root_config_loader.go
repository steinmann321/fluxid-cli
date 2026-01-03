package command

import (
	"errors"
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
	"fluxid-cli/internal/types"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

var (
	errTaskFileRequired    = errors.New("missing required --file=PATH flag for workflow execution")
	errTaskFileNotAbsolute = errors.New("task file path must be absolute")
	errTaskFileNotFound    = errors.New("task file not found")
	errTaskFileNotRegular  = errors.New("task file is not a regular file")
)

func loadAndResolveConfig() (types.Config, int) {
	emptyConfig := types.Config{
		Agent:               "",
		AgentArgs:           nil,
		SessionID:           "",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}

	// Load all configuration sources
	homeConfig, projectConfig, exitCode := loadAllConfigs()
	if exitCode != 0 {
		return emptyConfig, exitCode
	}

	// Parse command-line arguments
	args, err := ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return emptyConfig, 1
	}

	// Resolve configuration with precedence: CLI > project > home > defaults
	resolved := config.Resolve(
		projectConfig, homeConfig,
		args.CLIAgent,
		args.CLIIterations, args.CLIImplementRetries,
	)

	// Resolve and validate command files if configured
	commandFiles, err := config.ResolveCommandFiles(projectConfig, homeConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving command files: %v\n", err)
		return emptyConfig, 1
	}
	resolved.CommandFiles = commandFiles

	// Build final configuration
	cfg, err := buildFinalConfig(resolved, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return emptyConfig, 1
	}

	// Validate agent (skip in dry-run mode since we won't actually execute it)
	if !cfg.DryRun {
		if exitCode := validateAgent(cfg.Agent); exitCode != 0 {
			return emptyConfig, exitCode
		}
	}

	// Validate task file after agent validation (so agent errors are shown first)
	if !cfg.DryRun && args.CLIAgent != nil {
		if err := validateTaskFile(cfg.TaskFilePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return emptyConfig, 1
		}
	}

	return cfg, 0
}

func loadAllConfigs() (*config.HomeConfig, *config.ProjectConfig, int) {
	homeConfig, err := config.LoadHomeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading home configuration: %v\n", err)
		return nil, nil, 1
	}

	projectConfig, err := config.LoadProjectConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project configuration: %v\n", err)
		return nil, nil, 1
	}

	return homeConfig, projectConfig, 0
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

func validateTaskFile(taskPath string) error {
	if taskPath == "" {
		return errTaskFileRequired
	}
	if !filepath.IsAbs(taskPath) {
		return fmt.Errorf("%s: %w", taskPath, errTaskFileNotAbsolute)
	}
	info, err := os.Stat(taskPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: %w", taskPath, errTaskFileNotFound)
		}
		return fmt.Errorf("cannot access task file %s: %w", taskPath, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s: %w", taskPath, errTaskFileNotRegular)
	}
	// #nosec G304 -- taskPath is validated to be absolute and readable before this point
	f, err := os.Open(taskPath)
	if err != nil {
		return fmt.Errorf("cannot read task file %s: %w", taskPath, err)
	}
	if cerr := f.Close(); cerr != nil {
		return fmt.Errorf("failed to close task file %s: %w", taskPath, cerr)
	}
	return nil
}

func buildFinalConfig(resolved *config.ResolvedConfig, args *CLIArgs) (types.Config, error) {
	emptyConfig := types.Config{
		Agent:               "",
		AgentArgs:           nil,
		SessionID:           "",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
	}
	// Generate or use provided UUID v4 session ID
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	// Determine dry-run mode
	dryRun := args.CLIDryRun != nil && *args.CLIDryRun
	// Determine output format
	outputFormat := output.FormatText
	if args.CLIOutputFormat != nil {
		if err := output.ValidateFormat(*args.CLIOutputFormat); err != nil {
			return emptyConfig, fmt.Errorf("invalid output format: %w", err)
		}
		outputFormat = output.Format(*args.CLIOutputFormat)
	}
	// Extract task file path (validation happens in loadAndResolveConfig)
	var taskAbs string
	if args.CLITaskFilePath != nil {
		taskAbs = *args.CLITaskFilePath
	}
	return types.Config{
		Agent:               resolved.Agent,
		AgentArgs:           args.AgentArgs,
		SessionID:           sessionID,
		MaxReviewCycles:     resolved.Iterations,
		MaxImplementRetries: resolved.ImplementRetries,
		DryRun:              dryRun,
		CommandFiles:        resolved.CommandFiles,
		OutputFormat:        outputFormat,
		TaskFilePath:        taskAbs,
	}, nil
}
