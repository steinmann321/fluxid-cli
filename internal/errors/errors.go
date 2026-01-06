// Package errors provides centralized error handling for fluxid with
// consistent error formatting across all components.
//
//nolint:revive // Package name intentionally matches standard library for domain clarity
package errors

import (
	"fmt"
	"os"
)

// ComponentError represents an error from a specific component with standardized formatting.
type ComponentError struct {
	Component   string
	Description string
}

// Error implements the error interface with the format: "error: <component>: <description>".
func (e *ComponentError) Error() string {
	return fmt.Sprintf("error: %s: %s", e.Component, e.Description)
}

// NewConfigError creates a new configuration error.
func NewConfigError(description string) error {
	return &ComponentError{
		Component:   "config",
		Description: description,
	}
}

// NewArgsError creates a new argument parsing error.
func NewArgsError(description string) error {
	return &ComponentError{
		Component:   "args",
		Description: description,
	}
}

// NewWorkflowError creates a new workflow execution error.
func NewWorkflowError(description string) error {
	return &ComponentError{
		Component:   "workflow",
		Description: description,
	}
}

// NewIPCError creates a new IPC communication error.
func NewIPCError(description string) error {
	return &ComponentError{
		Component:   "ipc",
		Description: description,
	}
}

// LogError writes an error message to stderr with consistent formatting.
func LogError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
}
