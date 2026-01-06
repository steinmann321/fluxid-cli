//nolint:funlen // CLI command: comprehensive flag handling
package command

import (
	"errors"
	"fluxid-cli/internal/storage"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewReportCommand creates the report command with subcommands and flags.
//
// The report command provides file-based interface for external agents to:
// - Get the report file path (--get-file)
// - Validate report files (--validate)
// - Retrieve report schema (--get-schema)
//
// Per User Story 1 (FR-001, FR-002, FR-003): Enable agents to write reports.
func NewReportCommand() *cobra.Command {
	//nolint:exhaustruct // Cobra commands: only required fields initialized
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Manage workflow report files",
		Long: `The report command provides file-based interface for external agents to interact with workflow reports.

Reports contain the outcome of workflow phases (implement, review, commit) including:
- Status (PASS or FAIL)
- Categorized issues (blockers, defects, concerns, observations, enhancements)
- Next steps and summary

External agents write report files directly to session-specific paths obtained via --get-file.`,
		Example: `  # Get report file path for writing
  fluxid report --get-file

  # Validate report file before workflow reads it
  fluxid report --validate

  # Get report schema for programmatic parsing
  fluxid report --get-schema`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Check which flag was specified
			getFile, _ := cmd.Flags().GetBool("get-file")
			validate, _ := cmd.Flags().GetBool("validate")
			getSchema, _ := cmd.Flags().GetBool("get-schema")

			// Exactly one flag must be specified
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
				return errors.New("must specify one of: --get-file, --validate, or --get-schema") //nolint:err113 // CLI error
			}
			if flagCount > 1 {
				return errors.New("cannot specify multiple flags together") //nolint:err113 // CLI error
			}

			// Execute the appropriate handler
			if getFile {
				return handleReportGetFile()
			}
			if validate {
				return handleReportValidate()
			}
			if getSchema {
				return handleReportGetSchema()
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("get-file", false, "Get the absolute path to the report file for this session")
	cmd.Flags().Bool("validate", false, "Validate an existing report file against the schema")
	cmd.Flags().Bool("get-schema", false, "Output the report schema in YAML format")

	// Mark flags as mutually exclusive in help text
	cmd.MarkFlagsMutuallyExclusive("get-file", "validate", "get-schema")

	return cmd
}

// handleReportGetFile handles the --get-file flag.
//
// This function:
// 1. Reads FLUXID_SESSION_ID from environment (per FR-010)
// 2. Validates session ID format
// 3. Resolves session root directory
// 4. Constructs report file path: <session-root>/<session-id>/report.yaml
// 5. Creates session directory if it doesn't exist
// 6. Outputs absolute path to stdout
//
// Per FR-012, FR-013: Agent obtains file path via explicit CLI command.
//
// Returns error for missing session ID or path resolution failure.
func handleReportGetFile() error {
	// Step 1: Get session ID from environment
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		return &storage.PathValidationError{
			Field:      "FLUXID_SESSION_ID",
			Violation:  "environment variable not set",
			Constraint: "valid UUID",
			Value:      "<not set>",
		}
	}

	// Step 2: Resolve report file path
	// This validates session ID and creates directory if needed
	filePath, err := storage.ResolveSessionPath(sessionID, "report.yaml", os.Getenv("FLUXID_SESSION_ROOT"))
	if err != nil {
		return fmt.Errorf("failed to resolve report path: %w", err)
	}

	// Step 3: Output absolute path to stdout
	// Per FR-041: stdout reserved for data output only
	WriteData(filePath)

	// Silent success (no stderr output per FR-042)
	return nil
}

// handleReportValidate handles the --validate flag.
//
// This function:
// 1. Reads FLUXID_SESSION_ID from environment
// 2. Resolves report file path
// 3. Validates report file against schema
// 4. Returns validation errors if any
//
// Per FR-004, FR-005: Enable agents to validate reports before fluxid reads them.
//
// Returns error for missing session ID, path resolution, or validation failure.
func handleReportValidate() error {
	// Step 1: Get session ID from environment
	sessionID := os.Getenv("FLUXID_SESSION_ID")
	if sessionID == "" {
		return &storage.PathValidationError{
			Field:      "FLUXID_SESSION_ID",
			Violation:  "environment variable not set",
			Constraint: "valid UUID",
			Value:      "<not set>",
		}
	}

	// Step 2: Resolve report file path
	filePath, err := storage.ResolveSessionPath(sessionID, "report.yaml", os.Getenv("FLUXID_SESSION_ROOT"))
	if err != nil {
		return fmt.Errorf("failed to resolve report path: %w", err)
	}

	// Step 3: Validate report file
	if err := storage.ValidateReport(filePath); err != nil {
		return fmt.Errorf("report validation failed: %w", err)
	}

	// Silent success - no output to stderr per FR-042
	return nil
}

// handleReportGetSchema handles the --get-schema flag.
//
// This function outputs the embedded report schema to stdout in YAML format.
// Agents can parse this schema programmatically to understand report structure.
//
// Per FR-008, FR-009: Enable programmatic schema discovery.
//
// Returns error if schema cannot be loaded (should never happen).
func handleReportGetSchema() error {
	// Get embedded schema
	schema := storage.GetReportSchema()

	if schema == "" {
		// This should never happen (schema is embedded at compile time)
		// but handle it defensively
		return errors.New("failed to load embedded report schema") //nolint:err113 // Schema error
	}

	// Output schema to stdout per FR-041
	WriteData(schema)

	// Silent success (no stderr output per FR-042)
	return nil
}
