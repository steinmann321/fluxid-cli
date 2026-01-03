// Package command implements the CLI definition layer for fluxid.
package command

import (
	"errors"
	"fmt"
	"os"
)

var (
	errMultipleAgentFlags = errors.New(
		"multiple agent flags specified. Please use only one of: --claude, --codex, or --opencode",
	)
	errInvalidInteger     = errors.New("flag requires a valid integer")
	errNotPositiveInteger = errors.New("flag must be a positive integer (≥1)")
)

// CLIArgs represents parsed command-line arguments.
type CLIArgs struct {
	CLIAgent            *string
	AgentArgs           []string
	CLIIterations       *int
	CLIImplementRetries *int
	CLIDryRun           *bool
	CLIOutputFormat     *string
	CLIConfigPath       *string
	CLIImplementCommand *string
	CLIReviewCommand    *string
	CLICommitCommand    *string
	CLITaskFilePath     *string
}

// ParseArgs parses command-line arguments.
func ParseArgs() (*CLIArgs, error) {
	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           nil,
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
		CLIConfigPath:       nil,
		CLIImplementCommand: nil,
		CLIReviewCommand:    nil,
		CLICommitCommand:    nil,
		CLITaskFilePath:     nil,
	}

	var agentFlagCount int
	var currentAgent string

	// Manual argument parsing to allow passthrough of unknown flags
	for argIndex := 1; argIndex < len(os.Args); argIndex++ {
		arg := os.Args[argIndex]

		// Handle agent flags
		if isAgentFlag(arg) {
			agentFlagCount++
			currentAgent = arg[2:] // strip "--" prefix
			continue
		}

		// Handle fluxid-specific flags
		skipCount, handled, err := parseFluxidFlag(arg, argIndex, args)
		if err != nil {
			return nil, err
		}
		if handled {
			argIndex += skipCount
			continue
		}

		// Collect agent arguments
		if currentAgent != "" {
			args.AgentArgs = append(args.AgentArgs, arg)
		}
	}

	// Validate agent selection
	if err := validateAgentFlags(agentFlagCount, currentAgent, args); err != nil {
		return nil, err
	}

	return args, nil
}

func isAgentFlag(arg string) bool {
	return arg == "--claude" || arg == "--codex" || arg == "--opencode"
}

//nolint:cyclop,funlen,unparam // Flag parsing function with necessary complexity, return value for API consistency
func parseFluxidFlag(arg string, _ int, args *CLIArgs) (int, bool, error) {
	// Handle --output=value format
	if len(arg) >= len("--output=") && arg[:len("--output=")] == "--output=" {
		value := arg[len("--output="):]
		args.CLIOutputFormat = &value
		return 0, true, nil
	}

	// Handle --fluxid-output=value format
	if len(arg) >= len("--fluxid-output=") && arg[:len("--fluxid-output=")] == "--fluxid-output=" {
		value := arg[len("--fluxid-output="):]
		args.CLIOutputFormat = &value
		return 0, true, nil
	}

	// Handle --config=value format
	if len(arg) >= len("--config=") && arg[:len("--config=")] == "--config=" {
		// Check for multiple --config flags
		if args.CLIConfigPath != nil {
			//nolint:err113 // Simple validation error, sentinel error would be overkill
			return 0, true, errors.New("multiple --config flags not allowed")
		}
		value := arg[len("--config="):]
		args.CLIConfigPath = &value
		return 0, true, nil
	}

	// Handle --implement-command=value format
	if len(arg) >= len("--implement-command=") && arg[:len("--implement-command=")] == "--implement-command=" {
		value := arg[len("--implement-command="):]
		args.CLIImplementCommand = &value
		return 0, true, nil
	}

	// Handle --review-command=value format
	if len(arg) >= len("--review-command=") && arg[:len("--review-command=")] == "--review-command=" {
		value := arg[len("--review-command="):]
		args.CLIReviewCommand = &value
		return 0, true, nil
	}

	// Handle --commit-command=value format
	if len(arg) >= len("--commit-command=") && arg[:len("--commit-command=")] == "--commit-command=" {
		value := arg[len("--commit-command="):]
		args.CLICommitCommand = &value
		return 0, true, nil
	}

	// Handle --file=value format (task file)
	if len(arg) >= len("--file=") && arg[:len("--file=")] == "--file=" {
		value := arg[len("--file="):]
		args.CLITaskFilePath = &value
		return 0, true, nil
	}

	// Handle --fluxid-iterations=value format
	if len(arg) >= len("--fluxid-iterations=") && arg[:len("--fluxid-iterations=")] == "--fluxid-iterations=" {
		valueStr := arg[len("--fluxid-iterations="):]
		val, err := parsePositiveInt(valueStr, "--fluxid-iterations")
		if err != nil {
			return 0, true, err
		}
		args.CLIIterations = &val
		return 0, true, nil
	}

	// Handle --fluxid-implement-retries=value format
	if len(arg) >= len("--fluxid-implement-retries=") &&
		arg[:len("--fluxid-implement-retries=")] == "--fluxid-implement-retries=" {
		valueStr := arg[len("--fluxid-implement-retries="):]
		val, err := parsePositiveInt(valueStr, "--fluxid-implement-retries")
		if err != nil {
			return 0, true, err
		}
		args.CLIImplementRetries = &val
		return 0, true, nil
	}

	switch arg {
	case "--fluxid-iterations", "--fluxid-implement-retries", "--fluxid-output", "--output":
		// Reject space syntax for value flags
		//nolint:err113 // Dynamic error message includes arg name for better UX
		return 0, true, fmt.Errorf("%s requires equals syntax: use %s=<value> instead of %s <value>", arg, arg, arg)
	case "--fluxid-dry-run", "--dry-run":
		trueVal := true
		args.CLIDryRun = &trueVal
		return 0, true, nil
	default:
		return 0, false, nil
	}
}

func validateAgentFlags(count int, currentAgent string, args *CLIArgs) error {
	if count > 1 {
		return errMultipleAgentFlags
	}
	if count == 1 {
		args.CLIAgent = &currentAgent
	}
	return nil
}

func parsePositiveInt(value string, flagName string) (int, error) {
	var number int
	_, err := fmt.Sscanf(value, "%d", &number)
	if err != nil {
		return 0, fmt.Errorf("%s %w, got: %s", flagName, errInvalidInteger, value)
	}
	if number < 1 {
		return 0, fmt.Errorf("%s %w, got: %d", flagName, errNotPositiveInteger, number)
	}
	return number, nil
}
