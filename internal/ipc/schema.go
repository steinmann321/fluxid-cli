// Package ipc provides inter-process communication commands for fluxid.
package ipc

import (
	_ "embed"
	"fmt"
	"io"
)

//go:embed schema.yaml
var reportSchemaYAML string

// GetReportSchema returns the embedded report schema YAML content.
func GetReportSchema() string {
	return reportSchemaYAML
}

// WriteReportSchema writes the report schema to the provided writer.
func WriteReportSchema(w io.Writer) error {
	_, err := fmt.Fprint(w, GetReportSchema())
	if err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}
	return nil
}
