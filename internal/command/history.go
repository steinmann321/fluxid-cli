package command

import (
	"errors"
	"fluxid-cli/internal/storage"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewHistoryCommand creates the history subcommand with flags.
// Per User Story 4 (FR-004, FR-005, FR-006): Enable file-based history interface for agents.
func NewHistoryCommand() *cobra.Command {
	var getFile bool
	var validate bool
	var getSchema bool

	//nolint:exhaustruct // Cobra commands: only required fields initialized
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Manage history file interface for agents",
		Long: `Manage history file interface for agents.

Provides file-based API for agents to:
- Get history file path (--get-file)
- Validate history file (--validate)
- Get history schema (--get-schema)

Examples:
  fluxid history --get-file      # Get path to history file
  fluxid history --validate      # Validate history file
  fluxid history --get-schema    # Get history schema`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// Exactly one flag must be set
			flagCount := 0
			if getFile {
				flagCount++
			}
			if validate {
				flagCount++
			}
			if getSchema {
				flagCount++
			}

			if flagCount == 0 {
				return errors.New("must specify one of: --get-file, --validate, --get-schema") //nolint:err113 // CLI error
			}
			if flagCount > 1 {
				return errors.New("must specify exactly one flag") //nolint:err113 // CLI error
			}

			// Execute appropriate handler
			writer := NewErrorWriter()

			if getFile {
				return handleHistoryGetFile(writer)
			}
			if validate {
				return handleHistoryValidate(writer)
			}
			if getSchema {
				return handleHistoryGetSchema(writer)
			}

			return errors.New("no handler matched") //nolint:err113 // CLI error
		},
	}

	cmd.Flags().BoolVar(&getFile, "get-file", false, "Get history file path")
	cmd.Flags().BoolVar(&validate, "validate", false, "Validate history file")
	cmd.Flags().BoolVar(&getSchema, "get-schema", false, "Get history schema")
	cmd.MarkFlagsMutuallyExclusive("get-file", "validate", "get-schema")

	return cmd
}

// handleHistoryGetFile returns the path to the history file.
// Per FR-004: Agent gets history file path from environment.
// Returns error if session ID is missing or path resolution fails.
func handleHistoryGetFile(_ *ErrorWriter) error {
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		return errors.New("FLUXID_SESSION_ID environment variable not set") //nolint:err113 // Env error
	}

	filePath, err := storage.ResolveSessionPath(sessionID, "history.yaml", os.Getenv("FLUXID_SESSION_ROOT"))
	if err != nil {
		return fmt.Errorf("failed to resolve history path: %w", err)
	}

	// Ensure history file exists (create empty if not)
	if err := storage.EnsureFileExists(filePath); err != nil {
		return fmt.Errorf("failed to ensure history file exists: %w", err)
	}

	// Output path to stdout (data output per FR-042)
	WriteData(filePath)
	return nil
}

// handleHistoryValidate validates the history file structure.
// Per FR-005: Agent validates history against schema.
// Returns error if session ID is missing, path resolution fails, or validation fails.
func handleHistoryValidate(_ *ErrorWriter) error {
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		return errors.New("FLUXID_SESSION_ID environment variable not set") //nolint:err113 // Env error
	}

	filePath, err := storage.ResolveSessionPath(sessionID, "history.yaml", os.Getenv("FLUXID_SESSION_ROOT"))
	if err != nil {
		return fmt.Errorf("failed to resolve history path: %w", err)
	}

	// Validate history structure
	if err := storage.ValidateHistory(filePath); err != nil {
		// Include file path for FR-043: Errors must include sufficient context
		return fmt.Errorf("%s\nhistory validation failed: %w", filePath, err)
	}

	// Silent success per FR-042: no stdout, no stderr, exit 0
	return nil
}

// handleHistoryGetSchema returns the history schema.
// Per FR-006: Agent retrieves schema to understand history structure.
func handleHistoryGetSchema(_ *ErrorWriter) error {
	schema := storage.GetHistorySchema()

	// Output schema to stdout (data output per FR-042)
	WriteData(schema)
	return nil
}
