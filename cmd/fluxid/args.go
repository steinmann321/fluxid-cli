// Package main implements the fluxid CLI workflow controller for coding agents.
package main

import (
	"fmt"
	"os"
)

type cliArgs struct {
	claudeArgs          []string
	cliIterations       *int
	cliImplementRetries *int
	cliCommitEnabled    *bool
}

func parseArgs() (*cliArgs, error) {
	var claudeFlag bool
	args := &cliArgs{
		claudeArgs:          nil,
		cliIterations:       nil,
		cliImplementRetries: nil,
		cliCommitEnabled:    nil,
	}

	// Manual argument parsing to allow passthrough of unknown flags
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if arg == "--claude" {
			claudeFlag = true
			continue
		}

		if arg == "--fluxid-iterations" {
			if i+1 >= len(os.Args) {
				return nil, fmt.Errorf("--fluxid-iterations requires a value")
			}
			val, err := parsePositiveInt(os.Args[i+1], "--fluxid-iterations")
			if err != nil {
				return nil, err
			}
			args.cliIterations = &val
			i++ // skip the value
			continue
		}

		if arg == "--fluxid-implement-retries" {
			if i+1 >= len(os.Args) {
				return nil, fmt.Errorf("--fluxid-implement-retries requires a value")
			}
			val, err := parsePositiveInt(os.Args[i+1], "--fluxid-implement-retries")
			if err != nil {
				return nil, err
			}
			args.cliImplementRetries = &val
			i++ // skip the value
			continue
		}

		if arg == "--fluxid-no-commit" {
			falseVal := false
			args.cliCommitEnabled = &falseVal
			continue
		}

		if claudeFlag {
			args.claudeArgs = append(args.claudeArgs, arg)
		}
	}

	if !claudeFlag {
		return nil, fmt.Errorf("missing required --claude flag")
	}

	return args, nil
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
