package config

import (
	"strings"
	"testing"
)

func TestResolveWithAllCLIOverrides(t *testing.T) {
	t.Parallel()

	iterations := 30
	retries := 10
	commitEnabled := false

	resolved := Resolve(nil, nil, nil, nil, &iterations, &retries, &commitEnabled)

	if resolved.Iterations != 30 {
		t.Errorf("Expected Iterations=30, got %d", resolved.Iterations)
	}

	if resolved.ImplementRetries != 10 {
		t.Errorf("Expected ImplementRetries=10, got %d", resolved.ImplementRetries)
	}

	if resolved.CommitEnabled {
		t.Errorf("Expected CommitEnabled=false, got true")
	}

	if resolved.Sources["iterations"] != SourceCLI {
		t.Errorf("Expected iterations source=cli, got %s", resolved.Sources["iterations"])
	}

	if resolved.Sources["implement_retries"] != SourceCLI {
		t.Errorf("Expected implement_retries source=cli, got %s", resolved.Sources["implement_retries"])
	}

	if resolved.Sources["commit_enabled"] != SourceCLI {
		t.Errorf("Expected commit_enabled source=cli, got %s", resolved.Sources["commit_enabled"])
	}
}

func TestResolveWithMixedSources(t *testing.T) {
	t.Parallel()

	home := &HomeConfig{
		Agent:            strPtr("home-agent"),
		Iterations:       intPtr(10),
		ImplementRetries: intPtr(5),
		CommitEnabled:    boolPtr(false),
		Commands:         nil,
	}

	project := &ProjectConfig{
		Agent:            nil,
		Iterations:       intPtr(20),
		ImplementRetries: nil,
		CommitEnabled:    nil,
		Commands:         nil,
	}

	env := &EnvConfig{
		Agent:            nil,
		ImplementRetries: intPtr(7),
		Iterations:       nil,
		CommitEnabled:    nil,
	}

	cliIterations := 30

	resolved := Resolve(project, home, env, nil, &cliIterations, nil, nil)

	if resolved.Agent != "home-agent" {
		t.Errorf("Expected Agent=home-agent, got %s", resolved.Agent)
	}

	if resolved.Iterations != 30 {
		t.Errorf("Expected Iterations=30 (from CLI), got %d", resolved.Iterations)
	}

	if resolved.ImplementRetries != 7 {
		t.Errorf("Expected ImplementRetries=7 (from env), got %d", resolved.ImplementRetries)
	}

	if resolved.CommitEnabled {
		t.Errorf("Expected CommitEnabled=false (from home), got false")
	}

	if !strings.HasPrefix(resolved.Sources["agent"], SourceHome) {
		t.Errorf("Expected agent source=home (prefix), got %s", resolved.Sources["agent"])
	}

	if resolved.Sources["iterations"] != SourceCLI {
		t.Errorf("Expected iterations source=cli, got %s", resolved.Sources["iterations"])
	}

	if resolved.Sources["implement_retries"] != SourceEnv {
		t.Errorf("Expected implement_retries source=env, got %s", resolved.Sources["implement_retries"])
	}

	if !strings.HasPrefix(resolved.Sources["commit_enabled"], SourceHome) {
		t.Errorf("Expected commit_enabled source=home (prefix), got %s", resolved.Sources["commit_enabled"])
	}
}

