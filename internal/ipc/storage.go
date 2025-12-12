// Package ipc provides inter-process communication commands for fluxid.
package ipc

import (
	"fmt"
	"os"
	"path/filepath"
)

// getReportPath returns the file path for storing a session's report.
// Reports are stored in /tmp/fluxid-reports/<session-id>.yaml.
func getReportPath(sessionID string) string {
	dir := filepath.Join(os.TempDir(), "fluxid-reports")
	return filepath.Join(dir, sessionID+".yaml")
}

// WriteReport stores a report for the given session ID.
// Reports are stored as YAML files in /tmp/fluxid-reports/.
func WriteReport(sessionID string, reportYAML string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	// Ensure reports directory exists
	dir := filepath.Join(os.TempDir(), "fluxid-reports")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Write report to file
	reportPath := getReportPath(sessionID)
	if err := os.WriteFile(reportPath, []byte(reportYAML), 0o600); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// ReadReport retrieves the report for the given session ID.
// Returns an empty string if no report exists for the session.
func ReadReport(sessionID string) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}

	reportPath := getReportPath(sessionID)

	// Check if report file exists
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		return "", nil
	}

	// Read report from file
	data, err := os.ReadFile(reportPath) // #nosec G304 - reportPath is constructed from validated sessionID
	if err != nil {
		return "", fmt.Errorf("failed to read report file: %w", err)
	}

	return string(data), nil
}
