// Package workflow implements the core business logic for fluxid workflows.
package workflow

import (
	"encoding/json"
	"fluxid-cli/internal/output"
	"log"
	"strconv"
	"strings"
)

// Logger handles structured logging for workflow execution.
// It supports both human-readable text format and machine-readable JSON format.
type Logger struct {
	OutputFormat output.Format
}

// JSON event types for structured logging.
type stepStartEvent struct {
	Event      string `json:"event"`
	Step       string `json:"step"`
	Iteration  int    `json:"iteration"`
	Retry      int    `json:"retry"`
	MaxRetries int    `json:"max_retries"`
}

type stepCompleteEvent struct {
	Event      string `json:"event"`
	Step       string `json:"step"`
	Status     string `json:"status"`
	Retry      int    `json:"retry"`
	MaxRetries int    `json:"max_retries"`
}

type iterationStartEvent struct {
	Event         string `json:"event"`
	Iteration     int    `json:"iteration"`
	MaxIterations int    `json:"max_iterations"`
}

type iterationCompleteEvent struct {
	Event        string `json:"event"`
	Iteration    int    `json:"iteration"`
	ReviewStatus string `json:"review_status"`
}

type workflowCompleteEvent struct {
	Event          string `json:"event"`
	Success        bool   `json:"success"`
	FinalIteration int    `json:"final_iteration"`
}

type retriesExhaustedEvent struct {
	Event      string `json:"event"`
	Step       string `json:"step"`
	MaxRetries int    `json:"max_retries"`
}

// LogStepStart logs the start of a workflow step.
func (l *Logger) LogStepStart(stepName string, iteration int, retry int, maxRetries int) {
	if l.OutputFormat == output.FormatJSON {
		event := stepStartEvent{
			Event:      "step_start",
			Step:       stepName,
			Iteration:  iteration,
			Retry:      retry,
			MaxRetries: maxRetries,
		}
		//nolint:errchkjson // Struct with primitive types cannot fail to marshal
		jsonBytes, _ := json.Marshal(event)
		log.Println(string(jsonBytes))
	} else {
		separator := strings.Repeat("━", separatorWidth)
		log.Println(separator)
		log.Printf("▶ Starting step: %s (iteration %d, retry %d/%d)", stepName, iteration, retry, maxRetries)
		log.Println(separator)
	}
}

// LogStepComplete logs the completion of a workflow step.
func (l *Logger) LogStepComplete(stepName string, status string, retry int, maxRetries int) {
	if l.OutputFormat == output.FormatJSON {
		event := stepCompleteEvent{
			Event:      "step_complete",
			Step:       stepName,
			Status:     status,
			Retry:      retry,
			MaxRetries: maxRetries,
		}
		//nolint:errchkjson // Struct with primitive types cannot fail to marshal
		jsonBytes, _ := json.Marshal(event)
		log.Println(string(jsonBytes))
	} else {
		statusIcon := "✓"
		if status == statusFail {
			statusIcon = "✗"
		}
		log.Printf("%s Step complete: %s | Status: %s | Retry: %d/%d",
			statusIcon, stepName, status, retry, maxRetries)
	}
}

// LogIterationStart logs the start of a development iteration.
func (l *Logger) LogIterationStart(iteration int, maxIterations int) {
	if l.OutputFormat == output.FormatJSON {
		event := iterationStartEvent{
			Event:         "iteration_start",
			Iteration:     iteration,
			MaxIterations: maxIterations,
		}
		//nolint:errchkjson // Struct with primitive types cannot fail to marshal
		jsonBytes, _ := json.Marshal(event)
		log.Println(string(jsonBytes))
	} else {
		log.Println("")
		separator := strings.Repeat("━", separatorWidth)
		log.Println(separator)
		iterationLabel := strconv.Itoa(maxIterations)
		if maxIterations == 0 {
			iterationLabel = "∞"
		}
		log.Printf(" DEVELOPMENT ITERATION %d/%s", iteration, iterationLabel)
		log.Println(separator)
	}
}

// LogIterationComplete logs the completion of a development iteration.
func (l *Logger) LogIterationComplete(iteration int, reviewStatus string) {
	if l.OutputFormat == output.FormatJSON {
		event := iterationCompleteEvent{
			Event:        "iteration_complete",
			Iteration:    iteration,
			ReviewStatus: reviewStatus,
		}
		//nolint:errchkjson // Struct with primitive types cannot fail to marshal
		jsonBytes, _ := json.Marshal(event)
		log.Println(string(jsonBytes))
	} else {
		if reviewStatus == statusPass {
			log.Printf("✓ Iteration %d complete: Review PASSED", iteration)
		} else {
			log.Printf("✗ Iteration %d complete: Review FAILED (continuing to next iteration)", iteration)
		}
	}
}

// LogWorkflowComplete logs the final workflow completion status.
func (l *Logger) LogWorkflowComplete(success bool, finalIteration int) {
	if l.OutputFormat == output.FormatJSON {
		event := workflowCompleteEvent{
			Event:          "workflow_complete",
			Success:        success,
			FinalIteration: finalIteration,
		}
		//nolint:errchkjson // Struct with primitive types cannot fail to marshal
		jsonBytes, _ := json.Marshal(event)
		log.Println(string(jsonBytes))
	} else {
		separator := strings.Repeat("━", separatorWidth)
		log.Println("")
		log.Println(separator)
		if success {
			log.Printf("✓ Workflow completed successfully after %d iteration(s)", finalIteration)
		} else {
			log.Printf("⚠ Workflow completed: Maximum iterations (%d) exhausted", finalIteration)
		}
		log.Println(separator)
	}
}

// LogRetriesExhausted logs when a step exhausts all retry attempts.
func (l *Logger) LogRetriesExhausted(stepName string, maxRetries int) {
	if l.OutputFormat == output.FormatJSON {
		event := retriesExhaustedEvent{
			Event:      "retries_exhausted",
			Step:       stepName,
			MaxRetries: maxRetries,
		}
		//nolint:errchkjson // Struct with primitive types cannot fail to marshal
		jsonBytes, _ := json.Marshal(event)
		log.Println(string(jsonBytes))
	} else {
		log.Printf("⚠ Step '%s' exhausted all %d retries (continuing to next step)", stepName, maxRetries)
	}
}
