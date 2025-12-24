// Package command implements the CLI definition layer for fluxid.
package command

import (
	"errors"
	"fmt"
	"os"
)

var (
	errFlagMissingValue   = errors.New("flag requires a value")
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
	CLICommitEnabled    *bool
	CLIDryRun           *bool
	CLIOutputFormat     *string
}

// ParseArgs parses command-line arguments.
func ParseArgs() (*CLIArgs, error) {
	args := &CLIArgs{
		CLIAgent:            nil,
		AgentArgs:           nil,
		CLIIterations:       nil,
		CLIImplementRetries: nil,
		CLICommitEnabled:    nil,
		CLIDryRun:           nil,
		CLIOutputFormat:     nil,
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

func parseFluxidFlag(arg string, index int, args *CLIArgs) (int, bool, error) {
	// Handle --output=value format
	if len(arg) > len("--output=") && arg[:len("--output=")] == "--output=" {
		value := arg[len("--output="):]
		args.CLIOutputFormat = &value
		return 0, true, nil
	}

	switch arg {
	case "--fluxid-iterations":
		skip, err := parseIntFlag(index, arg, &args.CLIIterations)
		return skip, true, err
	case "--fluxid-implement-retries":
		skip, err := parseIntFlag(index, arg, &args.CLIImplementRetries)
		return skip, true, err
	case "--fluxid-commit-enabled":
		trueVal := true
		args.CLICommitEnabled = &trueVal
		return 0, true, nil
	case "--fluxid-no-commit":
		falseVal := false
		args.CLICommitEnabled = &falseVal
		return 0, true, nil
	case "--fluxid-dry-run", "--dry-run":
		trueVal := true
		args.CLIDryRun = &trueVal
		return 0, true, nil
	case "--fluxid-output", "--output":
		skip, err := parseStringFlag(index, arg, &args.CLIOutputFormat)
		return skip, true, err
	default:
		return 0, false, nil
	}
}

func parseIntFlag(index int, flagName string, dest **int) (int, error) {
	if index+1 >= len(os.Args) {
		return 0, fmt.Errorf("%s %w", flagName, errFlagMissingValue)
	}
	val, err := parsePositiveInt(os.Args[index+1], flagName)
	if err != nil {
		return 0, err
	}
	*dest = &val
	return 1, nil
}

func parseStringFlag(index int, flagName string, dest **string) (int, error) {
	if index+1 >= len(os.Args) {
		return 0, fmt.Errorf("%s %w", flagName, errFlagMissingValue)
	}
	value := os.Args[index+1]
	*dest = &value
	return 1, nil
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
