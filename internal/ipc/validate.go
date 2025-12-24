// Package ipc provides inter-process communication commands for fluxid.
package ipc

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	errReportEmpty      = errors.New("report is empty")
	errValidationFailed = errors.New("validation failed")
)

// Report represents the structure of a workflow report.
type Report struct {
	Command   string                 `yaml:"command"`
	Artifact  string                 `yaml:"artifact"`
	Timestamp string                 `yaml:"timestamp"`
	Status    string                 `yaml:"status"`
	Issues    Issues                 `yaml:"issues"`
	NextSteps []string               `yaml:"next_steps,omitempty"`
	Summary   string                 `yaml:"summary,omitempty"`
	Extra     map[string]interface{} `yaml:",inline"`
}

// Issues represents the categorized issues in a report.
type Issues struct {
	Blockers     []Issue `yaml:"blockers"`
	Defects      []Issue `yaml:"defects"`
	Concerns     []Issue `yaml:"concerns"`
	Observations []Issue `yaml:"observations"`
	Enhancements []Issue `yaml:"enhancements"`
}

// Issue represents a single issue entry.
type Issue struct {
	Message    string `yaml:"message"`
	Location   string `yaml:"location,omitempty"`
	Code       string `yaml:"code,omitempty"`
	Suggestion string `yaml:"suggestion,omitempty"`
	Reference  string `yaml:"reference,omitempty"`
}

// validateIssuesStructure checks if the issues field and its categories exist in the YAML.
func validateIssuesStructure(reportYAML string) []string {
	var errors []string
	var rawData map[string]interface{}

	if err := yaml.Unmarshal([]byte(reportYAML), &rawData); err != nil {
		return errors
	}

	issuesField, hasIssues := rawData["issues"]
	if !hasIssues {
		return append(errors, "missing required field: issues")
	}

	issues, ok := issuesField.(map[string]interface{})
	if !ok {
		return append(errors, "issues must be an object")
	}

	requiredCategories := []string{"blockers", "defects", "concerns", "observations", "enhancements"}
	for _, category := range requiredCategories {
		if _, exists := issues[category]; !exists {
			errors = append(errors, "issues missing required category: "+category)
		}
	}

	return errors
}

// ValidateReport parses and validates a report YAML string against the schema.
// Returns detailed validation errors if the report is invalid.
func ValidateReport(reportYAML string) error {
	if strings.TrimSpace(reportYAML) == "" {
		return errReportEmpty
	}

	var report Report
	if err := yaml.Unmarshal([]byte(reportYAML), &report); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Collect all validation errors
	var errors []string
	errors = append(errors, validateReportRequiredFields(report)...)
	errors = append(errors, validateIssuesStructure(reportYAML)...)
	errors = append(errors, validateIssuesMessages(report)...)

	if len(errors) > 0 {
		return fmt.Errorf("%w:\n  - %s", errValidationFailed, strings.Join(errors, "\n  - "))
	}

	return nil
}

func validateReportRequiredFields(report Report) []string {
	var errors []string

	if report.Command == "" {
		errors = append(errors, "missing required field: command")
	}

	if report.Artifact == "" {
		errors = append(errors, "missing required field: artifact")
	}

	if report.Timestamp == "" {
		errors = append(errors, "missing required field: timestamp")
	}

	if report.Status == "" {
		errors = append(errors, "missing required field: status")
	} else if report.Status != "PASS" && report.Status != "FAIL" {
		// Validate status enum
		errors = append(errors, fmt.Sprintf("invalid status value: %q (must be PASS or FAIL)", report.Status))
	}

	return errors
}

func validateIssuesMessages(report Report) []string {
	var errors []string

	// Validate each issue has required message field
	allIssues := []struct {
		name   string
		issues []Issue
	}{
		{"blockers", report.Issues.Blockers},
		{"defects", report.Issues.Defects},
		{"concerns", report.Issues.Concerns},
		{"observations", report.Issues.Observations},
		{"enhancements", report.Issues.Enhancements},
	}

	for _, category := range allIssues {
		for i, issue := range category.issues {
			if issue.Message == "" {
				errors = append(errors, fmt.Sprintf("issues.%s[%d]: missing required field 'message'", category.name, i))
			}
		}
	}

	return errors
}
