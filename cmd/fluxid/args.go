// Package main implements the fluxid CLI workflow controller for coding agents.
package main

import (
	"fmt"
	"os"
)

type cliArgs struct {
	cliAgent            *string
	agentArgs           []string
	cliIterations       *int
	cliImplementRetries *int
	cliCommitEnabled    *bool
	cliDryRun           *bool
	cliOutputFormat     *string
}

func parseArgs() (*cliArgs, error) {
	args := &cliArgs{
		cliAgent:            nil,
		agentArgs:           nil,
		cliIterations:       nil,
		cliImplementRetries: nil,
		cliCommitEnabled:    nil,
		cliDryRun:           nil,
		cliOutputFormat:     nil,
	}

	var agentFlagCount int
	var currentAgent string

	// Manual argument parsing to allow passthrough of unknown flags
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		// Handle agent flags
		if isAgentFlag(arg) {
			agentFlagCount++
			currentAgent = arg[2:] // strip "--" prefix
			continue
		}

		// Handle fluxid-specific flags
		skipCount, handled, err := parseFluxidFlag(arg, i, args)
		if err != nil {
			return nil, err
		}
		if handled {
			i += skipCount
			continue
		}

		// Collect agent arguments
		if currentAgent != "" {
			args.agentArgs = append(args.agentArgs, arg)
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

func parseFluxidFlag(arg string, index int, args *cliArgs) (int, bool, error) {
	switch arg {
	case "--fluxid-iterations":
		skip, err := parseIntFlag(index, arg, &args.cliIterations)
		return skip, true, err
	case "--fluxid-implement-retries":
		skip, err := parseIntFlag(index, arg, &args.cliImplementRetries)
		return skip, true, err
	case "--fluxid-commit-enabled":
		trueVal := true
		args.cliCommitEnabled = &trueVal
		return 0, true, nil
	case "--fluxid-no-commit":
		falseVal := false
		args.cliCommitEnabled = &falseVal
		return 0, true, nil
	case "--fluxid-dry-run":
		trueVal := true
		args.cliDryRun = &trueVal
		return 0, true, nil
	case "--fluxid-output":
		skip, err := parseStringFlag(index, arg, &args.cliOutputFormat)
		return skip, true, err
	default:
		return 0, false, nil
	}
}

func parseIntFlag(index int, flagName string, dest **int) (int, error) {
	if index+1 >= len(os.Args) {
		return 0, fmt.Errorf("%s requires a value", flagName)
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
		return 0, fmt.Errorf("%s requires a value", flagName)
	}
	value := os.Args[index+1]
	*dest = &value
	return 1, nil
}

func validateAgentFlags(count int, currentAgent string, args *cliArgs) error {
	if count > 1 {
		return fmt.Errorf("multiple agent flags specified. Please use only one of: --claude, --codex, or --opencode")
	}
	if count == 1 {
		args.cliAgent = &currentAgent
	}
	return nil
}

func parsePositiveInt(value string, flagName string) (int, error) {
	var n int
	_, err := fmt.Sscanf(value, "%d", &n)
	if err != nil {
		return 0, fmt.Errorf("%s requires a valid integer, got: %s", flagName, value)
	}
	if n < 1 {
		return 0, fmt.Errorf("%s must be a positive integer (≥1), got: %d", flagName, n)
	}
	return n, nil
}
