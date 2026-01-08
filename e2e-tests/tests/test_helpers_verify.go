package tests

import (
	"strings"
	"testing"
)

// verifyConfigLine checks that a line containing fieldPrefix also contains sourcePattern.
// This handles file paths in source strings like "source: home (/path/to/config.yaml)".
//
//nolint:unparam // sourcePattern parameter maintained for API flexibility
func verifyConfigLine(t *testing.T, output, fieldPrefix, sourcePattern string) {
	t.Helper()
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, fieldPrefix) && strings.Contains(line, sourcePattern) {
			return
		}
	}
	t.Errorf("Expected line containing %q and %q, got:\n%s", fieldPrefix, sourcePattern, output)
}

// verifyAgentArgsAndSource is a helper to verify agent arguments in test output.
// This reduces duplication in M06 tests for JSON and YAML validation.
// v2.0: source tracking removed (Phase 9).
func verifyAgentArgsAndSource(
	t *testing.T,
	agentArgs []string,
	agent string,
	expectedAgent string,
) {
	t.Helper()

	// Verify agent args are captured
	if len(agentArgs) != minExpectedArgCount {
		t.Errorf("Expected 2 agent args, got: %d (%v)", len(agentArgs), agentArgs)
	}
	if len(agentArgs) >= minExpectedArgCount {
		if agentArgs[0] != "arg1" || agentArgs[1] != "arg2" {
			t.Errorf("Expected agent args [arg1, arg2], got: %v", agentArgs)
		}
	}

	// Verify agent is as expected
	if agent != expectedAgent {
		t.Errorf("Expected agent %q, got: %q", expectedAgent, agent)
	}
}

// verifyConfigValues checks that max review cycles and implement retries match expected values.
// v2.0: source tracking removed (Phase 9).
func verifyConfigValues(
	t *testing.T,
	maxReviewCycles, maxImplementRetries int,
	expectedReviewCycles, expectedImplementRetries int,
) {
	t.Helper()

	if maxReviewCycles != expectedReviewCycles {
		t.Errorf("Expected max_review_cycles to be %d, got: %d", expectedReviewCycles, maxReviewCycles)
	}
	if maxImplementRetries != expectedImplementRetries {
		t.Errorf("Expected max_implement_retries to be %d, got: %d", expectedImplementRetries, maxImplementRetries)
	}
}

// verifyDefaultTextFormat verifies output is in text format and not structured format.
// unmarshalFunc should attempt to unmarshal the output as the structured format
// and return an error if it's not valid.
func verifyDefaultTextFormat(
	t *testing.T,
	output string,
	unmarshalFunc func(string) error,
	formatName string,
) {
	t.Helper()

	// Verify text format markers are present
	if !strings.Contains(output, "Workflow Initialization") {
		t.Errorf("Expected text format header in output, got:\n%s", output)
	}

	// Verify it's NOT the structured format (should fail to parse)
	if err := unmarshalFunc(output); err == nil {
		t.Errorf("Expected text output, but got valid %s:\n%s", formatName, output)
	}
}
