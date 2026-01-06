// Package types contains shared types used across fluxid packages.
//
//nolint:revive // "types" is a standard Go package name for shared type definitions.
package types

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/output"
)

// Config represents the complete fluxid configuration.
type Config struct {
	Agent               string
	AgentArgs           []string
	SessionID           string
	SessionRoot         string // Optional session root override from FLUXID_SESSION_ROOT
	MaxReviewCycles     int
	MaxImplementRetries int
	MaxCommitRetries    int
	DryRun              bool
	CommandFiles        *config.ResolvedCommandFiles
	OutputFormat        output.Format
	TaskFilePath        string
}
