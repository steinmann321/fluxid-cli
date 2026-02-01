// Package types contains shared types used across fluxid packages.
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
	Workflow            *Workflow // Config-driven workflow (replaces hardcoded 3-step loop)
}
