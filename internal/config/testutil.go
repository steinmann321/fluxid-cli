package config

// Test helper functions for config package tests.

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func equalHomeConfig(actual, expected *HomeConfig) bool {
	if actual == nil && expected == nil {
		return true
	}
	if actual == nil || expected == nil {
		return false
	}

	return equalStrPtr(actual.Agent, expected.Agent) &&
		equalIntPtr(actual.Iterations, expected.Iterations) &&
		equalIntPtr(actual.ImplementRetries, expected.ImplementRetries)
}

func equalProjectConfig(actual, expected *ProjectConfig) bool {
	if actual == nil && expected == nil {
		return true
	}
	if actual == nil || expected == nil {
		return false
	}

	return equalStrPtr(actual.Agent, expected.Agent) &&
		equalIntPtr(actual.Iterations, expected.Iterations) &&
		equalIntPtr(actual.ImplementRetries, expected.ImplementRetries)
}

func equalStrPtr(actual, expected *string) bool {
	if actual == nil && expected == nil {
		return true
	}
	if actual == nil || expected == nil {
		return false
	}
	return *actual == *expected
}

func equalIntPtr(actual, expected *int) bool {
	if actual == nil && expected == nil {
		return true
	}
	if actual == nil || expected == nil {
		return false
	}
	return *actual == *expected
}
