//nolint:paralleltest,exhaustruct // Tests involve shared state, test fixtures don't need all fields
package config

import (
	"testing"
)

func TestResolveAgentArgs_ProjectOnly(t *testing.T) {
	projectArgs := []string{"--model", "test"}
	result := resolveAgentArgs(&projectArgs, nil)
	if len(result) != 2 {
		t.Errorf("Expected 2 args, got %d", len(result))
	}
}

func TestResolveAgentArgs_HomeOnly(t *testing.T) {
	homeArgs := []string{"--verbose"}
	result := resolveAgentArgs(nil, &homeArgs)
	if len(result) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(result))
	}
}

func TestResolveAgentArgs_Both(t *testing.T) {
	projectArgs := []string{"--model", "test"}
	homeArgs := []string{"--verbose"}
	result := resolveAgentArgs(&projectArgs, &homeArgs)
	if len(result) != 2 {
		t.Errorf("Expected project args to take precedence, got %d", len(result))
	}
}

func TestResolveAgentArgs_Neither(t *testing.T) {
	result := resolveAgentArgs(nil, nil)
	if len(result) != 0 {
		t.Errorf("Expected empty args, got %d", len(result))
	}
}

func TestExtractAgentArgsValues_ProjectOnly(t *testing.T) {
	project := &ProjectConfig{
		AgentArgs: []string{"--model", "test"},
	}
	values := extractAgentArgsValues(project, nil)
	if values.project == nil {
		t.Error("Expected project args to be set")
	}
	if values.home != nil {
		t.Error("Expected home args to be nil")
	}
}

func TestExtractAgentArgsValues_HomeOnly(t *testing.T) {
	home := &HomeConfig{
		AgentArgs: []string{"--verbose"},
	}
	values := extractAgentArgsValues(nil, home)
	if values.home == nil {
		t.Error("Expected home args to be set")
	}
	if values.project != nil {
		t.Error("Expected project args to be nil")
	}
}

func TestExtractAgentArgsValues_Both(t *testing.T) {
	project := &ProjectConfig{
		AgentArgs: []string{"--model", "test"},
	}
	home := &HomeConfig{
		AgentArgs: []string{"--verbose"},
	}
	values := extractAgentArgsValues(project, home)
	if values.project == nil {
		t.Error("Expected project args to be set")
	}
	if values.home == nil {
		t.Error("Expected home args to be set")
	}
}

func TestExtractAgentArgsValues_EmptyArrays(t *testing.T) {
	project := &ProjectConfig{
		AgentArgs: []string{},
	}
	home := &HomeConfig{
		AgentArgs: []string{},
	}
	values := extractAgentArgsValues(project, home)
	if values.project != nil {
		t.Error("Expected nil for empty project args array")
	}
	if values.home != nil {
		t.Error("Expected nil for empty home args array")
	}
}

func TestGetHomeConfigPath_Default(t *testing.T) {
	path, err := GetHomeConfigPath()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
}

func TestGetProjectConfigPath_Default(t *testing.T) {
	path, err := GetProjectConfigPath()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
}
