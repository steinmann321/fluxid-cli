package storage

import (
	"fluxid-cli/internal/assets"
)

// GetReportSchema returns the embedded YAML schema for report files.
// The schema defines the structure, required fields, types, and constraints
// for report files written by external agents.
//
// Returns the schema as a YAML string that can be:
// - Output to stdout via `fluxid report --get-schema`
// - Used for validation via ValidateReport()
// - Parsed by external agents to understand report structure.
func GetReportSchema() string {
	return assets.ReportSchemaYAML
}

// GetHistorySchema returns the embedded YAML schema for history files.
// The schema defines the structure, required fields, types, and constraints
// for history files that track workflow events across sessions.
//
// Returns the schema as a YAML string that can be:
// - Output to stdout via `fluxid history --get-schema`
// - Used for validation via ValidateHistory()
// - Parsed by external agents to understand history structure.
func GetHistorySchema() string {
	return assets.HistorySchemaYAML
}
