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

	wf "fluxid-cli/internal/workflow"

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
		SessionRoot:         "",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		MaxCommitRetries:    0,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}

	// Parse command-line arguments first to get the config path
	args, err := ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return emptyConfig, 1
	}

	// Load all configuration sources (pass custom config path if provided)
	homeConfig, projectConfig, configDir, exitCode := loadAllConfigsWithDir(args.CLIConfigPath)
	if exitCode != 0 {
		return emptyConfig, exitCode
	}

	// Resolve configuration with precedence: CLI > project > home > defaults
	resolved := config.Resolve(
		projectConfig, homeConfig,
		args.CLIAgent,
		args.CLIIterations, args.CLIImplementRetries, nil, // nil for cliCommitRetries - use config/defaults
	)

	// Resolve and validate command files if configured
	commandFiles, err := config.ResolveCommandFiles(projectConfig, homeConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving command files: %v\n", err)
		return emptyConfig, 1
	}
	resolved.CommandFiles = commandFiles

	// Build final configuration
	cfg, err := buildFinalConfig(resolved, args, projectConfig, homeConfig, configDir)
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

func loadAllConfigs(customConfigPath *string) (*config.HomeConfig, *config.ProjectConfig, int) {
	homeConfig, projectConfig, _, exitCode := loadAllConfigsWithDir(customConfigPath)
	return homeConfig, projectConfig, exitCode
}

func loadAllConfigsWithDir(customConfigPath *string) (*config.HomeConfig, *config.ProjectConfig, string, int) {
	// If custom config path is provided, use it as the project config (takes precedence)
	if customConfigPath != nil && *customConfigPath != "" {
		homeConfig, projectConfig, exitCode := loadConfigsWithCustom(*customConfigPath)
		if exitCode != 0 {
			return nil, nil, "", exitCode
		}
		// Config dir is the directory containing the custom config file
		configDir := filepath.Dir(*customConfigPath)
		return homeConfig, projectConfig, configDir, 0
	}

	// Otherwise, load from default locations
	homeConfig, err := config.LoadHomeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading home configuration: %v\n", err)
		return nil, nil, "", 1
	}

	projectConfig, err := config.LoadProjectConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project configuration: %v\n", err)
		return nil, nil, "", 1
	}

	// Determine config dir: project config takes precedence if it has workflow
	configDir := ""
	if projectConfig != nil && projectConfig.Workflow != nil {
		projectConfigPath, _ := config.GetProjectConfigPath()
		configDir = filepath.Dir(projectConfigPath)
	} else if homeConfig != nil && homeConfig.Workflow != nil {
		homeConfigPath, _ := config.GetHomeConfigPath()
		configDir = filepath.Dir(homeConfigPath)
	}

	return homeConfig, projectConfig, configDir, 0
}

func loadConfigsWithCustom(customConfigPath string) (*config.HomeConfig, *config.ProjectConfig, int) {
	// Load the custom config
	customConfig, _, err := config.LoadCustomConfig(customConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading custom configuration from %s: %v\n", customConfigPath, err)
		return nil, nil, 1
	}

	// Convert CustomConfig to ProjectConfig (they have identical structure)
	projectConfig := customConfigToProjectConfig(customConfig)

	// Still load home config as fallback
	homeConfig, err := config.LoadHomeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading home configuration: %v\n", err)
		return nil, nil, 1
	}

	return homeConfig, projectConfig, 0
}

