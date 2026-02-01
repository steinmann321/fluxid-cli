// Package types contains shared types used across fluxid packages.
//
//nolint:revive // "types" is a standard Go package name for shared type definitions
package types

// WorkflowStep represents a workflow step during runtime execution.
type WorkflowStep struct {
	Name            string
	CommandFilePath string
	Retries         int
	IsReview        bool
	Order           int
}

// Workflow represents the complete workflow at runtime.
type Workflow struct {
	Steps            []WorkflowStep
	MaxIterations    int
	CurrentIteration int
}