func TestResolveDefaultValues(t *testing.T) {
	t.Parallel()

	resolved := Resolve(nil, nil, nil, nil, nil, nil, nil)

	if resolved.Agent != "claude" {
		t.Errorf("Expected Agent=claude (default), got %s", resolved.Agent)
	}

	if resolved.Iterations != 20 {
		t.Errorf("Expected Iterations=20 (default), got %d", resolved.Iterations)
	}

	if resolved.ImplementRetries != 3 {
		t.Errorf("Expected ImplementRetries=3 (default), got %d", resolved.ImplementRetries)
	}

	if resolved.CommitEnabled {
		t.Errorf("Expected CommitEnabled=false (default), got true")
	}

	if resolved.Sources["agent"] != SourceDefault {
		t.Errorf("Expected agent source=default, got %s", resolved.Sources["agent"])
	}

	if resolved.Sources["iterations"] != SourceDefault {
		t.Errorf("Expected iterations source=default, got %s", resolved.Sources["iterations"])
	}

	if resolved.Sources["implement_retries"] != SourceDefault {
		t.Errorf("Expected implement_retries source=default, got %s", resolved.Sources["implement_retries"])
	}

	if resolved.Sources["commit_enabled"] != SourceDefault {
		t.Errorf("Expected commit_enabled source=default, got %s", resolved.Sources["commit_enabled"])
	}
}

func TestResolveProjectOverridesHome(t *testing.T) {
	t.Parallel()

	home := &HomeConfig{
		Agent:            strPtr("home-agent"),
		Iterations:       intPtr(10),
		ImplementRetries: intPtr(5),
		CommitEnabled:    boolPtr(false),
		Commands:         nil,
	}

	project := &ProjectConfig{
		Agent:            strPtr("opencode"),
		Iterations:       intPtr(15),
		ImplementRetries: intPtr(7),
		CommitEnabled:    boolPtr(false),
		Commands:         nil,
	}

	resolved := Resolve(project, home, nil, nil, nil, nil, nil)

	if resolved.Agent != "opencode" {
		t.Errorf("Expected Agent=opencode, got %s", resolved.Agent)
	}

	if resolved.Iterations != 15 {
		t.Errorf("Expected Iterations=15, got %d", resolved.Iterations)
	}

	if resolved.ImplementRetries != 7 {
		t.Errorf("Expected ImplementRetries=7, got %d", resolved.ImplementRetries)
	}

	if resolved.CommitEnabled {
		t.Errorf("Expected CommitEnabled=false, got true")
	}

	// All should be from project
	for key, want := range map[string]string{
		"agent":             SourceProject,
		"iterations":        SourceProject,
		"implement_retries": SourceProject,
		"commit_enabled":    SourceProject,
	} {
		if !strings.HasPrefix(resolved.Sources[key], want) {
			t.Errorf("Expected %s source=%s (prefix), got %s", key, want, resolved.Sources[key])
		}
	}
}

func TestResolveEnvOverridesProjectAndHome(t *testing.T) {
	t.Parallel()

	home := &HomeConfig{
		Agent:            nil,
		Iterations:       intPtr(10),
		ImplementRetries: intPtr(5),
		CommitEnabled:    nil,
		Commands:         nil,
	}

	project := &ProjectConfig{
		Agent:            nil,
		Iterations:       intPtr(15),
		ImplementRetries: intPtr(7),
		CommitEnabled:    nil,
		Commands:         nil,
	}

	env := &EnvConfig{
		Agent:            nil,
		Iterations:       intPtr(25),
		ImplementRetries: intPtr(9),
		CommitEnabled:    boolPtr(false),
	}

	resolved := Resolve(project, home, env, nil, nil, nil, nil)

	if resolved.Iterations != 25 {
		t.Errorf("Expected Iterations=25 (from env), got %d", resolved.Iterations)
	}

	if resolved.ImplementRetries != 9 {
		t.Errorf("Expected ImplementRetries=9 (from env), got %d", resolved.ImplementRetries)
	}

	if resolved.CommitEnabled {
		t.Errorf("Expected CommitEnabled=false (from env), got false")
	}

	if resolved.Sources["iterations"] != SourceEnv {
		t.Errorf("Expected iterations source=env, got %s", resolved.Sources["iterations"])
	}

	if resolved.Sources["implement_retries"] != SourceEnv {
		t.Errorf("Expected implement_retries source=env, got %s", resolved.Sources["implement_retries"])
	}

	if resolved.Sources["commit_enabled"] != SourceEnv {
		t.Errorf("Expected commit_enabled source=env, got %s", resolved.Sources["commit_enabled"])
	}
}
