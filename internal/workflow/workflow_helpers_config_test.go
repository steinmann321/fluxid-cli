//nolint:paralleltest,exhaustruct,funlen,nestif // Test file with test fixtures
package workflow

import (
	"fluxid-cli/internal/config"
	"fluxid-cli/internal/types"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetMaxIterations(t *testing.T) {
	tests := []struct {
		name     string
		cfg      types.Config
		expected int
	}{
		{
			name: "Use workflow max iterations",
			cfg: types.Config{
				Workflow: &types.Workflow{
					MaxIterations: 5,
				},
				MaxReviewCycles: 10,
			},
			expected: 5,
		},
		{
			name: "Fallback to MaxReviewCycles",
			cfg: types.Config{
				Workflow: &types.Workflow{
					MaxIterations: 0,
				},
				MaxReviewCycles: 10,
			},
			expected: 10,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := getMaxIterations(testCase.cfg)
			if result != testCase.expected {
				t.Errorf("Expected %d, got %d", testCase.expected, result)
			}
		})
	}
}

func TestPrepareConfigForStep(t *testing.T) {
	tests := []struct {
		name          string
		cfg           types.Config
		step          types.WorkflowStep
		expectedField string
		expectedValue string
	}{
		{
			name: "Implement step",
			cfg: types.Config{
				CommandFiles: &config.ResolvedCommandFiles{},
			},
			step: types.WorkflowStep{
				Name:            "implement",
				CommandFilePath: "/path/to/implement.txt",
			},
			expectedField: "ImplementPath",
			expectedValue: "/path/to/implement.txt",
		},
		{
			name: "Review step",
			cfg: types.Config{
				CommandFiles: &config.ResolvedCommandFiles{},
			},
			step: types.WorkflowStep{
				Name:            "review",
				CommandFilePath: "/path/to/review.txt",
			},
			expectedField: "ReviewPath",
			expectedValue: "/path/to/review.txt",
		},
		{
			name: "Nil CommandFiles",
			cfg: types.Config{
				CommandFiles: nil,
			},
			step: types.WorkflowStep{
				Name:            "implement",
				CommandFilePath: "/path/to/implement.txt",
			},
			expectedField: "ImplementPath",
			expectedValue: "/path/to/implement.txt",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := prepareConfigForStep(testCase.cfg, testCase.step)
			if result.CommandFiles == nil {
				t.Fatal("CommandFiles should not be nil")
			}

			switch testCase.expectedField {
			case "ImplementPath":
				if result.CommandFiles.ImplementPath != testCase.expectedValue {
					t.Errorf("Expected %s, got %s", testCase.expectedValue, result.CommandFiles.ImplementPath)
				}
			case "ReviewPath":
				if result.CommandFiles.ReviewPath != testCase.expectedValue {
					t.Errorf("Expected %s, got %s", testCase.expectedValue, result.CommandFiles.ReviewPath)
				}
			}
		})
	}
}

func TestGetReviewStatus(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		reportContent  string
		expectedStatus string
		shouldExist    bool
	}{
		{
			name: "PASS status",
			reportContent: `command: "test command"
artifact: "test artifact"
timestamp: "2026-01-01T00:00:00Z"
status: PASS
summary: "Test passed"
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`,
			expectedStatus: "PASS",
			shouldExist:    true,
		},
		{
			name: "FAIL status",
			reportContent: `command: "test command"
artifact: "test artifact"
timestamp: "2026-01-01T00:00:00Z"
status: FAIL
summary: "Test failed"
issues:
  blockers:
    - message: "Test error"
  defects: []
  concerns: []
  observations: []
  enhancements: []
`,
			expectedStatus: "FAIL",
			shouldExist:    true,
		},
		{
			name:           "Missing report",
			reportContent:  "",
			expectedStatus: "FAIL",
			shouldExist:    false,
		},
	}

	for i, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Use valid UUID format for session ID
			sessionID := fmt.Sprintf("00000000-0000-4000-8000-%012d", i+100)
			// SessionRoot should be the directory where session subdirectories are created
			sessionRoot := tempDir

			if testCase.shouldExist {
				// Create report at sessionRoot/<sessionID>/report.yaml
				reportPath := filepath.Join(sessionRoot, sessionID, "report.yaml")
				if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
					t.Fatalf("Failed to create report dir: %v", err)
				}
				if err := os.WriteFile(reportPath, []byte(testCase.reportContent), 0o600); err != nil {
					t.Fatalf("Failed to write report: %v", err)
				}
			}

			cfg := types.Config{
				SessionID:   sessionID,
				SessionRoot: sessionRoot,
			}

			status := getReviewStatus(cfg)
			if status != testCase.expectedStatus {
				t.Errorf("Expected status %s, got %s", testCase.expectedStatus, status)
			}
		})
	}
}

func TestBuildWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.WorkflowConfig
		configDir     string
		maxIterations int
		expectError   bool
	}{
		{
			name:          "Nil config",
			cfg:           nil,
			configDir:     "/test",
			maxIterations: 5,
			expectError:   true,
		},
		{
			name: "Valid config",
			cfg: &config.WorkflowConfig{
				Steps: []config.WorkflowStepConfig{
					{
						Name:    "implement",
						Command: "implement.txt",
						Retries: 2,
					},
				},
				Review: config.ReviewStepConfig{
					Command: "review.txt",
					Retries: 1,
				},
			},
			configDir:     "/test",
			maxIterations: 5,
			expectError:   false,
		},
		{
			name: "Default retries (0)",
			cfg: &config.WorkflowConfig{
				Steps: []config.WorkflowStepConfig{
					{
						Name:    "implement",
						Command: "implement.txt",
						Retries: 0, // Test default retries
					},
				},
				Review: config.ReviewStepConfig{
					Command: "review.txt",
					Retries: 0, // Test default retries
				},
			},
			configDir:     "/test",
			maxIterations: 5,
			expectError:   false,
		},
		{
			name: "Infinite iterations (0)",
			cfg: &config.WorkflowConfig{
				Steps: []config.WorkflowStepConfig{
					{
						Name:    "implement",
						Command: "implement.txt",
						Retries: 1,
					},
				},
				Review: config.ReviewStepConfig{
					Command: "review.txt",
					Retries: 1,
				},
			},
			configDir:     "/test",
			maxIterations: 0, // Test infinite iterations
			expectError:   false,
		},
		{
			name: "Absolute command path",
			cfg: &config.WorkflowConfig{
				Steps: []config.WorkflowStepConfig{
					{
						Name:    "implement",
						Command: "/absolute/path/implement.txt", // Test absolute path
						Retries: 1,
					},
				},
				Review: config.ReviewStepConfig{
					Command: "/absolute/path/review.txt", // Test absolute path
					Retries: 1,
				},
			},
			configDir:     "/test",
			maxIterations: 5,
			expectError:   false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			workflow, err := BuildWorkflow(testCase.cfg, testCase.configDir, testCase.maxIterations)
			if testCase.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if workflow == nil {
					t.Fatal("Expected workflow, got nil")
				}
				if len(workflow.Steps) != len(testCase.cfg.Steps)+1 {
					t.Errorf("Expected %d steps, got %d", len(testCase.cfg.Steps)+1, len(workflow.Steps))
				}
			}
		})
	}
}