func customConfigToProjectConfig(customConfig *config.CustomConfig) *config.ProjectConfig {
	if customConfig == nil {
		return nil
	}
	return &config.ProjectConfig{
		Agent:            customConfig.Agent,
		AgentArgs:        customConfig.AgentArgs,
		ImplementRetries: customConfig.ImplementRetries,
		CommitRetries:    customConfig.CommitRetries,
		Iterations:       customConfig.Iterations,
		Commands:         customConfig.Commands,
		Workflow:         customConfig.Workflow,
	}
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

func buildFinalConfig(
	resolved *config.ResolvedConfig,
	args *CLIArgs,
	projectConfig *config.ProjectConfig,
	homeConfig *config.HomeConfig,
	configDir string,
) (types.Config, error) {
	emptyConfig := getEmptyConfig()

	// Resolve basic config values
	sessionID := getSessionID()
	sessionRoot := os.Getenv("FLUXID_SESSION_ROOT")
	dryRun := args.CLIDryRun != nil && *args.CLIDryRun

	// Resolve output format
	outputFormat, err := resolveOutputFormat(args)
	if err != nil {
		return emptyConfig, err
	}

	// Resolve task file path and agent args
	taskAbs := getTaskFilePath(args)
	agentArgs := resolveAgentArgs(resolved, args)

	// Build workflow if configured
	builtWorkflow, err := buildWorkflowIfConfigured(projectConfig, homeConfig, configDir, resolved.Iterations)
	if err != nil {
		return emptyConfig, err
	}

	return types.Config{
		Agent:               resolved.Agent,
		AgentArgs:           agentArgs,
		SessionID:           sessionID,
		SessionRoot:         sessionRoot,
		MaxReviewCycles:     resolved.Iterations,
		MaxImplementRetries: resolved.ImplementRetries,
		MaxCommitRetries:    resolved.CommitRetries,
		DryRun:              dryRun,
		CommandFiles:        resolved.CommandFiles,
		OutputFormat:        outputFormat,
		TaskFilePath:        taskAbs,
		Workflow:            builtWorkflow,
	}, nil
}

// getEmptyConfig returns an empty Config for error cases.
func getEmptyConfig() types.Config {
	return types.Config{
		Agent:               "",
		AgentArgs:           nil,
		SessionID:           "",
		SessionRoot:         "",
		MaxReviewCycles:     0,
		MaxImplementRetries: 0,
		MaxCommitRetries:    0,
		DryRun:              false,
		CommandFiles:        nil,
		OutputFormat:        output.FormatText,
		TaskFilePath:        "",
		Workflow:            nil,
	}
}

// getSessionID returns the session ID from env or generates a new one.
func getSessionID() string {
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	return sessionID
}

// resolveOutputFormat resolves and validates the output format.
func resolveOutputFormat(args *CLIArgs) (output.Format, error) {
	outputFormat := output.FormatText
	if args.CLIOutputFormat != nil {
		if err := output.ValidateFormat(*args.CLIOutputFormat); err != nil {
			return outputFormat, fmt.Errorf("invalid output format: %w", err)
		}
		outputFormat = output.Format(*args.CLIOutputFormat)
	}
	return outputFormat, nil
}

// getTaskFilePath extracts the task file path from CLI args.
func getTaskFilePath(args *CLIArgs) string {
	if args.CLITaskFilePath != nil {
		return *args.CLITaskFilePath
	}
	return ""
}

// resolveAgentArgs resolves agent args with CLI precedence.
func resolveAgentArgs(resolved *config.ResolvedConfig, args *CLIArgs) []string {
	agentArgs := resolved.AgentArgs
	if len(args.AgentArgs) > 0 {
		agentArgs = args.AgentArgs
	}
	return agentArgs
}

// buildWorkflowIfConfigured builds the workflow if it's configured.
func buildWorkflowIfConfigured(
	projectConfig *config.ProjectConfig,
	homeConfig *config.HomeConfig,
	configDir string,
	iterations int,
) (*types.Workflow, error) {
	// Get workflow config with precedence: project > home
	var workflowCfg *config.WorkflowConfig
	if projectConfig != nil && projectConfig.Workflow != nil {
		workflowCfg = projectConfig.Workflow
	} else if homeConfig != nil && homeConfig.Workflow != nil {
		workflowCfg = homeConfig.Workflow
	}

	if workflowCfg == nil {
		return nil, nil //nolint:nilnil // nil workflow is valid when not configured
	}

	// Validate workflow config at startup (fail-fast)
	if err := config.ValidateWorkflowConfig(workflowCfg, configDir); err != nil {
		return nil, fmt.Errorf("workflow configuration error: %w", err)
	}

	// Build runtime workflow
	wf := wf.BuildWorkflow
	workflow, err := wf(workflowCfg, configDir, iterations)
	if err != nil {
		return nil, fmt.Errorf("failed to build workflow: %w", err)
	}

	return workflow, nil
}
