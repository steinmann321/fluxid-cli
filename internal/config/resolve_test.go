package config

import (
	"testing"
)

func TestResolveWithAllCLIOverrides(t *testing.T) {
	t.Parallel()

	iterations := 30
	retries := 10

	resolved := Resolve(nil, nil, nil, &iterations, &retries, nil)

	if resolved.Iterations != 30 {
		t.Errorf("Expected Iterations=30, got %d", resolved.Iterations)
	}

	if resolved.ImplementRetries != 10 {
		t.Errorf("Expected ImplementRetries=10, got %d", resolved.ImplementRetries)
	}
}

func TestResolveWithMixedSources(t *testing.T) {
	t.Parallel()

	home := &HomeConfig{
		Agent:            strPtr("home-agent"),
		Iterations:       intPtr(10),
		ImplementRetries: intPtr(5),
		CommitRetries:    nil,
		Commands:         nil,
	}

	project := &ProjectConfig{
		Agent:            nil,
		Iterations:       intPtr(20),
		ImplementRetries: nil,
		CommitRetries:    nil,
		Commands:         nil,
	}

	cliIterations := 30

	resolved := Resolve(project, home, nil, &cliIterations, nil, nil)

	if resolved.Agent != "home-agent" {
		t.Errorf("Expected Agent=home-agent, got %s", resolved.Agent)
	}

	if resolved.Iterations != 30 {
		t.Errorf("Expected Iterations=30 (from CLI), got %d", resolved.Iterations)
	}

	if resolved.ImplementRetries != 5 {
		t.Errorf("Expected ImplementRetries=5 (from home), got %d", resolved.ImplementRetries)
	}
}

func TestResolveDefaultValues(t *testing.T) {
	t.Parallel()

	resolved := Resolve(nil, nil, nil, nil, nil, nil)

	if resolved.Agent != "claude" {
		t.Errorf("Expected Agent=claude (default), got %s", resolved.Agent)
	}

	if resolved.Iterations != 20 {
		t.Errorf("Expected Iterations=20 (default), got %d", resolved.Iterations)
	}

	if resolved.ImplementRetries != 3 {
		t.Errorf("Expected ImplementRetries=3 (default), got %d", resolved.ImplementRetries)
	}

	if resolved.CommitRetries != 100 {
		t.Errorf("Expected CommitRetries=100 (default), got %d", resolved.CommitRetries)
	}
}

func TestResolveProjectOverridesHome(t *testing.T) {
	t.Parallel()

	home := &HomeConfig{
		Agent:            strPtr("home-agent"),
		Iterations:       intPtr(10),
		ImplementRetries: intPtr(5),
		CommitRetries:    nil,
		Commands:         nil,
	}

	project := &ProjectConfig{
		Agent:            strPtr("opencode"),
		Iterations:       intPtr(15),
		ImplementRetries: intPtr(7),
		CommitRetries:    nil,
		Commands:         nil,
	}

	resolved := Resolve(project, home, nil, nil, nil, nil)

	if resolved.Agent != "opencode" {
		t.Errorf("Expected Agent=opencode, got %s", resolved.Agent)
	}

	if resolved.Iterations != 15 {
		t.Errorf("Expected Iterations=15, got %d", resolved.Iterations)
	}

	if resolved.ImplementRetries != 7 {
		t.Errorf("Expected ImplementRetries=7, got %d", resolved.ImplementRetries)
	}
}

// TestResolveEnvOverridesProjectAndHome removed - environment variable support removed in v2.0
