//nolint:paralleltest // E2E tests use shared infrastructure
package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestM05E03ClaudeParserIntegration validates Claude parser with stream-json format.
//
//nolint:dupl // Similar test structure is intentional for independent e2e tests
func TestM05E03ClaudeParserIntegration(t *testing.T) {
	// NOTE: Not using t.Parallel() because tests share agent stubs in root/bin
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create Claude-specific stub that emits stream-json format
	createClaudeFormatStub(t, root)

	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "claude")

	binPath := filepath.Join(root, "bin", "fluxid")
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("test task"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(t.Context(), binPath, "--claude", "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fluxid --claude failed: %v\nOutput:\n%s", err, string(output))
	}

	outputStr := string(output)

	// Verify Claude-specific stream-json was parsed correctly
	if !strings.Contains(outputStr, "Claude streaming message") {
		t.Errorf("Expected Claude stream-json content to be parsed and displayed")
	}

	// Verify agent selection
	if !strings.Contains(outputStr, "Agent: claude") {
		t.Errorf("Expected agent: claude in output")
	}

	// Verify workflow completion
	if !strings.Contains(outputStr, "Status: SUCCESS") {
		t.Errorf("Expected successful workflow completion, got:\n%s", outputStr)
	}

	// Verify all phases ran
	if !strings.Contains(outputStr, "Starting phase: implement") {
		t.Errorf("Missing implement phase")
	}
	if !strings.Contains(outputStr, "Starting phase: review") {
		t.Errorf("Missing review phase")
	}
}

// TestM05E03CodexParserIntegration validates Codex parser with JSONL format.
//
//nolint:dupl // Similar test structure is intentional for independent e2e tests
func TestM05E03CodexParserIntegration(t *testing.T) {
	// NOTE: Not using t.Parallel() because tests share agent stubs in root/bin
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create Codex-specific stub that emits JSONL format
	createCodexFormatStub(t, root)

	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "codex")

	binPath := filepath.Join(root, "bin", "fluxid")
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("test task"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(t.Context(), binPath, "--codex", "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fluxid --codex failed: %v\nOutput:\n%s", err, string(output))
	}

	outputStr := string(output)

	// Verify Codex-specific JSONL was parsed correctly
	if !strings.Contains(outputStr, "Codex agent message") {
		t.Errorf("Expected Codex JSONL content to be parsed and displayed")
	}

	// Verify agent selection
	if !strings.Contains(outputStr, "Agent: codex") {
		t.Errorf("Expected agent: codex in output")
	}

	// Verify workflow completion
	if !strings.Contains(outputStr, "Status: SUCCESS") {
		t.Errorf("Expected successful workflow completion, got:\n%s", outputStr)
	}

	// Verify all phases ran
	if !strings.Contains(outputStr, "Starting phase: implement") {
		t.Errorf("Missing implement phase")
	}
	if !strings.Contains(outputStr, "Starting phase: review") {
		t.Errorf("Missing review phase")
	}
}

// TestM05E03OpencodeParserIntegration validates Opencode parser with JSON format.
//
//nolint:dupl // Similar test structure is intentional for independent e2e tests
func TestM05E03OpencodeParserIntegration(t *testing.T) {
	// NOTE: Not using t.Parallel() because tests share agent stubs in root/bin
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create Opencode-specific stub that emits JSON format
	createOpencodeFormatStub(t, root)

	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "opencode")

	binPath := filepath.Join(root, "bin", "fluxid")
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("test task"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(t.Context(), binPath, "--opencode", "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fluxid --opencode failed: %v\nOutput:\n%s", err, string(output))
	}

	outputStr := string(output)

	// Verify Opencode-specific JSON was parsed correctly
	if !strings.Contains(outputStr, "Opencode text output") {
		t.Errorf("Expected Opencode JSON content to be parsed and displayed")
	}

	// Verify agent selection
	if !strings.Contains(outputStr, "Agent: opencode") {
		t.Errorf("Expected agent: opencode in output")
	}

	// Verify workflow completion
	if !strings.Contains(outputStr, "Status: SUCCESS") {
		t.Errorf("Expected successful workflow completion, got:\n%s", outputStr)
	}

	// Verify all phases ran
	if !strings.Contains(outputStr, "Starting phase: implement") {
		t.Errorf("Missing implement phase")
	}
	if !strings.Contains(outputStr, "Starting phase: review") {
		t.Errorf("Missing review phase")
	}
}

