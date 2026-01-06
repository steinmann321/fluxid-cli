package assets

import (
	_ "embed"
)

// Embed schema files into the binary using Go embed directive.
// These schemas are loaded from internal/assets/templates/ at compile time.

// ReportSchemaYAML contains the embedded report schema YAML.
//
//go:embed templates/report-schema.yaml
var ReportSchemaYAML string

// HistorySchemaYAML contains the embedded history schema YAML.
//
//go:embed templates/history-schema.yaml
var HistorySchemaYAML string