// TestM05E03GeminiParserIntegration validates Gemini parser with stream-json format.
func TestM05E03GeminiParserIntegration(t *testing.T) {
	// NOTE: Not using t.Parallel() because tests share agent stubs in root/bin
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create Gemini-specific stub that emits stream-json format
	createGeminiFormatStub(t, root)

	tmpHome := t.TempDir()
	setupConfigWithCommands(t, tmpHome, "gemini")

	binPath := filepath.Join(root, "bin", "fluxid")
	taskPath := filepath.Join(tmpHome, "task.txt")
	if err := os.WriteFile(taskPath, []byte("test task"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(t.Context(), binPath, "--gemini", "--fluxid-iterations=1", "--file="+taskPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
		"HOME="+tmpHome,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fluxid --gemini failed: %v\nOutput:\n%s", err, string(output))
	}

	outputStr := string(output)

	// Verify Gemini-specific stream-json was parsed correctly
	if !strings.Contains(outputStr, "Gemini assistant response") {
		t.Errorf("Expected Gemini JSON content to be parsed and displayed\nActual output:\n%s", outputStr)
	}

	// Verify agent selection
	if !strings.Contains(outputStr, "Agent: gemini") {
		t.Errorf("Expected agent: gemini in output")
	}

	// Verify workflow completion
	if !strings.Contains(outputStr, "Status: SUCCESS") {
		t.Errorf("Expected successful workflow completion, got:\n%s", outputStr)
	}

	// Verify all phases ran
	if !strings.Contains(outputStr, "Starting phase: implement") {
		t.Errorf("Missing implement phase")
	}
	if !strings.Contains(outputStr, "Starting phase: review") {
		t.Errorf("Missing review phase")
	}
}

// TestM05E03MultiAgentReportGeneration validates report generation works for all agents.
//
//nolint:cyclop,tparallel,goconst,funlen // Test loops over multiple agents with full validations
func TestM05E03MultiAgentReportGeneration(t *testing.T) {
	// NOTE: Not using t.Parallel() because tests share agent stubs in root/bin
	root := getProjectRoot(t)
	buildFluxid(t, root)

	agents := []string{"claude", "codex", "opencode", "gemini"}

	for _, agent := range agents {
		t.Run(agent, func(t *testing.T) {
			t.Parallel()

			// Create format-specific stub
			switch agent {
			case "claude":
				createClaudeFormatStub(t, root)
			case "codex":
				createCodexFormatStub(t, root)
			case "opencode":
				createOpencodeFormatStub(t, root)
			case "gemini":
				createGeminiFormatStub(t, root)
			}

			tmpHome := t.TempDir()
			setupConfigWithCommands(t, tmpHome, agent)

			binPath := filepath.Join(root, "bin", "fluxid")
			taskPath := filepath.Join(tmpHome, "task.txt")
			if err := os.WriteFile(taskPath, []byte("test task"), 0o644); err != nil {
				t.Fatal(err)
			}

			cmd := exec.CommandContext(
				t.Context(),
				binPath,
				"--"+agent,
				"--fluxid-iterations=1",
				"--file="+taskPath,
			)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
				"HOME="+tmpHome,
			)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("fluxid --%s failed: %v\nOutput:\n%s", agent, err, string(output))
			}

			outputStr := string(output)

			// Verify report was generated and contains expected fields
			if !strings.Contains(outputStr, "Status: SUCCESS") {
				t.Errorf("Expected SUCCESS status in report for %s agent", agent)
			}

			// Verify workflow phases completed
			if !strings.Contains(outputStr, "PHASE COMPLETED: IMPLEMENT") {
				t.Errorf("Expected implement phase completion for %s agent", agent)
			}

			if !strings.Contains(outputStr, "PHASE COMPLETED: REVIEW CYCLE") {
				t.Errorf("Expected review phase completion for %s agent", agent)
			}
		})
	}
}

// TestM05E03MultiAgentFailureHandling validates FAIL report handling for all agents.
//
//nolint:cyclop,tparallel // Test loops over multiple agents; no t.Parallel due to shared stubs
func TestM05E03MultiAgentFailureHandling(t *testing.T) {
	// NOTE: Not using t.Parallel() because FAIL stubs conflict with PASS stubs
	// when they overwrite the same agent binaries in root/bin
	root := getProjectRoot(t)
	buildFluxid(t, root)

	agents := []string{"claude", "codex", "opencode", "gemini"}

	for _, agent := range agents {
		t.Run(agent+"_fail", func(t *testing.T) {
			t.Parallel()

			// Create format-specific stub that generates FAIL reports
			switch agent {
			case "claude":
				createClaudeFormatStubFail(t, root)
			case "codex":
				createCodexFormatStubFail(t, root)
			case "opencode":
				createOpencodeFormatStubFail(t, root)
			case "gemini":
				createGeminiFormatStubFail(t, root)
			}

			tmpHome := t.TempDir()
			setupConfigWithCommands(t, tmpHome, agent)

			binPath := filepath.Join(root, "bin", "fluxid")
			taskPath := filepath.Join(tmpHome, "task.txt")
			if err := os.WriteFile(taskPath, []byte("test task"), 0o644); err != nil {
				t.Fatal(err)
			}

			// Use --fluxid-iterations=2 and --max-commit-retries=3 to test retry behavior
			cmd := exec.CommandContext(
				t.Context(),
				binPath,
				"--"+agent,
				"--fluxid-iterations=2",
				"--max-commit-retries=3",
				"--file="+taskPath,
			)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
				"HOME="+tmpHome,
			)

			output, err := cmd.CombinedOutput()
			// Corrected behavior: Command completes successfully even with FAIL reports
			// The workflow continues through all phases, logging failures
			if err != nil {
				t.Fatalf("Expected success for %s agent (workflow completes all phases), got error: %v\nOutput:\n%s",
					agent, err, string(output))
			}

			outputStr := string(output)

			// Verify FAIL status was recognized
			if !strings.Contains(outputStr, "Status:  FAIL") && !strings.Contains(outputStr, "Status: FAIL") {
				t.Errorf("Expected FAIL status in output for %s agent, got:\n%s", agent, outputStr)
			}

			// Verify workflow attempted retries (development iterations)
			if !strings.Contains(outputStr, "DEVELOPMENT ITERATION") {
				t.Errorf("Expected development iteration cycles for %s agent\nActual output:\n%s", agent, outputStr)
			}
		})
	}
}
